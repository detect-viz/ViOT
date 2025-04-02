package inputs

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"strconv"
	"time"

	"viot/logger"
	"viot/models"
	"viot/pkg/collector/core"

	"github.com/grid-x/modbus"
)

// ModbusInput Modbus 輸入插件
type ModbusInput struct {
	*core.BaseDeviceInput
	client  modbus.Client
	handler *modbus.TCPClientHandler
	config  *models.CollectorConfig
}

// NewModbusInput 創建新的 Modbus 輸入插件
func NewModbusInput(config *models.CollectorConfig, logger logger.Logger) *ModbusInput {
	return &ModbusInput{
		BaseDeviceInput: core.NewBaseDeviceInput("modbus", config, logger),
		config:          config,
	}
}

// Start 啟動 Modbus 輸入插件
func (m *ModbusInput) Start(ctx context.Context) error {
	if err := m.BaseDeviceInput.Start(ctx); err != nil {
		return fmt.Errorf("failed to start base device input: %v", err)
	}

	// 創建 Modbus TCP 客戶端
	address := fmt.Sprintf("%s:%d", m.config.Protocols.Modbus.Host, m.config.Protocols.Modbus.Port)
	handler := modbus.NewTCPClientHandler(address)
	handler.SlaveID = m.config.Protocols.Modbus.SlaveID
	handler.Timeout = time.Duration(m.config.Protocols.Modbus.Timeout) * time.Second

	if err := handler.Connect(); err != nil {
		m.UpdateError(fmt.Errorf("failed to connect modbus client: %v", err))
		return err
	}

	m.handler = handler
	m.client = modbus.NewClient(handler)
	return nil
}

// Stop 停止 Modbus 輸入插件
func (m *ModbusInput) Stop() error {
	if m.handler != nil {
		m.handler.Close()
		m.handler = nil
		m.client = nil
	}
	return m.BaseDeviceInput.Stop()
}

