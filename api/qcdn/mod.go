// SPDX-License-Identifier: GPL-3.0-or-later

package qcdn

import (
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/qiniu/go-sdk/v7/auth"

	"github.com/xen0n/qiniu-cert-refresher/api/common"
)

const defaultHost = "https://api.qiniu.com"

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
	hasQ := false
	if len(marker) > 0 {
		hasQ = true
		q.Set("marker", marker)
	}
	if limit != defaultListCertsPageSize {
		hasQ = true
		q.Set("limit", strconv.Itoa(limit))
	}

	if hasQ {
		sb.WriteRune('?')
		sb.WriteString(q.Encode())
	}

	return common.RequestWithBody[*RespListCerts](mac, sb.String(), nil)
}

func UploadCert(mac *auth.Credentials, req *ReqUploadCert) (id string, err error) {
	resp, err := common.RequestWithBody[RespUploadCert](mac, defaultHost+"/sslcert", req)
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

	_, err := common.RequestWithBody[struct{}](mac, sb.String(), nil, http.MethodDelete)
	return err
}
