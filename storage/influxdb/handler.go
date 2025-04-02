package influxdb

import (
	"context"
	"fmt"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"go.uber.org/zap"

	"viot/logger"
	"viot/models"
	"viot/storage/recovery"
)

// 採用接口模式，以便在測試時可以模擬InfluxDB客戶端
type influxClient interface {
	Close()
	Ping(ctx context.Context) (bool, error)
}

type influxWriteAPI interface {
	WritePoint(point *Point)
	Flush()
	Errors() <-chan error
}

type influxQueryAPI interface {
	Query(ctx context.Context, query string) (QueryResult, error)
}

type QueryResult interface {
	Next() bool
	Record() Record
	Err() error
	Close()
}

type Record interface {
	Time() time.Time
	ValueByKey(key string) interface{}
}

// Point 表示一個數據點
type Point struct {
	Measurement string
	Tags        map[string]string
	Fields      map[string]interface{}
	Timestamp   time.Time
}

// NewPoint 創建一個新的數據點
func NewPoint(measurement string, tags map[string]string, fields map[string]interface{}, timestamp time.Time) *Point {
	return &Point{
		Measurement: measurement,
		Tags:        tags,
		Fields:      fields,
		Timestamp:   timestamp,
	}
}

// Config InfluxDB配置
type Config struct {
	URL          string `json:"url"`
	Token        string `json:"token"`
	Organization string `json:"organization"`
	Bucket       string `json:"bucket"`
	BatchSize    int    `json:"batch_size"`
}

// Handler InfluxDB處理程式
type Handler struct {
	client      influxdb2.Client
	writeAPI    api.WriteAPI
	queryAPI    api.QueryAPI
	fallback    recovery.FallbackStrategy
	retryPolicy recovery.RetryPolicy
	config      Config
	logger      logger.Logger
	connected   bool
	bucketTags  map[string]map[string]string // 每個桶的默認標籤
}

// NewHandler 創建一個新的InfluxDB處理程式
func NewHandler(config Config, fallbackStrategy recovery.FallbackStrategy, retryPolicy recovery.RetryPolicy, log logger.Logger) (*Handler, error) {
	if config.URL == "" || config.Token == "" {
		return nil, fmt.Errorf("missing required InfluxDB configuration")
	}

	// 初始化客戶端
	client := influxdb2.NewClientWithOptions(
		config.URL,
		config.Token,
		influxdb2.DefaultOptions().SetBatchSize(uint(config.BatchSize)),
	)

	// 測試連接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx)
	if err != nil {
		log.Error("InfluxDB連接失敗", zap.String("url", config.URL), zap.Error(err))
		return nil, fmt.Errorf("無法連接到InfluxDB: %w", err)
	}

	// 獲取寫入API
	writeAPI := client.WriteAPI(config.Organization, config.Bucket)

	// 獲取查詢API
	queryAPI := client.QueryAPI(config.Organization)

	log.Info("InfluxDB連接成功", zap.String("url", config.URL))

	// 返回處理程式
	return &Handler{
		client:      client,
		writeAPI:    writeAPI,
		queryAPI:    queryAPI,
		fallback:    fallbackStrategy,
		retryPolicy: retryPolicy,
		config:      config,
		logger:      log.Named("influxdb"),
		connected:   true,
		bucketTags:  make(map[string]map[string]string),
	}, nil
}

// Close 關閉InfluxDB連接
func (h *Handler) Close() {
	if h.client != nil {
		h.client.Close()
	}
}

