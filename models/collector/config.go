package collector

import "time"

// CollectorConfig 收集器配置
type CollectorConfig struct {
	// 基本配置
	Enabled  bool          `json:"enabled" yaml:"enabled"`
	Interval time.Duration `json:"interval" yaml:"interval"`

	// IP掃描配置
	IPRangeFile    string        `json:"ip_range_file" yaml:"ip_range_file"`
	IPRanges       []IPRange     `json:"ip_ranges" yaml:"ip_ranges"`
	ScanInterval   time.Duration `json:"scan_interval" yaml:"scan_interval"`
	ScanTimeout    time.Duration `json:"scan_timeout" yaml:"scan_timeout"`
	ScanConcurrent int           `json:"scan_concurrent" yaml:"scan_concurrent"`

	// 部署配置
	AutoDeploy       bool   `json:"auto_deploy" yaml:"auto_deploy"`
	PendingListPath  string `json:"pending_list_path" yaml:"pending_list_path"`
	RegistryPath     string `json:"registry_path" yaml:"registry_path"`
	DeployConfigPath string `json:"deploy_config_path" yaml:"deploy_config_path"`

	// 工作目錄配置
	WorkDir string `json:"work_dir" yaml:"work_dir"`
	LogDir  string `json:"log_dir" yaml:"log_dir"`
	DataDir string `json:"data_dir" yaml:"data_dir"`

	// 協議配置
	Protocols struct {
		SNMP struct {
			Community string `json:"community" yaml:"community"`
			Version   string `json:"version" yaml:"version"`
			Port      int    `json:"port" yaml:"port"`
			Timeout   int    `json:"timeout" yaml:"timeout"`
			Retries   int    `json:"retries" yaml:"retries"`
		} `json:"snmp" yaml:"snmp"`
		Modbus struct {
			Host    string `json:"host" yaml:"host"`
			Port    int    `json:"port" yaml:"port"`
			SlaveID byte   `json:"slave_id" yaml:"slave_id"`
			Timeout int    `json:"timeout" yaml:"timeout"`
		} `json:"modbus" yaml:"modbus"`
	} `json:"protocols" yaml:"protocols"`

	Name   string `json:"name"`
	Type   string `json:"type"`
	Buffer struct {
		Size          int           `json:"size"`
		WorkerCount   int           `json:"worker_count"`
		FlushInterval time.Duration `json:"flush_interval"`
	} `json:"buffer"`
	Manager struct {
		Queue struct {
			BatchSize int `json:"batch_size"`
		} `json:"queue"`
	} `json:"manager"`
}

// SNMPCollectorConfig SNMP 收集器配置
type SNMPCollectorConfig struct {
	Community    string        `json:"community" yaml:"community"`
	Version      string        `json:"version" yaml:"version"`
	Port         int           `json:"port" yaml:"port"`
	Timeout      time.Duration `json:"timeout" yaml:"timeout"`
	Retries      int           `json:"retries" yaml:"retries"`
	OIDs         []string      `json:"oids" yaml:"oids"`
	PollInterval string        `json:"poll_interval" yaml:"poll_interval"`
}

// Protocols 協議配置
type Protocols struct {
	SNMP   SNMPCollectorConfig   `yaml:"snmp"`
	Modbus ModbusCollectorConfig `yaml:"modbus"`
}
