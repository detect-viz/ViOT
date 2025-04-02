package manager

import (
	"context"
	"fmt"

	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"viot/models"

	"go.uber.org/zap"
)

// DataManager 負責數據的流轉、備份與恢復
type DataManager struct {
	config          *models.ManagerConfig
	influxDBConfig  *models.InfluxDBConfig
	processorConfig *models.ProcessorConfig
	logger          *zap.Logger
	configManager   *ConfigManager
	processors      map[string]models.ProcessorInterface
	ctx             context.Context
	cancel          context.CancelFunc
	recoveryQueue   chan []byte
	mutex           sync.RWMutex
}

// NewDataManager 創建新的數據管理器
func NewDataManager(
	configManager *ConfigManager,
	logger *zap.Logger,
) *DataManager {
	ctx, cancel := context.WithCancel(context.Background())

	// 將 manager 配置轉換為通用模型配置
	processorConfig := &models.ProcessorConfig{
		// 轉換配置...
	}

	return &DataManager{
		configManager:   configManager,
		config:          configManager.GetManagerConfig(),
		influxDBConfig:  configManager.GetInfluxDBConfig(),
		processorConfig: processorConfig,
		logger:          logger,
		processors:      make(map[string]models.ProcessorInterface),
		ctx:             ctx,
		cancel:          cancel,
		recoveryQueue:   make(chan []byte, 1000),
	}
}

// RegisterProcessor 註冊數據處理器
func (dm *DataManager) RegisterProcessor(name string, proc models.ProcessorInterface) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	dm.processors[name] = proc
	dm.logger.Info("註冊處理器", zap.String("name", name))
}

// Start 啟動數據管理器
func (dm *DataManager) Start(ctx context.Context) error {
	dm.ctx, dm.cancel = context.WithCancel(ctx)

	// 檢查並創建必要的目錄
	if err := dm.ensureDirectories(); err != nil {
		return fmt.Errorf("確保目錄存在失敗: %w", err)
	}

	// 啟動恢復處理
	go dm.processRecoveryQueue()

	// 啟動自動恢復
	go dm.startRecovery()

	dm.logger.Info("數據管理器啟動成功")
	return nil
}

// Stop 停止數據管理器
func (dm *DataManager) Stop() error {
	if dm.cancel != nil {
		dm.cancel()
	}

	dm.logger.Info("數據管理器已停止")
	return nil
}

// ensureDirectories 確保所有必要的目錄存在
func (dm *DataManager) ensureDirectories() error {
	dirs := []string{
		dm.config.ServiceDataPath,
		dm.influxDBConfig.DataPolicy.FailWritePath,
	}

	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			dm.logger.Info("創建目錄", zap.String("path", dir))
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("創建目錄失敗 %s: %w", dir, err)
			}
		}
	}

	return nil
}

// HandleDataFailure 處理數據寫入失敗的情況
func (dm *DataManager) HandleDataFailure(data []byte, dataType string) error {
	// 檢查配置是否可用
	if dm.influxDBConfig == nil || dm.influxDBConfig.DataPolicy.FailWritePath == "" {
		return fmt.Errorf("未配置數據備份路徑")
	}

	// 檢查磁盤空間
	if !dm.checkDiskSpace() {
		dm.logger.Warn("磁盤空間不足，無法備份數據")
		return fmt.Errorf("磁盤空間不足")
	}

	// 確保目錄存在
	backupDir := dm.influxDBConfig.DataPolicy.FailWritePath
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("創建備份目錄失敗: %w", err)
	}

	// 創建備份文件
	timestamp := time.Now().Format("20060102-150405.000")
	filename := filepath.Join(backupDir, fmt.Sprintf("%s-%s.json", dataType, timestamp))

	// 寫入數據
	if err := os.WriteFile(filename, data, 0644); err != nil {
		dm.logger.Error("寫入備份文件失敗",
			zap.String("filename", filename),
			zap.Error(err),
		)
		return fmt.Errorf("寫入備份文件失敗: %w", err)
	}

	dm.logger.Info("數據備份成功",
		zap.String("filename", filename),
		zap.String("type", dataType),
	)

	return nil
}

