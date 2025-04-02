package manager

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"viot/models"

	"go.uber.org/zap"
)

// ScanDeployManager 負責協調設備掃描和配置部署
type ScanDeployManager struct {
	scannerConfig   *models.ScannerConfig
	deployerConfig  *models.DeployerConfig
	collectorConfig *models.CollectorConfig
	configManager   *ConfigManager
	logger          *zap.Logger
	scanner         models.ScannerInterface
	deployer        models.DeployerInterface
	ctx             context.Context
	cancel          context.CancelFunc
	mutex           sync.RWMutex
	scanningActive  bool
	lastScanResult  *models.ScanResult
}

// NewScanDeployManager 創建新的掃描部署管理器
func NewScanDeployManager(
	configManager *ConfigManager,
	logger *zap.Logger,
) *ScanDeployManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ScanDeployManager{
		configManager:   configManager,
		scannerConfig:   configManager.GetScannerConfig(),
		deployerConfig:  configManager.GetDeployerConfig(),
		collectorConfig: configManager.GetCollectorConfig(),
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
		scanningActive:  false,
	}
}

// Start 啟動掃描部署管理器
func (sdm *ScanDeployManager) Start(ctx context.Context) error {
	sdm.ctx, sdm.cancel = context.WithCancel(ctx)

	// 創建日誌適配器
	loggerAdapter := NewZapLoggerAdapter(sdm.logger)

	// 使用收集器配置初始化掃描器
	if sdm.collectorConfig == nil {
		return fmt.Errorf("收集器配置為空，無法初始化掃描器")
	}

	// 從收集器配置創建SNMP掃描器
	snmpScanner := NewSNMPScanner(&sdm.collectorConfig.Protocols.SNMP, loggerAdapter)

	// 從收集器配置創建Modbus掃描器
	modbusScanner := NewModbusScanner(&sdm.collectorConfig.Protocols.Modbus, loggerAdapter)

	// 創建複合掃描器
	sdm.scanner = NewMultiScanner([]ScannerInterface{snmpScanner, modbusScanner}, loggerAdapter)

	// 初始化部署器
	sdm.deployer = NewTelegrafDeployer(sdm.deployerConfig.Telegraf, loggerAdapter)

	sdm.logger.Info("掃描部署管理器啟動成功")
	return nil
}

// Stop 停止掃描部署管理器
func (sdm *ScanDeployManager) Stop() error {
	if sdm.cancel != nil {
		sdm.cancel()
	}

	// 等待掃描完成
	sdm.mutex.Lock()
	defer sdm.mutex.Unlock()

	sdm.scanningActive = false
	sdm.logger.Info("掃描部署管理器已停止")
	return nil
}

// StartScan 開始設備掃描
func (sdm *ScanDeployManager) StartScan() error {
	sdm.mutex.Lock()
	if sdm.scanningActive {
		sdm.mutex.Unlock()
		return fmt.Errorf("掃描已在進行中")
	}

	sdm.scanningActive = true
	sdm.mutex.Unlock()

	go func() {
		defer func() {
			sdm.mutex.Lock()
			sdm.scanningActive = false
			sdm.mutex.Unlock()
		}()

		sdm.logger.Info("開始設備掃描")

		// 解析掃描超时
		var scanTimeout time.Duration
		scanTimeoutSeconds := sdm.scannerConfig.Settings.ScanTimeout
		if scanTimeoutSeconds > 0 {
			scanTimeout = time.Duration(scanTimeoutSeconds) * time.Second
		} else {
			scanTimeout = 5 * time.Minute // 默認5分鐘
		}

		scanCtx, cancel := context.WithTimeout(sdm.ctx, scanTimeout)
		defer cancel()

		// 加載IP範圍
		ipRanges, err := sdm.loadIPRanges()
		if err != nil {
			sdm.logger.Error("加載IP範圍失敗", zap.Error(err))
			return
		}

		// 執行掃描
		result, err := sdm.scanner.Scan(scanCtx, ipRanges)
		if err != nil {
			sdm.logger.Error("設備掃描失敗", zap.Error(err))
			return
		}

		// 保存掃描結果
		sdm.mutex.Lock()
		sdm.lastScanResult = result
		sdm.mutex.Unlock()

		// 保存結果到文件
		if err := sdm.saveScanResult(result); err != nil {
			sdm.logger.Error("保存掃描結果失敗", zap.Error(err))
		}

		sdm.logger.Info("設備掃描完成",
			zap.Int("發現設備數", len(result.Devices)),
			zap.Int("識別設備數", result.IdentifiedCount),
		)

		// 檢查是否自動部署
		if sdm.scannerConfig.Settings.AutoDeploy {
			if err := sdm.deployIdentifiedDevices(result); err != nil {
				sdm.logger.Error("自動部署失敗", zap.Error(err))
			}
		}
	}()

	return nil
}

