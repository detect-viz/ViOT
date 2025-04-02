package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"viot/domain/common"
	"viot/logger"
)

// Handler 處理 CSV 文件操作
type Handler struct {
	basePath string
	logger   logger.Logger
}

// NewHandler 創建一個新的 CSV 處理器
func NewHandler(basePath string, logger logger.Logger) *Handler {
	return &Handler{
		basePath: basePath,
		logger:   logger.Named("csv"),
	}
}

// GetBasePath 獲取基礎路徑
func (h *Handler) GetBasePath() string {
	return h.basePath
}

// ReadPDUData 從 CSV 文件讀取 PDU 數據
func (h *Handler) ReadPDUData(filename, room string) ([]common.PDUData, error) {
	filePath := filepath.Join(h.basePath, filename+".csv")
	h.logger.Debug("讀取 PDU 數據", logger.String("file", filePath), logger.String("room", room))

	// 檢查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.logger.Warn("PDU 數據文件不存在", logger.String("file", filePath))
		return []common.PDUData{}, nil
	}

	// 打開文件
	file, err := os.Open(filePath)
	if err != nil {
		h.logger.Error("打開 PDU 數據文件時出錯", logger.String("file", filePath), logger.Any("error", err))
		return nil, fmt.Errorf("打開 PDU 數據文件時出錯: %w", err)
	}
	defer file.Close()

	// 創建 CSV 讀取器
	reader := csv.NewReader(file)

	// 讀取標題行
	header, err := reader.Read()
	if err != nil {
		h.logger.Error("讀取標題行時出錯", logger.String("file", filePath), logger.Any("error", err))
		return nil, fmt.Errorf("讀取標題行時出錯: %w", err)
	}

	// 為標題創建索引映射
	headerMap := make(map[string]int)
	for i, name := range header {
		headerMap[name] = i
	}

	// 檢查必要的列
	requiredColumns := []string{"Name", "Room", "Status", "IP"}
	for _, col := range requiredColumns {
		if _, ok := headerMap[col]; !ok {
			h.logger.Error("CSV 文件缺少必要的列", logger.String("file", filePath), logger.String("column", col))
			return nil, fmt.Errorf("CSV 文件缺少必要的列: %s", col)
		}
	}

	// 讀取所有記錄
	records, err := reader.ReadAll()
	if err != nil {
		h.logger.Error("讀取記錄時出錯", logger.String("file", filePath), logger.Any("error", err))
		return nil, fmt.Errorf("讀取記錄時出錯: %w", err)
	}

	// 解析記錄
	var pduList []common.PDUData
	for _, record := range records {
		// 如果指定了房間，則僅處理該房間的記錄
		recordRoom := record[headerMap["Room"]]
		if room != "" && recordRoom != room {
			continue
		}

		// 創建 PDU 數據對象
		pdu := common.PDUData{
			Name:      record[headerMap["Name"]],
			Room:      recordRoom,
			Status:    record[headerMap["Status"]],
			Timestamp: time.Now(), // 使用當前時間作為時間戳
			Tags:      make(map[string]string),
		}

		// 添加標籤
		if tagsIndex, ok := headerMap["Tags"]; ok && tagsIndex < len(record) {
			tagsStr := record[tagsIndex]
			if tagsStr != "" {
				tagsList := strings.Split(tagsStr, ",")
				for _, tag := range tagsList {
					pdu.Tags[tag] = tag
				}
			}
		}

		pduList = append(pduList, pdu)
	}

	h.logger.Debug("成功讀取 PDU 數據", logger.String("file", filePath), logger.Int("count", len(pduList)))
	return pduList, nil
}