// checkDiskSpace 檢查磁盤空間是否充足
func (dm *DataManager) checkDiskSpace() bool {
	if dm.influxDBConfig.DataPolicy.LowSpaceThreshold == 0 {
		// 未設置閾值，不檢查
		return true
	}

	// 實際應該使用系統API獲取可用磁盤空間
	// 這裡僅作為示例
	// 實際實現可使用 syscall.Statfs 或類似方法

	// TODO: 實現實際的磁盤空間檢查

	return true
}

// startRecovery 開始恢復處理
func (dm *DataManager) startRecovery() {
	dm.logger.Info("開始檢查需要恢復的數據")

	if dm.influxDBConfig.DataPolicy.FailWritePath == "" {
		dm.logger.Info("未配置數據恢復路徑，跳過恢復")
		return
	}

	// 掃描恢復目錄
	files, err := os.ReadDir(dm.influxDBConfig.DataPolicy.FailWritePath)
	if err != nil {
		dm.logger.Error("掃描恢復目錄失敗", zap.Error(err))
		return
	}

	// 按文件名排序（通常包含時間戳）
	// TODO: 實現文件排序

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(dm.influxDBConfig.DataPolicy.FailWritePath, file.Name())

		// 讀取文件內容
		data, err := os.ReadFile(filePath)
		if err != nil {
			dm.logger.Error("讀取恢復文件失敗",
				zap.String("file", filePath),
				zap.Error(err),
			)
			continue
		}

		// 將數據添加到恢復隊列
		select {
		case dm.recoveryQueue <- data:
			dm.logger.Info("添加數據到恢復隊列", zap.String("file", filePath))
		default:
			dm.logger.Warn("恢復隊列已滿，稍後再試", zap.String("file", filePath))
			return // 隊列已滿，下次再嘗試
		}
	}
}

// processRecoveryQueue 處理恢復隊列中的數據
func (dm *DataManager) processRecoveryQueue() {
	for {
		select {
		case <-dm.ctx.Done():
			return
		case data := <-dm.recoveryQueue:
			// 處理恢復的數據
			// 將數據發送到InfluxDB
			if err := dm.sendToInfluxDB(data); err != nil {
				dm.logger.Error("恢復數據到InfluxDB失敗", zap.Error(err))
				// 可以考慮重新放回隊列或重新寫入文件系統
			} else {
				dm.logger.Info("數據恢復成功")
				// 成功恢復後，可以刪除對應的文件
				// 需要記錄哪個數據來自哪個文件
			}
		}
	}
}

// sendToInfluxDB 將數據發送到InfluxDB
func (dm *DataManager) sendToInfluxDB(data []byte) error {
	// TODO: 實現實際的InfluxDB寫入邏輯
	// 這裡應該使用InfluxDB客戶端發送數據

	return nil
}

// ScheduleTask 根據配置調度任務
func (dm *DataManager) ScheduleTask(taskName string) error {
	switch taskName {
	case "recover_check":
		go dm.startRecovery()
	case "disk_check":
		// 檢查磁盤空間
		available := dm.checkDiskSpace()
		dm.logger.Info("磁盤空間檢查", zap.Bool("available", available))
	default:
		return fmt.Errorf("未知的任務: %s", taskName)
	}

	return nil
}

// ProcessAndOutput 處理並輸出數據
func (m *DataManager) ProcessAndOutput(points []models.RestAPIPoint) error {
	// 實現處理和輸出數據的邏輯
	return nil
}

// RecoverBackupData 恢復備份數據
func (m *DataManager) RecoverBackupData() error {
	// 實現恢復備份數據的邏輯
	return nil
}
