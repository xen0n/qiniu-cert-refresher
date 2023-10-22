// SPDX-License-Identifier: GPL-3.0-or-later

package common

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/qiniu/go-sdk/v7/auth"
)

// RequestWithBody 带body对api发出请求并且返回response body
//
// copied and adapted from qiniu go-sdk (MIT-licensed)
func RequestWithBody[Resp any](mac *auth.Credentials, path string, body any) (Resp, error) {
	var req *http.Request
	var zeroResp Resp
	if body != nil {
		reqData, err := json.Marshal(body)
		if err != nil {
			return zeroResp, err
		}

		req, err = http.NewRequest(http.MethodPost, path, bytes.NewReader(reqData))
		if err != nil {
			return zeroResp, err
		}
	} else {
		var err error
		req, err = http.NewRequest(http.MethodGet, path, nil)
		if err != nil {
			return zeroResp, err
		}
	}

	accessToken, err := mac.SignRequest(req)
	if err != nil {
		return zeroResp, err
	}

	req.Header.Add("Authorization", "QBox "+accessToken)
	if body != nil {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return zeroResp, err
	}
	defer resp.Body.Close()

	d := json.NewDecoder(resp.Body)
	var result Resp
	err = d.Decode(&result)
	if err != nil {
		return zeroResp, err
	}

	return result, nil
}
