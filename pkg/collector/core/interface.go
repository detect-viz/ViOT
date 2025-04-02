package core

import (
	"context"

	"viot/models"
)

// Input 定義輸入插件的接口
type Input interface {
	// Name 返回輸入插件的名稱
	Name() string
	// Start 啟動輸入插件
	Start(ctx context.Context) error
	// Stop 停止輸入插件
	Stop() error
	// Status 返回輸入插件的狀態
	Status() models.CollectorStatus
	// Collect 收集數據
	Collect(ctx context.Context) ([]models.RestAPIPoint, error)
	// Deploy 部署配置
	Deploy(ctx context.Context) error
}

// DataHandler 定義數據處理器的接口
type DataHandler interface {
	// HandleData 處理收集到的數據
	HandleData(ctx context.Context, data []models.RestAPIPoint) error
}

// DeviceDiscoverer 設備發現接口
type DeviceDiscoverer interface {
	// ScanIP 掃描IP是否可達
	ScanIP(ctx context.Context, ipRange models.IPRange) ([]string, error)

	// ScanPort 掃描端口是否開放
	ScanPort(ctx context.Context, ip string, ports []int) (map[int]bool, error)

	// GetDeviceInfo 獲取設備信息
	GetDeviceInfo(ctx context.Context, ip string, protocol string) (*models.Device, error)
}

// DeviceDeployer 設備部署接口
type DeviceDeployer interface {
	// Deploy 部署設備配置
	Deploy(device models.Device) error

	// Verify 驗證設備配置
	Verify(device models.Device) error

	// Rollback 回滾設備配置
	Rollback(device models.Device) error
}

// DeviceMatcher 設備匹配接口
type DeviceMatcher interface {
	// Match 匹配設備
	Match(device models.Device, matchFields []string) (bool, error)

	// GetCollectFields 獲取採集字段
	GetCollectFields(device models.Device) ([]models.CollectField, error)
}
