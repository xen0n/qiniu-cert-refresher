// SPDX-License-Identifier: GPL-3.0-or-later

package common

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/qiniu/go-sdk/v7/auth"
)

// RequestWithBody 带body对api发出请求并且返回response body
//
// copied and adapted from qiniu go-sdk (MIT-licensed)
func RequestWithBody[Resp any](
	mac *auth.Credentials,
	path string,
	body any,
	overrideMethod ...string,
) (Resp, error) {
	// workaround the fact that `zero` isn't available; never write to this value
	var zeroResp Resp

	var method string
	if len(overrideMethod) > 0 {
		method = overrideMethod[0]
	} else {
		if body != nil {
			method = http.MethodPost
		} else {
			method = http.MethodGet
		}
	}

	var bodyReader io.Reader
	if body != nil {
		reqData, err := json.Marshal(body)
		if err != nil {
			return zeroResp, err
		}

		bodyReader = bytes.NewReader(reqData)
	}

	req, err := http.NewRequest(method, path, bodyReader)
	if err != nil {
		return zeroResp, err
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
