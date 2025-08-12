// SPDX-License-Identifier: GPL-3.0-or-later

package qcdn

import (
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/qiniu/go-sdk/v7/auth"

	"github.com/xen0n/qiniu-cert-refresher/api/qiniuCommon"
)

const defaultHost = "https://api.qiniu.com"

///////////////////////////////////////////////////////////////////////////////

func GetDomain(mac *auth.Credentials, domain string) (*Domain, error) {
	var sb strings.Builder
	sb.WriteString(defaultHost)
	sb.WriteString("/domain/")
	sb.WriteString(url.PathEscape(domain))

	return qiniuCommon.RequestWithBody[*Domain](mac, sb.String(), nil)
}

func ListAllDomainsByCertID(mac *auth.Credentials, certID string) ([]*Domain, error) {
	var result []*Domain

	marker := ""
	for {
		req := ReqListDomains{
			CertID: certID,
			Marker: marker,
		}
		resp, err := listDomains(mac, &req)
		if err != nil {
			return nil, err
		}

		if len(resp.Domains) == 0 {
			break
		}

		marker = resp.Marker
		result = append(result, resp.Domains...)
	}

	return result, nil
}

func listDomains(mac *auth.Credentials, req *ReqListDomains) (*RespListDomains, error) {
	var sb strings.Builder
	sb.WriteString(defaultHost)
	sb.WriteString("/domain")

	q := url.Values{}
	for _, t := range req.Types {
		q.Add("types", string(t))
	}
	if len(req.CertID) > 0 {
		q.Set("certId", req.CertID)
	}
	if len(req.SourceTypes) > 0 {
		q["sourceTypes"] = req.SourceTypes
	}
	if len(req.SourceQiniuBucket) > 0 {
		q.Set("sourceQiniuBucket", req.SourceQiniuBucket)
	}
	if len(req.SourceIP) > 0 {
		q.Set("sourceIp", req.SourceIP)
	}
	if len(req.Marker) > 0 {
		q.Set("marker", req.Marker)
	}

	if req.Limit < 1 || req.Limit > 1000 {
		req.Limit = 10
	}
	if req.Limit != 10 {
		q.Set("limit", strconv.Itoa(req.Limit))
	}

	if len(q) > 0 {
		sb.WriteRune('?')
		sb.WriteString(q.Encode())
	}

	return qiniuCommon.RequestWithBody[*RespListDomains](mac, sb.String(), nil)
}

func UpdateHTTPSConfig(mac *auth.Credentials, domain string, newConf *HTTPSConfig) error {
	var sb strings.Builder
	sb.WriteString(defaultHost)
	sb.WriteString("/domain/")
	sb.WriteString(url.PathEscape(domain))
	sb.WriteString("/httpsconf")

	_, err := qiniuCommon.RequestWithBody[struct{}](mac, sb.String(), newConf, http.MethodPut)
	return err
}

///////////////////////////////////////////////////////////////////////////////

const defaultListCertsPageSize = 100

func ListAllCerts(mac *auth.Credentials) ([]*Cert, error) {
	var result []*Cert

	marker := ""
	for {
		slog.Debug("listing certs", "marker", marker)
		resp, err := listCerts(mac, marker, defaultListCertsPageSize)
		if err != nil {
			return nil, err
		}

		slog.Debug("got cert list", "newMarker", resp.Marker, "numCerts", len(resp.Certs))
		marker = resp.Marker
		result = append(result, resp.Certs...)

		if len(resp.Certs) == 0 {
			break
		}
	}

	return result, nil
}

func listCerts(mac *auth.Credentials, marker string, limit int) (*RespListCerts, error) {
	var sb strings.Builder
	sb.WriteString(defaultHost)
	sb.WriteString("/sslcert")

	q := url.Values{}
	if len(marker) > 0 {
		q.Set("marker", marker)
	}
	if limit != defaultListCertsPageSize {
		q.Set("limit", strconv.Itoa(limit))
	}

	if len(q) > 0 {
		sb.WriteRune('?')
		sb.WriteString(q.Encode())
	}

	return qiniuCommon.RequestWithBody[*RespListCerts](mac, sb.String(), nil)
}

func UploadCert(mac *auth.Credentials, req *ReqUploadCert) (id string, err error) {
	resp, err := qiniuCommon.RequestWithBody[RespUploadCert](mac, defaultHost+"/sslcert", req)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func DeleteCert(mac *auth.Credentials, id string) error {
	var sb strings.Builder
	sb.WriteString(defaultHost)
	sb.WriteString("/sslcert/")
	sb.WriteString(id)

	_, err := qiniuCommon.RequestWithBody[struct{}](mac, sb.String(), nil, http.MethodDelete)
	return err
}
