package processor

import (
	"context"
	"sync"
	"time"

	"viot/interfaces"
	"viot/logger"
	"viot/models"
)

// BaseProcessor 處理器基類
type BaseProcessor struct {
	name            string
	config          models.ProcessorConfig
	logger          logger.Logger
	outputRouter    models.OutputRouter
	status          models.ProcessorStatus
	ctx             context.Context
	cancel          context.CancelFunc
	mutex           sync.RWMutex
	running         bool
	lastProcessTime time.Time
	processedCount  int64
	errorCount      int64
	lastError       string
}

// ProcessorInstance 處理器實例
type ProcessorInstance struct {
	Processor   interfaces.Processor   `json:"-"`
	Config      models.ProcessorConfig `json:"config"`
	Status      models.ProcessorStatus `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	LastUpdated time.Time              `json:"last_updated"`
}

// NewBaseProcessor 創建基礎處理器
func NewBaseProcessor(name string, config models.ProcessorConfig, logger logger.Logger) *BaseProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseProcessor{
		name:   name,
		config: config,
		logger: logger.Named(name),
		status: models.ProcessorStatus{
			Running:        false,
			StartTime:      time.Time{},
			LastProcessed:  time.Time{},
			ProcessedCount: 0,
			ErrorCount:     0,
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Name 返回處理器名稱
func (p *BaseProcessor) Name() string {
	return p.name
}

// SetOutputRouter 設置輸出路由器
func (p *BaseProcessor) SetOutputRouter(router models.OutputRouter) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.outputRouter = router
}

// Start 啟動處理器
func (p *BaseProcessor) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.status.Running {
		return nil
	}

	p.status.Running = true
	p.status.StartTime = time.Now()
	return nil
}

// Stop 停止處理器
func (p *BaseProcessor) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.status.Running {
		return nil
	}

	if p.cancel != nil {
		p.cancel()
	}

	p.status.Running = false
	return nil
}

// Status 返回處理器狀態
func (p *BaseProcessor) Status() models.ProcessorStatus {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.status
}

// Process 基礎處理方法，應被子類覆蓋
func (p *BaseProcessor) Process(data []models.DeviceData) error {
	// 基礎實現，應由子類覆蓋
	return nil
}

// IncrementProcessedCount 增加處理數據計數
func (p *BaseProcessor) IncrementProcessedCount(count int64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.status.ProcessedCount += count
	p.status.LastProcessed = time.Now()
}

// IncrementErrorCount 增加錯誤計數
func (p *BaseProcessor) IncrementErrorCount(count int64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.status.ErrorCount += count
}

// GetConfig 獲取處理器配置
func (p *BaseProcessor) GetConfig() models.ProcessorConfig {
	return p.config
}

// GetLogger 獲取日誌記錄器
func (p *BaseProcessor) GetLogger() logger.Logger {
	return p.logger
}

// IsRunning 檢查處理器是否正在運行
func (p *BaseProcessor) IsRunning() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.status.Running
}

// GetLastProcessTime 獲取最後處理時間
func (p *BaseProcessor) GetLastProcessTime() time.Time {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.status.LastProcessed
}

// GetProcessedCount 獲取處理計數
func (p *BaseProcessor) GetProcessedCount() int64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.status.ProcessedCount
}

// GetErrorCount 獲取錯誤計數
func (p *BaseProcessor) GetErrorCount() int64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.status.ErrorCount
}

// GetLastError 獲取最後錯誤
func (p *BaseProcessor) GetLastError() string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.status.LastError
}

// SetRunning 設置運行狀態
func (p *BaseProcessor) SetRunning(running bool) {
	p.running = running
}

// SetLastProcessTime 設置最後處理時間
func (p *BaseProcessor) SetLastProcessTime(t time.Time) {
	p.lastProcessTime = t
}
