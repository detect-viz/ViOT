package services

import (
	"fmt"
	"time"
	"viot-scanner/global"
	"viot-scanner/models"

	"github.com/grid-x/modbus"
)

// NewModbusClient 創建 Modbus 客戶端
func NewModbusClient(ip string, port int, mode string, slaveID byte) (modbus.Client, modbus.ClientHandler, error) {
	//fmt.Printf("正在連接到 Modbus 設備 %s:%d...\n", ip, port)

	var handler modbus.ClientHandler
	switch mode {
	case "RTUOverTCP":
		handler = modbus.NewRTUOverTCPClientHandler(fmt.Sprintf("%s:%d", ip, port))
	case "TCP":
		handler = modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", ip, port))
	default:
		return nil, nil, fmt.Errorf("不支援的通訊模式: %s", mode)
	}

	switch h := handler.(type) {
	case *modbus.TCPClientHandler:
		h.SlaveID = slaveID
		h.Timeout = time.Duration(global.Config.Modbus.Timeout) * time.Second
	case *modbus.RTUClientHandler:
		h.SlaveID = slaveID
		h.Timeout = time.Duration(global.Config.Modbus.Timeout) * time.Second
	}

	// 確保在返回錯誤時不會留下未關閉的連接
	if err := handler.Connect(); err != nil {
		handler.Close()
		return nil, nil, fmt.Errorf("連接失敗: %v", err)
	}

	return modbus.NewClient(handler), handler, nil
}

// GetModbusDeviceInfo 獲取 Modbus 設備信息
func GetModbusValue(session models.ModbusSession) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	client, handler, err := NewModbusClient(session.Conn.IP, session.Conn.Port, session.Conn.Mode, session.Conn.SlaveID)

	if err != nil {
		fmt.Printf("🛜 [%s:%d:%d] 連接失敗: %v\n", session.Conn.IP, session.Conn.Port, session.Conn.SlaveID, err)
		return nil, fmt.Errorf("創建 Modbus 客戶端失敗: %v", err)
	}
	defer handler.Close()

	// 使用閉包來處理錯誤，確保在返回錯誤前關閉連接
	readError := func() error {
		for _, request := range session.Requests {
			var results []byte
			var err error

			// 每次讀取前重新設置 SlaveID
			handler.SetSlave(session.Conn.SlaveID)

			switch request.RegisterType {
			case "input":
				results, err = client.ReadInputRegisters(request.Address, request.Length)
			case "holding":
				results, err = client.ReadHoldingRegisters(request.Address, request.Length)
			default:
				return fmt.Errorf("不支援的寄存器類型: %s", request.RegisterType)
			}

			if err != nil {
				return fmt.Errorf("讀取 %s 寄存器失敗: %v", request.RegisterType, err)
			}

			// 檢查是否為空或全零
			if len(results) == 0 || allZeros(results) {
				fmt.Printf("⚠️ [%s:%d:%d] 讀取到的數據為空或全零\n", session.Conn.IP, session.Conn.Port, session.Conn.SlaveID)
				return fmt.Errorf("讀取到的數據無效")
			}

			data[request.RegisterName] = string(results)
		}
		return nil
	}()

	if readError != nil {
		return nil, readError
	}

	return data, nil
}

// allZeros 檢查字節切片是否全為零
func allZeros(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// TestModbusConnect 測試 Modbus 連接
func TestModbusConnect(ip string, port int, mode string, slaveID byte) bool {
	fmt.Printf("正在測試 Modbus 設備連接 %s:%d (SlaveID: %d)...\n", ip, port, slaveID)

	client, handler, err := NewModbusClient(ip, port, mode, slaveID)
	if err != nil {
		fmt.Printf("連接失敗: %v\n", err)
		return false
	}
	defer handler.Close()

	// 每次讀取前重新設置 SlaveID
	handler.SetSlave(slaveID)

	// 嘗試讀取一個保持寄存器來測試連接
	results, err := client.ReadHoldingRegisters(768, 8)
	if err != nil {
		fmt.Printf("讀取寄存器失敗: %v\n", err)
		return false
	}

	if len(results) == 0 || allZeros(results) {
		fmt.Printf("讀取到的數據為空或全零\n")
		return false
	}

	fmt.Printf("Modbus 連接測試成功，SlaveID: %d\n", slaveID)
	return true
}
