package interfaces

import (
	"context"
	"viot/models"
	"viot/models/processor"
)

// Processor 處理器接口
type Processor interface {
	Name() string
	Start(ctx context.Context) error
	Stop() error
	Status() processor.ProcessorStatus
	Process(data []models.DeviceData) error
	HandlePDUData(ctx context.Context, data []models.PDUData) error
	SetOutputRouter(router OutputRouter)
}

// OutputRouter 輸出路由器接口
type OutputRouter interface {
	RegisterHandler(handler OutputHandler) error
	RoutePDUData(ctx context.Context, data ...interface{}) error
}

// OutputHandler 輸出處理器接口
type OutputHandler interface {
	HandlePDUData(ctx context.Context, data []models.PDUData) error
}

// TelegrafProcessor 定義了 Telegraf 數據處理器的接口
type TelegrafProcessor interface {
	Processor
	// HandleTelegrafData 處理 Telegraf 數據
	HandleTelegrafData(ctx context.Context, data []models.RestAPIPoint) error
}
