package models

import (
	"time"
)

// RestAPIPoint API 接收數據點
type RestAPIPoint struct {
	Metric    string                 `json:"metric"`    // 指標名稱
	Fields    map[string]interface{} `json:"fields"`    // 字段
	Value     float64                `json:"value"`     // 指標值
	Tags      map[string]string      `json:"tags"`      // 標籤
	Timestamp time.Time              `json:"timestamp"` // 時間戳
}

// Device 設備信息
type Device struct {
	IP           string            `json:"ip"`
	Factory      string            `json:"factory"`
	Phase        string            `json:"phase"`
	Datacenter   string            `json:"datacenter"`
	Room         string            `json:"room"`
	Protocol     string            `json:"protocol"`
	Type         string            `json:"type"`
	MAC          string            `json:"mac"`
	Manufacturer string            `json:"manufacturer"`
	Model        string            `json:"model"`
	Version      string            `json:"version"`
	SerialNumber string            `json:"serial_number"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Tags         map[string]string `json:"tags,omitempty"`
}

// DeviceData 設備數據
type DeviceData struct {
	Name      string                 `json:"name"`      // 設備名稱
	Metric    string                 `json:"metric"`    // 指標名稱
	Value     float64                `json:"value"`     // 指標值
	Timestamp time.Time              `json:"timestamp"` // 時間戳
	Tags      map[string]string      `json:"tags"`      // 標籤
	Fields    map[string]interface{} `json:"fields"`    // 字段
}


// PDUScale PDU數據比例因子
type PDUScale struct {
	Manufacturer string
	Current      float64
	Voltage      float64
	Power        float64
	Energy       float64
}

// PDUManufacturerHandler PDU製造商處理器介面
type PDUManufacturerHandler interface {
	CanHandle(manufacturer, model string) bool
	ProcessFields(point PDUPoint) (map[string]float64, error)
	GetScaleFactor() PDUScale
}

// PDUPoint PDU數據點
type PDUPoint struct {
	Name      string                 `json:"name"`
	Timestamp time.Time              `json:"timestamp"`
	Tags      map[string]string      `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
}


// Branch 分支數據
type Branch struct {
	ID      string  `json:"id"`
	Current float64 `json:"current,omitempty"`
	Voltage float64 `json:"voltage,omitempty"`
	Power   float64 `json:"power,omitempty"`
	Energy  float64 `json:"energy,omitempty"`
}

// Phase 相位數據
type Phase struct {
	ID      string  `json:"id"`
	Current float64 `json:"current,omitempty"`
	Voltage float64 `json:"voltage,omitempty"`
	Power   float64 `json:"power,omitempty"`
	Energy  float64 `json:"energy,omitempty"`
}

// DeviceConfig 設備配置
type DeviceConfig struct {
	Name          string
	Protocol      string
	Manufacturer  string
	Model         string
	Type          string
	MatchField    string
	MatchKeyword  string
	AutoDeploy    bool
	CollectFields []CollectField
}

// CollectField 採集字段
type CollectField struct {
	Name        string
	OID         string
	Register    int
	DataType    string
	Scale       float64
	Description string
}

// DeviceInfo 設備信息
type DeviceInfo struct {
	IP           string
	Factory      string
	Phase        string
	Datacenter   string
	Room         string
	Manufacturer string
	Model        string
	Version      string
	SerialNumber string
	Type         string
}
