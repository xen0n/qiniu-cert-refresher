// SPDX-License-Identifier: GPL-3.0-or-later

package qiniucommon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/qiniu/go-sdk/v7/auth"
)

type RespError struct {
	Code     int    `json:"code"`
	ErrorMsg string `json:"error"`
}

var _ error = (*RespError)(nil)

func (e *RespError) Error() string {
	return fmt.Sprintf("Qiniu error %d: %s", e.Code, e.ErrorMsg)
}

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

	slog.Debug("about to make HTTP call", "req", req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return zeroResp, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return zeroResp, err
	}

	slog.Debug("got HTTP response", "req", req, "statusCode", resp.StatusCode, "body", respBody)

	if resp.StatusCode >= 400 {
		// this is an error
		var errObj RespError
		err = json.Unmarshal(respBody, &errObj)
		if err != nil {
			return zeroResp, err
		}
		return zeroResp, &errObj
	}

	var result Resp
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return zeroResp, err
	}

	return result, nil
}
