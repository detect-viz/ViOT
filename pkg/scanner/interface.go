package scanner

import (
	"context"

	"viot/models"
)

// Device 表示掃描到的設備信息
type Device struct {
	IP           string
	Port         int
	Protocol     string // "snmp", "modbus"
	Manufacturer string
	Model        string
	Type         string // "pdu", "switch", "ap", "iot"
	Status       string // "online", "offline"
	Details      map[string]interface{}
}

// Scanner 掃描器接口
type Scanner interface {
	// Start 開始掃描
	Start(ctx context.Context) error
	// Stop 停止掃描
	Stop() error
	// Scan 掃描特定IP和端口
	Scan(ctx context.Context, ip string, port int) (*models.DeviceInfo, error)
}

// ScanManager 定義掃描管理器介面
type ScanManager interface {
	// Init 初始化掃描管理器
	Init() error

	// RegisterScanner 註冊掃描器
	RegisterScanner(protocol string, scanner Scanner)

	// StartScan 開始掃描
	StartScan(config *models.ScanSetting) error

	// StopScan 停止掃描
	StopScan() error

	// GetScanStatus 獲取當前掃描狀態
	GetScanStatus() models.ScanStatus

	// SetResultHandler 設置掃描結果處理器
	SetResultHandler(handler func(models.DeviceInfo))
}
