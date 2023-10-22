// SPDX-License-Identifier: GPL-3.0-or-later

package qcdn

// Keep the definitions here synced with the official docs:
//
// https://developer.qiniu.com/fusion/8593/interface-related-certificate

// Cert 证书
type Cert struct {
	// ID 证书ID
	ID string `json:"certid"`
	// Name 证书名称
	Name string `json:"name"`
	// CommonName 通用名称
	CommonName string `json:"common_name"`
	// DNSNames DNS 域名
	DNSNames []string `json:"dnsnames"`
	// NotBefore 生效时间
	NotBefore int64 `json:"not_before"`
	// NotAfter 过期时间
	NotAfter int64 `json:"not_after"`
	// PEM 证书私钥
	PEM string `json:"pri"`
	// CA 证书内容
	CA string `json:"ca"`
	// CreateTime 创建时间
	CreateTime int64 `json:"create_time"`
}

type ReqUploadCert struct {
	// Name 证书名称
	Name string `json:"name"`
	// CommonName 通用名称
	CommonName string `json:"common_name"`
	// PEM 证书私钥
	PEM string `json:"pri"`
	// CA 证书内容
	CA string `json:"ca"`
}

type RespListCerts struct {
	// Marker 用于标示下一次从哪个位置开始获取证书列表
	Marker string  `json:"marker"`
	Certs  []*Cert `json:"certs"`
}

const (
	// ErrValidityPeriodTooShort https证书有效期太短
	ErrValidityPeriodTooShort = 400322
	// ErrFailedToVerifyCertChain 验证https证书链失败
	ErrFailedToVerifyCertChain = 400323
	// ErrCertAlreadyExpired https证书过期
	ErrCertAlreadyExpired = 400329
	// ErrNoSuchCert 无此证书
	ErrNoSuchCert = 400401
	// ErrCertQuotaExceeded 超过用户绑定证书最大额度
	ErrCertQuotaExceeded = 400500
	// ErrStillBoundToCDNDomain 证书已绑定CDN域名
	ErrStillBoundToCDNDomain = 400611
	// ErrStillBoundToStorageDomain 证书已绑定存储域名
	ErrStillBoundToStorageDomain = 400911
	// ErrFailedToParseCert https证书解码失败
	ErrFailedToParseCert = 404906
	// ErrUnauthorizedForThisCert 无权操作该证书
	ErrUnauthorizedForThisCert = 404908
)
