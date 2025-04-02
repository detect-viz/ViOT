package process

import (
	"time"

	"viot/models"
)

// 定義一些默認常量，用於在配置不存在時提供默認值
const (
	DefaultMonitorInterval = 10 * time.Second
	DefaultStopDelay       = 5 * time.Second
	DefaultShutdownTimeout = 30 * time.Second
	DefaultLogLevel        = "info"
)

// ProcessConfigAdapter 進程配置適配器
type ProcessConfigAdapter struct {
	managerConfig *models.ProcessMonitorConfig
}

// NewProcessConfigAdapter 創建新的進程配置適配器
func NewProcessConfigAdapter(config *models.ProcessMonitorConfig) *ProcessConfigAdapter {
	return &ProcessConfigAdapter{
		managerConfig: config,
	}
}

// GetMonitorInterval 獲取監控間隔
func (a *ProcessConfigAdapter) GetMonitorInterval() time.Duration {
	if a.managerConfig == nil {
		return 5 * time.Second
	}
	return a.managerConfig.MonitorInterval
}

// GetStopDelay 獲取停止延遲
func (a *ProcessConfigAdapter) GetStopDelay() time.Duration {
	if a.managerConfig == nil {
		return 5 * time.Second
	}
	return a.managerConfig.StopDelay
}

// GetSubprocessConfig 獲取子進程配置
func (a *ProcessConfigAdapter) GetSubprocessConfig(name string) (models.SubprocessConfig, bool) {
	if a.managerConfig == nil {
		return models.SubprocessConfig{}, false
	}
	if a.managerConfig.Subprocesses == nil {
		return models.SubprocessConfig{}, false
	}
	if subConfig, exists := a.managerConfig.Subprocesses[name]; exists {
		return subConfig, true
	}
	return models.SubprocessConfig{}, false
}

// GetStartupOrder 獲取啟動順序
func (a *ProcessConfigAdapter) GetStartupOrder() []string {
	if a.managerConfig == nil {
		return nil
	}
	return a.managerConfig.StartupOrder
}

// GetShutdownOrder 獲取關閉順序
func (a *ProcessConfigAdapter) GetShutdownOrder() []string {
	if a.managerConfig == nil {
		return nil
	}
	return a.managerConfig.ShutdownOrder
}

// GetHealthCheck 獲取健康檢查配置
func (a *ProcessConfigAdapter) GetHealthCheck() models.HealthCheckConfig {
	if a.managerConfig == nil {
		return models.HealthCheckConfig{
			Enabled:  true,
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
		}
	}
	return a.managerConfig.HealthCheck
}