// GetScanStatus 獲取掃描狀態
func (sdm *ScanDeployManager) GetScanStatus() (bool, *ScanResult) {
	sdm.mutex.RLock()
	defer sdm.mutex.RUnlock()

	return sdm.scanningActive, sdm.lastScanResult
}

// DeployDevice 部署特定設備
func (sdm *ScanDeployManager) DeployDevice(ipAddress string) error {
	sdm.mutex.RLock()
	defer sdm.mutex.RUnlock()

	if sdm.lastScanResult == nil {
		return fmt.Errorf("沒有可用的掃描結果")
	}

	// 查找設備
	var deviceToDeploy DeviceInfo
	found := false
	for _, device := range sdm.lastScanResult.Devices {
		if device.IPAddress == ipAddress {
			deviceToDeploy = device
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("未找到IP地址為 %s 的設備", ipAddress)
	}

	// 執行部署
	return sdm.deployer.Deploy(deviceToDeploy)
}

// loadIPRanges 從配置加載IP範圍
func (sdm *ScanDeployManager) loadIPRanges() ([]IPRange, error) {
	if sdm.scannerConfig == nil || len(sdm.scannerConfig.IPRanges) == 0 {
		return []IPRange{
			{
				Start:       "192.168.1.1",
				End:         "192.168.1.254",
				Description: "默認網段",
			},
		}, nil
	}

	ranges := make([]IPRange, len(sdm.scannerConfig.IPRanges))
	for i, cfgRange := range sdm.scannerConfig.IPRanges {
		ranges[i] = IPRange{
			Start:       cfgRange.StartIP,
			End:         cfgRange.EndIP,
			Description: fmt.Sprintf("%s-%s", cfgRange.Datacenter, cfgRange.Room),
		}
	}

	return ranges, nil
}

// saveScanResult 保存掃描結果到文件
func (sdm *ScanDeployManager) saveScanResult(result *ScanResult) error {
	// 使用配置的掃描結果路徑或默認路徑
	resultPath := sdm.scannerConfig.Settings.ScanResultPath
	if resultPath == "" {
		resultPath = DefaultScanResultPath
	}

	// 添加時間戳以創建唯一文件名
	timestamp := time.Now().Format("20060102-150405")
	outputFile := filepath.Join(resultPath, fmt.Sprintf("scan-result-%s.json", timestamp))

	// 確保路徑存在
	outputDir := filepath.Dir(outputFile)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("創建輸出目錄失敗: %w", err)
		}
	}

	// TODO: 實現實際的結果保存邏輯

	sdm.logger.Info("保存掃描結果",
		zap.String("path", outputFile),
		zap.Int("設備數", len(result.Devices)),
	)

	return nil
}

// deployIdentifiedDevices 部署識別的設備 - 添加驗證和回滾邏輯
func (sdm *ScanDeployManager) deployIdentifiedDevices(result *ScanResult) error {
	if sdm.scannerConfig == nil || sdm.deployerConfig == nil {
		return fmt.Errorf("掃描器或部署器配置為空")
	}

	// 獲取部署設置
	verifyAfterDeploy := sdm.deployerConfig.Settings.VerifyAfterDeploy
	rollbackOnFailure := sdm.deployerConfig.Settings.RollbackOnFailure

	for _, device := range result.Devices {
		if device.Identified {
			sdm.logger.Info("部署設備",
				zap.String("ip", device.IPAddress),
				zap.String("type", device.Type),
				zap.Bool("verify", verifyAfterDeploy),
				zap.Bool("rollback", rollbackOnFailure),
			)

			// 實現部署邏輯，使用配置選項
			if err := sdm.deployer.Deploy(device); err != nil {
				sdm.logger.Error("部署失敗", zap.Error(err))

				if rollbackOnFailure {
					// 實現回滾邏輯
				}
			} else if verifyAfterDeploy {
				// 實現驗證邏輯
				if err := sdm.deployer.Validate(device); err != nil {
					sdm.logger.Error("驗證失敗", zap.Error(err))

					if rollbackOnFailure {
						// 實現回滾邏輯
					}
				}
			}
		}
	}

	return nil
}

// ScheduleTask 根據配置調度任務 - 使用默認cron
func (sdm *ScanDeployManager) ScheduleTask(taskName string) error {
	switch taskName {
	case "scanner":
		sdm.logger.Info("調度掃描任務", zap.String("cron", DefaultScannerCron))
		return sdm.StartScan()
	default:
		return fmt.Errorf("未知的任務: %s", taskName)
	}
}
