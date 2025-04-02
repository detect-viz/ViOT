package manager

import (
	"sync"

	"viot/models"

	"go.uber.org/zap"
)

var (
	configManager *ConfigManager
	configLock    sync.RWMutex
)

// InitConfigManager 初始化全局配置管理器
func InitConfigManager(configPath string, logger *zap.Logger) error {
	configLock.Lock()
	defer configLock.Unlock()

	cm := NewConfigManager(configPath, logger)
	if err := cm.Init(); err != nil {
		return err
	}

	configManager = cm
	return nil
}

// GetConfigManager 獲取全局配置管理器
func GetConfigManager() *ConfigManager {
	configLock.RLock()
	defer configLock.RUnlock()

	return configManager
}

// GetProcessorConfig 獲取處理器配置的便捷方法
func GetProcessorConfig() *models.ProcessorConfig {
	cm := GetConfigManager()
	if cm == nil {
		return nil
	}
	return cm.GetProcessorConfig()
}

// GetInfluxDBConfig 獲取InfluxDB配置的便捷方法
func GetInfluxDBConfig() *models.InfluxDBConfig {
	cm := GetConfigManager()
	if cm == nil {
		return nil
	}
	return cm.GetInfluxDBConfig()
}

// GetProcessMonitorConfig 獲取進程監控配置的便捷方法
func GetProcessMonitorConfig() *models.ProcessMonitorConfig {
	cm := GetConfigManager()
	if cm == nil {
		return nil
	}
	return cm.GetProcessMonitorConfig()
}

// ConfigFactory 配置工廠接口
type ConfigFactory interface {
	GetSystemConfig() *models.SystemConfig
	GetProcessMonitorConfig() *models.ProcessMonitorConfig
	GetWebConfig() *models.WebConfig
	GetProcessorConfig() *models.ProcessorConfig
	GetManagerConfig() *models.ManagerConfig
	GetServerConfig() *models.ServerConfig
	GetSchedulerConfig() *models.SchedulerConfig
	GetLoggingConfig() *models.LoggingConfig
}

// configFactoryImpl 配置工廠實現
type configFactoryImpl struct {
	configManager *ConfigManager
}

// NewConfigFactory 創建新的配置工廠
func NewConfigFactory(configManager *ConfigManager) ConfigFactory {
	return &configFactoryImpl{
		configManager: configManager,
	}
}

// GetSystemConfig 獲取系統配置
func (f *configFactoryImpl) GetSystemConfig() *models.SystemConfig {
	return f.configManager.GetSystemConfig()
}

// GetProcessMonitorConfig 獲取進程監控配置
func (f *configFactoryImpl) GetProcessMonitorConfig() *models.ProcessMonitorConfig {
	return f.configManager.GetProcessMonitorConfig()
}

// GetWebConfig 獲取 Web 配置
func (f *configFactoryImpl) GetWebConfig() *models.WebConfig {
	return f.configManager.GetWebConfig()
}

// GetProcessorConfig 獲取處理器配置
func (f *configFactoryImpl) GetProcessorConfig() *models.ProcessorConfig {
	return f.configManager.GetProcessorConfig()
}

// GetManagerConfig 獲取管理器配置
func (f *configFactoryImpl) GetManagerConfig() *models.ManagerConfig {
	return f.configManager.GetManagerConfig()
}

// GetServerConfig 獲取服務器配置
func (f *configFactoryImpl) GetServerConfig() *models.ServerConfig {
	return f.configManager.GetServerConfig()
}

// GetSchedulerConfig 獲取調度器配置
func (f *configFactoryImpl) GetSchedulerConfig() *models.SchedulerConfig {
	return f.configManager.GetSchedulerConfig()
}

// GetLoggingConfig 獲取日誌配置
func (f *configFactoryImpl) GetLoggingConfig() *models.LoggingConfig {
	return f.configManager.GetLoggingConfig()
}