// ReadPDUData 從InfluxDB讀取PDU數據
func (h *Handler) ReadPDUData(filename, room string) ([]models.PDUData, error) {
	// 檢查連接狀態
	if !h.connected {
		h.logger.Warn("InfluxDB未連接，無法讀取PDU數據")
		return nil, fmt.Errorf("InfluxDB未連接")
	}

	h.logger.Debug("嘗試從InfluxDB讀取PDU數據", zap.String("room", room))

	// 構建Flux查詢
	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: -24h)
			|> filter(fn: (r) => r._measurement == "pdu" and r.room == "%s")
			|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> last()
	`, h.config.Bucket, room)

	// 執行查詢
	result, err := h.queryAPI.Query(context.Background(), query)
	if err != nil {
		h.logger.Error("InfluxDB查詢失敗", zap.String("room", room), zap.Error(err))
		return nil, fmt.Errorf("InfluxDB查詢失敗: %w", err)
	}
	defer result.Close()

	// 解析結果
	var pduData []models.PDUData
	for result.Next() {
		record := result.Record()
		row, _ := strconv.Atoi(record.ValueByKey("row").(string))
		col, _ := strconv.Atoi(record.ValueByKey("col").(string))
		monitored := record.ValueByKey("monitored").(string) == "true"

		// 解析標籤
		var tags []string
		if tagsStr, ok := record.ValueByKey("tags").(string); ok && tagsStr != "" {
			tags = parseTags(tagsStr)
		}

		// 構建PDU數據
		pdu := models.PDUData{
			Name: record.ValueByKey("name").(string),

			Status: record.ValueByKey("status").(string),

			Tags: tags,
		}

		pduData = append(pduData, pdu)
	}

	if err := result.Err(); err != nil {
		h.logger.Error("解析InfluxDB查詢結果失敗", zap.String("room", room), zap.Error(err))
		return nil, fmt.Errorf("解析InfluxDB查詢結果失敗: %w", err)
	}

	return pduData, nil
}

// WritePDUData 將PDU數據寫入InfluxDB
func (h *Handler) WritePDUData(filename string, data []models.PDUData) error {
	if !h.connected {
		// 如果未連接，嘗試備份數據
		h.logger.Warn("InfluxDB未連接，將數據寫入備份", zap.String("filename", filename))
		return h.fallbackToDisk(filename, data)
	}

	h.logger.Debug("嘗試寫入PDU數據到InfluxDB", zap.String("filename", filename), zap.Int("count", len(data)))

	// 寫入數據到InfluxDB
	for _, pdu := range data {
		// 創建新的Point
		point := influxdb2.NewPoint(
			"pdu", // measurement
			map[string]string{
				"name": pdu.Name,
				"room": pdu.Room,
				"side": pdu.Side,
				"type": pdu.Type,
				"ip":   pdu.IP,
			},
			map[string]interface{}{
				"row":         pdu.Row,
				"col":         pdu.Col,
				"status":      pdu.Status,
				"monitored":   pdu.Monitored,
				"tags":        joinTags(pdu.Tags),
				"last_update": pdu.LastUpdate,
			},
			time.Now(),
		)

		// 添加點到批次
		h.writeAPI.WritePoint(point)
	}

	// 刷新以確保數據寫入
	h.writeAPI.Flush()

	// 檢查錯誤
	if writeErr := h.checkWriteError(); writeErr != nil {
		h.connected = false
		h.logger.Error("寫入InfluxDB失敗，將數據寫入備份", zap.String("filename", filename), zap.Error(writeErr))
		return h.fallbackToDisk(filename, data)
	}

	h.logger.Info("成功寫入PDU數據到InfluxDB", zap.String("filename", filename), zap.Int("count", len(data)))
	return nil
}

// UpdatePDUStatus 更新PDU狀態
func (h *Handler) UpdatePDUStatus(filename, pduName, status string) error {
	if !h.connected {
		h.logger.Warn("InfluxDB未連接，無法更新PDU狀態", zap.String("name", pduName))
		return fmt.Errorf("InfluxDB未連接")
	}

	h.logger.Debug("嘗試更新PDU狀態", zap.String("name", pduName), zap.String("status", status))

	// 創建新的Point更新狀態
	point := influxdb2.NewPoint(
		"pdu_status", // measurement
		map[string]string{
			"name": pduName,
		},
		map[string]interface{}{
			"status":      status,
			"update_time": time.Now().Format(time.RFC3339),
		},
		time.Now(),
	)

	h.writeAPI.WritePoint(point)
	h.writeAPI.Flush()

	// 檢查錯誤
	if writeErr := h.checkWriteError(); writeErr != nil {
		h.connected = false
		h.logger.Error("更新InfluxDB中的PDU狀態失敗", zap.String("name", pduName), zap.Error(writeErr))
		return fmt.Errorf("更新InfluxDB中的PDU狀態失敗: %w", writeErr)
	}

	h.logger.Info("成功更新PDU狀態", zap.String("name", pduName), zap.String("status", status))
	return nil
}

// UpdatePDUTags 更新PDU標籤
func (h *Handler) UpdatePDUTags(filename, pduName string, tags []string) error {
	if !h.connected {
		h.logger.Warn("InfluxDB未連接，無法更新PDU標籤", zap.String("name", pduName))
		return fmt.Errorf("InfluxDB未連接")
	}

	h.logger.Debug("嘗試更新PDU標籤", zap.String("name", pduName), zap.Any("tags", tags))

	// 創建新的Point更新標籤
	point := influxdb2.NewPoint(
		"pdu_tags", // measurement
		map[string]string{
			"name": pduName,
		},
		map[string]interface{}{
			"tags":        joinTags(tags),
			"update_time": time.Now().Format(time.RFC3339),
		},
		time.Now(),
	)

	h.writeAPI.WritePoint(point)
	h.writeAPI.Flush()

	// 檢查錯誤
	if writeErr := h.checkWriteError(); writeErr != nil {
		h.connected = false
		h.logger.Error("更新InfluxDB中的PDU標籤失敗", zap.String("name", pduName), zap.Error(writeErr))
		return fmt.Errorf("更新InfluxDB中的PDU標籤失敗: %w", writeErr)
	}

	h.logger.Info("成功更新PDU標籤", zap.String("name", pduName), zap.Any("tags", tags))
	return nil
}

// GetPDUTags 獲取PDU標籤
func (h *Handler) GetPDUTags(filename, pduName string) ([]string, error) {
	if !h.connected {
		h.logger.Warn("InfluxDB未連接，無法獲取PDU標籤", zap.String("name", pduName))
		return nil, fmt.Errorf("InfluxDB未連接")
	}

	h.logger.Debug("嘗試獲取PDU標籤", zap.String("name", pduName))

	// 構建Flux查詢
	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: -24h)
			|> filter(fn: (r) => r._measurement == "pdu_tags" and r.name == "%s")
			|> last()
	`, h.config.Bucket, pduName)

	// 執行查詢
	result, err := h.queryAPI.Query(context.Background(), query)
	if err != nil {
		h.logger.Error("InfluxDB查詢失敗", zap.String("name", pduName), zap.Error(err))
		return nil, fmt.Errorf("InfluxDB查詢失敗: %w", err)
	}
	defer result.Close()

	// 解析結果
	var tags []string
	for result.Next() {
		record := result.Record()
		if tagsStr, ok := record.ValueByKey("tags").(string); ok && tagsStr != "" {
			tags = parseTags(tagsStr)
		}
	}

	if err := result.Err(); err != nil {
		h.logger.Error("解析InfluxDB查詢結果失敗", zap.String("name", pduName), zap.Error(err))
		return nil, fmt.Errorf("解析InfluxDB查詢結果失敗: %w", err)
	}

	return tags, nil
}