// WritePDUData 將 PDU 數據寫入 CSV 文件
func (h *Handler) WritePDUData(filename string, data []common.PDUData) error {
	filePath := filepath.Join(h.basePath, filename+".csv")
	h.logger.Debug("寫入 PDU 數據", logger.String("file", filePath), logger.Int("count", len(data)))

	// 創建目錄（如果不存在）
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.logger.Error("創建目錄時出錯", logger.String("dir", dir), logger.Any("error", err))
		return fmt.Errorf("創建目錄時出錯: %w", err)
	}

	// 創建文件
	file, err := os.Create(filePath)
	if err != nil {
		h.logger.Error("創建 PDU 數據文件時出錯", logger.String("file", filePath), logger.Any("error", err))
		return fmt.Errorf("創建 PDU 數據文件時出錯: %w", err)
	}
	defer file.Close()

	// 創建 CSV 寫入器
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 寫入標題行
	header := []string{"Name", "Room", "Status", "Current", "Voltage", "Power", "Energy", "Tags", "UpdatedAt"}
	if err := writer.Write(header); err != nil {
		h.logger.Error("寫入標題行時出錯", logger.String("file", filePath), logger.Any("error", err))
		return fmt.Errorf("寫入標題行時出錯: %w", err)
	}

	// 寫入數據
	for _, pdu := range data {
		// 將標籤轉換為逗號分隔的字符串
		var tags []string
		for tag := range pdu.Tags {
			tags = append(tags, tag)
		}
		tagsStr := strings.Join(tags, ",")

		// 創建記錄
		record := []string{
			pdu.Name,
			pdu.Room,
			pdu.Status,
			strconv.FormatFloat(pdu.Current, 'f', 2, 64),
			strconv.FormatFloat(pdu.Voltage, 'f', 2, 64),
			strconv.FormatFloat(pdu.Power, 'f', 2, 64),
			strconv.FormatFloat(pdu.Energy, 'f', 2, 64),
			tagsStr,
			pdu.Timestamp.Format(time.RFC3339),
		}

		// 寫入記錄
		if err := writer.Write(record); err != nil {
			h.logger.Error("寫入記錄時出錯", logger.String("file", filePath), logger.Any("error", err))
			return fmt.Errorf("寫入記錄時出錯: %w", err)
		}
	}

	h.logger.Debug("成功寫入 PDU 數據", logger.String("file", filePath), logger.Int("count", len(data)))
	return nil
}

// ReadDeviceData 從 CSV 文件讀取空調機架數據
func (h *Handler) ReadDeviceData(filename string, deviceName string) (common.RoomLayout, error) {
	filePath := filepath.Join(h.basePath, filename)
	h.logger.Debug("讀取空調機架數據", logger.String("file", filePath), logger.String("deviceName", deviceName))

	// 檢查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.logger.Error("機架數據文件不存在", logger.String("file", filePath), logger.Any("error", err))
		return common.RoomLayout{}, fmt.Errorf("機架數據文件不存在: %s", filePath)
	}

	// 打開文件
	file, err := os.Open(filePath)
	if err != nil {
		h.logger.Error("無法打開機架數據文件", logger.String("file", filePath), logger.Any("error", err))
		return common.RoomLayout{}, fmt.Errorf("無法打開機架數據文件: %w", err)
	}
	defer file.Close()

	// 創建 CSV 讀取器
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // 允許不同行有不同數量的字段

	// 讀取所有記錄
	records, err := reader.ReadAll()
	if err != nil {
		h.logger.Error("讀取 CSV 記錄時出錯", logger.String("file", filePath), logger.Any("error", err))
		return common.RoomLayout{}, fmt.Errorf("讀取 CSV 記錄時出錯: %w", err)
	}

	// 解析記錄
	var rackData []common.RoomLayout
	for i, record := range records {
		// 跳過標題行
		if i == 0 {
			continue
		}

		// 確保記錄有足夠的字段
		if len(record) < 5 {
			h.logger.Error("記錄格式無效", logger.String("file", filePath), logger.Int("line", i+1))
			return common.RoomLayout{}, fmt.Errorf("第 %d 行的記錄格式無效", i+1)
		}

		// 解析行和列
		row, err := strconv.Atoi(record[2])
		if err != nil {
			h.logger.Error("行號無效", logger.String("file", filePath), logger.Int("line", i+1), logger.Any("error", err))
			return common.RoomLayout{}, fmt.Errorf("第 %d 行的行號無效: %w", i+1, err)
		}

		col, err := strconv.Atoi(record[3])
		if err != nil {
			h.logger.Error("列號無效", logger.String("file", filePath), logger.Int("line", i+1), logger.Any("error", err))
			return common.RoomLayout{}, fmt.Errorf("第 %d 行的列號無效: %w", i+1, err)
		}

		// 如果指定了設備名稱，則只返回該設備的數據
		if deviceName != "" && record[0] != deviceName {
			continue
		}

		// 創建機架數據
		rack := common.DevicePosition{
			Position: common.PositionData{
				Key:  record[0],
				Type: record[4],
				Row:  row,
				Col:  col,
			},
			Status: record[5],
		}

		rackData = append(rackData, common.RoomLayout{
			Room:    record[1],
			Devices: []common.DevicePosition{rack},
		})
	}

	h.logger.Debug("成功讀取空調機架數據", logger.String("file", filePath), logger.Int("count", len(rackData)))
	if len(rackData) == 0 {
		return common.RoomLayout{}, fmt.Errorf("no data found")
	}
	return rackData[0], nil
}

