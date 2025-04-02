package processor

import (
	"context"
	"fmt"
	"sync"

	"viot/models"

	"go.uber.org/zap"
)

// OutputRouterImpl 實現OutputRouter接口
type OutputRouterImpl struct {
	handlers     []models.OutputHandler
	handlerMutex sync.RWMutex
	logger       *zap.Logger
}

// NewOutputRouter 創建新的輸出路由器
func NewOutputRouter(logger *zap.Logger) *OutputRouterImpl {
	return &OutputRouterImpl{
		handlers: make([]models.OutputHandler, 0),
		logger:   logger.Named("output-router"),
	}
}

// RegisterHandler 註冊輸出處理器
func (r *OutputRouterImpl) RegisterHandler(handler models.OutputHandler) error {
	r.handlerMutex.Lock()
	defer r.handlerMutex.Unlock()

	r.handlers = append(r.handlers, handler)
	r.logger.Info("註冊輸出處理程序",
		zap.String("handler", fmt.Sprintf("%T", handler)))
	return nil
}

// RoutePDUData 路由PDU數據到適當的輸出處理器
func (r *OutputRouterImpl) RoutePDUData(ctx context.Context, data ...interface{}) error {
	r.handlerMutex.RLock()
	defer r.handlerMutex.RUnlock()

	if len(r.handlers) == 0 {
		return nil
	}

	// 將interface{}轉換為PDUData類型
	var pduData []models.PDUData
	for _, item := range data {
		switch v := item.(type) {
		case models.PDUData:
			pduData = append(pduData, v)
		case []models.PDUData:
			pduData = append(pduData, v...)
		default:
			r.logger.Warn("無法處理的PDU數據類型",
				zap.String("type", fmt.Sprintf("%T", v)))
		}
	}

	if len(pduData) == 0 {
		return nil
	}

	// 路由到所有處理程序
	for _, handler := range r.handlers {
		if err := handler.HandlePDUData(ctx, pduData); err != nil {
			r.logger.Error("處理PDU數據失敗",
				zap.String("handler", fmt.Sprintf("%T", handler)),
				zap.Error(err))
		}
	}

	return nil
}

// InfluxDBOutputHandler InfluxDB輸出處理程序
type InfluxDBOutputHandler struct {
	logger *zap.Logger
	// TODO: 添加InfluxDB客戶端配置
}

// NewInfluxDBOutputHandler 創建InfluxDB輸出處理程序
func NewInfluxDBOutputHandler(logger *zap.Logger) *InfluxDBOutputHandler {
	return &InfluxDBOutputHandler{
		logger: logger.Named("influxdb-output"),
	}
}

// HandlePDUData 處理PDU數據
func (h *InfluxDBOutputHandler) HandlePDUData(ctx context.Context, data []models.PDUData) error {
	for _, pdu := range data {
		h.logger.Debug("處理PDU數據",
			zap.String("name", pdu.Name),
			zap.Time("timestamp", pdu.Timestamp))

		// TODO: 將PDU數據寫入InfluxDB
	}

	return nil
}

// LoggingOutputHandler 日誌輸出處理程序
type LoggingOutputHandler struct {
	logger *zap.Logger
}

// NewLoggingOutputHandler 創建日誌輸出處理程序
func NewLoggingOutputHandler(logger *zap.Logger) *LoggingOutputHandler {
	return &LoggingOutputHandler{
		logger: logger.Named("logging-output"),
	}
}

// HandlePDUData 處理PDU數據
func (h *LoggingOutputHandler) HandlePDUData(ctx context.Context, data []models.PDUData) error {
	for _, pdu := range data {
		h.logger.Info("收到PDU數據",
			zap.String("name", pdu.Name),
			zap.Time("timestamp", pdu.Timestamp),
			zap.Float64("current", pdu.Current),
			zap.Float64("voltage", pdu.Voltage),
			zap.Float64("power", pdu.Power),
			zap.Int("branches", len(pdu.Branches)),
			zap.Int("phases", len(pdu.Phases)))
	}

	return nil
}
