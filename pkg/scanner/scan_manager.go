package scanner

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	"viot/models"
)

// scanManagerImpl 實現掃描管理器介面
type scanManagerImpl struct {
	logger        *zap.Logger
	scanners      map[string]Scanner
	config        *models.ScannerConfig
	ipZoneMap     map[string]*models.IPRangeConfig
	statusLock    sync.RWMutex
	status        models.ScanStatus
	ctx           context.Context
	cancelFunc    context.CancelFunc
	resultHandler func(models.DeviceInfo)
}

// NewScanManager 創建新的掃描管理器
func NewScanManager(logger *zap.Logger, config *models.ScannerConfig) ScanManager {
	return &scanManagerImpl{
		logger:    logger.Named("scan-manager"),
		scanners:  make(map[string]Scanner),
		config:    config,
		ipZoneMap: make(map[string]*models.IPRangeConfig),
	}
}

// Init 初始化掃描管理器
func (s *scanManagerImpl) Init() error {
	// 加載IP區域映射
	if s.config != nil && s.config.IPRanges != nil {
		for i, zone := range s.config.IPRanges {
			key := fmt.Sprintf("%s-%s", zone.StartIP, zone.EndIP)
			// 創建指針的副本，避免循環迭代問題
			zoneCopy := s.config.IPRanges[i]
			s.ipZoneMap[key] = &zoneCopy
		}
	}

	// 註冊默認掃描器
	s.RegisterScanner("snmp", NewSNMPScanner(s.logger, s.config))
	s.RegisterScanner("modbus", NewModbusScanner(s.logger, s.config))

	return nil
}

// RegisterScanner 註冊掃描器
func (s *scanManagerImpl) RegisterScanner(protocol string, scanner Scanner) {
	s.scanners[protocol] = scanner
}

// StartScan 開始掃描指定的 IP 範圍
func (s *scanManagerImpl) StartScan(config *models.ScanSetting) error {
	s.statusLock.Lock()
	defer s.statusLock.Unlock()

	if s.status.Running {
		return fmt.Errorf("已有掃描任務正在運行")
	}

	// 設置默認值
	if len(config.SNMPPorts) == 0 {
		config.SNMPPorts = []int{161}
	}
	if len(config.ModbusPorts) == 0 {
		config.ModbusPorts = []int{502}
	}
	if config.Interval <= 0 {
		config.Interval = 10 // 默認10秒
	}
	if config.Duration <= 0 {
		config.Duration = 60 // 默認60分鐘
	}

	// 驗證至少有一個IP範圍
	if len(config.IPRanges) == 0 {
		return fmt.Errorf("沒有指定IP範圍")
	}

	// 創建新的掃描上下文
	s.ctx, s.cancelFunc = context.WithTimeout(
		context.Background(),
		time.Duration(config.Duration)*time.Minute,
	)

	// 初始化掃描狀態 - 實際掃描總數將在快速掃描後更新
	s.status = models.ScanStatus{
		Running:       true,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(time.Duration(config.Duration) * time.Minute),
		TotalIPs:      0, // 將在快速掃描後更新
		ScannedIPs:    0,
		FoundDevices:  0,
		Progress:      0,
		RemainingTime: config.Duration * 60, // 轉換為秒
	}

	// 啟動掃描協程
	go s.runScan(s.ctx, config)

	return nil
}

// StopScan 停止當前掃描
func (s *scanManagerImpl) StopScan() error {
	s.statusLock.Lock()
	defer s.statusLock.Unlock()

	if !s.status.Running {
		return fmt.Errorf("當前沒有掃描任務運行")
	}

	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	s.status.Running = false
	s.status.EndTime = time.Now()

	return nil
}

// GetScanStatus 獲取當前掃描狀態
func (s *scanManagerImpl) GetScanStatus() models.ScanStatus {
	s.statusLock.RLock()
	defer s.statusLock.RUnlock()

	modelStatus := models.ScanStatus{
		Running:       s.status.Running,
		StartTime:     s.status.StartTime,
		EndTime:       s.status.EndTime,
		TotalIPs:      s.status.TotalIPs,
		ScannedIPs:    s.status.ScannedIPs,
		FoundDevices:  s.status.FoundDevices,
		Progress:      s.status.Progress,
		RemainingTime: s.status.RemainingTime,
	}

	return modelStatus
}

