// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/xen0n/qiniu-cert-refresher/api/qcdn"
)

func cmdInfo(cCtx *cli.Context) error {
	cfg := getConfig(cCtx.Context)
	slog.Debug("invoked the info command")

	for i, acc := range cfg.Accounts {
		if i > 0 {
			fmt.Println("")
		}

		err := showAccount(acc)
		if err != nil {
			slog.Error("failed to show one account", "account", acc.DisplayName)
		}
	}

	return nil
}

func showAccount(acc *AccountConfig) error {
	slog.Debug("querying certs", "account", acc.DisplayName)
	allCerts, err := qcdn.ListAllCerts(acc.qiniuCreds)
	if err != nil {
		slog.Error("failed to list certs", "account", acc.DisplayName, "err", err)
		return err
	}

	// TODO: parallelize
	domainsByCertID := make(map[string][]*qcdn.Domain)
	for _, cert := range allCerts {
		certID := cert.ID

		slog.Debug("querying domains associated with cert", "account", acc.DisplayName, "certID", certID)
		domains, err := qcdn.ListAllDomainsByCertID(acc.qiniuCreds, certID)
		if err != nil {
			slog.Error("failed to list domains", "account", acc.DisplayName, "certID", certID, "err", err)
			return err
		}

		domainsByCertID[certID] = domains
	}

	fmt.Printf("# Account %s\n", acc.DisplayName)
	for _, cert := range allCerts {
		fmt.Printf("\n- ID:         %s\n  Name:       %s\n  CommonName: %s\n", cert.ID, cert.Name, cert.CommonName)
		fmt.Printf("  NotBefore:  %s\n", time.Unix(cert.NotBefore, 0).Format(time.RFC3339))
		fmt.Printf("  NotAfter:   %s\n", time.Unix(cert.NotAfter, 0).Format(time.RFC3339))
		fmt.Printf("  CreatedAt:  %s\n", time.Unix(cert.CreateTime, 0).Format(time.RFC3339))

		if len(cert.DNSNames) == 0 {
			fmt.Println("  DNSNames: []")
		} else {
			fmt.Println("  DNSNames:")
			for _, n := range cert.DNSNames {
				fmt.Printf("    - %s\n", n)
			}
		}

		associatedDomains := domainsByCertID[cert.ID]
		if len(associatedDomains) == 0 {
			fmt.Println("  Domains: []")
		} else {
			fmt.Println("  Domains:")
			for _, d := range associatedDomains {
				fmt.Printf("    - %s\n", d.Name)
			}
		}
	}

	return nil
}
