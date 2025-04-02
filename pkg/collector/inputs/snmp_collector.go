package inputs

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"viot/models"

	"github.com/gosnmp/gosnmp"
)

// SNMPCollector SNMP 收集器
type SNMPCollector struct {
	config *models.SNMPConfig // 改用 SNMPConfig 而不是 SNMPCollectorConfig
	client *gosnmp.GoSNMP
	status models.CollectorStatus
}

// NewSNMPCollector 創建新的 SNMP 收集器
func NewSNMPCollector(collectorConfig *models.SNMPCollectorConfig) (*SNMPCollector, error) {
	// 使用 MergeWithDefault 獲取完整配置
	config := collectorConfig.MergeWithDefault()

	// 設置 MIB 路徑
	if config.MIBPath != "" {
		if err := os.Setenv("MIBDIRS", config.MIBPath); err != nil {
			return nil, fmt.Errorf("failed to set MIBDIRS: %v", err)
		}
		if err := os.Setenv("MIBS", "ALL"); err != nil {
			return nil, fmt.Errorf("failed to set MIBS: %v", err)
		}
	}

	// 將版本字符串轉換為 gosnmp.SnmpVersion
	var snmpVersion gosnmp.SnmpVersion
	switch config.Version {
	case "v1", "1":
		snmpVersion = gosnmp.Version1
	case "v3", "3":
		snmpVersion = gosnmp.Version3
	default:
		snmpVersion = gosnmp.Version2c
	}

	client := &gosnmp.GoSNMP{
		Target:         config.Host,
		Port:           uint16(config.Port),
		Community:      config.Community,
		Version:        snmpVersion,
		Timeout:        config.Timeout,
		Retries:        config.Retries,
		MaxRepetitions: uint32(config.MaxRepetitions),
		ContextName:    config.ContextName,
	}

	// 配置 SNMPv3
	if snmpVersion == gosnmp.Version3 {
		// 設置安全級別
		var msgFlags gosnmp.SnmpV3MsgFlags
		switch config.SecLevel {
		case "authPriv":
			msgFlags = gosnmp.AuthPriv
		case "authNoPriv":
			msgFlags = gosnmp.AuthNoPriv
		default:
			msgFlags = gosnmp.NoAuthNoPriv
		}

		// 設置認證協議
		var authProtocol gosnmp.SnmpV3AuthProtocol
		switch config.AuthProtocol {
		case "SHA":
			authProtocol = gosnmp.SHA
		case "SHA224":
			authProtocol = gosnmp.SHA224
		case "SHA256":
			authProtocol = gosnmp.SHA256
		case "SHA384":
			authProtocol = gosnmp.SHA384
		case "SHA512":
			authProtocol = gosnmp.SHA512
		default:
			authProtocol = gosnmp.MD5
		}

		// 設置加密協議
		var privProtocol gosnmp.SnmpV3PrivProtocol
		switch config.PrivProtocol {
		case "AES":
			privProtocol = gosnmp.AES
		case "AES192":
			privProtocol = gosnmp.AES192
		case "AES256":
			privProtocol = gosnmp.AES256
		default:
			privProtocol = gosnmp.DES
		}

		// 配置 SNMPv3 參數
		client.SecurityModel = gosnmp.UserSecurityModel
		client.MsgFlags = msgFlags
		client.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 config.SecName,
			AuthenticationProtocol:   authProtocol,
			AuthenticationPassphrase: config.AuthPassword,
			PrivacyProtocol:          privProtocol,
			PrivacyPassphrase:        config.PrivPassword,
		}
	}

	collector := &SNMPCollector{
		config: &config,
		client: client,
		status: models.CollectorStatus{
			Running: false,
		},
	}

	return collector, nil
}

// connect 建立 SNMP 連接
func (c *SNMPCollector) connect() error {
	if c.client == nil {
		return fmt.Errorf("SNMP client is not initialized")
	}

	// 如果已經連接,先斷開
	if c.client.Conn != nil {
		c.client.Conn.Close()
	}

	return c.client.Connect()
}

// Status 返回收集器狀態
func (c *SNMPCollector) Status() models.CollectorStatus {
	return c.status
}

// Start 啟動收集器
func (c *SNMPCollector) Start(ctx context.Context) error {
	c.status.Running = true
	return nil
}

