package recovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"viot/logger"
	"viot/models"

	"go.uber.org/zap"
)

// JSONFallbackStrategy 實現基於JSON文件的備份策略
type JSONFallbackStrategy struct {
	basePath     string
	pduDataMutex sync.Mutex
	acRackMutex  sync.Mutex
	logger       logger.Logger
}

// NewJSONFallbackStrategy 創建一個新的JSON文件備份策略
func NewJSONFallbackStrategy(basePath string, logger logger.Logger) *JSONFallbackStrategy {
	// 確保備份目錄存在
	backupDir := filepath.Join(basePath, "backup")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		logger.Error("無法創建備份目錄", zap.String("dir", backupDir), zap.String("error", err.Error()))
	}

	backupPDUDir := filepath.Join(backupDir, "pdu")
	if err := os.MkdirAll(backupPDUDir, 0755); err != nil {
		logger.Error("無法創建PDU備份目錄", zap.String("dir", backupPDUDir), zap.Any("error", err))
	}

	backupACRackDir := filepath.Join(backupDir, "acrack")
	if err := os.MkdirAll(backupACRackDir, 0755); err != nil {
		logger.Error("無法創建AC架構備份目錄", zap.String("dir", backupACRackDir), zap.Any("error", err))
	}

	return &JSONFallbackStrategy{
		basePath: basePath,
		logger:   logger.Named("json-fallback"),
	}
}

// SavePDUData 將PDU數據保存到JSON備份文件
func (j *JSONFallbackStrategy) SavePDUData(filename string, data []models.PDUData) error {
	j.pduDataMutex.Lock()
	defer j.pduDataMutex.Unlock()

	// 格式化時間戳，用於文件名
	timestamp := time.Now().Format("20060102-150405")
	backupFilename := fmt.Sprintf("%s_%s.json", filepath.Base(filename), timestamp)
	backupPath := filepath.Join(j.basePath, "backup", "pdu", backupFilename)

	// 建立備份索引文件
	indexPath := filepath.Join(j.basePath, "backup", "pdu", "pending_index.json")
	var pendingIndex map[string][]string
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		pendingIndex = make(map[string][]string)
	} else {
		indexFile, err := os.Open(indexPath)
		if err != nil {
			j.logger.Error("無法開啟備份索引文件", zap.String("file", indexPath), zap.Any("error", err))
			pendingIndex = make(map[string][]string)
		} else {
			defer indexFile.Close()
			if err := json.NewDecoder(indexFile).Decode(&pendingIndex); err != nil {
				j.logger.Error("解析備份索引文件失敗", zap.String("file", indexPath), zap.Any("error", err))
				pendingIndex = make(map[string][]string)
			}
		}
	}

	// 將數據序列化為JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		j.logger.Error("序列化PDU數據失敗", zap.String("file", backupPath), zap.Any("error", err))
		return fmt.Errorf("序列化PDU數據失敗: %w", err)
	}

	// 寫入備份文件
	if err := os.WriteFile(backupPath, jsonData, 0644); err != nil {
		j.logger.Error("寫入PDU備份文件失敗", zap.String("file", backupPath), zap.Any("error", err))
		return fmt.Errorf("寫入PDU備份文件失敗: %w", err)
	}

	// 更新索引
	if _, exists := pendingIndex[filename]; !exists {
		pendingIndex[filename] = []string{}
	}
	pendingIndex[filename] = append(pendingIndex[filename], backupFilename)

	// 保存索引文件
	indexData, err := json.Marshal(pendingIndex)
	if err != nil {
		j.logger.Error("序列化備份索引失敗", zap.String("file", indexPath), zap.Any("error", err))
		return fmt.Errorf("序列化備份索引失敗: %w", err)
	}

	if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
		j.logger.Error("寫入備份索引文件失敗", zap.String("file", indexPath), zap.Any("error", err))
		return fmt.Errorf("寫入備份索引文件失敗: %w", err)
	}

	j.logger.Info("成功保存PDU備份數據", zap.String("file", backupPath))
	return nil
}

