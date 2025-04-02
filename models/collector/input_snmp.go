package collector

import (
	"fmt"
	"time"
)

// SNMPVersion SNMP 版本
type SNMPVersion string

const (
	// SNMPv1 版本 1
	SNMPv1 SNMPVersion = "1"
	// SNMPv2c 版本 2c
	SNMPv2c SNMPVersion = "2c"
	// SNMPv3 版本 3
	SNMPv3 SNMPVersion = "3"
)

// SNMPSecLevel SNMPv3 安全級別
type SNMPSecLevel string

const (
	// NoAuthNoPriv 無認證無加密
	NoAuthNoPriv SNMPSecLevel = "noAuthNoPriv"
	// AuthNoPriv 有認證無加密
	AuthNoPriv SNMPSecLevel = "authNoPriv"
	// AuthPriv 有認證有加密
	AuthPriv SNMPSecLevel = "authPriv"
)

// SNMPAuthProtocol SNMPv3 認證協議
type SNMPAuthProtocol string

const (
	// AuthMD5 MD5 認證
	AuthMD5 SNMPAuthProtocol = "MD5"
	// AuthSHA SHA 認證
	AuthSHA SNMPAuthProtocol = "SHA"
	// AuthSHA224 SHA224 認證
	AuthSHA224 SNMPAuthProtocol = "SHA224"
	// AuthSHA256 SHA256 認證
	AuthSHA256 SNMPAuthProtocol = "SHA256"
	// AuthSHA384 SHA384 認證
	AuthSHA384 SNMPAuthProtocol = "SHA384"
	// AuthSHA512 SHA512 認證
	AuthSHA512 SNMPAuthProtocol = "SHA512"
)

// SNMPPrivProtocol SNMPv3 加密協議
type SNMPPrivProtocol string

const (
	// PrivDES DES 加密
	PrivDES SNMPPrivProtocol = "DES"
	// PrivAES AES 加密
	PrivAES SNMPPrivProtocol = "AES"
	// PrivAES192 AES192 加密
	PrivAES192 SNMPPrivProtocol = "AES192"
	// PrivAES256 AES256 加密
	PrivAES256 SNMPPrivProtocol = "AES256"
)

// DefaultSNMPConfig 默認 SNMP 配置
var DefaultSNMPConfig = SNMPConfig{
	Port:      161,
	Community: "public",
	Version:   string(SNMPv2c),
	Timeout:   5 * time.Second,
	Retries:   3,
}

// SNMPConfig SNMP 配置
type SNMPConfig struct {
	// 基本配置
	Host      string        `json:"host" yaml:"host"`
	Port      int           `json:"port" yaml:"port"`
	Community string        `json:"community" yaml:"community"`
	Version   string        `json:"version" yaml:"version"`
	Timeout   time.Duration `json:"timeout" yaml:"timeout"`
	Retries   int           `json:"retries" yaml:"retries"`
	MIBPath   string        `json:"mib_path" yaml:"mib_path"`

	// 高級配置
	UnconnectedUDPSocket bool   `json:"unconnected_udp_socket" yaml:"unconnected_udp_socket"` // 是否接受來自任何地址的響應
	MaxRepetitions       int    `json:"max_repetitions" yaml:"max_repetitions"`               // GETBULK 最大重複次數
	ContextName          string `json:"context_name" yaml:"context_name"`                     // SNMPv3 上下文名稱

	// SNMPv3 配置
	SecName      string `json:"sec_name" yaml:"sec_name"`           // 安全名稱
	AuthProtocol string `json:"auth_protocol" yaml:"auth_protocol"` // 認證協議
	AuthPassword string `json:"auth_password" yaml:"auth_password"` // 認證密碼
	SecLevel     string `json:"sec_level" yaml:"sec_level"`         // 安全級別
	PrivProtocol string `json:"priv_protocol" yaml:"priv_protocol"` // 加密協議
	PrivPassword string `json:"priv_password" yaml:"priv_password"` // 加密密碼
}

// MergeWithDefault 將收集器配置與默認配置合併
func (c *SNMPCollectorConfig) MergeWithDefault() SNMPConfig {
	config := DefaultSNMPConfig

	// 只覆蓋非空值
	if c.Community != "" {
		config.Community = c.Community
	}
	if c.Version != "" {
		config.Version = c.Version
	}
	if c.Port > 0 {
		config.Port = c.Port
	}
	if c.Timeout > 0 {
		config.Timeout = c.Timeout
	}
	if c.Retries > 0 {
		config.Retries = c.Retries
	}

	return config
}

// Validate 驗證配置是否有效
func (c *SNMPConfig) Validate() error {
	// 檢查必填字段
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}

	// 檢查版本
	switch SNMPVersion(c.Version) {
	case SNMPv1, SNMPv2c:
		if c.Community == "" {
			return fmt.Errorf("community is required for SNMP v1/v2c")
		}
	case SNMPv3:
		if c.SecName == "" {
			return fmt.Errorf("sec_name is required for SNMP v3")
		}
		// 檢查安全級別
		switch SNMPSecLevel(c.SecLevel) {
		case NoAuthNoPriv:
			// 無需額外驗證
		case AuthNoPriv:
			if c.AuthPassword == "" {
				return fmt.Errorf("auth_password is required for authNoPriv security level")
			}
		case AuthPriv:
			if c.AuthPassword == "" {
				return fmt.Errorf("auth_password is required for authPriv security level")
			}
			if c.PrivPassword == "" {
				return fmt.Errorf("priv_password is required for authPriv security level")
			}
		default:
			return fmt.Errorf("unsupported security level: %s", c.SecLevel)
		}
	default:
		return fmt.Errorf("unsupported SNMP version: %s", c.Version)
	}

	return nil
}