// ReadDeviceData 從InfluxDB讀取空調機架數據
func (h *Handler) ReadDeviceData(filename, room string) ([]models.DeviceData, error) {
	if !h.connected {
		h.logger.Warn("InfluxDB未連接，無法讀取空調機架數據")
		return nil, fmt.Errorf("InfluxDB未連接")
	}

	h.logger.Debug("嘗試從InfluxDB讀取空調機架數據", zap.String("room", room))

	// 構建Flux查詢
	query := fmt.Sprintf(`
		from(bucket: "%s")
			|> range(start: -24h)
			|> filter(fn: (r) => r._measurement == "acrack" and r.room == "%s")
			|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
			|> last()
	`, h.config.Bucket, room)

	// 執行查詢
	result, err := h.queryAPI.Query(context.Background(), query)
	if err != nil {
		h.logger.Error("InfluxDB查詢失敗", zap.String("room", room), zap.Error(err))
		return nil, fmt.Errorf("InfluxDB查詢失敗: %w", err)
	}
	defer result.Close()

	// 解析結果
	var acRackData []models.DeviceData
	for result.Next() {
		record := result.Record()
		row, _ := strconv.Atoi(record.ValueByKey("row").(string))
		col, _ := strconv.Atoi(record.ValueByKey("col").(string))

		// 構建空調機架數據
		acRack := models.DeviceData{
			Name: record.ValueByKey("name").(string),
			Room: record.ValueByKey("room").(string),
			Row:  row,
			Col:  col,
			Type: record.ValueByKey("type").(string),
		}

		acRackData = append(acRackData, acRack)
	}

	if err := result.Err(); err != nil {
		h.logger.Error("解析InfluxDB查詢結果失敗", zap.String("room", room), zap.Error(err))
		return nil, fmt.Errorf("解析InfluxDB查詢結果失敗: %w", err)
	}

	return acRackData, nil
}

