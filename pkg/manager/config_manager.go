package manager

import (
	"fmt"

	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"viot/models"
)

// ConfigManager 負責加載和管理系統所有配置
type ConfigManager struct {
	configDir       string
	logger          *zap.Logger
	config          *models.SystemConfig
	webConfig       *models.WebConfig
	processorConfig *models.ProcessorConfig
	deployerConfig  *models.DeployerConfig
	collectorConfig *models.CollectorConfig
	scannerConfig   *models.ScannerConfig
	configLock      sync.RWMutex // 添加讀寫鎖保護配置訪問
}

// NewConfigManager 創建新的配置管理器
func NewConfigManager(configDir string, logger *zap.Logger) *ConfigManager {
	return &ConfigManager{
		configDir: configDir,
		logger:    logger,
	}
}

// Init 初始化配置管理器，加載所有配置文件
func (cm *ConfigManager) Init() error {
	cm.configLock.Lock()
	defer cm.configLock.Unlock()

	// 首先加載主配置
	if err := cm.loadSystemConfig(); err != nil {
		return fmt.Errorf("加載系統配置失敗: %w", err)
	}

	// 檢查config_paths是否存在
	if (models.ConfigPaths{}) == cm.config.Manager.ConfigPaths {
		return fmt.Errorf("主配置中缺少config_paths設置")
	}

	// 加載各模塊配置
	if err := cm.loadModuleConfigs(); err != nil {
		return fmt.Errorf("加載模塊配置失敗: %w", err)
	}

	// 驗證所有配置的有效性
	if err := cm.validateAllConfigs(); err != nil {
		return fmt.Errorf("配置驗證失敗: %w", err)
	}

	cm.logger.Info("所有配置加載成功")
	return nil
}

// loadSystemConfig 加載系統配置文件
func (cm *ConfigManager) loadSystemConfig() error {
	configPath := filepath.Join(cm.configDir, "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("讀取配置文件失敗: %w", err)
	}

	cm.config = &models.SystemConfig{}
	if err := yaml.Unmarshal(data, cm.config); err != nil {
		return fmt.Errorf("解析系統配置文件失敗: %w", err)
	}

	cm.logger.Info("系統配置加載成功", zap.String("path", configPath))
	return nil
}

// loadModuleConfigs 加載所有模塊配置
func (cm *ConfigManager) loadModuleConfigs() error {
	// 加載Web配置
	if path := cm.config.Manager.ConfigPaths.Web; path != "" {
		fullPath := cm.resolvePath(path)
		if err := cm.loadWebConfig(fullPath); err != nil {
			return fmt.Errorf("加載Web配置失敗: %w", err)
		}
	}

	// 加載Processor配置
	if path := cm.config.Manager.ConfigPaths.Processor; path != "" {
		fullPath := cm.resolvePath(path)
		if err := cm.loadProcessorConfig(fullPath); err != nil {
			return fmt.Errorf("加載Processor配置失敗: %w", err)
		}
	}

	// 加載Deployer配置
	if path := cm.config.Manager.ConfigPaths.Deployer; path != "" {
		fullPath := cm.resolvePath(path)
		if err := cm.loadDeployerConfig(fullPath); err != nil {
			return fmt.Errorf("加載Deployer配置失敗: %w", err)
		}
	}

	// 加載Collector配置
	if path := cm.config.Manager.ConfigPaths.Collector; path != "" {
		fullPath := cm.resolvePath(path)
		if err := cm.loadCollectorConfig(fullPath); err != nil {
			return fmt.Errorf("加載Collector配置失敗: %w", err)
		}
	}

	// 加載Scanner配置
	if path := cm.config.Manager.ConfigPaths.Scanner; path != "" {
		fullPath := cm.resolvePath(path)
		if err := cm.loadScannerConfig(fullPath); err != nil {
			return fmt.Errorf("加載Scanner配置失敗: %w", err)
		}
	}

	return nil
}

// resolvePath 解析路徑
func (cm *ConfigManager) resolvePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(cm.configDir, path)
}

// validateAllConfigs 驗證所有配置的有效性
func (cm *ConfigManager) validateAllConfigs() error {
	// 驗證系統配置
	if cm.config == nil {
		return fmt.Errorf("系統配置未加載")
	}

	// 驗證必要的配置路徑
	if err := cm.validatePaths(); err != nil {
		return err
	}

	return nil
}