// SaveDeviceData 將空調機架數據保存到JSON備份文件
func (j *JSONFallbackStrategy) SaveDeviceData(filename string, data []models.DeviceData) error {
	j.acRackMutex.Lock()
	defer j.acRackMutex.Unlock()

	// 格式化時間戳，用於文件名
	timestamp := time.Now().Format("20060102-150405")
	backupFilename := fmt.Sprintf("%s_%s.json", filepath.Base(filename), timestamp)
	backupPath := filepath.Join(j.basePath, "backup", "acrack", backupFilename)

	// 建立備份索引文件
	indexPath := filepath.Join(j.basePath, "backup", "acrack", "pending_index.json")
	var pendingIndex map[string][]string
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		pendingIndex = make(map[string][]string)
	} else {
		indexFile, err := os.Open(indexPath)
		if err != nil {
			j.logger.Error("無法開啟備份索引文件", zap.String("file", indexPath), zap.Any("error", err))
			pendingIndex = make(map[string][]string)
		} else {
			defer indexFile.Close()
			if err := json.NewDecoder(indexFile).Decode(&pendingIndex); err != nil {
				j.logger.Error("解析備份索引文件失敗", zap.String("file", indexPath), zap.Any("error", err))
				pendingIndex = make(map[string][]string)
			}
		}
	}

	// 將數據序列化為JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		j.logger.Error("序列化AC架構數據失敗", zap.String("file", backupPath), zap.Any("error", err))
		return fmt.Errorf("序列化AC架構數據失敗: %w", err)
	}

	// 寫入備份文件
	if err := os.WriteFile(backupPath, jsonData, 0644); err != nil {
		j.logger.Error("寫入AC架構備份文件失敗", zap.String("file", backupPath), zap.Any("error", err))
		return fmt.Errorf("寫入AC架構備份文件失敗: %w", err)
	}

	// 更新索引
	if _, exists := pendingIndex[filename]; !exists {
		pendingIndex[filename] = []string{}
	}
	pendingIndex[filename] = append(pendingIndex[filename], backupFilename)

	// 保存索引文件
	indexData, err := json.Marshal(pendingIndex)
	if err != nil {
		j.logger.Error("序列化備份索引失敗", zap.String("file", indexPath), zap.Any("error", err))
		return fmt.Errorf("序列化備份索引失敗: %w", err)
	}

	if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
		j.logger.Error("寫入備份索引文件失敗", zap.String("file", indexPath), zap.Any("error", err))
		return fmt.Errorf("寫入備份索引文件失敗: %w", err)
	}

	j.logger.Info("成功保存AC架構備份數據", zap.String("file", backupPath))
	return nil
}

// GetPendingPDUData 獲取待處理的PDU數據
func (j *JSONFallbackStrategy) GetPendingPDUData() (map[string][]models.PDUData, error) {
	j.pduDataMutex.Lock()
	defer j.pduDataMutex.Unlock()

	result := make(map[string][]models.PDUData)
	indexPath := filepath.Join(j.basePath, "backup", "pdu", "pending_index.json")

	// 檢查索引文件是否存在
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return result, nil
	}

	// 讀取索引文件
	indexFile, err := os.Open(indexPath)
	if err != nil {
		j.logger.Error("無法開啟備份索引文件", zap.String("file", indexPath), zap.Any("error", err))
		return result, fmt.Errorf("無法開啟備份索引文件: %w", err)
	}
	defer indexFile.Close()

	// 解析索引
	var pendingIndex map[string][]string
	if err := json.NewDecoder(indexFile).Decode(&pendingIndex); err != nil {
		j.logger.Error("解析備份索引文件失敗", zap.String("file", indexPath), zap.Any("error", err))
		return result, fmt.Errorf("解析備份索引文件失敗: %w", err)
	}

	// 讀取並合併所有備份文件
	for filename, backupFiles := range pendingIndex {
		for _, backupFile := range backupFiles {
			backupPath := filepath.Join(j.basePath, "backup", "pdu", backupFile)

			// 讀取備份文件
			data, err := os.ReadFile(backupPath)
			if err != nil {
				j.logger.Error("讀取備份文件失敗", zap.String("file", backupPath), zap.Any("error", err))
				continue
			}

			// 解析PDU數據
			var pduData []models.PDUData
			if err := json.Unmarshal(data, &pduData); err != nil {
				j.logger.Error("解析備份PDU數據失敗", zap.String("file", backupPath), zap.Any("error", err))
				continue
			}

			// 合併數據
			if _, exists := result[filename]; !exists {
				result[filename] = []models.PDUData{}
			}
			result[filename] = append(result[filename], pduData...)
		}
	}

	return result, nil
}

// GetPendingDeviceData 獲取待處理的空調機架數據
func (j *JSONFallbackStrategy) GetPendingDeviceData() (map[string][]models.DeviceData, error) {
	j.acRackMutex.Lock()
	defer j.acRackMutex.Unlock()

	result := make(map[string][]models.DeviceData)
	indexPath := filepath.Join(j.basePath, "backup", "acrack", "pending_index.json")

	// 檢查索引文件是否存在
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return result, nil
	}

	// 讀取索引文件
	indexFile, err := os.Open(indexPath)
	if err != nil {
		j.logger.Error("無法開啟備份索引文件", zap.String("file", indexPath), zap.Any("error", err))
		return result, fmt.Errorf("無法開啟備份索引文件: %w", err)
	}
	defer indexFile.Close()

	// 解析索引
	var pendingIndex map[string][]string
	if err := json.NewDecoder(indexFile).Decode(&pendingIndex); err != nil {
		j.logger.Error("解析備份索引文件失敗", zap.String("file", indexPath), zap.Any("error", err))
		return result, fmt.Errorf("解析備份索引文件失敗: %w", err)
	}

	// 讀取並合併所有備份文件
	for filename, backupFiles := range pendingIndex {
		for _, backupFile := range backupFiles {
			backupPath := filepath.Join(j.basePath, "backup", "acrack", backupFile)

			// 讀取備份文件
			data, err := os.ReadFile(backupPath)
			if err != nil {
				j.logger.Error("讀取備份文件失敗", zap.String("file", backupPath), zap.Any("error", err))
				continue
			}

			// 解析AC架構數據
			var acRackData []models.DeviceData
			if err := json.Unmarshal(data, &acRackData); err != nil {
				j.logger.Error("解析備份AC架構數據失敗", zap.String("file", backupPath), zap.Any("error", err))
				continue
			}

			// 合併數據
			if _, exists := result[filename]; !exists {
				result[filename] = []models.DeviceData{}
			}
			result[filename] = append(result[filename], acRackData...)
		}
	}

	return result, nil
}