// Stop 停止收集器
func (c *SNMPCollector) Stop() error {
	if c.client != nil {
		c.client.Conn.Close()
	}
	c.status.Running = false
	return nil
}

// Name 返回收集器名稱
func (c *SNMPCollector) Name() string {
	return "snmp"
}

// Collect 收集設備數據
func (c *SNMPCollector) Collect(ctx context.Context, device models.Device) ([]models.DeviceData, error) {
	// 確保連接
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	defer c.client.Conn.Close()

	// 收集系統運行時間
	uptime, err := c.getSysUpTime()
	if err != nil {
		return nil, fmt.Errorf("failed to get system uptime: %v", err)
	}

	// 收集設備元數據
	model, version, serialNumber, err := c.getDeviceMetadata(device)
	if err != nil {
		return nil, fmt.Errorf("failed to get device metadata: %v", err)
	}

	// 組合所有數據
	data := []models.DeviceData{
		{
			Name:      device.IP,
			Metric:    "device_info",
			Value:     1.0,
			Timestamp: time.Now(),
			Tags: map[string]string{
				"manufacturer": device.Manufacturer,
				"model":        model,
				"type":         device.Type,
			},
			Fields: map[string]interface{}{
				"uptime":        uptime,
				"version":       version,
				"serial_number": serialNumber,
			},
		},
	}

	return data, nil
}

// getSysUpTime 獲取系統運行時間
func (c *SNMPCollector) getSysUpTime() (int64, error) {
	oid := ".1.3.6.1.2.1.1.3.0" // SNMPv2-MIB::sysUpTime.0
	result, err := c.client.Get([]string{oid})
	if err != nil {
		return 0, err
	}

	if len(result.Variables) == 0 {
		return 0, fmt.Errorf("no data returned for sysUpTime")
	}

	// 將 uint32 轉換為 int64
	uptime := int64(result.Variables[0].Value.(uint32))
	return uptime, nil
}

// getDeviceMetadata 獲取設備元數據
func (c *SNMPCollector) getDeviceMetadata(device models.Device) (string, string, string, error) {
	var model, version, serialNumber string
	var err error

	switch device.Manufacturer {
	case "Delta":
		model = "PDUE428"
		version, err = c.getDeltaVersion()
		serialNumber, err = c.getDeltaSerialNumber()
	case "Vertiv":
		model = "6PS56"
		version, err = c.getVertivVersion()
		serialNumber, err = c.getVertivSerialNumber()
	default:
		return "", "", "", fmt.Errorf("unsupported manufacturer: %s", device.Manufacturer)
	}

	if err != nil {
		return "", "", "", err
	}

	return model, version, serialNumber, nil
}

// getDeltaVersion 獲取 Delta PDU 版本
func (c *SNMPCollector) getDeltaVersion() (string, error) {
	oid := ".1.3.6.1.4.1.1718.3.2.3.1.1.0" // DeltaPDU-MIB::dpduIdentAgentSoftwareVersion.0
	result, err := c.client.Get([]string{oid})
	if err != nil {
		return "", err
	}

	if len(result.Variables) == 0 {
		return "", fmt.Errorf("no data returned for Delta version")
	}

	return result.Variables[0].Value.(string), nil
}

// getDeltaSerialNumber 獲取 Delta PDU 序列號
func (c *SNMPCollector) getDeltaSerialNumber() (string, error) {
	oid := ".1.3.6.1.4.1.1718.3.2.3.1.2.1" // DeltaPDU-MIB::dpduIdentSerialNumber.1
	result, err := c.client.Get([]string{oid})
	if err != nil {
		return "", err
	}

	if len(result.Variables) == 0 {
		return "", fmt.Errorf("no data returned for Delta serial number")
	}

	return result.Variables[0].Value.(string), nil
}

// getVertivVersion 獲取 Vertiv PDU 版本
func (c *SNMPCollector) getVertivVersion() (string, error) {
	oid := ".1.3.6.1.4.1.21239.2.1.1.1.0" // VERTIV-V5-MIB::productVersion.0
	result, err := c.client.Get([]string{oid})
	if err != nil {
		return "", err
	}

	if len(result.Variables) == 0 {
		return "", fmt.Errorf("no data returned for Vertiv version")
	}

	return result.Variables[0].Value.(string), nil
}