// Collect 收集 Modbus 設備數據
func (m *ModbusInput) Collect(ctx context.Context) ([]models.RestAPIPoint, error) {
	if m.client == nil {
		return nil, fmt.Errorf("modbus client not initialized")
	}

	// 讀取保持寄存器
	results, err := m.client.ReadHoldingRegisters(uint16(0), uint16(10))
	if err != nil {
		m.UpdateError(fmt.Errorf("failed to read holding registers: %v", err))
		return nil, err
	}

	// 創建數據點
	point := models.RestAPIPoint{
		Metric:    "modbus_data",
		Value:     0.0, // 將在後面更新
		Tags:      make(map[string]string),
		Fields:    make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	// 添加標籤
	point.Tags["type"] = "modbus"
	point.Tags["host"] = fmt.Sprintf("%s:%d", m.config.Protocols.Modbus.Host, m.config.Protocols.Modbus.Port)

	// 解析寄存器數據
	for i, value := range results {
		point.Fields[fmt.Sprintf("register_%d", i)] = uint16(value)
	}

	// 設置第一個寄存器的值作為主要值
	if len(results) > 0 {
		point.Value = float64(results[0])
	}

	return []models.RestAPIPoint{point}, nil
}

// Deploy 部署 Modbus 設備配置
func (m *ModbusInput) Deploy(ctx context.Context) error {
	if m.client == nil {
		return fmt.Errorf("modbus client not initialized")
	}

	// 寫入保持寄存器
	_, err := m.client.WriteSingleRegister(uint16(0), uint16(1))
	if err != nil {
		m.UpdateError(fmt.Errorf("failed to write holding register: %v", err))
		return err
	}

	return nil
}

// ModbusCollector Modbus 收集器
type ModbusCollector struct {
	config    *models.ModbusConfig
	registers []models.ModbusRegister
	client    modbus.Client
	handler   *modbus.TCPClientHandler
	status    models.CollectorStatus
}

// NewModbusCollector 創建新的 Modbus 收集器
func NewModbusCollector(collectorConfig *models.ModbusCollectorConfig, collectorHost string) (*ModbusCollector, error) {
	// 使用 MergeWithDefault 獲取完整配置
	config := collectorConfig.MergeWithDefault()

	// 注入收集器主機地址
	if collectorHost != "" {
		config.Host = collectorHost
	}

	collector := &ModbusCollector{
		config:    &config,
		registers: collectorConfig.Registers,
		status: models.CollectorStatus{
			Running: false,
		},
	}

	return collector, nil
}

// connect 建立 Modbus 連接
func (c *ModbusCollector) connect() error {
	if c.client != nil {
		return nil
	}

	handler := modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", c.config.Host, c.config.Port))
	handler.SlaveID = c.config.SlaveID
	handler.Timeout = c.config.Timeout

	if err := handler.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.handler = handler
	c.client = modbus.NewClient(handler)
	return nil
}

// Status 返回收集器狀態
func (c *ModbusCollector) Status() models.CollectorStatus {
	return c.status
}

// Start 啟動收集器
func (c *ModbusCollector) Start(ctx context.Context) error {
	c.status.Running = true
	return nil
}

// Stop 停止收集器
func (c *ModbusCollector) Stop() error {
	if c.handler != nil {
		c.handler.Close()
		c.handler = nil
		c.client = nil
	}
	c.status.Running = false
	return nil
}

// Name 返回收集器名稱
func (c *ModbusCollector) Name() string {
	return "modbus"
}

// Collect 收集設備數據
func (c *ModbusCollector) Collect(ctx context.Context, device models.Device) ([]models.DeviceData, error) {
	// 確保連接
	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	var data []models.DeviceData

	// 遍歷所有寄存器
	for _, reg := range c.registers {
		value, err := c.readRegister(reg)
		if err != nil {
			return nil, fmt.Errorf("failed to read register %s: %w", reg.Name, err)
		}

		// 創建數據點
		point := models.DeviceData{
			Name:      device.IP,
			Metric:    fmt.Sprintf("modbus_%s", reg.Name),
			Value:     value,
			Timestamp: time.Now(),
			Tags: map[string]string{
				"type":          "modbus",
				"register":      reg.Name,
				"address":       fmt.Sprintf("%d", reg.Address),
				"register_type": string(reg.Type),
				"data_type":     string(reg.DataType),
				"unit":          reg.Unit,
			},
			Fields: map[string]interface{}{
				"raw_value": value,
				"scale":     reg.Scale,
			},
		}

		data = append(data, point)
	}

	return data, nil
}

// readRegister 讀取寄存器值
func (c *ModbusCollector) readRegister(reg models.ModbusRegister) (float64, error) {
	var results []byte
	var err error

	// 根據寄存器類型讀取數據
	switch reg.Type {
	case models.Coil:
		var coils []byte
		coils, err = c.client.ReadCoils(reg.Address, 1)
		if err == nil && len(coils) > 0 {
			if coils[0] != 0 {
				return 1, nil
			}
			return 0, nil
		}
	case models.DiscreteInput:
		var inputs []byte
		inputs, err = c.client.ReadDiscreteInputs(reg.Address, 1)
		if err == nil && len(inputs) > 0 {
			if inputs[0] != 0 {
				return 1, nil
			}
			return 0, nil
		}
	case models.HoldingRegister:
		results, err = c.client.ReadHoldingRegisters(reg.Address, c.getRegisterLength(reg.DataType))
	case models.InputRegister:
		results, err = c.client.ReadInputRegisters(reg.Address, c.getRegisterLength(reg.DataType))
	default:
		return 0, fmt.Errorf("unsupported register type: %s", reg.Type)
	}

	if err != nil {
		return 0, err
	}

	// 解析數據
	value, err := c.parseValue(results, reg.DataType, reg.ByteOrder)
	if err != nil {
		return 0, err
	}

	// 應用縮放因子
	return value * reg.Scale, nil
}

// getRegisterLength 獲取寄存器長度
func (c *ModbusCollector) getRegisterLength(dataType models.ModbusDataType) uint16 {
	switch dataType {
	case models.INT16, models.UINT16:
		return 1
	case models.INT32, models.UINT32, models.FLOAT32:
		return 2
	case models.INT64, models.UINT64, models.FLOAT64:
		return 4
	case models.STRING:
		return 8 // 默認字符串長度
	default:
		return 1
	}
}

// parseValue 解析寄存器值
func (c *ModbusCollector) parseValue(data []byte, dataType models.ModbusDataType, byteOrder string) (float64, error) {
	if len(data) == 0 {
		return 0, fmt.Errorf("no data to parse")
	}

	// 設置字節序
	var order binary.ByteOrder = binary.BigEndian
	if byteOrder == "little-endian" {
		order = binary.LittleEndian
	}

	switch dataType {
	case models.INT16:
		return float64(int16(order.Uint16(data))), nil
	case models.UINT16:
		return float64(order.Uint16(data)), nil
	case models.INT32:
		return float64(int32(order.Uint32(data))), nil
	case models.UINT32:
		return float64(order.Uint32(data)), nil
	case models.INT64:
		return float64(int64(order.Uint64(data))), nil
	case models.UINT64:
		return float64(order.Uint64(data)), nil
	case models.FLOAT32:
		bits := order.Uint32(data)
		return float64(math.Float32frombits(bits)), nil
	case models.FLOAT64:
		bits := order.Uint64(data)
		return math.Float64frombits(bits), nil
	case models.STRING:
		return 0, fmt.Errorf("string type not supported for numeric value")
	default:
		return 0, fmt.Errorf("unsupported data type: %s", dataType)
	}
}

// Deploy 部署設備配置
func (c *ModbusCollector) Deploy(device models.Device) error {
	// 檢查設備是否支持 Modbus
	if device.Type != "modbus" {
		return fmt.Errorf("設備不支持 Modbus 協議: %s", device.Type)
	}

	// 更新收集器配置
	c.config.Host = device.IP
	if port, ok := device.Tags["port"]; ok {
		if p, err := strconv.Atoi(port); err == nil && p > 0 {
			c.config.Port = p
		}
	}
	if slaveID, ok := device.Tags["slave_id"]; ok {
		if id, err := strconv.Atoi(slaveID); err == nil && id > 0 {
			c.config.SlaveID = byte(id)
		}
	}

	// 重新初始化 Modbus 客戶端
	if err := c.connect(); err != nil {
		return fmt.Errorf("modbus 連接失敗: %w", err)
	}

	return nil
}

// Match 匹配設備
func (c *ModbusCollector) Match(device models.Device) (bool, error) {
	// 檢查設備類型是否為 Modbus
	if device.Type != "modbus" {
		return false, nil
	}

	// 檢查必要的標籤
	if _, ok := device.Tags["port"]; !ok {
		return false, nil
	}
	if _, ok := device.Tags["slave_id"]; !ok {
		return false, nil
	}

	return true, nil
}

// Discover 發現設備
func (c *ModbusCollector) Discover(ctx context.Context, ipRange models.IPRange) ([]models.Device, error) {
	var devices []models.Device

	// 遍歷IP範圍
	startIP := net.ParseIP(ipRange.StartIP)
	endIP := net.ParseIP(ipRange.EndIP)

	if startIP == nil || endIP == nil {
		return nil, fmt.Errorf("無效的IP範圍")
	}

	for ip := startIP; bytes4ToUint32(ip) <= bytes4ToUint32(endIP); incrementIP(ip) {
		// 嘗試連接
		c.config.Host = ip.String()
		if err := c.connect(); err != nil {
			continue
		}

		// 嘗試讀取第一個保持寄存器
		_, err := c.client.ReadHoldingRegisters(0, 1)
		if err != nil {
			continue
		}

		// 添加發現的設備
		devices = append(devices, models.Device{
			IP:   ip.String(),
			Type: "modbus",
			Tags: map[string]string{
				"port":     fmt.Sprintf("%d", c.config.Port),
				"slave_id": fmt.Sprintf("%d", c.config.SlaveID),
			},
		})
	}

	return devices, nil
}

// bytes4ToUint32 將IPv4地址轉換為uint32
func bytes4ToUint32(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

// incrementIP 增加IP地址
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