// validatePaths 驗證配置中的路徑是否存在或可創建
func (cm *ConfigManager) validatePaths() error {
	paths := []string{
		cm.config.Manager.ServiceDataPath,
		filepath.Dir(cm.config.Manager.Logging.OutputPath),
	}

	if cm.config.Manager.InfluxDB.BackupPolicy.Path != "" {
		paths = append(paths, cm.config.Manager.InfluxDB.BackupPolicy.Path)
	}

	if cm.config.Manager.InfluxDB.DataPolicy.FailWritePath != "" {
		paths = append(paths, cm.config.Manager.InfluxDB.DataPolicy.FailWritePath)
	}

	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			cm.logger.Info("創建目錄", zap.String("path", path))
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("創建目錄失敗 %s: %w", path, err)
			}
		}
	}

	return nil
}

// GetSystemConfig 獲取系統配置
func (cm *ConfigManager) GetSystemConfig() *models.SystemConfig {
	return cm.config
}

// GetWebConfig 獲取Web配置
func (cm *ConfigManager) GetWebConfig() *models.WebConfig {
	cm.configLock.RLock()
	defer cm.configLock.RUnlock()
	return cm.webConfig
}

// GetProcessorConfig 獲取Processor配置
func (cm *ConfigManager) GetProcessorConfig() *models.ProcessorConfig {
	cm.configLock.RLock()
	defer cm.configLock.RUnlock()
	return cm.processorConfig
}

// GetDeployerConfig 獲取Deployer配置
func (cm *ConfigManager) GetDeployerConfig() *models.DeployerConfig {
	cm.configLock.RLock()
	defer cm.configLock.RUnlock()
	return cm.deployerConfig
}

// GetCollectorConfig 獲取Collector配置
func (cm *ConfigManager) GetCollectorConfig() *models.CollectorConfig {
	cm.configLock.RLock()
	defer cm.configLock.RUnlock()
	return cm.collectorConfig
}

// GetScannerConfig 獲取Scanner配置
func (cm *ConfigManager) GetScannerConfig() *models.ScannerConfig {
	cm.configLock.RLock()
	defer cm.configLock.RUnlock()
	return cm.scannerConfig
}

// GetManagerConfig 獲取管理器配置
func (cm *ConfigManager) GetManagerConfig() *models.ManagerConfig {
	return &cm.config.Manager
}

// GetServerConfig 獲取服務器配置
func (cm *ConfigManager) GetServerConfig() *models.ServerConfig {
	return &cm.config.Server
}

// GetProcessMonitorConfig 獲取進程監控配置
func (cm *ConfigManager) GetProcessMonitorConfig() *models.ProcessMonitorConfig {
	if cm.config == nil {
		return nil
	}
	return &cm.config.ProcessMonitor
}

// GetUIConfig 獲取UI配置
func (cm *ConfigManager) GetUIConfig() *models.UIConfig {
	return &cm.config.UI
}

// GetSchedulerConfig 獲取任務調度配置
func (cm *ConfigManager) GetSchedulerConfig() *models.SchedulerConfig {
	return &cm.config.Scheduler
}

// GetLoggingConfig 獲取日誌配置
func (cm *ConfigManager) GetLoggingConfig() *models.LoggingConfig {
	return &cm.config.Manager.Logging
}

// GetInfluxDBConfig 獲取InfluxDB配置
func (cm *ConfigManager) GetInfluxDBConfig() *models.InfluxDBConfig {
	return &cm.config.Manager.InfluxDB
}

// ParseDuration 解析字符串為時間間隔
func ParseDuration(durationStr string) (time.Duration, error) {
	return time.ParseDuration(durationStr)
}

// ReloadConfigs 重新加載所有配置
func (cm *ConfigManager) ReloadConfigs() error {
	cm.logger.Info("開始重載所有配置")
	return cm.Init()
}

