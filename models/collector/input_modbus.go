package collector

import "time"

// ModbusDataType Modbus 數據類型
type ModbusDataType string

const (
	// INT16 16位整數
	INT16 ModbusDataType = "INT16"
	// UINT16 16位無符號整數
	UINT16 ModbusDataType = "UINT16"
	// INT32 32位整數
	INT32 ModbusDataType = "INT32"
	// UINT32 32位無符號整數
	UINT32 ModbusDataType = "UINT32"
	// INT64 64位整數
	INT64 ModbusDataType = "INT64"
	// UINT64 64位無符號整數
	UINT64 ModbusDataType = "UINT64"
	// FLOAT32 32位浮點數
	FLOAT32 ModbusDataType = "FLOAT32"
	// FLOAT64 64位浮點數
	FLOAT64 ModbusDataType = "FLOAT64"
	// STRING 字符串
	STRING ModbusDataType = "STRING"
)

// ModbusRegisterType Modbus 寄存器類型
type ModbusRegisterType string

const (
	// Coil 線圈寄存器
	Coil ModbusRegisterType = "coil"
	// DiscreteInput 離散輸入寄存器
	DiscreteInput ModbusRegisterType = "discrete_input"
	// HoldingRegister 保持寄存器
	HoldingRegister ModbusRegisterType = "holding_register"
	// InputRegister 輸入寄存器
	InputRegister ModbusRegisterType = "input_register"
)

// ModbusRegister Modbus 寄存器配置
type ModbusRegister struct {
	Address     uint16             `json:"address" yaml:"address"`         // 寄存器地址
	Type        ModbusRegisterType `json:"type" yaml:"type"`               // 寄存器類型
	DataType    ModbusDataType     `json:"data_type" yaml:"data_type"`     // 數據類型
	Name        string             `json:"name" yaml:"name"`               // 寄存器名稱
	ByteOrder   string             `json:"byte_order" yaml:"byte_order"`   // 字節順序 (big-endian/little-endian)
	WordOrder   string             `json:"word_order" yaml:"word_order"`   // 字順序 (big-endian/little-endian)
	Scale       float64            `json:"scale" yaml:"scale"`             // 縮放因子
	Unit        string             `json:"unit" yaml:"unit"`               // 單位
	Description string             `json:"description" yaml:"description"` // 描述
}

// DefaultModbusConfig 默認 Modbus 配置
var DefaultModbusConfig = ModbusConfig{
	Port:    502,
	SlaveID: 1,
	Timeout: 5 * time.Second,
	Retries: 3,
}

// ModbusConfig Modbus 配置
type ModbusConfig struct {
	Host      string        `json:"host" yaml:"host"`             // 主機地址
	Port      int           `json:"port" yaml:"port"`             // 端口
	SlaveID   byte          `json:"slave_id" yaml:"slave_id"`     // 從站 ID
	Timeout   time.Duration `json:"timeout" yaml:"timeout"`       // 超時時間
	Retries   int           `json:"retries" yaml:"retries"`       // 重試次數
	ByteOrder string        `json:"byte_order" yaml:"byte_order"` // 默認字節順序
	WordOrder string        `json:"word_order" yaml:"word_order"` // 默認字順序
}

// ModbusCollectorConfig Modbus 收集器配置
type ModbusCollectorConfig struct {
	Host         string           `json:"host" yaml:"host"`
	Port         int              `json:"port" yaml:"port"`
	SlaveID      byte             `json:"slave_id" yaml:"slave_id"`
	Timeout      time.Duration    `json:"timeout" yaml:"timeout"`
	Retries      int              `json:"retries" yaml:"retries"`
	PollInterval string           `json:"poll_interval" yaml:"poll_interval"`
	Registers    []ModbusRegister `json:"registers" yaml:"registers"`
}

// MergeWithDefault 將收集器配置與默認配置合併
func (c *ModbusCollectorConfig) MergeWithDefault() ModbusConfig {
	config := DefaultModbusConfig

	// 只覆蓋非空值
	if c.Host != "" {
		config.Host = c.Host
	}
	if c.Port > 0 {
		config.Port = c.Port
	}
	if c.SlaveID > 0 {
		config.SlaveID = c.SlaveID
	}
	if c.Timeout > 0 {
		config.Timeout = c.Timeout
	}
	if c.Retries > 0 {
		config.Retries = c.Retries
	}

	return config
}
