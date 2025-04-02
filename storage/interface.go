package storage

import (
	"viot/domain/common"
)

// PDUStorage 定義 PDU 存儲接口
type PDUStorage interface {
	// ReadPDUData 從存儲中讀取 PDU 數據
	ReadPDUData(filename, room string) ([]common.PDUData, error)

	// WritePDUData 將 PDU 數據寫入存儲
	WritePDUData(filename string, data []common.PDUData) error

	// UpdatePDUStatus 更新 PDU 狀態
	UpdatePDUStatus(filename, pduName, status string) error

	// UpdatePDUTags 更新 PDU 標籤
	UpdatePDUTags(filename, pduName string, tags []string) error

	// GetPDUTags 獲取 PDU 標籤
	GetPDUTags(filename, pduName string) ([]string, error)

	// ClearPDUData 清除 PDU 數據
	ClearPDUData(filename string) error
}

type DCStorage interface {
	// GetRooms 獲取所有房間
	GetRooms(filename string) ([]string, error)
}

// Storage 定義綜合存儲接口
type Storage interface {
	PDUStorage
	DCStorage
	WritePDUData(filename string, data []common.PDUData) error
	ReadDeviceData(filename string, deviceName string) (common.RoomLayout, error)
}
