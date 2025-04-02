package manager

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"viot/models"

	"go.uber.org/zap"
)

// WebManager Web 服務管理器
type WebManager struct {
	config        *models.WebConfig
	configManager *ConfigManager
	logger        *zap.Logger
	storage       any // 實際使用適當的存儲接口
	statusCache   *StatusCache
	statusSync    *StatusSyncService
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewWebManager 創建一個新的 Web 服務管理器
func NewWebManager(configManager *ConfigManager, logger *zap.Logger) *WebManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebManager{
		config:        configManager.GetWebConfig(),
		configManager: configManager,
		logger:        logger.Named("web-manager"),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Init 初始化Web服務管理器
func (wm *WebManager) Init() error {
	// 初始化狀態緩存
	wm.statusCache = NewStatusCache()

	// 初始化狀態同步服務
	wm.statusSync = NewStatusSyncService(wm.config.Status, wm.logger)

	// 啟動狀態同步服務
	if err := wm.statusSync.Start(wm.ctx); err != nil {
		return err
	}

	wm.logger.Info("Web管理器初始化完成")
	return nil
}

// Stop 停止Web服務管理器
func (wm *WebManager) Stop() {
	if wm.statusSync != nil {
		wm.statusSync.Stop()
	}
	wm.cancel()
	wm.logger.Info("Web管理器已停止")
}

// GetPDUData 獲取 PDU 數據
func (wm *WebManager) GetPDUData(room string) ([]common.PDUData, error) {
	wm.logger.Debug("獲取 PDU 數據", zap.String("room", room))

	if wm.statusCache == nil {
		return []common.PDUData{}, nil
	}

	// 從狀態緩存獲取PDU數據
	statuses := wm.statusCache.GetAllPDUStatuses()
	result := make([]common.PDUData, 0)

	// 轉換為PDUData格式
	for name, status := range statuses {
		// 過濾不屬於指定房間的PDU
		// 這裡需要根據實際情況實現房間過濾邏輯

		pduData := common.PDUData{
			Name:   name,
			Status: strconv.FormatBool(status.Status),
			// 其他字段根據實際情況填充
			Tags: map[string]string{
				"ip": status.IP,
			},
		}
		result = append(result, pduData)
	}

	return result, nil
}

// UpdatePDUStatus 更新 PDU 狀態
func (wm *WebManager) UpdatePDUStatus(pduName, status string) error {
	wm.logger.Debug("更新 PDU 狀態", zap.String("pdu", pduName), zap.String("status", status))

	// 實現更新 PDU 狀態的邏輯
	return nil
}

// UpdatePDUTags 更新 PDU 標籤
func (wm *WebManager) UpdatePDUTags(pduName string, tags []string) error {
	wm.logger.Debug("更新 PDU 標籤", zap.String("pdu", pduName), zap.Any("tags", tags))

	// 實現更新 PDU 標籤的邏輯
	return nil
}

// GetPDUTags 獲取 PDU 標籤
func (wm *WebManager) GetPDUTags(pduName string) ([]string, error) {
	wm.logger.Debug("獲取 PDU 標籤", zap.String("pdu", pduName))

	// 實現獲取 PDU 標籤的邏輯
	return []string{}, nil
}

// GetRooms 獲取所有房間
func (wm *WebManager) GetRooms() ([]string, error) {
	wm.logger.Debug("獲取所有房間")

	// 實現獲取所有房間的邏輯
	return []string{}, nil
}

// GetRoomsByDatacenter 獲取指定數據中心的所有房間
func (wm *WebManager) GetRoomsByDatacenter(datacenter string) ([]string, error) {
	wm.logger.Debug("獲取指定數據中心的所有房間", zap.String("datacenter", datacenter))

	// 實現獲取指定數據中心所有房間的邏輯
	return []string{}, nil
}

// GetDatacenters 獲取所有數據中心
func (wm *WebManager) GetDatacenters() ([]string, error) {
	wm.logger.Debug("獲取所有數據中心")

	// 實現獲取所有數據中心的邏輯
	return []string{}, nil
}

// GetPhases 獲取所有階段
func (wm *WebManager) GetPhases() ([]string, error) {
	wm.logger.Debug("獲取所有階段")

	// 實現獲取所有階段的邏輯
	return []string{}, nil
}

// GetFactories 獲取所有工廠
func (wm *WebManager) GetFactories() ([]string, error) {
	wm.logger.Debug("獲取所有工廠")

	// 實現獲取所有工廠的邏輯
	return []string{}, nil
}

// GetRacksByRoom 獲取指定房間的所有機架
func (wm *WebManager) GetRacksByRoom(factory, phase, datacenter, room string) ([]common.PDUData, error) {
	wm.logger.Debug("獲取指定房間的所有機架",
		zap.String("factory", factory),
		zap.String("phase", phase),
		zap.String("datacenter", datacenter),
		zap.String("room", room))

	// 實現獲取指定房間所有機架的邏輯
	return []common.PDUData{}, nil
}

// ForceStatusSync 強制同步狀態
func (wm *WebManager) ForceStatusSync() error {
	wm.logger.Debug("強制同步狀態")
	if wm.statusSync == nil {
		return nil
	}
	return wm.statusSync.ForceSync()
}

// StatusCache 設備狀態緩存，支持並發訪問
type StatusCache struct {
	mu             sync.RWMutex
	pduStatuses    map[string]common.StatusData
	deviceStatuses map[string]common.StatusData
	lastUpdated    time.Time
}

// NewStatusCache 創建新的設備狀態緩存
func NewStatusCache() *StatusCache {
	return &StatusCache{
		mu:             sync.RWMutex{},
		pduStatuses:    make(map[string]common.StatusData),
		deviceStatuses: make(map[string]common.StatusData),
		lastUpdated:    time.Now(),
	}
}

// LoadPDUStatuses 從CSV文件加載PDU狀態
func (c *StatusCache) LoadPDUStatuses(filepath string) error {
	statuses, err := LoadStatusesFromCSV(filepath)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.pduStatuses = statuses
	c.lastUpdated = time.Now()
	return nil
}

// GetPDUStatus 獲取PDU狀態
func (c *StatusCache) GetPDUStatus(name string) (common.StatusData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status, exists := c.pduStatuses[name]
	return status, exists
}

// GetAllPDUStatuses 獲取所有PDU狀態
func (c *StatusCache) GetAllPDUStatuses() map[string]common.StatusData {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 創建一個副本避免並發問題
	statuses := make(map[string]common.StatusData, len(c.pduStatuses))
	for k, v := range c.pduStatuses {
		statuses[k] = v
	}
	return statuses
}

// GetLastUpdated 獲取最後更新時間
func (c *StatusCache) GetLastUpdated() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastUpdated
}

// StatusSyncService 設備狀態同步服務
type StatusSyncService struct {
	config      models.WebserviceConfig
	cache       *StatusCache
	logger      *zap.Logger
	stopCh      chan struct{}
	wg          sync.WaitGroup
	initialized bool
	mu          sync.Mutex
}

// NewStatusSyncService 創建新的狀態同步服務
func NewStatusSyncService(config models.WebserviceConfig, logger *zap.Logger) *StatusSyncService {
	return &StatusSyncService{
		config:      config,
		cache:       NewStatusCache(),
		logger:      logger.Named("status-sync"),
		stopCh:      make(chan struct{}),
		initialized: false,
	}
}

// Start 啟動狀態同步服務
func (s *StatusSyncService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 避免重複啟動
	if s.initialized {
		return nil
	}

	// 立即執行一次同步
	if err := s.syncStatuses(); err != nil {
		s.logger.Error("初始同步狀態失敗", zap.Error(err))
		return err
	}

	s.wg.Add(1)
	go s.runSync(ctx)
	s.initialized = true

	s.logger.Info("狀態同步服務已啟動",
		zap.String("pdu_file", s.config.PDUStatusPath),
		zap.String("device_file", s.config.DeviceStatusPath),
		zap.Duration("interval", s.config.SyncInterval))

	return nil
}

// Stop 停止狀態同步服務
func (s *StatusSyncService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		return
	}

	close(s.stopCh)
	s.wg.Wait()
	s.initialized = false
	s.logger.Info("狀態同步服務已停止")
}

