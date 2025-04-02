package scanner

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"viot/logger"
	"viot/models"
	"viot/pkg/collector/inputs"
	"viot/pkg/deployer"

	"go.uber.org/zap"
)

// scannerImpl 掃描器實現
type scannerImpl struct {
	config     *models.ScannerConfig
	deployer   *deployer.Deployer
	logger     logger.Logger
	collectors map[string]models.Collector
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewScanner 創建掃描器
func NewScanner(config *models.ScannerConfig, deployer *deployer.Deployer, logger logger.Logger) Scanner {
	return &scannerImpl{
		config:     config,
		deployer:   deployer,
		logger:     logger,
		collectors: make(map[string]models.Collector),
	}
}

// Start 開始掃描
func (s *scannerImpl) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	// 載入IP範圍
	ranges, err := s.loadIPRanges()
	if err != nil {
		return err
	}

	// 創建任務通道
	tasks := make(chan string, len(ranges))

	// 啟動工作協程
	for i := 0; i < s.config.MaxConcurrentScan; i++ {
		go s.scanWorker(s.ctx, tasks)
	}

	// 發送掃描任務
	for _, ipRange := range ranges {
		select {
		case <-s.ctx.Done():
			return nil
		case tasks <- fmt.Sprintf("%s-%s", ipRange.StartIP, ipRange.EndIP):
		}
	}

	close(tasks)
	return nil
}

// Stop 停止掃描
func (s *scannerImpl) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

// Scan 掃描特定IP和端口
func (s *scannerImpl) Scan(ctx context.Context, ip string, port int) (*models.DeviceInfo, error) {
	// 檢查SNMP端口
	if port == 161 {
		return s.scanSNMP(ctx, ip)
	}

	// 檢查Modbus端口
	if port == 502 {
		return s.scanModbus(ctx, ip)
	}

	return nil, fmt.Errorf("不支持的端口: %d", port)
}

// loadIPRanges 載入IP範圍配置
func (s *scannerImpl) loadIPRanges() ([]models.IPRange, error) {
	// TODO: 實現從CSV文件讀取IP範圍配置
	return nil, nil
}

// mergeIPRanges 合併IP範圍
func (s *scannerImpl) mergeIPRanges(ranges []models.IPRange) []string {
	// TODO: 實現IP範圍合併邏輯
	return nil
}

// scanWorker 掃描工作協程
func (s *scannerImpl) scanWorker(ctx context.Context, tasks <-chan string) {
	for {
		select {
		case <-ctx.Done():
			return
		case ipRange, ok := <-tasks:
			if !ok {
				return
			}
			s.scanRange(ctx, ipRange)
		}
	}
}

// scanRange 掃描IP範圍
func (s *scannerImpl) scanRange(ctx context.Context, ipRange string) {
	// 使用fping掃描
	aliveIPs := s.fpingScan(ipRange)

	// 對每個存活的IP進行協議掃描
	for _, ip := range aliveIPs {
		select {
		case <-ctx.Done():
			return
		default:
			s.scanIP(ctx, ip)
		}
	}
}

// fpingScan 使用fping進行掃描
func (s *scannerImpl) fpingScan(ipRange string) []string {
	cmd := exec.Command("fping", "-a", "-g", ipRange)
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.Error("fping掃描失敗", zap.Error(err))
		return nil
	}

	// 解析輸出
	aliveIPs := strings.Split(strings.TrimSpace(string(output)), "\n")
	return aliveIPs
}

// scanIP 掃描單個IP
func (s *scannerImpl) scanIP(ctx context.Context, ip string) {
	// 檢查SNMP端口
	if s.checkPorts(ip, s.config.Protocol.SNMP.Ports) {
		s.scanSNMP(ctx, ip)
	}

	// 檢查Modbus端口
	if s.checkPorts(ip, s.config.Protocol.Modbus.Ports) {
		s.scanModbus(ctx, ip)
	}
}

// checkPorts 檢查端口是否開放
func (s *scannerImpl) checkPorts(ip string, ports []int) bool {
	// TODO: 實現端口檢查邏輯
	return false
}

