package core

import (
	"context"
	"fmt"
	"sync"

	"viot/logger"
	"viot/models"

	"go.uber.org/zap"
)

// InputManager 管理所有輸入插件
type InputManager struct {
	inputs   map[string]Input
	handlers []DataHandler
	config   *models.CollectorConfig
	logger   logger.Logger
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewInputManager 創建一個新的輸入管理器
func NewInputManager(config *models.CollectorConfig, logger logger.Logger) *InputManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &InputManager{
		inputs:   make(map[string]Input),
		handlers: make([]DataHandler, 0),
		config:   config,
		logger:   logger.Named("input-manager"),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// CreateDefaultInputs 創建默認的輸入插件
func (m *InputManager) CreateDefaultInputs() error {
	// 創建 SNMP 輸入插件
	snmpInput := NewBaseDeviceInput("snmp", m.config, m.logger)
	if err := m.RegisterInput(snmpInput); err != nil {
		return fmt.Errorf("註冊 SNMP 輸入插件失敗: %w", err)
	}

	// 創建 Modbus 輸入插件
	modbusInput := NewBaseDeviceInput("modbus", m.config, m.logger)
	if err := m.RegisterInput(modbusInput); err != nil {
		return fmt.Errorf("註冊 Modbus 輸入插件失敗: %w", err)
	}

	return nil
}

// RegisterInput 註冊一個輸入插件
func (m *InputManager) RegisterInput(input Input) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	name := input.Name()
	if _, exists := m.inputs[name]; exists {
		return fmt.Errorf("輸入插件 %s 已經註冊", name)
	}

	m.inputs[name] = input
	m.logger.Info("註冊輸入插件",
		zap.String("name", name),
		zap.String("type", fmt.Sprintf("%T", input)))

	return nil
}

// GetInput 獲取指定名稱的輸入插件
func (m *InputManager) GetInput(name string) (Input, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	input, exists := m.inputs[name]
	if !exists {
		return nil, fmt.Errorf("輸入插件 %s 不存在", name)
	}

	return input, nil
}

// StartInput 啟動指定的輸入插件
func (m *InputManager) StartInput(ctx context.Context, name string) error {
	input, err := m.GetInput(name)
	if err != nil {
		return err
	}

	m.logger.Info("啟動輸入插件", zap.String("name", name))
	return input.Start(ctx)
}

// StopInput 停止指定的輸入插件
func (m *InputManager) StopInput(name string) error {
	input, err := m.GetInput(name)
	if err != nil {
		return err
	}

	m.logger.Info("停止輸入插件", zap.String("name", name))
	return input.Stop()
}

// GetInputStatus 獲取指定輸入插件的狀態
func (m *InputManager) GetInputStatus(name string) (models.CollectorStatus, error) {
	input, err := m.GetInput(name)
	if err != nil {
		return models.CollectorStatus{}, err
	}

	return input.Status(), nil
}

// ListInputs 列出所有已註冊的輸入插件
func (m *InputManager) ListInputs() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	inputs := make([]string, 0, len(m.inputs))
	for name := range m.inputs {
		inputs = append(inputs, name)
	}

	return inputs
}

// StartAll 啟動所有輸入插件
func (m *InputManager) StartAll(ctx context.Context) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for name, input := range m.inputs {
		if err := input.Start(ctx); err != nil {
			m.logger.Error("啟動輸入插件失敗",
				zap.String("name", name),
				zap.Error(err))
			return fmt.Errorf("啟動輸入插件 %s 失敗: %w", name, err)
		}
		m.logger.Info("輸入插件已啟動", zap.String("name", name))
	}

	return nil
}

// StopAll 停止所有輸入插件
func (m *InputManager) StopAll() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.cancel != nil {
		m.cancel()
	}

	for name, input := range m.inputs {
		if err := input.Stop(); err != nil {
			m.logger.Error("停止輸入插件失敗",
				zap.String("name", name),
				zap.Error(err))
			return fmt.Errorf("停止輸入插件 %s 失敗: %w", name, err)
		}
		m.logger.Info("輸入插件已停止", zap.String("name", name))
	}

	return nil
}

// RegisterHandler 註冊一個數據處理器
func (m *InputManager) RegisterHandler(handler DataHandler) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.handlers = append(m.handlers, handler)
	m.logger.Info("註冊數據處理器",
		zap.String("handler", fmt.Sprintf("%T", handler)))
}

// HandleData 處理收集到的數據
func (m *InputManager) HandleData(ctx context.Context, data []models.RestAPIPoint) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for _, handler := range m.handlers {
		if err := handler.HandleData(ctx, data); err != nil {
			m.logger.Error("處理數據失敗", zap.Error(err))
		}
	}

	return nil
}