// GetCache 獲取設備狀態緩存
func (s *StatusSyncService) GetCache() *StatusCache {
	return s.cache
}

// 執行同步循環
func (s *StatusSyncService) runSync(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.syncStatuses(); err != nil {
				s.logger.Error("同步狀態失敗", zap.Error(err))
			}
		case <-ctx.Done():
			s.logger.Info("同步服務停止 (ctx取消)")
			return
		case <-s.stopCh:
			s.logger.Info("同步服務停止 (主動停止)")
			return
		}
	}
}

// 同步PDU和設備狀態
func (s *StatusSyncService) syncStatuses() error {
	// 使用絕對路徑處理文件路徑
	pduFilePath := s.config.PDUStatusPath
	deviceFilePath := s.config.DeviceStatusPath

	// 如果是相對路徑，則轉換為絕對路徑
	if !filepath.IsAbs(pduFilePath) {
		var err error
		pduFilePath, err = filepath.Abs(pduFilePath)
		if err != nil {
			s.logger.Error("無法獲取PDU狀態文件的絕對路徑", zap.Error(err))
		}
	}

	if !filepath.IsAbs(deviceFilePath) {
		var err error
		deviceFilePath, err = filepath.Abs(deviceFilePath)
		if err != nil {
			s.logger.Error("無法獲取設備狀態文件的絕對路徑", zap.Error(err))
		}
	}

	// 同步PDU狀態
	if err := s.cache.LoadPDUStatuses(pduFilePath); err != nil {
		s.logger.Error("載入PDU狀態失敗", zap.Error(err), zap.String("file", pduFilePath))
	} else {
		s.logger.Debug("PDU狀態已同步", zap.String("file", pduFilePath),
			zap.Int("count", len(s.cache.GetAllPDUStatuses())))
	}

	// 同步設備狀態
	if err := s.cache.LoadPDUStatuses(deviceFilePath); err != nil {
		s.logger.Error("載入設備狀態失敗", zap.Error(err), zap.String("file", deviceFilePath))
	} else {
		s.logger.Debug("設備狀態已同步", zap.String("file", deviceFilePath),
			zap.Int("count", len(s.cache.GetAllPDUStatuses())))
	}

	return nil
}

