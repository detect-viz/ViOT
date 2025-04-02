package services

import (
	"testing"
	"viot-scanner/models"
)

func TestModbusSlaveID(t *testing.T) {
	// 測試參數
	testCases := []struct {
		name     string
		ip       string
		port     int
		mode     string
		slaveID  byte
		expected bool
	}{
		{
			name:     "測試 SlaveID 0",
			ip:       "10.1.249.34",
			port:     7000,
			mode:     "RTUOverTCP",
			slaveID:  0,
			expected: true,
		},
		{
			name:     "測試 SlaveID 1",
			ip:       "10.1.249.34",
			port:     7000,
			mode:     "RTUOverTCP",
			slaveID:  1,
			expected: true, // 修改為 true，看看是否真的可以連接
		},
		{
			name:     "測試 SlaveID 2",
			ip:       "10.1.249.34",
			port:     7000,
			mode:     "RTUOverTCP",
			slaveID:  2,
			expected: true, // 修改為 true，看看是否真的可以連接
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 創建測試用的 session
			session := models.ModbusSession{
				Conn: models.ModbusConn{
					IP:      tc.ip,
					Port:    tc.port,
					Mode:    tc.mode,
					SlaveID: tc.slaveID,
				},
				Requests: []models.ModbusRequest{
					{
						RegisterType: "holding",
						Address:      768,
						Length:       8,
						RegisterName: "test",
					},
				},
			}

			// 測試連接
			data, err := GetModbusValue(session)
			if tc.expected {
				if err != nil {
					t.Errorf("預期成功但失敗: %v", err)
				}
				if data == nil {
					t.Error("預期有數據但為空")
				}
				if _, ok := data["test"]; !ok {
					t.Error("預期有 test 數據但不存在")
				}
			} else {
				if err == nil {
					t.Errorf("預期失敗但成功，SlaveID: %d", tc.slaveID)
				}
			}
		})
	}
}

func TestModbusConnection(t *testing.T) {
	// 測試參數
	testCases := []struct {
		name     string
		ip       string
		port     int
		mode     string
		slaveID  byte
		expected bool
	}{
		{
			name:     "測試有效連接 (SlaveID 0)",
			ip:       "10.1.249.34",
			port:     7000,
			mode:     "RTUOverTCP",
			slaveID:  0,
			expected: true,
		},
		{
			name:     "測試無效 SlaveID 1",
			ip:       "10.1.249.34",
			port:     7000,
			mode:     "RTUOverTCP",
			slaveID:  1,
			expected: false,
		},
		{
			name:     "測試無效 SlaveID 2",
			ip:       "10.1.249.34",
			port:     7000,
			mode:     "RTUOverTCP",
			slaveID:  2,
			expected: false,
		},
		{
			name:     "測試無效 IP",
			ip:       "192.168.1.999",
			port:     7000,
			mode:     "RTUOverTCP",
			slaveID:  0,
			expected: false,
		},
		{
			name:     "測試無效端口",
			ip:       "10.1.249.34",
			port:     9999,
			mode:     "RTUOverTCP",
			slaveID:  0,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			success := TestModbusConnect(tc.ip, tc.port, tc.mode, tc.slaveID)
			if success != tc.expected {
				t.Errorf("預期結果 %v 但得到 %v (SlaveID: %d)", tc.expected, success, tc.slaveID)
			}
		})
	}
}