// SetResultHandler 設置掃描結果處理器
func (s *scanManagerImpl) SetResultHandler(handler func(models.DeviceInfo)) {
	s.resultHandler = handler
}

// runScan 在後台運行掃描任務
func (s *scanManagerImpl) runScan(ctx context.Context, config *models.ScanSetting) {
	s.logger.Info("開始掃描",
		zap.Strings("ipRanges", config.IPRanges),
		zap.Int("duration", config.Duration))

	var scannedCount int
	var foundCount int

	// 使用fping快速檢查活躍的IP
	aliveIPs := []string{}
	for _, ipRange := range config.IPRanges {
		// 以500 IP為一組進行快速掃描
		timeout := 1 * time.Second // 設置fping超時時間
		ips, err := FastPingScan(ctx, s.logger, []string{ipRange}, timeout)
		if err != nil {
			s.logger.Error("快速掃描失敗",
				zap.String("range", ipRange),
				zap.Error(err))
			continue
		}
		aliveIPs = append(aliveIPs, ips...)
	}

	// 更新掃描狀態的總IP數
	s.statusLock.Lock()
	s.status.TotalIPs = len(aliveIPs)
	s.statusLock.Unlock()

	s.logger.Info("完成快速IP掃描",
		zap.Int("aliveIPs", len(aliveIPs)))

	// 對每個活躍的IP進行設備識別
	for _, ip := range aliveIPs {
		select {
		case <-ctx.Done():
			s.logger.Info("掃描已停止")
			s.updateScanStatus(scannedCount, foundCount, true)
			return
		default:
			// 識別設備
			deviceInfo, err := s.identifyDevice(ip, config)
			if err != nil {
				s.logger.Debug("識別設備失敗",
					zap.String("ip", ip),
					zap.Error(err))
			} else {
				// 設備識別成功
				s.logger.Info("發現設備",
					zap.String("ip", ip),
					zap.String("type", deviceInfo.Type),
					zap.String("manufacturer", deviceInfo.Manufacturer),
					zap.String("model", deviceInfo.Model))

				// 處理掃描結果
				if s.resultHandler != nil {
					s.resultHandler(*deviceInfo)
				}

				foundCount++
			}

			scannedCount++
			s.updateScanStatus(scannedCount, foundCount, false)

			// 等待間隔時間
			select {
			case <-ctx.Done():
				s.logger.Info("掃描已停止")
				s.updateScanStatus(scannedCount, foundCount, true)
				return
			case <-time.After(time.Duration(config.Interval) * time.Second):
				// 繼續下一個IP
			}
		}
	}

	// 掃描完成
	s.updateScanStatus(scannedCount, foundCount, true)
	s.logger.Info("掃描完成",
		zap.Int("scannedIPs", scannedCount),
		zap.Int("foundDevices", foundCount))
}

// updateScanStatus 更新掃描狀態
func (s *scanManagerImpl) updateScanStatus(scannedCount, foundCount int, completed bool) {
	s.statusLock.Lock()
	defer s.statusLock.Unlock()

	s.status.ScannedIPs = scannedCount
	s.status.FoundDevices = foundCount

	if s.status.TotalIPs > 0 {
		s.status.Progress = float64(scannedCount) / float64(s.status.TotalIPs) * 100
	}

	if completed {
		s.status.Running = false
		s.status.EndTime = time.Now()
		s.status.RemainingTime = 0
		s.status.Progress = 100
	}
}

// findZoneForIP 查找IP所屬的區域
func (s *scanManagerImpl) findZoneForIP(ip string) *models.IPRangeConfig {
	ipObj := net.ParseIP(ip)
	if ipObj == nil {
		return nil
	}

	for _, zone := range s.ipZoneMap {
		startIP := net.ParseIP(zone.StartIP)
		endIP := net.ParseIP(zone.EndIP)

		if startIP == nil || endIP == nil {
			continue
		}

		if bytes4ToUint32(ipObj) >= bytes4ToUint32(startIP) &&
			bytes4ToUint32(ipObj) <= bytes4ToUint32(endIP) {
			return zone
		}
	}

	return nil
}
