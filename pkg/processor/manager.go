package processor

import (
	"fmt"
	"sync"
	"time"

	"viot/interfaces"
	"viot/logger"
	"viot/models"

	"go.uber.org/zap"
)

// ProcessorManager 處理器管理器，負責管理所有處理器
type ProcessorManager struct {
	processors     map[string]*ProcessorInstance
	processorMutex sync.RWMutex
	logger         logger.Logger
}

// NewProcessorManager 創建處理器管理器
func NewProcessorManager(logger logger.Logger) *ProcessorManager {
	return &ProcessorManager{
		processors: make(map[string]*models.ProcessorInstance),
		logger:     logger,
	}
}

// RegisterProcessor 註冊處理器
func (m *ProcessorManager) RegisterProcessor(name string, processor interfaces.Processor, config models.ProcessorConfig) error {
	m.processorMutex.Lock()
	defer m.processorMutex.Unlock()

	if _, exists := m.processors[name]; exists {
		return fmt.Errorf("處理器 %s 已經註冊", name)
	}

	m.processors[name] = &models.ProcessorInstance{
		Processor:   processor,
		Config:      config,
		Status:      models.ProcessorStatus{},
		StartTime:   time.Now(),
		LastUpdated: time.Now(),
	}

	m.logger.Info("處理器註冊成功", zap.String("name", name))
	return nil
}

// GetProcessor 獲取處理器
func (m *ProcessorManager) GetProcessor(name string) (interfaces.Processor, error) {
	m.processorMutex.RLock()
	defer m.processorMutex.RUnlock()

	instance, exists := m.processors[name]
	if !exists {
		return nil, fmt.Errorf("處理器 %s 不存在", name)
	}

	return instance.Processor, nil
}

// GetAllProcessors 獲取所有處理器
func (m *ProcessorManager) GetAllProcessors() map[string]interfaces.Processor {
	m.processorMutex.RLock()
	defer m.processorMutex.RUnlock()

	result := make(map[string]interfaces.Processor)
	for name, instance := range m.processors {
		result[name] = instance.Processor
	}

	return result
}

// GetProcessorStatus 獲取處理器狀態
func (m *ProcessorManager) GetProcessorStatus(name string) (models.ProcessorStatus, error) {
	m.processorMutex.RLock()
	defer m.processorMutex.RUnlock()

	instance, exists := m.processors[name]
	if !exists {
		return models.ProcessorStatus{}, fmt.Errorf("處理器 %s 不存在", name)
	}

	// 更新處理器實例的狀態
	if pduProc, ok := interfaces.Processor.(*PDUProcessor); ok {
		instance.Status = pduProc.GetStatus()
	}
	instance.LastUpdated = time.Now()

	return instance.Status, nil
}

// GetAllProcessorStatus 獲取所有處理器狀態
func (m *ProcessorManager) GetAllProcessorStatus() map[string]models.ProcessorStatus {
	m.processorMutex.RLock()
	defer m.processorMutex.RUnlock()

	result := make(map[string]models.ProcessorStatus)
	for name, instance := range m.processors {
		// 更新處理器實例的狀態
		if pduProc, ok := interfaces.Processor.(*PDUProcessor); ok {
			instance.Status = pduProc.GetStatus()
		}
		instance.LastUpdated = time.Now()

		result[name] = instance.Status
	}

	return result
}

// RemoveProcessor 移除處理器
func (m *ProcessorManager) RemoveProcessor(name string) error {
	m.processorMutex.Lock()
	defer m.processorMutex.Unlock()

	if _, exists := m.processors[name]; !exists {
		return fmt.Errorf("處理器 %s 不存在", name)
	}

	delete(m.processors, name)
	m.logger.Info("處理器移除成功", zap.String("name", name))
	return nil
}

// CreateDefaultProcessors 創建默認處理器
func (m *ProcessorManager) CreateDefaultProcessors() error {
	// 創建PDU處理器
	pduConfig := models.ProcessorConfig{
		Enabled:   true,
		Type:      "pdu",
		Interval:  5,
		BatchSize: 100,
	}

	pduProcessor := NewPDUProcessor("pdu", pduConfig, m.logger)
	if err := m.RegisterProcessor("pdu", pduProcessor, pduConfig); err != nil {
		m.logger.Error("註冊PDU處理器失敗", zap.Error(err))
		return err
	}

	return nil
}

// SetupOutputRouter 設置所有處理器的輸出路由器
func (m *ProcessorManager) SetupOutputRouter(router models.OutputRouter) error {
	for name, instance := range m.processors {
		if pduProc, ok := interfaces.Processor.(*PDUProcessor); ok {
			pduProc.SetOutputRouter(router)
			m.logger.Info("設置PDU處理器輸出路由器成功", zap.String("processor", name))
		} else {
			m.logger.Warn("處理器不支持設置輸出路由器", zap.String("processor", name))
		}
	}

	return nil
}