// getVertivSerialNumber 獲取 Vertiv PDU 序列號
func (c *SNMPCollector) getVertivSerialNumber() (string, error) {
	oid := ".1.3.6.1.4.1.21239.2.1.1.2.0" // VERTIV-V5-MIB::productSerialNumber.0
	result, err := c.client.Get([]string{oid})
	if err != nil {
		return "", err
	}

	if len(result.Variables) == 0 {
		return "", fmt.Errorf("no data returned for Vertiv serial number")
	}

	return result.Variables[0].Value.(string), nil
}

// Deploy 部署設備配置
func (c *SNMPCollector) Deploy(device models.Device) error {
	// 檢查設備是否支持 SNMP
	if device.Type != "snmp" {
		return fmt.Errorf("設備不支持 SNMP 協議: %s", device.Type)
	}

	// 更新收集器配置
	c.config.Host = device.IP
	if port, ok := device.Tags["port"]; ok {
		if p, err := strconv.Atoi(port); err == nil && p > 0 {
			c.config.Port = p
		}
	}
	if community, ok := device.Tags["community"]; ok {
		c.config.Community = community
	}
	if version, ok := device.Tags["version"]; ok {
		c.config.Version = version
	}

	// 重新初始化 SNMP 客戶端
	if err := c.connect(); err != nil {
		return fmt.Errorf("SNMP 連接失敗: %w", err)
	}

	return nil
}

// Discover 發現設備
func (c *SNMPCollector) Discover(ctx context.Context, ipRange models.IPRange) ([]models.Device, error) {
	var devices []models.Device

	// 解析 IP 範圍
	startIP := ipRange.StartIP
	endIP := ipRange.EndIP

	// 使用默認的 SNMP 配置進行掃描
	scanConfig := *c.config
	scanConfig.Timeout = 1 * time.Second // 掃描時使用較短的超時時間
	scanConfig.Retries = 1               // 掃描時減少重試次數

	// 遍歷 IP 範圍
	currentIP := startIP
	for {
		// 檢查上下文是否取消
		select {
		case <-ctx.Done():
			return devices, ctx.Err()
		default:
		}

		// 嘗試連接設備
		scanConfig.Host = currentIP
		client := &gosnmp.GoSNMP{
			Target:    scanConfig.Host,
			Port:      uint16(scanConfig.Port),
			Community: scanConfig.Community,
			Version:   gosnmp.Version2c,
			Timeout:   scanConfig.Timeout,
			Retries:   scanConfig.Retries,
		}

		if err := client.Connect(); err == nil {
			// 嘗試獲取系統描述
			result, err := client.Get([]string{".1.3.6.1.2.1.1.1.0"}) // SNMPv2-MIB::sysDescr.0
			if err == nil && len(result.Variables) > 0 {
				// 創建設備對象
				device := models.Device{
					IP:        currentIP,
					Type:      "snmp",
					UpdatedAt: time.Now(),
					Tags: map[string]string{
						"port":      fmt.Sprintf("%d", scanConfig.Port),
						"community": scanConfig.Community,
						"version":   scanConfig.Version,
					},
				}

				// 嘗試獲取更多設備信息
				if sysDescr, ok := result.Variables[0].Value.(string); ok {
					device.Tags["system_description"] = sysDescr
				}

				devices = append(devices, device)
			}
			client.Conn.Close()
		}

		// 移動到下一個 IP
		if currentIP == endIP {
			break
		}
		// TODO: 實現 IP 遞增邏輯
		// 這裡需要一個輔助函數來處理 IP 地址的遞增
	}

	return devices, nil
}

// Match 匹配設備
func (c *SNMPCollector) Match(device models.Device) (bool, error) {
	// 檢查設備類型是否為 SNMP
	if device.Type != "snmp" {
		return false, nil
	}

	// 檢查必要的標籤
	if _, ok := device.Tags["port"]; !ok {
		return false, nil
	}
	if _, ok := device.Tags["community"]; !ok {
		return false, nil
	}
	if _, ok := device.Tags["version"]; !ok {
		return false, nil
	}

	return true, nil
}
