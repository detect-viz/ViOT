package inputs

import (
	"context"
	"fmt"

	"viot/logger"
	"viot/models"
)

// TelegrafInput Telegraf 輸入插件
type TelegrafInput struct {
	*BaseDeviceInput
	processor models.TelegrafProcessor
	points    []models.RestAPIPoint
}

// NewTelegrafInput 創建新的 Telegraf 輸入插件
func NewTelegrafInput(config *models.CollectorConfig, logger logger.Logger, processor models.TelegrafProcessor) *TelegrafInput {
	return &TelegrafInput{
		BaseDeviceInput: NewBaseDeviceInput("telegraf", config, logger),
		processor:       processor,
		points:          make([]models.RestAPIPoint, 0),
	}
}

// Start 啟動 Telegraf 輸入插件
func (t *TelegrafInput) Start(ctx context.Context) error {
	if err := t.BaseDeviceInput.Start(ctx); err != nil {
		return fmt.Errorf("failed to start base device input: %v", err)
	}

	t.logger.Info(fmt.Sprintf("Telegraf 輸入插件啟動"))
	return nil
}

// Stop 停止 Telegraf 輸入插件
func (t *TelegrafInput) Stop() error {
	if err := t.BaseDeviceInput.Stop(); err != nil {
		return fmt.Errorf("failed to stop base device input: %v", err)
	}

	t.logger.Info(fmt.Sprintf("Telegraf 輸入插件停止"))
	return nil
}

// Collect 收集 Telegraf 設備數據
func (t *TelegrafInput) Collect(ctx context.Context, device models.Device) ([]models.DeviceData, error) {
	if device.Type != "telegraf" {
		return nil, fmt.Errorf("device type is not telegraf")
	}

	// 轉換為設備數據
	var deviceData []models.DeviceData
	for _, point := range t.points {
		data := models.DeviceData{
			Name:      device.IP,
			Timestamp: point.Timestamp,
			Tags:      point.Tags,
			Fields:    point.Fields,
		}
		deviceData = append(deviceData, data)
	}

	return deviceData, nil
}

// Match 匹配 Telegraf 設備
func (t *TelegrafInput) Match(device models.Device) (bool, error) {
	return device.Type == "telegraf", nil
}

// Deploy 部署 Telegraf 設備配置
func (t *TelegrafInput) Deploy(device models.Device) error {
	if device.Type != "telegraf" {
		return fmt.Errorf("device type is not telegraf")
	}

	// 部署 Telegraf 配置
	// TODO: 實現 Telegraf 配置部署邏輯

	return nil
}

// HandleTelegrafData 處理 Telegraf 數據
func (t *TelegrafInput) HandleTelegrafData(ctx context.Context, data []models.RestAPIPoint) error {
	t.points = data
	return nil
}
