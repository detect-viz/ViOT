package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"viot/models"
)

// BaseInput 提供基本的輸入插件功能
type BaseInput struct {
	name      string
	status    models.CollectorStatus
	startTime time.Time
	stopTime  time.Time
	mutex     sync.RWMutex
}

// NewBaseInput 創建一個新的基本輸入插件
func NewBaseInput(name string) *BaseInput {
	return &BaseInput{
		name:   name,
		status: models.CollectorStatus{Running: false},
	}
}

// Start 啟動輸入插件
func (b *BaseInput) Start(ctx context.Context) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.startTime = time.Now()
	b.status = models.CollectorStatus{
		Running:       true,
		LastCollected: b.startTime,
	}
	return nil
}

// Stop 停止輸入插件
func (b *BaseInput) Stop() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.stopTime = time.Now()
	b.status = models.CollectorStatus{
		Running:       false,
		LastCollected: b.stopTime,
	}
	return nil
}

// Status 返回輸入插件狀態
func (b *BaseInput) Status() models.CollectorStatus {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.status
}

// Name 返回輸入插件名稱
func (b *BaseInput) Name() string {
	return b.name
}

// Collect 收集設備數據（基礎實現）
func (b *BaseInput) Collect(ctx context.Context, device models.Device) ([]models.DeviceData, error) {
	return nil, fmt.Errorf("collect not implemented")
}

// Discover 發現設備（基礎實現）
func (b *BaseInput) Discover(ctx context.Context, ipRange models.IPRange) ([]models.Device, error) {
	return nil, fmt.Errorf("discover not implemented")
}

// Match 匹配設備（基礎實現）
func (b *BaseInput) Match(device models.Device) (bool, error) {
	return false, fmt.Errorf("match not implemented")
}

// Deploy 部署設備配置（基礎實現）
func (b *BaseInput) Deploy(device models.Device) error {
	return fmt.Errorf("deploy not implemented")
}