// ReprocessBackups 重新處理備份數據
func (h *Handler) ReprocessBackups() error {
	// 獲取待處理的PDU數據
	pendingPDUData, err := h.fallback.GetPendingPDUData()
	if err != nil {
		h.logger.Error("獲取待處理PDU數據失敗", zap.Error(err))
		return fmt.Errorf("獲取待處理PDU數據失敗: %w", err)
	}

	// 獲取待處理的空調機架數據
	pendingDeviceData, err := h.fallback.GetPendingDeviceData()
	if err != nil {
		h.logger.Error("獲取待處理空調機架數據失敗", zap.Error(err))
		return fmt.Errorf("獲取待處理空調機架數據失敗: %w", err)
	}

	// 檢查連接狀態
	if !h.connected {
		// 嘗試重新連接
		_, err := h.client.Ping(context.Background())
		if err != nil {
			h.logger.Error("重新連接InfluxDB失敗", zap.Error(err))
			return fmt.Errorf("重新連接InfluxDB失敗: %w", err)
		}
		h.connected = true
	}

	// 處理PDU數據
	for filename, dataList := range pendingPDUData {
		if err := h.WritePDUData(filename, dataList); err != nil {
			h.logger.Error("重新處理PDU數據失敗", zap.String("filename", filename), zap.Error(err))
			continue
		}
		// 標記數據已處理
		if err := h.fallback.MarkPDUDataProcessed(filename); err != nil {
			h.logger.Error("標記PDU數據已處理失敗", zap.String("filename", filename), zap.Error(err))
		}
	}

	// 處理空調機架數據
	for filename, dataList := range pendingDeviceData {
		if err := h.WriteDeviceData(filename, dataList); err != nil {
			h.logger.Error("重新處理空調機架數據失敗", zap.String("filename", filename), zap.Error(err))
			continue
		}
		// 標記數據已處理
		if err := h.fallback.MarkDeviceDataProcessed(filename); err != nil {
			h.logger.Error("標記空調機架數據已處理失敗", zap.String("filename", filename), zap.Error(err))
		}
	}

	return nil
}

// 將PDU數據備份到磁盤
func (h *Handler) fallbackToDisk(filename string, data []models.PDUData) error {
	if h.fallback == nil {
		h.logger.Error("未配置備份策略，無法備份PDU數據", zap.String("filename", filename))
		return fmt.Errorf("未配置備份策略，無法備份PDU數據")
	}

	if err := h.fallback.SavePDUData(filename, data); err != nil {
		h.logger.Error("備份PDU數據失敗", zap.String("filename", filename), zap.Error(err))
		return fmt.Errorf("備份PDU數據失敗: %w", err)
	}

	h.logger.Info("成功備份PDU數據到磁盤", zap.String("filename", filename), zap.Int("count", len(data)))
	return nil
}

// 將空調機架數據備份到磁盤
func (h *Handler) fallbackACRackToDisk(filename string, data []models.DeviceData) error {
	if h.fallback == nil {
		h.logger.Error("未配置備份策略，無法備份空調機架數據", zap.String("filename", filename))
		return fmt.Errorf("未配置備份策略，無法備份空調機架數據")
	}

	if err := h.fallback.SaveDeviceData(filename, data); err != nil {
		h.logger.Error("備份空調機架數據失敗", zap.String("filename", filename), zap.Error(err))
		return fmt.Errorf("備份空調機架數據失敗: %w", err)
	}

	h.logger.Info("成功備份空調機架數據到磁盤", zap.String("filename", filename), zap.Int("count", len(data)))
	return nil
}

// 檢查寫入錯誤
func (h *Handler) checkWriteError() error {
	// 通過InfluxDB client的錯誤通道獲取寫入錯誤
	select {
	case err := <-h.writeAPI.Errors():
		return err
	default:
		return nil
	}
}

// 連接標籤到字符串
func joinTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	result := ""
	for i, tag := range tags {
		if i > 0 {
			result += ","
		}
		result += tag
	}
	return result
}

// 解析標籤字符串
func parseTags(tagsStr string) []string {
	if tagsStr == "" {
		return []string{}
	}
	return []string{tagsStr}
}
