package models

// ScannerSettings 掃描器設置
type Config struct {
	Scanner      ScannerSettings `yaml:"scanner"`
	Modbus       ModbusSettings  `yaml:"modbus"`
	SNMP         SNMPSettings    `yaml:"snmp"`
	Dependencies []Dependencies  `yaml:"dependencies"`
}

// ScannerSettings 掃描器設置
type ScannerSettings struct {
	ScanInterval  int    `yaml:"scan_interval"`  // 掃描間隔（秒）
	IPRangeFile   string `yaml:"ip_range_file"`  // IP 範圍檔案路徑
	TargetFile    string `yaml:"target_file"`    // 目標檔案路徑
	MaxConcurrent int    `yaml:"max_concurrent"` // 最大並發數
	FPingTimeout  int    `yaml:"fping_timeout"`  // fping 超時時間（毫秒）
	FPingRetries  int    `yaml:"fping_retries"`  // fping 重試次數

}

// ModbusSettings Modbus 協議設置
type ModbusSettings struct {
	Port    []int  `yaml:"port"`
	Mode    string `yaml:"mode"`
	Timeout int    `yaml:"timeout"`
	Retry   int    `yaml:"retry"`
	SlaveID []int  `yaml:"slave_id"`
}

// SNMPSettings SNMP 協議設置
type SNMPSettings struct {
	Port      int    `yaml:"port"`
	Mode      string `yaml:"mode"`
	Community string `yaml:"community"`
	Version   int    `yaml:"version"`
	Timeout   int    `yaml:"timeout"`
	Retry     int    `yaml:"retry"`
}

// ModbusDependency 設備依賴關係
type Dependencies struct {
	Protocol     string   `yaml:"protocol"`
	SourceTarget string   `yaml:"source_target"`
	LinkedTarget []string `yaml:"linked_target"`
}

// TargetSettings 目標設備配置 YAML
type TargetSettings struct {
	Targets []Target `yaml:"targets"`
}

type Target struct {
	Name         string `yaml:"name"`
	InstanceType string `yaml:"instance_type"`
	Manufacturer string `yaml:"manufacturer"`
	Model        string `yaml:"model"`
	Protocol     string `yaml:"protocol"`
	SNMP         struct {
		MatchValueContains string `yaml:"match_value_contains"`
		MatchOID           string `yaml:"match_oid"`
	}
	Modbus struct {
		MatchValueContains string `yaml:"match_value_contains"`
		MatchRegisterType  string `yaml:"match_register_type"` // modbus
		MatchAddress       int    `yaml:"match_address"`       // modbus
		MatchLength        int    `yaml:"match_length"`        // modbus
	}
}

// IPRange IP 範圍配置
type IPRange struct {
	Factory    string `yaml:"factory"`
	Phase      string `yaml:"phase"`
	Datacenter string `yaml:"datacenter"`
	Room       string `yaml:"room"`
	Range      string `yaml:"range"`
	StartIP    string `yaml:"start"`
	EndIP      string `yaml:"end"`
}
