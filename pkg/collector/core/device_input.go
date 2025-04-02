package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"viot/logger"
	"viot/models"
)

// BaseDeviceInput 是所有設備輸入插件的基礎實現
type BaseDeviceInput struct {
	name   string
	config *models.CollectorConfig
	logger logger.Logger
	status models.CollectorStatus
	mutex  sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	points []models.RestAPIPoint
}

// NewBaseDeviceInput 創建一個新的基礎設備輸入插件
func NewBaseDeviceInput(name string, config *models.CollectorConfig, logger logger.Logger) *BaseDeviceInput {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseDeviceInput{
		name:   name,
		config: config,
		logger: logger.Named(fmt.Sprintf("%s-input", name)),
		status: models.CollectorStatus{
			Running:       false,
			LastCollected: time.Now(),
			DataPoints:    0,
			ErrorCount:    0,
			SuccessRate:   0.0,
		},
		ctx:    ctx,
		cancel: cancel,
		points: make([]models.RestAPIPoint, 0),
	}
}

// Name 返回輸入插件的名稱
func (b *BaseDeviceInput) Name() string {
	return b.name
}

// Start 啟動輸入插件
func (b *BaseDeviceInput) Start(ctx context.Context) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.status.Running = true
	b.status.LastCollected = time.Now()
	b.logger.Info("輸入插件已啟動")
	return nil
}

// Stop 停止輸入插件
func (b *BaseDeviceInput) Stop() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.cancel != nil {
		b.cancel()
	}

	b.status.Running = false
	b.status.LastCollected = time.Now()
	b.logger.Info("輸入插件已停止")
	return nil
}

// Status 返回輸入插件的狀態
func (b *BaseDeviceInput) Status() models.CollectorStatus {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.status
}

// Collect 收集數據
func (b *BaseDeviceInput) Collect(ctx context.Context) ([]models.RestAPIPoint, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if !b.status.Running {
		return nil, fmt.Errorf("輸入插件未運行")
	}

	return b.points, nil
}

// Deploy 部署配置
func (b *BaseDeviceInput) Deploy(ctx context.Context) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.logger.Info("部署配置")
	return nil
}

// HandleData 處理收集到的數據
func (b *BaseDeviceInput) HandleData(ctx context.Context, points []models.RestAPIPoint) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.points = points
	b.status.LastCollected = time.Now()
	b.status.DataPoints += int64(len(points))

	// 更新成功率
	totalPoints := b.status.DataPoints + b.status.ErrorCount
	if totalPoints > 0 {
		b.status.SuccessRate = float64(b.status.DataPoints) / float64(totalPoints) * 100
	}

	return nil
}

// UpdateError 更新錯誤信息
func (b *BaseDeviceInput) UpdateError(err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if err != nil {
		b.status.ErrorCount++
	}
	b.status.LastCollected = time.Now()

	// 更新成功率
	totalPoints := b.status.DataPoints + b.status.ErrorCount
	if totalPoints > 0 {
		b.status.SuccessRate = float64(b.status.DataPoints) / float64(totalPoints) * 100
	}
}
