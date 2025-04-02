package webservice

// WebConfig Web服務配置
type WebserviceConfig struct {
	Host       string          `json:"host"`
	Port       int             `json:"port"`
	APIPrefix  string          `json:"api_prefix"`
	StaticPath string          `json:"static_path"`
	SSL        SSLConfig       `json:"ssl"`
	Auth       AuthConfig      `json:"auth"`
	RateLimit  RateLimitConfig `json:"rate_limit"`
	CORS       CORSConfig      `json:"cors"`
}

// SSLConfig SSL配置
type SSLConfig struct {
	Enabled    bool   `json:"enabled"`
	CertFile   string `json:"cert_file"`
	KeyFile    string `json:"key_file"`
	MinVersion string `json:"min_version"`
	MaxVersion string `json:"max_version"`
}

// AuthConfig 認證配置
type AuthConfig struct {
	Enabled  bool   `json:"enabled"`
	Type     string `json:"type"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	Enabled bool `json:"enabled"`
	Rate    int  `json:"rate"`
	Burst   int  `json:"burst"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	Enabled     bool     `json:"enabled"`
	Origins     []string `json:"origins"`
	Methods     []string `json:"methods"`
	Headers     []string `json:"headers"`
	Credentials bool     `json:"credentials"`
}