// WriteDeviceData 將空調機架數據寫入 CSV 文件
func (h *Handler) WriteDeviceData(filename string, data []common.RoomLayout) error {
	filePath := filepath.Join(h.basePath, filename)
	h.logger.Debug("寫入空調機架數據", logger.String("file", filePath), logger.Int("count", len(data)))

	// 創建目錄（如果不存在）
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.logger.Error("創建目錄時出錯", logger.String("dir", dir), logger.Any("error", err))
		return fmt.Errorf("創建目錄時出錯: %w", err)
	}

	// 創建文件
	file, err := os.Create(filePath)
	if err != nil {
		h.logger.Error("創建機架數據文件時出錯", logger.String("file", filePath), logger.Any("error", err))
		return fmt.Errorf("創建機架數據文件時出錯: %w", err)
	}
	defer file.Close()

	// 創建 CSV 寫入器
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 寫入標題行
	header := []string{"Name", "Room", "Row", "Col", "Type"}
	if err := writer.Write(header); err != nil {
		h.logger.Error("寫入標題行時出錯", logger.String("file", filePath), logger.Any("error", err))
		return fmt.Errorf("寫入標題行時出錯: %w", err)
	}

	// 寫入數據
	for _, rack := range data {
		// 創建記錄
		record := []string{
			rack.Room,
			strconv.Itoa(rack.Devices[0].Position.Row),
			strconv.Itoa(rack.Devices[0].Position.Col),
			rack.Devices[0].Position.Type,
		}

		// 寫入記錄
		if err := writer.Write(record); err != nil {
			h.logger.Error("寫入記錄時出錯", logger.String("file", filePath), logger.Any("error", err))
			return fmt.Errorf("寫入記錄時出錯: %w", err)
		}
	}

	h.logger.Debug("成功寫入空調機架數據", logger.String("file", filePath), logger.Int("count", len(data)))
	return nil
}