// scanSNMP 掃描SNMP設備
func (s *scannerImpl) scanSNMP(ctx context.Context, ip string) (*models.DeviceInfo, error) {
	// 獲取SNMP收集器
	collector := s.getCollector("snmp")
	if collector == nil {
		return nil, fmt.Errorf("無法獲取SNMP收集器")
	}

	// 掃描每個設備配置
	for _, device := range s.config.Devices {
		if device.Protocol != "snmp" {
			continue
		}

		// 收集設備信息
		deviceData, err := collector.Collect(ctx, models.Device{
			IP:           ip,
			Type:         device.Type,
			Manufacturer: device.Manufacturer,
			Model:        device.Model,
		})
		if err != nil {
			s.logger.Error("SNMP數據採集失敗",
				zap.String("ip", ip),
				zap.String("device", device.Name),
				zap.Error(err))
			continue
		}

		// 檢查是否匹配
		if len(deviceData) > 0 && s.matchDeviceData(device, deviceData[0]) {
			// 獲取設備位置信息
			zone := s.findZoneForIP(ip)
			if zone == nil {
				return nil, fmt.Errorf("無法找到IP對應的區域信息")
			}

			return &models.DeviceInfo{
				IP:           ip,
				Factory:      zone.Factory,
				Phase:        zone.Phase,
				Datacenter:   zone.Datacenter,
				Room:         zone.Room,
				Manufacturer: device.Manufacturer,
				Model:        device.Model,
				Type:         device.Type,
			}, nil
		}
	}

	return nil, fmt.Errorf("未找到匹配的SNMP設備")
}

