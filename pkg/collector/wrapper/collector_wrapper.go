package collector

import (
	"context"
	"fmt"

	"viot/models"
	"viot/pkg/processor"

	"go.uber.org/zap"
)

// ProcessorHandler 將收集的數據發送到處理器的處理程序
type ProcessorHandler struct {
	processor *processor.BaseProcessor
	logger    *zap.Logger
}

// NewProcessorHandler 創建一個將數據發送到處理器的處理器
func NewProcessorHandler(proc *processor.BaseProcessor, logger *zap.Logger) *ProcessorHandler {
	return &ProcessorHandler{
		processor: proc,
		logger:    logger.Named("processor-handler"),
	}
}

// HandleData 實現DataHandler接口，處理收集到的數據並發送到處理器
func (h *ProcessorHandler) HandleData(ctx context.Context, data []models.RestAPIPoint) error {
	if len(data) == 0 {
		return nil
	}

	h.logger.Debug("處理收集到的數據點", zap.Int("count", len(data)))

	// 將數據轉換為設備數據格式
	deviceData := make([]models.DeviceData, 0, len(data))
	for _, point := range data {
		// 從 Tags 中提取設備名稱
		deviceName, ok := point.Tags["device"]
		if !ok {
			h.logger.Warn("數據點缺少設備名稱標籤", zap.Any("tags", point.Tags))
			continue
		}

		// 轉換 RestAPIPoint 為 DeviceData
		dd := models.DeviceData{
			Name:      deviceName,
			Metric:    point.Metric,
			Value:     point.Value,
			Timestamp: point.Timestamp,
			Tags:      point.Tags,
			Fields:    point.Fields,
		}
		deviceData = append(deviceData, dd)
	}

	if len(deviceData) == 0 {
		h.logger.Warn("沒有有效的設備數據需要處理")
		return nil
	}

	// 發送到處理器
	if err := h.processor.Process(deviceData); err != nil {
		h.logger.Error("處理數據失敗", zap.Error(err))
		h.processor.IncrementErrorCount(1)
		return fmt.Errorf("處理數據失敗: %w", err)
	}

	h.processor.IncrementProcessedCount(int64(len(deviceData)))
	h.logger.Debug("數據處理完成",
		zap.Int("count", len(deviceData)),
		zap.Int64("total_processed", h.processor.GetProcessedCount()),
		zap.Time("last_processed", h.processor.GetLastProcessTime()))

	return nil
}
