package processor

import (
	"viot/logger"
	"viot/models"
)

// TelegrafProcessor Telegraf 數據處理器
type TelegrafProcessor struct {
	*BaseProcessor
}

// NewTelegrafProcessor 創建 Telegraf 處理器
func NewTelegrafProcessor(config models.ProcessorConfig, logger logger.Logger) *TelegrafProcessor {
	return &TelegrafProcessor{
		BaseProcessor: NewBaseProcessor("telegraf", config, logger),
	}
}

// Process 處理設備數據
func (p *TelegrafProcessor) Process(data []models.DeviceData) error {
	// 將設備數據轉換為 Telegraf 數據點
	var points []models.RestAPIPoint
	for _, d := range data {
		points = append(points, models.RestAPIPoint{
			Metric:    d.Metric,
			Value:     d.Value,
			Timestamp: d.Timestamp,
			Tags:      d.Tags,
			Fields:    d.Fields,
		})
	}

	return p.ProcessTelegraf(points)
}

// ProcessTelegraf 處理 Telegraf 數據
func (p *TelegrafProcessor) ProcessTelegraf(points []models.RestAPIPoint) error {
	// 將 Telegraf 數據點轉換為設備數據
	var deviceData []models.DeviceData
	for _, point := range points {
		deviceData = append(deviceData, models.DeviceData{
			Name:      point.Tags["device"],
			Metric:    point.Metric,
			Value:     point.Value,
			Timestamp: point.Timestamp,
			Tags:      point.Tags,
			Fields:    point.Fields,
		})
	}

	// 調用基礎處理方法
	return p.BaseProcessor.Process(deviceData)
}
