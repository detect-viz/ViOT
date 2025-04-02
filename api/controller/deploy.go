package controller

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"viot/api/response"
	"viot/infra/collector"
	"viot/infra/collector/modbus"
	"viot/infra/collector/snmp"
	"viot/logger"
	"viot/pkg/processor"

	"github.com/gin-gonic/gin"
)

// ScanRequest 表示掃描請求
type ScanRequest struct {
	Scope    []string `json:"scope" binding:"required"`
	Interval int      `json:"interval" binding:"required,min=1"`
	Duration int      `json:"duration" binding:"required,min=1"`
}

// DeployController 處理設備掃描和部署相關的 API 請求
type DeployController struct {
	processorManager *processor.ProcessorManager
	logger           logger.Logger
	scanStatus       processor.ProcessorStatus
	devices          map[string]*collector.DeviceInfo
	mu               sync.RWMutex
	cancelScan       context.CancelFunc
}

// NewDeployController 創建一個新的部署控制器
func NewDeployController(logger logger.Logger) *DeployController {
	// 創建收集器管理器
	collectorManager := collector.NewCollectorManager(logger)

	// 註冊收集器工廠
	collectorManager.RegisterFactory(snmp.NewSNMPCollectorFactory(logger))
	collectorManager.RegisterFactory(modbus.NewModbusCollectorFactory(logger))

	return &DeployController{
		collectorManager: collectorManager,
		logger:           logger.Named("deploy-controller"),
		devices:          make(map[string]*collector.DeviceInfo),
		scanStatus: ScanStatus{
			IsScanning: false,
		},
	}
}

// GetScanStatus 獲取掃描狀態
// @Summary 獲取掃描狀態
// @Description 獲取當前掃描狀態
// @Tags Deploy
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/scan/status [get]
func (c *DeployController) GetScanStatus(ctx *gin.Context) {
	c.mu.RLock()
	status := c.scanStatus
	c.mu.RUnlock()

	response.Success(ctx, "獲取掃描狀態成功", status)
}

// StartScan 開始掃描
// @Summary 開始掃描
// @Description 開始掃描網絡中的設備
// @Tags Deploy
// @Accept json
// @Produce json
// @Param request body ScanRequest true "掃描請求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/scan [post]
func (c *DeployController) StartScan(ctx *gin.Context) {
	var req ScanRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("請求參數無效", logger.Any("error", err))
		response.Fail(ctx, http.StatusBadRequest, "請求參數無效", err.Error())
		return
	}

	c.mu.Lock()
	if c.scanStatus.IsScanning {
		c.mu.Unlock()
		response.Fail(ctx, http.StatusBadRequest, "掃描已在進行中", "")
		return
	}

	// 更新掃描狀態
	now := time.Now()
	c.scanStatus = ScanStatus{
		IsScanning:    true,
		StartTime:     now,
		EndTime:       now.Add(time.Duration(req.Duration) * time.Hour),
		RemainingTime: req.Duration * 60,
		TotalIPs:      0,
		ScannedIPs:    0,
		FoundDevices:  0,
	}
	c.mu.Unlock()

	// 清空設備列表
	c.mu.Lock()
	c.devices = make(map[string]*collector.DeviceInfo)
	c.mu.Unlock()

	// 開始掃描
	scanCtx, cancel := context.WithCancel(context.Background())
	c.cancelScan = cancel

	go c.runScan(scanCtx, req)

	response.Success(ctx, "掃描已開始", c.scanStatus)
}

// StopScan 停止掃描
// @Summary 停止掃描
// @Description 停止當前掃描
// @Tags Deploy
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/scan/stop [post]
func (c *DeployController) StopScan(ctx *gin.Context) {
	c.mu.Lock()
	if !c.scanStatus.IsScanning {
		c.mu.Unlock()
		response.Fail(ctx, http.StatusBadRequest, "沒有正在進行的掃描", "")
		return
	}

	// 停止掃描
	if c.cancelScan != nil {
		c.cancelScan()
	}

	// 更新掃描狀態
	c.scanStatus.IsScanning = false
	c.scanStatus.EndTime = time.Now()
	c.scanStatus.RemainingTime = 0
	c.mu.Unlock()

	response.Success(ctx, "掃描已停止", c.scanStatus)
}

// GetDevices 獲取發現的設備
// @Summary 獲取發現的設備
// @Description 獲取掃描發現的設備
// @Tags Deploy
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/devices [get]
func (c *DeployController) GetDevices(ctx *gin.Context) {
	c.mu.RLock()
	devices := make([]*collector.DeviceInfo, 0, len(c.devices))
	for _, device := range c.devices {
		devices = append(devices, device)
	}
	c.mu.RUnlock()

	response.Success(ctx, "獲取設備列表成功", devices)
}

