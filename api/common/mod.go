// SPDX-License-Identifier: GPL-3.0-or-later

package common

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/qiniu/go-sdk/v7/auth"
)

const defaultHost = "https://api.qiniu.com"

// RequestWithBody 带body对api发出请求并且返回response body
//
// copied and adapted from qiniu go-sdk (MIT-licensed)
func RequestWithBody(mac *auth.Credentials, path string, body interface{}) ([]byte, error) {
	reqData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", path, bytes.NewReader(reqData))
	if err != nil {
		return nil, err
	}

	accessToken, err := mac.SignRequest(req)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "QBox "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
