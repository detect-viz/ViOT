package inputs

import (
	"context"
	"fmt"
	"time"

	"viot/logger"
	"viot/models"
	"viot/pkg/collector/core"
)

// BaseDeviceInput 基礎設備輸入插件
type BaseDeviceInput struct {
	*core.BaseInput
	config *models.CollectorConfig
	logger logger.Logger
}

// NewBaseDeviceInput 創建新的基礎設備輸入插件
func NewBaseDeviceInput(name string, config *models.CollectorConfig, logger logger.Logger) *BaseDeviceInput {
	return &BaseDeviceInput{
		BaseInput: core.NewBaseInput(name),
		config:    config,
		logger:    logger,
	}
}

// Start 啟動設備輸入插件
func (d *BaseDeviceInput) Start(ctx context.Context) error {
	if err := d.BaseInput.Start(ctx); err != nil {
		return fmt.Errorf("failed to start base input: %v", err)
	}

	d.logger.Info(fmt.Sprintf("設備輸入插件 %s 啟動", d.Name()))
	return nil
}

// Stop 停止設備輸入插件
func (d *BaseDeviceInput) Stop() error {
	if err := d.BaseInput.Stop(); err != nil {
		return fmt.Errorf("failed to stop base input: %v", err)
	}

	d.logger.Info(fmt.Sprintf("設備輸入插件 %s 停止", d.Name()))
	return nil
}

// Collect 收集設備數據
func (d *BaseDeviceInput) Collect(ctx context.Context, device models.Device) ([]models.DeviceData, error) {
	d.logger.Debug(fmt.Sprintf("收集設備數據: %s", device.IP))
	return nil, fmt.Errorf("collect not implemented")
}

// Discover 發現設備
func (d *BaseDeviceInput) Discover(ctx context.Context, ipRange models.IPRange) ([]models.Device, error) {
	d.logger.Debug(fmt.Sprintf("發現設備: %s - %s", ipRange.StartIP, ipRange.EndIP))
	return nil, fmt.Errorf("discover not implemented")
}

// Match 匹配設備
func (d *BaseDeviceInput) Match(device models.Device) (bool, error) {
	d.logger.Debug(fmt.Sprintf("匹配設備: %s", device.IP))
	return false, fmt.Errorf("match not implemented")
}

// Deploy 部署設備配置
func (d *BaseDeviceInput) Deploy(device models.Device) error {
	d.logger.Debug(fmt.Sprintf("部署設備配置: %s", device.IP))
	return fmt.Errorf("deploy not implemented")
}

// UpdateStatus 更新狀態
func (d *BaseDeviceInput) UpdateStatus(dataPoints int64, errorCount int64) {
	status := d.Status()
	status.DataPoints += dataPoints
	status.ErrorCount += errorCount

	total := status.DataPoints + status.ErrorCount
	if total > 0 {
		status.SuccessRate = float64(status.DataPoints) / float64(total) * 100
	}

	status.LastCollected = time.Now()
}