// DeployDevice 部署設備
// @Summary 部署設備
// @Description 為指定設備部署 Telegraf 配置
// @Tags Deploy
// @Accept json
// @Produce json
// @Param ip path string true "設備 IP"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/devices/{ip}/deploy [post]
func (c *DeployController) DeployDevice(ctx *gin.Context) {
	ip := ctx.Param("ip")

	c.mu.RLock()
	device, ok := c.devices[ip]
	c.mu.RUnlock()

	if !ok {
		response.Fail(ctx, http.StatusNotFound, "設備不存在", "")
		return
	}

	// 這裡應該實現部署邏輯
	// 例如，根據設備型號生成 Telegraf 配置並部署

	c.logger.Info("部署設備",
		logger.String("ip", ip),
		logger.String("model", device.Model),
		logger.String("protocol", device.Protocol))

	// 模擬部署成功
	response.Success(ctx, "設備部署成功", device)
}

// runScan 執行掃描
func (c *DeployController) runScan(ctx context.Context, req ScanRequest) {
	c.logger.Info("開始掃描",
		logger.Any("scope", req.Scope),
		logger.Int("interval", req.Interval),
		logger.Int("duration", req.Duration))

	// 根據掃描範圍生成 IP 列表
	ipList := c.generateIPList(req.Scope)

	c.mu.Lock()
	c.scanStatus.TotalIPs = len(ipList)
	c.mu.Unlock()

	// 創建工作池
	workerCount := 10
	ipChan := make(chan string, len(ipList))
	var wg sync.WaitGroup

	// 啟動工作協程
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range ipChan {
				select {
				case <-ctx.Done():
					return
				default:
					c.scanIP(ctx, ip)

					// 更新掃描狀態
					c.mu.Lock()
					c.scanStatus.ScannedIPs++
					c.mu.Unlock()

					// 等待指定的間隔時間
					time.Sleep(time.Duration(req.Interval) * time.Second)
				}
			}
		}()
	}

	// 發送 IP 到通道
	for _, ip := range ipList {
		select {
		case <-ctx.Done():
			break
		default:
			ipChan <- ip
		}
	}
	close(ipChan)

	// 等待所有工作協程完成
	wg.Wait()

	// 更新掃描狀態
	c.mu.Lock()
	c.scanStatus.IsScanning = false
	c.scanStatus.EndTime = time.Now()
	c.scanStatus.RemainingTime = 0
	c.mu.Unlock()

	c.logger.Info("掃描完成",
		logger.Int("total_ips", c.scanStatus.TotalIPs),
		logger.Int("scanned_ips", c.scanStatus.ScannedIPs),
		logger.Int("found_devices", c.scanStatus.FoundDevices))
}

// scanIP 掃描單個 IP
func (c *DeployController) scanIP(ctx context.Context, ip string) {
	c.logger.Debug("掃描 IP", logger.String("ip", ip))

	// 嘗試使用 SNMP 協議
	snmpConfig := map[string]interface{}{
		"name":       "scan-" + ip,
		"ip_address": ip,
		"port":       161,
		"community":  "public",
		"version":    "2c",
		"timeout":    2, // 2 秒超時
		"retries":    1,
	}

	snmpCollector, err := c.collectorManager.CreateCollector("snmp", snmpConfig)
	if err == nil {
		info, err := snmpCollector.Collect(ctx)
		if err == nil && info != nil {
			c.logger.Info("發現 SNMP 設備",
				logger.String("ip", ip),
				logger.String("model", info.Model),
				logger.String("serial", info.SerialNumber))

			c.mu.Lock()
			c.devices[ip] = info
			c.scanStatus.FoundDevices++
			c.mu.Unlock()
			return
		}
	}

	// 嘗試使用 Modbus 協議
	modbusConfig := map[string]interface{}{
		"name":       "scan-" + ip,
		"ip_address": ip,
		"port":       502, // Modbus TCP 默認端口
		"unit_id":    1,
		"timeout":    2, // 2 秒超時
		"retries":    1,
	}

	modbusCollector, err := c.collectorManager.CreateCollector("modbus", modbusConfig)
	if err == nil {
		info, err := modbusCollector.Collect(ctx)
		if err == nil && info != nil {
			c.logger.Info("發現 Modbus 設備",
				logger.String("ip", ip),
				logger.String("model", info.Model),
				logger.String("serial", info.SerialNumber))

			c.mu.Lock()
			c.devices[ip] = info
			c.scanStatus.FoundDevices++
			c.mu.Unlock()
			return
		}
	}
}

// generateIPList 根據掃描範圍生成 IP 列表
func (c *DeployController) generateIPList(scope []string) []string {
	var ipList []string

	// 這裡應該根據實際情況實現 IP 列表生成邏輯
	// 例如，根據不同的機房或區域生成不同的 IP 範圍

	// 示例：為每個範圍生成一些 IP
	for _, s := range scope {
		switch s {
		case "F12":
			// 假設 F12 區域的 IP 範圍是 10.1.12.0/24
			for i := 1; i <= 10; i++ {
				ipList = append(ipList, "10.1.12."+strconv.Itoa(i))
			}
		case "P7":
			// 假設 P7 區域的 IP 範圍是 10.1.7.0/24
			for i := 1; i <= 10; i++ {
				ipList = append(ipList, "10.1.7."+strconv.Itoa(i))
			}
		case "DC1":
			// 假設 DC1 區域的 IP 範圍是 10.1.1.0/24
			for i := 1; i <= 10; i++ {
				ipList = append(ipList, "10.1.1."+strconv.Itoa(i))
			}
		}
	}

	return ipList
}
