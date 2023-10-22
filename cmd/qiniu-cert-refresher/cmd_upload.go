// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"github.com/xen0n/qiniu-cert-refresher/api/qcdn"
)

func cmdUpload(cCtx *cli.Context) error {
	certPath := cCtx.Path("cert")
	pemPath := cCtx.Path("pem")
	key := cCtx.Args().First()

	if len(certPath) == 0 {
		return errors.New("a certificate chain file must be given")
	}
	if len(pemPath) == 0 {
		return errors.New("a private key file must be given")
	}
	if len(key) == 0 {
		return errors.New("a tracing key for the certificate must be specified")
	}
	slog.Debug("invoked the upload command", "cert", certPath, "pem", pemPath, "key", key)

	payloadBase, err := preparePartialUploadPayload(certPath, pemPath)
	if err != nil {
		slog.Error("failed to prepare the upload", "err", err)
		return err
	}

	cfg := getConfig(cCtx.Context)
	for _, acc := range cfg.Accounts {
		payload := *payloadBase
		payload.Name = deriveCertNameForAccount(acc, key)

		err := uploadAndRefreshForAccount(acc, key, &payload)
		if err != nil {
			slog.Error("failed to upload and refresh", "account", acc.DisplayName, "key", key, "err", err)
			return err
		}
	}

	return nil
}

func readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}

func preparePartialUploadPayload(
	certPath string,
	pemPath string,
) (*qcdn.ReqUploadCert, error) {
	cert, err := readFile(certPath)
	if err != nil {
		return nil, err
	}

	pem, err := readFile(pemPath)
	if err != nil {
		return nil, err
	}

	cn, err := getCommonNameFromCert(cert)
	if err != nil {
		return nil, err
	}

	return &qcdn.ReqUploadCert{
		Name:       "", // to be filled in later with per-account customization
		CommonName: cn,
		PEM:        string(pem),
		CA:         string(cert),
	}, nil
}

func getCommonNameFromCert(pemCerts []byte) (string, error) {
	// this is resembling (*crypto/x509.CertPool).AppendCertsFromPEM
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		certBytes := block.Bytes
		cert, err := x509.ParseCertificate(certBytes)
		if err != nil {
			return "", err
		}

		return cert.Subject.CommonName, nil
	}

	return "", errors.New("no certificate in input")
}

func deriveCertNameForAccount(acc *AccountConfig, key string) string {
	return fmt.Sprintf("%s %s (%d)", acc.ManagedCertNamePrefix, key, time.Now().UnixNano())
}

func uploadAndRefreshForAccount(
	acc *AccountConfig,
	key string,
	payload *qcdn.ReqUploadCert,
) error {
	slog.Debug("about to upload cert", "account", acc.DisplayName, "certName", payload.Name)
	newCertID, err := qcdn.UploadCert(acc.qiniuCreds, payload)
	if err != nil {
		slog.Error("failed to upload cert", "account", acc.DisplayName, "err", err)
		return err
	}

	return refreshForAccount(acc, key, newCertID)
}

func refreshForAccount(acc *AccountConfig, key string, newCertID string) error {
	slog.Debug("about to refresh domains", "account", acc.DisplayName, "key", key, "newCertID", newCertID)

	relevantCerts, err := listAllCertsWithTracingKey(acc, key)
	if err != nil {
		return err
	}

	// if newCertID == "": refresh to the latest-expiring certificate that's valid
	// (i.e. already past its NotBefore)
	//
	// if newCertID != "": just use it
	if newCertID == "" {
		targetCert := findLatestNonExpiringValidCert(relevantCerts, time.Now())
		newCertID = targetCert.ID
	}

	// sanity check: is newCertID actually belonging to this account & tracing key?
	if !isCertIDInList(newCertID, relevantCerts) {
		return fmt.Errorf("the specified cert ID '%s' seems irrelevant to tracing key '%s'", newCertID, key)
	}

	certIDsToSupersede := lo.FilterMap(relevantCerts, func(c *qcdn.Cert, _ int) (string, bool) {
		if c.ID == newCertID {
			return "", false
		}
		return c.ID, true
	})

	// TODO: parallelize (while respecting some global concurrency limit)
	for _, oldCertID := range certIDsToSupersede {
		err := replaceDomainCerts(acc, oldCertID, newCertID)
		if err != nil {
			return err
		}
	}

	return nil
}

func makeTracingKeyMatcher(prefix string, key string) (*regexp.Regexp, error) {
	var sb strings.Builder
	sb.WriteRune('^')
	sb.WriteString(regexp.QuoteMeta(prefix))
	sb.WriteString(`\s*`)
	sb.WriteString(regexp.QuoteMeta(key))
	return regexp.Compile(sb.String())
}

func listAllCertsWithTracingKey(acc *AccountConfig, key string) ([]*qcdn.Cert, error) {
	allCerts, err := qcdn.ListAllCerts(acc.qiniuCreds)
	if err != nil {
		return nil, err
	}

	matcher, err := makeTracingKeyMatcher(acc.ManagedCertNamePrefix, key)
	if err != nil {
		return nil, err
	}

	return lo.Filter(allCerts, func(c *qcdn.Cert, _ int) bool {
		return matcher.MatchString(c.Name)
	}), nil
}

func findLatestNonExpiringValidCert(certs []*qcdn.Cert, epoch time.Time) *qcdn.Cert {
	epochUnix := epoch.Unix()

	var result *qcdn.Cert
	for _, c := range certs {
		if epochUnix < c.NotBefore || c.NotAfter < epochUnix {
			// not yet entered its validity period || already expired
			continue
		}

		if result == nil || result.NotAfter < c.NotAfter {
			result = c
			continue
		}
	}

	return result
}

func isCertIDInList(certID string, certs []*qcdn.Cert) bool {
	for _, c := range certs {
		if c.ID == certID {
			return true
		}
	}
	return false
}

func replaceDomainCerts(acc *AccountConfig, oldCertID string, newCertID string) error {
	slog.Debug(
		"about to replace domain certs",
		"account",
		acc.DisplayName,
		"oldCertID",
		oldCertID,
		"newCertID",
		newCertID,
	)

	domains, err := qcdn.ListAllDomainsByCertID(acc.qiniuCreds, oldCertID)
	if err != nil {
		return err
	}

	var eg errgroup.Group
	for _, d := range domains {
		d := d
		// TODO: throttle
		eg.Go(func() error {
			return replaceCertForOneDomain(acc, d.Name, newCertID)
		})
	}

	err = eg.Wait()
	if err != nil {
		return err
	}

	return nil
}

func replaceCertForOneDomain(acc *AccountConfig, domain string, newCertID string) error {
	// get old https config
	details, err := qcdn.GetDomain(acc.qiniuCreds, domain)
	if err != nil {
		return err
	}

	cfg := details.HTTPS
	slog.Debug("got current HTTPS config", "account", acc.DisplayName, "domain", domain, "cfg", cfg)

	cfg.CertID = newCertID

	slog.Debug("about to update HTTPS config", "account", acc.DisplayName, "domain", domain, "cfg", cfg)
	return qcdn.UpdateHTTPSConfig(acc.qiniuCreds, domain, cfg)
}
