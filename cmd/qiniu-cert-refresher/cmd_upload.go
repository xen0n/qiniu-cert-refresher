// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"log/slog"

	"github.com/urfave/cli/v2"
)

func cmdUpload(cCtx *cli.Context) error {
	cfg := getConfig(cCtx.Context)
	slog.Debug("invoked the upload command")
	_ = cfg
	return nil
}