// ForceSync 強制立即同步狀態
func (s *StatusSyncService) ForceSync() error {
	return s.syncStatuses()
}

// LoadStatusesFromCSV 從CSV文件加載狀態數據
func LoadStatusesFromCSV(filepath string) (map[string]common.StatusData, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// 讀取標題行
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	// 檢查標題行是否符合預期
	expectedHeaders := []string{"name", "ip", "monitored", "status", "uptime", "update_at"}
	if len(header) < len(expectedHeaders) {
		return nil, fmt.Errorf("CSV標題格式不正確，預期: %v, 實際: %v", expectedHeaders, header)
	}

	statuses := make(map[string]common.StatusData)

	// 讀取數據行
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(record) < 6 {
			continue // 跳過格式不正確的行
		}

		// 解析boolean字段
		monitored, _ := strconv.ParseBool(record[2])
		status, _ := strconv.ParseBool(record[3])
		uptime, _ := strconv.ParseInt(record[4], 10, 64)

		// 解析時間字段
		updatedAt, err := time.Parse(time.RFC3339, record[5])
		if err != nil {
			// 嘗試其他常見時間格式
			updatedAt, err = time.Parse("2006-01-02T15:04:05Z", record[5])
			if err != nil {
				updatedAt = time.Now() // 如果無法解析，使用當前時間
			}
		}

		// 創建設備狀態
		statusData := common.StatusData{
			Key:       record[0],
			IP:        record[1],
			Monitored: monitored,
			Status:    status,
			Uptime:    uptime,
			UpdatedAt: updatedAt,
		}

		// 以設備名稱為鍵存儲
		statuses[record[0]] = statusData
	}

	return statuses, nil
}

// GetPDUStatusByName 獲取指定PDU的狀態
func (wm *WebManager) GetPDUStatusByName(name string) (*common.PDUData, error) {
	wm.logger.Debug("獲取PDU狀態", zap.String("name", name))

	if wm.statusCache == nil {
		return nil, fmt.Errorf("狀態緩存未初始化")
	}

	// 獲取PDU狀態
	status, exists := wm.statusCache.GetPDUStatus(name)
	if !exists {
		return nil, fmt.Errorf("找不到名為 %s 的PDU", name)
	}

	// 創建PDUData
	pduData := &common.PDUData{
		Name:   name,
		Status: strconv.FormatBool(status.Status),
		Tags: map[string]string{
			"ip": status.IP,
		},
	}

	return pduData, nil
}