// GetRooms 獲取所有房間
func (h *Handler) GetRooms(filename string) ([]string, error) {
	filePath := filepath.Join(h.basePath, filename)
	h.logger.Debug("獲取所有房間", logger.String("file", filePath))

	// 檢查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		h.logger.Error("PDU 數據文件不存在", logger.String("file", filePath), logger.Any("error", err))
		return nil, fmt.Errorf("PDU 數據文件不存在: %s", filePath)
	}

	// 打開文件
	file, err := os.Open(filePath)
	if err != nil {
		h.logger.Error("無法打開 PDU 數據文件", logger.String("file", filePath), logger.Any("error", err))
		return nil, fmt.Errorf("無法打開 PDU 數據文件: %w", err)
	}
	defer file.Close()

	// 創建 CSV 讀取器
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // 允許不同行有不同數量的字段

	// 讀取所有記錄
	records, err := reader.ReadAll()
	if err != nil {
		h.logger.Error("讀取 CSV 記錄時出錯", logger.String("file", filePath), logger.Any("error", err))
		return nil, fmt.Errorf("讀取 CSV 記錄時出錯: %w", err)
	}

	// 使用 map 來去重
	roomMap := make(map[string]bool)
	for i, record := range records {
		// 跳過標題行
		if i == 0 {
			continue
		}

		// 確保記錄有足夠的字段
		if len(record) < 2 {
			h.logger.Error("記錄格式無效", logger.String("file", filePath), logger.Int("line", i+1))
			return nil, fmt.Errorf("第 %d 行的記錄格式無效", i+1)
		}

		// 添加房間
		if record[1] != "" {
			roomMap[record[1]] = true
		}
	}

	// 將 map 轉換為切片
	rooms := make([]string, 0, len(roomMap))
	for room := range roomMap {
		rooms = append(rooms, room)
	}

	h.logger.Debug("成功獲取所有房間", logger.String("file", filePath), logger.Int("count", len(rooms)))
	return rooms, nil
}

// UpdatePDUStatus 更新 PDU 狀態
func (h *Handler) UpdatePDUStatus(filename, pduName, status string) error {
	// 讀取所有 PDU 數據
	pduList, err := h.ReadPDUData(filename, "")
	if err != nil {
		return err
	}

	// 更新指定 PDU 的狀態
	found := false
	for i, pdu := range pduList {
		if pdu.Name == pduName {
			pduList[i].Status = status
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("未找到 PDU: %s", pduName)
	}

	// 寫回所有 PDU 數據
	return h.WritePDUData(filename, pduList)
}

// UpdatePDUTags 更新 PDU 標籤
func (h *Handler) UpdatePDUTags(filename, pduName string, tags []string) error {
	// 讀取所有 PDU 數據
	pduList, err := h.ReadPDUData(filename, "")
	if err != nil {
		return err
	}

	// 更新指定 PDU 的標籤
	found := false
	for i, pdu := range pduList {
		if pdu.Name == pduName {
			// 創建新的標籤映射
			newTags := make(map[string]string)
			for _, tag := range tags {
				newTags[tag] = tag
			}
			pduList[i].Tags = newTags
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("未找到 PDU: %s", pduName)
	}

	// 寫回所有 PDU 數據
	return h.WritePDUData(filename, pduList)
}

// GetPDUTags 獲取 PDU 標籤
func (h *Handler) GetPDUTags(filename, pduName string) ([]string, error) {
	// 讀取所有 PDU 數據
	pduList, err := h.ReadPDUData(filename, "")
	if err != nil {
		return nil, err
	}

	// 查找指定 PDU 並返回其標籤
	for _, pdu := range pduList {
		if pdu.Name == pduName {
			var tags []string
			for tag := range pdu.Tags {
				tags = append(tags, tag)
			}
			return tags, nil
		}
	}

	return nil, fmt.Errorf("未找到 PDU: %s", pduName)
}

// ClearPDUData 清除 PDU 數據
func (h *Handler) ClearPDUData(filename string) error {
	filePath := filepath.Join(h.basePath, filename+".csv")

	// 檢查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 文件不存在，視為已清除
		return nil
	}

	// 刪除文件
	if err := os.Remove(filePath); err != nil {
		h.logger.Error("刪除備份檔案失敗", logger.String("file", filePath), logger.Any("error", err))
		return fmt.Errorf("刪除備份檔案失敗: %w", err)
	}

	h.logger.Info("成功清除備份檔案", logger.String("file", filePath))
	return nil
}
