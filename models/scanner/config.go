package scanner

import "time"

// ScannerConfig 掃描器配置
type ScannerConfig struct {
	ScanTimeout       int             `yaml:"scan_timeout"`
	MaxConcurrentScan int             `json:"max_concurrent_scan"`
	CollectorHost     string          `json:"collector_host"`
	Protocol          ProtocolConfig  `json:"protocol"`
	Devices           []DeviceConfig  `json:"devices"`
	IPRanges          []IPRangeConfig `json:"ip_ranges"`
}

// ProtocolConfig 協議配置
type ProtocolConfig struct {
	SNMP struct {
		Ports     []int  `json:"ports"`
		Community string `json:"community"`
		Version   string `json:"version"`
		Timeout   int    `json:"timeout"`
		Retries   int    `json:"retries"`
	} `json:"snmp"`
	Modbus struct {
		Ports   []int `json:"ports"`
		SlaveID []int `json:"slave_id"`
		Timeout int   `json:"timeout"`
		Retries int   `json:"retries"`
	} `json:"modbus"`
}

// IPRangeConfig IP範圍配置
type IPRangeConfig struct {
	StartIP    string
	EndIP      string
	Factory    string
	Phase      string
	Datacenter string
	Room       string
}

// IPRange 表示 IP 掃描範圍 (collector)
type IPRange struct {
	StartIP  string        `json:"start_ip" yaml:"start_ip"`
	EndIP    string        `json:"end_ip" yaml:"end_ip"`
	Protocol string        `json:"protocol" yaml:"protocol"`
	Ports    []int         `json:"ports" yaml:"ports"`
	Enabled  bool          `json:"enabled" yaml:"enabled"`
	Interval time.Duration `json:"interval" yaml:"interval"`
}
