package interfaces

import (
	"context"
	"viot/models"
	"viot/models/collector"
	"viot/models/scanner"
)

// Collector 收集器接口
type Collector interface {
	// Start 啟動收集器
	Start(ctx context.Context) error
	// Stop 停止收集器
	Stop() error
	// Status 獲取收集器狀態
	Status() collector.CollectorStatus
	// Name 返回收集器名稱
	Name() string
	// Collect 收集設備數據
	Collect(ctx context.Context, device models.Device) ([]models.DeviceData, error)
	// Discover 發現設備
	Discover(ctx context.Context, ipRange scanner.IPRange) ([]models.Device, error)
	// Match 匹配設備
	Match(device models.Device) (bool, error)
	// Deploy 部署設備配置
	Deploy(device models.Device) error
}
