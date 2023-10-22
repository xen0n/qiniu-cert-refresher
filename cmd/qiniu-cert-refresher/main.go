// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.App{
		Name:  "qiniu-cert-refresher",
		Usage: "keeps Qiniu CDN synced with your locally-renewed certificates",
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "force using configuration from this file (supported format: toml)",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "enable debug output",
			},
			&cli.BoolFlag{
				Name:    "env-config",
				Aliases: []string{"e"},
				Usage:   "force using configuration from environment variables",
			},
			&cli.BoolFlag{
				Name:    "jsonlog",
				Aliases: []string{"j"},
				Usage:   "produce log messages in JSON",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "info",
				Aliases: []string{"i"},
				Usage:   "queries and shows the current state of configured accounts",
				Before:  beforeCmd,
				Action:  cmdInfo,
			},
			{
				Name:      "upload",
				Aliases:   []string{"u"},
				Usage:     "uploads a new certificate, refreshing all associated domains",
				ArgsUsage: "<TRACING-KEY-OF-THE-CERT>",
				Before:    beforeCmd,
				Action:    cmdUpload,
				Flags: []cli.Flag{
					&cli.PathFlag{
						Name:  "cert",
						Usage: "path to the certificate file",
					},
					&cli.PathFlag{
						Name:  "pem",
						Usage: "path to the private key file",
					},
				},
			},
		},
	}

	err := app.RunContext(context.Background(), os.Args)
	if err != nil {
		slog.Error("command failed", "err", err)
		os.Exit(1)
	}
}

func beforeCmd(cCtx *cli.Context) error {
	initLogging(cCtx)
	return initConfig(cCtx)
}

func initLogging(cCtx *cli.Context) {
	var opts slog.HandlerOptions

	if cCtx.Bool("debug") {
		opts.Level = slog.LevelDebug
		if cCtx.Count("debug") >= 2 {
			opts.AddSource = true
		}
	}

	var h slog.Handler
	if cCtx.Bool("jsonlog") {
		h = slog.NewJSONHandler(os.Stderr, &opts)
	} else {
		h = slog.NewTextHandler(os.Stderr, &opts)
	}

	slog.SetDefault(slog.New(h))
}

func initConfig(cCtx *cli.Context) error {
	forceEnv := cCtx.Bool("env-config")
	forceFilePath := cCtx.Path("config")
	forceFile := len(forceFilePath) > 0
	if forceEnv && forceFile {
		return errors.New("cannot force configuration from both environment and file")
	}

	var cfg *Config
	if forceEnv {
		var err error
		slog.Debug("forced configuration from environment")
		cfg, err = configFromEnv()
		if err != nil {
			return err
		}
		if cfg == nil {
			return errors.New("forced to config from environment, but no config found")
		}
	} else if forceFile {
		var err error
		slog.Debug("forced configuration from file", "path", forceFilePath)
		cfg, err = configFromTOML(forceFilePath)
		if err != nil {
			return err
		}
	} else {
		var err error
		// try env first, then read from default config path
		slog.Debug("try configuring from environment")
		cfg, err = configFromEnv()
		if err != nil {
			return err
		}
		if cfg == nil {
			slog.Debug("fallback to default config file", "path", defaultConfigPath)
			// env doesn't contain config
			cfg, err = configFromTOML(defaultConfigPath)
			if err != nil {
				return err
			}
		}
	}

	cCtx.Context = setConfig(cCtx.Context, cfg)
	return nil
}
