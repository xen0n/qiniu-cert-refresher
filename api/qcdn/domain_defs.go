// SPDX-License-Identifier: GPL-3.0-or-later

package qcdn

import "time"

// Keep the definitions here synced with the official docs:
//
// https://developer.qiniu.com/fusion/4246/the-domain-name

// HTTPSConfig 域名 HTTPS 配置
//
// https://developer.qiniu.com/fusion/4249/product-features#7
type HTTPSConfig struct {
	// CertID 证书ID, 如果Protocol是https, 则需要填写该证书id
	CertID string `json:"certId"`
	// ForceHTTPS 是否开启强制https访问，即http访问时30x到https
	ForceHTTPS bool `json:"forceHttps"`
	// HTTP2Enabled http2功能是否启用，false为关闭，true为开启
	HTTP2Enabled bool `json:"http2Enable"`
}

type DomainType string

const (
	// DomainTypeNormal 普通域名
	DomainTypeNormal DomainType = "normal"
	// DomainTypeWildcard 泛域名
	DomainTypeWildcard DomainType = "wildcard"
	// DomainTypePan 泛子域名
	DomainTypePan DomainType = "pan"
	// DomainTypeEvaluation 测试域名
	DomainTypeEvaluation DomainType = "test"
)

type IPType uint

const (
	IPTypeV4   IPType = 0x1
	IPTypeV6   IPType = 0x2
	IPTypeV4V6 IPType = IPTypeV4 | IPTypeV6
)

// OpKind 域名操作类型
type OpKind string

const (
	OpKindCreateDomain   OpKind = "create_domain"
	OpKindOfflineDomain  OpKind = "offline_domain"
	OpKindOnlineDomain   OpKind = "online_domain"
	OpKindModifySource   OpKind = "modify_source"
	OpKindModifyReferer  OpKind = "modify_referer"
	OpKindModifyCache    OpKind = "modify_cache"
	OpKindFreezeDomain   OpKind = "freeze_domain"
	OpKindUnfreezeDomain OpKind = "unfreeze_domain"
	OpKindModifyTimeACL  OpKind = "modify_timeacl" // 修改时间戳防盗链
	OpKindModifyHTTPSCrt OpKind = "modify_https_crt"
	OpKindSSLize         OpKind = "sslize"         // 升级HTTPS
	OpKindModifyBsauth   OpKind = "modify_bsauth"  // 修改回源鉴权
	OpKindOfflineBsauth  OpKind = "offline_bsauth" // 删除回源鉴权
)

// OpStatus 域名操作状态
type OpStatus string

const (
	OpStatusProcessing OpStatus = "processing"
	OpStatusSuccessful OpStatus = "success"
	OpStatusFailed     OpStatus = "failed"
	OpStatusFrozen     OpStatus = "frozen"
	OpStatusOfflined   OpStatus = "offlined"
)

type Domain struct {
	// Name 域名, 如果是泛域名，必须以点号 . 开头
	Name string `json:"name"`
	// Type 域名类型
	Type DomainType `json:"type"`
	// CName 创建域名成功后七牛生成的域名，用户需要把 Name cname 到这个域名
	CName string `json:"cname"`
	// TestURLPath 域名的测试资源，需要保证这个资源是可访问的
	TestURLPath string `json:"testURLPath"`
	Platform    string `json:"platform"`
	GeoCover    string `json:"geoCover"`
	Protocol    string `json:"protocol"`
	IPTypes     IPType `json:"ipTypes"`
	// TagList 域名的标签列表
	TagList []string `json:"tagList"`
	// Source 结构请参考 回源配置
	Source any `json:"source"`
	// Cache 结构请参考 缓存策略
	Cache any `json:"cache"`
	// Referer 结构请参考 referer防盗链
	Referer any `json:"referer"`
	// IPACL 结构请参考 ip黑白名单
	IPACL any `json:"ipACL"`
	// TimeACL 结构请参考 时间戳防盗链
	TimeACL any `json:"timeACL"`
	// Bsauth 结构请参考 回源鉴权
	Bsauth any `json:"bsauth"`
	// LastOp 域名最近一次操作类型
	LastOp OpKind `json:"operationType"`
	// LastOpStatus 域名最近一次的操作状态
	LastOpStatus       OpStatus `json:"operatingState"`
	OperatingStateDesc string   `json:"operatingStateDesc"`
	// CreatedAt 域名创建时间
	CreatedAt time.Time `json:"createAt"`
	// ModifiedAt 域名最后一次修改时间
	ModifiedAt time.Time `json:"modifyAt"`
	// ParentDomain 父域名,属于泛子域名字段
	ParentDomain string `json:"pareDomain"`
	// HTTPS HTTPS 配置
	HTTPS *HTTPSConfig `json:"https"`
}

type ReqListDomains struct {
	// Types 域名类型，可选normal（普通域名）、wildcard（泛域名）、pan（泛子域名）、test（测试域名）中的一个或多个，不填默认查询全部域名。
	Types []DomainType
	// CertID 证书ID，不填默认查询全部域名。
	CertID string
	// SourceTypes 回源类型，可选domain、ip、qiniuBucket、advanced中的一个或多个，不填默认查询全部域名
	//
	// 如果指定了SourceQiniuBucket参数，SourceTypes只能指定为qiniuBucket一种回源类型，否则SourceQiniuBucket参数将不生效；
	// 如果指定了SourceIp参数，SourceTypes只能指定为ip一种回源类型，否则SourceIp参数将不生效；
	// 同时获取多种回源类型域名的请求url示例：http://api.qiniu.com/domain?sourceTypes=domain&sourceTypes=ip
	SourceTypes []string
	// SourceQiniuBucket 七牛存储空间名称，不填默认查询全部域名。
	//
	// 请求url示例：http://api.qiniu.com/domain?sourceTypes=qiniuBucket&sourceQiniuBucket=test
	SourceQiniuBucket string
	// SourceIP 回源IP, 不填默认查询全部域名。
	//
	// 请求url示例：http://api.qiniu.com/domain?sourceTypes=ip&sourceIp=1.1.1.1
	SourceIP string
	// Marker 用于标示从哪个位置开始获取域名列表，不填或空表示从头开始。
	Marker string
	// Limit 返回的最大域名个数。1~1000, 不填默认为 10
	Limit int
}

type RespListDomains struct {
	// Marker 用于标示下一次从哪个位置开始获取域名列表
	Marker  string    `json:"marker"`
	Domains []*Domain `json:"domains"`
}