// loadWebConfig 加載Web配置文件
func (cm *ConfigManager) loadWebConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("讀取Web配置文件失敗: %w", err)
	}

	// 因為web.yaml有一個頂層"web"鍵
	var webYaml struct {
		Web *models.WebConfig `yaml:"web"`
	}

	if err := yaml.Unmarshal(data, &webYaml); err != nil {
		return fmt.Errorf("解析Web配置文件失敗: %w", err)
	}

	if webYaml.Web == nil {
		return fmt.Errorf("Web配置格式錯誤: 缺少web頂層鍵")
	}

	// 處理SyncInterval字符串到時間間隔的轉換
	if webYaml.Web.Status.SyncInterval == 0 && webYaml.Web.Status.PDUStatusPath != "" {
		// 創建一個臨時的結構來獲取原始字符串值
		var tempConfig struct {
			Web struct {
				Status struct {
					SyncInterval string `yaml:"sync_interval"`
				} `yaml:"status"`
			} `yaml:"web"`
		}

		if err := yaml.Unmarshal(data, &tempConfig); err == nil && tempConfig.Web.Status.SyncInterval != "" {
			duration, err := time.ParseDuration(tempConfig.Web.Status.SyncInterval)
			if err != nil {
				cm.logger.Warn("無法解析同步間隔，使用默認值5分鐘", zap.Error(err), zap.String("value", tempConfig.Web.Status.SyncInterval))
				duration = 5 * time.Minute
			}
			webYaml.Web.Status.SyncInterval = duration
		} else {
			// 設置默認值為5分鐘
			webYaml.Web.Status.SyncInterval = 5 * time.Minute
		}
	}

	cm.webConfig = webYaml.Web
	cm.logger.Info("Web配置加載成功", zap.String("path", path))
	return nil
}

// loadProcessorConfig 加載處理器配置文件
func (cm *ConfigManager) loadProcessorConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("讀取處理器配置文件失敗: %w", err)
	}

	config := &models.ProcessorConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("解析處理器配置文件失敗: %w", err)
	}

	cm.processorConfig = config
	cm.logger.Info("處理器配置加載成功", zap.String("path", path))
	return nil
}

// loadDeployerConfig 加載部署器配置文件
func (cm *ConfigManager) loadDeployerConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("讀取部署器配置文件失敗: %w", err)
	}

	config := &models.DeployerConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("解析部署器配置文件失敗: %w", err)
	}

	cm.deployerConfig = config
	cm.logger.Info("部署器配置加載成功", zap.String("path", path))
	return nil
}

// loadCollectorConfig 加載收集器配置文件
func (cm *ConfigManager) loadCollectorConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("讀取收集器配置文件失敗: %w", err)
	}

	config := &models.CollectorConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("解析收集器配置文件失敗: %w", err)
	}

	cm.collectorConfig = config
	cm.logger.Info("收集器配置加載成功", zap.String("path", path))
	return nil
}

// loadScannerConfig 加載掃描器配置文件
func (cm *ConfigManager) loadScannerConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("讀取掃描器配置文件失敗: %w", err)
	}

	config := &models.ScannerConfig{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("解析掃描器配置文件失敗: %w", err)
	}

	cm.scannerConfig = config
	cm.logger.Info("掃描器配置加載成功", zap.String("path", path))
	return nil
}

// GetProcessorConfigAsModel 獲取處理器配置作為模型
func (cm *ConfigManager) GetProcessorConfigAsModel() *models.ProcessorConfig {
	cm.configLock.RLock()
	defer cm.configLock.RUnlock()

	// 將本地配置轉換為模型配置
	if cm.processorConfig == nil {
		return nil
	}

	// 構建模型配置
	modelConfig := &models.ProcessorConfig{
		Enabled:      cm.processorConfig.Enabled,
		Name:         cm.processorConfig.Name,
		Type:         cm.processorConfig.Type,
		InputFormat:  cm.processorConfig.InputFormat,
		OutputFormat: cm.processorConfig.OutputFormat,
	}

	// 這裡可以添加更多字段映射
	// 根據需要從內部配置結構映射到模型配置

	return modelConfig
}

// SaveConfig 保存配置
func (cm *ConfigManager) SaveConfig() error {
	if cm.config == nil {
		return fmt.Errorf("配置未加載")
	}

	data, err := yaml.Marshal(cm.config)
	if err != nil {
		return fmt.Errorf("序列化配置失敗: %w", err)
	}

	configPath := filepath.Join(cm.configDir, "config.yaml")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("保存配置文件失敗: %w", err)
	}

	return nil
}
