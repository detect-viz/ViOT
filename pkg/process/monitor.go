package process

import (
	"context"
	"sync"
	"time"

	// 引入 manager 包
	"viot/models"

	"go.uber.org/zap"
)

// ProcessMonitor 進程監控器
type ProcessMonitor struct {
	configAdapter *ProcessConfigAdapter
	logger        *zap.Logger
	processes     map[string]*Process
	statuses      map[string]models.ProcessStatus
	statusMutex   sync.RWMutex
	processMutex  sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	notifiers     []Notifier
}

// Process 表示一個子進程
type Process struct {
	Config models.SubprocessConfig
	status models.ProcessStatus
}

// Notifier 通知接口
type Notifier interface {
	Notify(event string, process string, message string) error
}

// NewProcessMonitor 創建新的進程監控器
func NewProcessMonitor(config *models.ProcessMonitorConfig, logger *zap.Logger) *ProcessMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &ProcessMonitor{
		configAdapter: NewProcessConfigAdapter(config),
		logger:        logger,
		processes:     make(map[string]*Process),
		statuses:      make(map[string]models.ProcessStatus),
		ctx:           ctx,
		cancel:        cancel,
		notifiers:     []Notifier{},
	}
}

// 添加初始化方法，從 manager 配置轉換成內部使用的格式
func (m *ProcessMonitor) Init() error {
	// 初始化子進程映射
	m.processes = make(map[string]*Process)

	// 如果配置為空，使用默認值
	if m.configAdapter == nil {
		m.logger.Warn("使用默認配置初始化進程監控器")
		return nil
	}

	// 從 manager 配置獲取子進程配置
	for name := range m.configAdapter.managerConfig.Subprocesses {
		subConfig, exists := m.configAdapter.GetSubprocessConfig(name)
		if !exists {
			continue
		}

		process := &Process{
			Config: subConfig,
			status: models.ProcessStatus{
				Name:    name,
				Running: false,
				Healthy: true,
			},
		}

		m.processes[name] = process
	}

	// 添加日誌通知器
	m.notifiers = append(m.notifiers, &LogNotifier{
		Logger: m.logger,
		Level:  "info",
	})

	m.logger.Info("進程監控器初始化完成",
		zap.Int("子進程數", len(m.processes)),
		zap.String("監控間隔", m.configAdapter.GetMonitorInterval().String()))

	return nil
}

// Start 啟動進程監控器
func (m *ProcessMonitor) Start() error {
	m.logger.Info("進程監控器啟動")

	// 按照配置的啟動順序啟動進程
	startupOrder := m.configAdapter.GetStartupOrder()
	if len(startupOrder) > 0 {
		// 使用配置的啟動順序
		m.logger.Info("按照設定順序啟動進程")
		for _, name := range startupOrder {
			if process, exists := m.processes[name]; exists && process.Config.Enabled && process.Config.AutoStart {
				m.startProcess(name, process)
			}
		}
	} else {
		// 無需依賴檢查，直接啟動所有自動啟動的進程
		m.logger.Info("啟動所有自動啟動進程")
		for name, process := range m.processes {
			if process.Config.Enabled && process.Config.AutoStart {
				m.startProcess(name, process)
			}
		}
	}

	// 啟動監控線程
	go m.monitorLoop()

	return nil
}

// Stop 停止進程監控器
func (m *ProcessMonitor) Stop() error {
	m.logger.Info("進程監控器停止")

	// 取消上下文，停止監控線程
	if m.cancel != nil {
		m.cancel()
	}

	// 按照關閉順序停止進程
	shutdownOrder := m.configAdapter.GetShutdownOrder()
	if len(shutdownOrder) > 0 {
		// 使用配置的關閉順序
		for _, name := range shutdownOrder {
			if process, exists := m.processes[name]; exists && process.Config.Enabled {
				m.stopProcess(name, process)
			}
		}
	} else {
		// 無依賴順序，直接停止所有進程
		for name, process := range m.processes {
			if process.Config.Enabled {
				m.stopProcess(name, process)
			}
		}
	}

	return nil
}

// StartProcess 啟動特定進程
func (m *ProcessMonitor) StartProcess(name string) error {
	m.processMutex.Lock()
	defer m.processMutex.Unlock()

	process, exists := m.processes[name]
	if !exists {
		m.logger.Error("無法找到指定進程", zap.String("process", name))
		return nil
	}

	return m.startProcess(name, process)
}

// StopProcess 停止特定進程
func (m *ProcessMonitor) StopProcess(name string) error {
	m.processMutex.Lock()
	defer m.processMutex.Unlock()

	process, exists := m.processes[name]
	if !exists {
		m.logger.Error("無法找到指定進程", zap.String("process", name))
		return nil
	}

	return m.stopProcess(name, process)
}

// 內部啟動進程方法
func (m *ProcessMonitor) startProcess(name string, process *Process) error {
	if process.status.Running {
		m.logger.Info("進程已在運行中", zap.String("process", name))
		return nil
	}

	// 這裡會實現實際的進程啟動邏輯
	// 使用 os/exec 等包來啟動子進程

	m.logger.Info("啟動進程", zap.String("process", name))

	// 更新進程狀態
	process.status.Running = true
	process.status.StartTime = time.Now()

	// 通知
	for _, notifier := range m.notifiers {
		notifier.Notify("process_started", name, "進程已啟動")
	}

	return nil
}