// scanModbus 掃描Modbus設備
func (s *scannerImpl) scanModbus(ctx context.Context, ip string) (*models.DeviceInfo, error) {
	// 獲取Modbus收集器
	collector := s.getCollector("modbus")
	if collector == nil {
		return nil, fmt.Errorf("無法獲取Modbus收集器")
	}

	// 掃描每個設備配置
	for _, device := range s.config.Devices {
		if device.Protocol != "modbus" {
			continue
		}

		// 掃描每個slave ID
		for _, slaveID := range s.config.Protocol.Modbus.SlaveID {
			// 收集設備信息
			deviceData, err := collector.Collect(ctx, models.Device{
				IP:           ip,
				Type:         device.Type,
				Manufacturer: device.Manufacturer,
				Model:        device.Model,
				Tags: map[string]string{
					"slave_id": fmt.Sprintf("%d", slaveID),
				},
			})
			if err != nil {
				s.logger.Error("Modbus數據採集失敗",
					zap.String("ip", ip),
					zap.String("device", device.Name),
					zap.Int("slave_id", slaveID),
					zap.Error(err))
				continue
			}

			// 檢查是否匹配
			if len(deviceData) > 0 && s.matchDeviceData(device, deviceData[0]) {
				// 獲取設備位置信息
				zone := s.findZoneForIP(ip)
				if zone == nil {
					return nil, fmt.Errorf("無法找到IP對應的區域信息")
				}

				return &models.DeviceInfo{
					IP:           ip,
					Factory:      zone.Factory,
					Phase:        zone.Phase,
					Datacenter:   zone.Datacenter,
					Room:         zone.Room,
					Manufacturer: device.Manufacturer,
					Model:        device.Model,
					Type:         device.Type,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("未找到匹配的Modbus設備")
}

// getCollector 獲取收集器
func (s *scannerImpl) getCollector(protocol string) models.Collector {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if c, ok := s.collectors[protocol]; ok {
		return c
	}

	// 創建新的收集器
	var c models.Collector
	switch protocol {
	case "snmp":
		// 從配置中獲取SNMP參數
		snmpConfig := s.config.Protocol.SNMP
		if len(snmpConfig.Community) == 0 {
			s.logger.Error("SNMP community 未配置")
			return nil
		}

		snmpCollector, err := inputs.NewSNMPCollector(
			&models.SNMPCollectorConfig{
				Community: snmpConfig.Community,
			},
		)
		if err != nil {
			s.logger.Error("創建SNMP收集器失敗", zap.Error(err))
			return nil
		}
		c = snmpCollector

	case "modbus":
		// 從配置中獲取Modbus參數
		modbusConfig := s.config.Protocol.Modbus
		if len(modbusConfig.Ports) == 0 {
			s.logger.Error("Modbus端口未配置")
			return nil
		}

		modbusCollector, err := inputs.NewModbusCollector(
			&models.ModbusCollectorConfig{
				Port:    modbusConfig.Ports[0],         // 使用第一個配置的端口
				SlaveID: byte(modbusConfig.SlaveID[0]), // 使用第一個從站ID
				Timeout: time.Duration(modbusConfig.Timeout) * time.Second,
				Retries: modbusConfig.Retries,
			},
			s.config.CollectorHost,
		)
		if err != nil {
			s.logger.Error("創建Modbus收集器失敗", zap.Error(err))
			return nil
		}
		c = modbusCollector
	}

	if c != nil {
		s.mu.Lock()
		s.collectors[protocol] = c
		s.mu.Unlock()
	}

	return c
}

// matchDeviceData 檢查設備數據是否匹配
func (s *scannerImpl) matchDeviceData(device models.DeviceConfig, data models.DeviceData) bool {
	value, ok := data.Fields[device.MatchField]
	if !ok {
		return false
	}
	if str, ok := value.(string); ok {
		return strings.Contains(str, device.MatchKeyword)
	}
	return false
}

// handleMatchedDevice 處理匹配的設備
func (s *scannerImpl) handleMatchedDevice(device models.DeviceConfig, ip string, data map[string]string) {
	// 獲取MAC地址
	mac := ""
	if device.Protocol == "snmp" {
		// TODO: 實現MAC地址獲取
	}

	// 創建掃描結果
	result := &models.ScanResult{
		IP:            ip,
		Protocol:      device.Protocol,
		DeviceConfig:  &device,
		CollectedData: data,
		MAC:           mac,
		ScanTime:      time.Now(),
	}

	// 根據auto_deploy決定處理方式
	if device.AutoDeploy {
		s.handleAutoDeploy(result)
	} else {
		s.handlePending(result)
	}
}

// handleAutoDeploy 處理自動部署
func (s *scannerImpl) handleAutoDeploy(result *models.ScanResult) {
	// 生成設備名稱
	name := fmt.Sprintf("%s_%s_%s", result.DeviceConfig.Type, result.DeviceConfig.Manufacturer, result.IP)

	// 生成部署配置
	config := &models.DeployConfig{
		VerifyAfterDeploy:  true,
		RollbackOnFailure:  true,
		ConcurrentDeploys:  1,
		DeploymentTimeout:  "5m",
		RestartServices:    true,
		BackupBeforeDeploy: true,
		NotifyOnDeploy:     true,
		Telegraf: models.TelegrafConfig{
			ConfigPath:     "conf/telegraf/telegraf.conf",
			DeployPath:     "conf/telegraf/telegraf.d",
			TemplatesPath:  "conf/settings/depoly",
			BackupPath:     "conf/telegraf/backups",
			BinaryPath:     "bin/telegraf",
			VerifyCommand:  "telegraf --test --config",
			RestartCommand: "systemctl restart telegraf",
		},
	}

	// 部署設備
	if err := s.deployer.Deploy(name, config); err != nil {
		s.logger.Error("設備部署失敗",
			zap.String("name", name),
			zap.Error(err))
		return
	}

	// 記錄到註冊表
	s.saveToRegistry(result)
}

// handlePending 處理待部署設備
func (s *scannerImpl) handlePending(result *models.ScanResult) {
	// 記錄到待處理列表
	s.saveToPending(result)
}

// saveToRegistry 保存到註冊表
func (s *scannerImpl) saveToRegistry(result *models.ScanResult) {
	// TODO: 實現註冊表保存邏輯
}

// saveToPending 保存到待處理列表
func (s *scannerImpl) saveToPending(result *models.ScanResult) {
	// TODO: 實現待處理列表保存邏輯
}

// findZoneForIP 查找IP所屬的區域
func (s *scannerImpl) findZoneForIP(ip string) *models.IPRangeConfig {
	ipObj := net.ParseIP(ip)
	if ipObj == nil {
		return nil
	}

	for _, zone := range s.config.IPRanges {
		startIP := net.ParseIP(zone.StartIP)
		endIP := net.ParseIP(zone.EndIP)

		if startIP == nil || endIP == nil {
			continue
		}

		if bytes4ToUint32(ipObj) >= bytes4ToUint32(startIP) &&
			bytes4ToUint32(ipObj) <= bytes4ToUint32(endIP) {
			return &zone
		}
	}

	return nil
}
