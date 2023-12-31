// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/qiniu/go-sdk/v7/auth"
)

const defaultManagedCertNamePrefix = "[QCR-Managed]"
const defaultConfigPath = "qcr-config.toml"

type Config struct {
	Accounts []*AccountConfig `toml:"accounts"`
}

type AccountConfig struct {
	// AK AccessKey
	AK string `toml:"ak"`
	// SK SecretKey
	SK string `toml:"sk"`
	// DisplayName 此账号的名称，仅用于日志、调试信息等显示用途
	// 可以留空，该账号在展示时将仅体现 AK 的头尾几个字符
	DisplayName string `toml:"display_name"`
	// ManagedCertNamePrefix 由本工具管理的证书名称的前缀，用于自动识别这部分证书记录与相关的域名
	// 可以留空，意为取工具默认值
	ManagedCertNamePrefix string `toml:"managed_cert_name_prefix"`

	qiniuCreds *auth.Credentials
}

func defaultDisplayNameFromAK(ak string) string {
	return fmt.Sprintf("%s***%s", ak[:3], ak[len(ak)-3:])
}

func (x *AccountConfig) postinit() {
	if x.DisplayName == "" {
		x.DisplayName = defaultDisplayNameFromAK(x.AK)
	}
	if x.ManagedCertNamePrefix == "" {
		x.ManagedCertNamePrefix = defaultManagedCertNamePrefix
	}

	x.qiniuCreds = auth.New(x.AK, x.SK)
}

//////////////////////////////////////////////////////////////////////////////

type ctxConfigKey struct{}

func setConfig(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, ctxConfigKey{}, cfg)
}

func getConfig(ctx context.Context) *Config {
	return ctx.Value(ctxConfigKey{}).(*Config)
}

//////////////////////////////////////////////////////////////////////////////

func configFromTOML(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var cfg Config
	_, err = toml.Decode(string(data), &cfg)
	if err != nil {
		return nil, err
	}

	for _, acc := range cfg.Accounts {
		acc.postinit()
	}

	return &cfg, nil
}

func configFromEnv() (*Config, error) {
	numAccountsStr := os.Getenv("QCR_NUM_ACCOUNTS")
	if len(numAccountsStr) == 0 {
		return nil, nil
	}

	numAccounts, err := strconv.ParseInt(numAccountsStr, 10, 0)
	if err != nil {
		return nil, err
	}

	var accounts []*AccountConfig
	for i := 1; i <= int(numAccounts); i++ {
		accCfg, err := accountConfigFromEnv(i)
		if err != nil {
			return nil, err
		}
		accCfg.postinit()
		accounts = append(accounts, accCfg)
	}

	return &Config{Accounts: accounts}, nil
}

func envKey(idx int, kind string) string {
	return fmt.Sprintf("QCR_ACCOUNT_%d_%s", idx, kind)
}

func getenvForAccount(idx int, kind string) string {
	return os.Getenv(envKey(idx, kind))
}

func accountConfigFromEnv(idx int) (*AccountConfig, error) {
	ak := getenvForAccount(idx, "AK")
	if len(ak) == 0 {
		return nil, fmt.Errorf("no AK for account #%d", idx)
	}

	sk := getenvForAccount(idx, "SK")
	if len(sk) == 0 {
		return nil, fmt.Errorf("no SK for account #%d", idx)
	}

	displayName := getenvForAccount(idx, "DISPLAY_NAME")
	prefix := getenvForAccount(idx, "MANAGED_CERT_NAME_PREFIX")

	return &AccountConfig{
		AK:                    ak,
		SK:                    sk,
		DisplayName:           displayName,
		ManagedCertNamePrefix: prefix,
	}, nil
}