// MarkPDUDataProcessed 標記PDU數據已處理
func (j *JSONFallbackStrategy) MarkPDUDataProcessed(filename string) error {
	j.pduDataMutex.Lock()
	defer j.pduDataMutex.Unlock()

	indexPath := filepath.Join(j.basePath, "backup", "pdu", "pending_index.json")

	// 檢查索引文件是否存在
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return nil // 沒有待處理數據
	}

	// 讀取索引文件
	indexFile, err := os.Open(indexPath)
	if err != nil {
		j.logger.Error("無法開啟備份索引文件", zap.String("file", indexPath), zap.Any("error", err))
		return fmt.Errorf("無法開啟備份索引文件: %w", err)
	}

	// 解析索引
	var pendingIndex map[string][]string
	if err := json.NewDecoder(indexFile).Decode(&pendingIndex); err != nil {
		indexFile.Close()
		j.logger.Error("解析備份索引文件失敗", zap.String("file", indexPath), zap.Any("error", err))
		return fmt.Errorf("解析備份索引文件失敗: %w", err)
	}
	indexFile.Close()

	// 刪除對應的備份文件
	if backupFiles, exists := pendingIndex[filename]; exists {
		for _, backupFile := range backupFiles {
			backupPath := filepath.Join(j.basePath, "backup", "pdu", backupFile)
			if err := os.Remove(backupPath); err != nil {
				j.logger.Error("刪除備份文件失敗", zap.String("file", backupPath), zap.Any("error", err))
				// 繼續處理其他文件
			}
		}
		// 從索引中移除
		delete(pendingIndex, filename)
	}

	// 保存更新後的索引
	indexData, err := json.Marshal(pendingIndex)
	if err != nil {
		j.logger.Error("序列化備份索引失敗", zap.String("file", indexPath), zap.Any("error", err))
		return fmt.Errorf("序列化備份索引失敗: %w", err)
	}

	if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
		j.logger.Error("寫入備份索引文件失敗", zap.String("file", indexPath), zap.Any("error", err))
		return fmt.Errorf("寫入備份索引文件失敗: %w", err)
	}

	j.logger.Info("成功標記PDU數據已處理", zap.String("filename", filename))
	return nil
}

// MarkDeviceDataProcessed 標記空調機架數據已處理
func (j *JSONFallbackStrategy) MarkDeviceDataProcessed(filename string) error {
	j.acRackMutex.Lock()
	defer j.acRackMutex.Unlock()

	indexPath := filepath.Join(j.basePath, "backup", "acrack", "pending_index.json")

	// 檢查索引文件是否存在
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return nil // 沒有待處理數據
	}

	// 讀取索引文件
	indexFile, err := os.Open(indexPath)
	if err != nil {
		j.logger.Error("無法開啟備份索引文件", zap.String("file", indexPath), zap.Any("error", err))
		return fmt.Errorf("無法開啟備份索引文件: %w", err)
	}

	// 解析索引
	var pendingIndex map[string][]string
	if err := json.NewDecoder(indexFile).Decode(&pendingIndex); err != nil {
		indexFile.Close()
		j.logger.Error("解析備份索引文件失敗", zap.String("file", indexPath), zap.Any("error", err))
		return fmt.Errorf("解析備份索引文件失敗: %w", err)
	}
	indexFile.Close()

	// 刪除對應的備份文件
	if backupFiles, exists := pendingIndex[filename]; exists {
		for _, backupFile := range backupFiles {
			backupPath := filepath.Join(j.basePath, "backup", "acrack", backupFile)
			if err := os.Remove(backupPath); err != nil {
				j.logger.Error("刪除備份文件失敗", zap.String("file", backupPath), zap.Any("error", err))
				// 繼續處理其他文件
			}
		}
		// 從索引中移除
		delete(pendingIndex, filename)
	}

	// 保存更新後的索引
	indexData, err := json.Marshal(pendingIndex)
	if err != nil {
		j.logger.Error("序列化備份索引失敗", zap.String("file", indexPath), zap.Any("error", err))
		return fmt.Errorf("序列化備份索引失敗: %w", err)
	}

	if err := os.WriteFile(indexPath, indexData, 0644); err != nil {
		j.logger.Error("寫入備份索引文件失敗", zap.String("file", indexPath), zap.Any("error", err))
		return fmt.Errorf("寫入備份索引文件失敗: %w", err)
	}

	j.logger.Info("成功標記AC架構數據已處理", zap.String("filename", filename))
	return nil
}