// 內部停止進程方法
func (m *ProcessMonitor) stopProcess(name string, process *Process) error {
	if !process.status.Running {
		m.logger.Info("進程未運行", zap.String("process", name))
		return nil
	}

	// 這裡會實現實際的進程停止邏輯
	// 使用 cmd.Process.Kill() 等方法停止子進程

	m.logger.Info("停止進程", zap.String("process", name))

	// 更新進程狀態
	process.status.Running = false

	// 通知
	for _, notifier := range m.notifiers {
		notifier.Notify("process_stopped", name, "進程已停止")
	}

	return nil
}

// 監控循環
func (m *ProcessMonitor) monitorLoop() {
	ticker := time.NewTicker(m.configAdapter.GetMonitorInterval())
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkProcesses()
		}
	}
}

// 檢查所有進程
func (m *ProcessMonitor) checkProcesses() {
	m.processMutex.Lock()
	defer m.processMutex.Unlock()

	for name, process := range m.processes {
		if !process.Config.Enabled {
			continue
		}

		// 檢查進程是否還在運行
		if process.status.Running {
			// 更新進程狀態（CPU、內存等）
			m.updateProcessStats(name, process)

			// 檢查健康狀況
			healthCheck := m.configAdapter.GetHealthCheck()
			if healthCheck.Enabled &&
				time.Since(process.status.HealthTime) > process.Config.HealthCheckInterval {
				m.checkProcessHealth(name, process)
			}
		} else if process.Config.RestartOnFailure {
			// 進程應該運行但未運行，自動重啟
			if process.status.RestartCount < process.Config.MaxRestarts {
				m.logger.Info("重啟進程",
					zap.String("process", name),
					zap.Int("restarts", process.status.RestartCount+1),
					zap.Int("max_restarts", process.Config.MaxRestarts),
				)

				m.startProcess(name, process)
				process.status.RestartCount++
			} else {
				m.logger.Error("進程重啟次數超過上限，不再嘗試重啟",
					zap.String("process", name),
					zap.Int("restarts", process.status.RestartCount),
					zap.Int("max_restarts", process.Config.MaxRestarts),
				)

				// 通知
				for _, notifier := range m.notifiers {
					notifier.Notify("restart_limit_exceeded", name, "進程重啟次數超過上限")
				}
			}
		}
	}
}

// 更新進程統計信息
func (m *ProcessMonitor) updateProcessStats(name string, process *Process) {
	// 這裡實現實際獲取進程CPU、內存等統計信息的邏輯
	// 可以通過 os.Process.Pid 獲取進程ID，再通過系統工具(如 ps)獲取詳細信息

	// 更新狀態
	uptime := time.Since(process.status.StartTime)
	process.status.Uptime = uptime.String()

	// TODO: 實現實際統計數據收集邏輯
	// process.status.CPU = ...
	// process.status.Memory = ...
}

// 檢查進程健康狀況
func (m *ProcessMonitor) checkProcessHealth(name string, process *Process) {
	// 這裡實現健康檢查邏輯
	// 可以執行特定命令或檢查進程響應

	// 更新健康檢查時間
	process.status.HealthTime = time.Now()

	// TODO: 實現實際健康檢查邏輯
	// 例如執行 process.Config.HealthCheckCommand

	// 假設檢查成功
	process.status.Healthy = true
}

// GetProcessStatus 獲取特定進程狀態
func (m *ProcessMonitor) GetProcessStatus(name string) (models.ProcessStatus, bool) {
	if status, exists := m.statuses[name]; exists {
		return status, true
	}
	return models.ProcessStatus{}, false
}

// GetAllProcessStatuses 獲取所有進程狀態
func (m *ProcessMonitor) GetAllProcessStatuses() map[string]models.ProcessStatus {
	return m.statuses
}

// LogNotifier 日誌通知器
type LogNotifier struct {
	Logger *zap.Logger
	Level  string
}

// Notify 發送通知
func (n *LogNotifier) Notify(event string, process string, message string) error {
	switch n.Level {
	case "debug":
		n.Logger.Debug(message, zap.String("event", event), zap.String("process", process))
	case "info":
		n.Logger.Info(message, zap.String("event", event), zap.String("process", process))
	case "warn":
		n.Logger.Warn(message, zap.String("event", event), zap.String("process", process))
	case "error":
		n.Logger.Error(message, zap.String("event", event), zap.String("process", process))
	default:
		n.Logger.Info(message, zap.String("event", event), zap.String("process", process))
	}
	return nil
}

// WebhookNotifier Webhook通知器
type WebhookNotifier struct {
	URL     string
	Method  string
	Headers map[string]string
}

// Notify 發送通知
func (n *WebhookNotifier) Notify(event string, process string, message string) error {
	// 這裡實現HTTP請求邏輯
	// 使用 http.Client 發送通知

	return nil
}
