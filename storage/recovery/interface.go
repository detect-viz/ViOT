package recovery

import "viot/models"

// FallbackStrategy 定義資料儲存失敗時的備份策略接口
type FallbackStrategy interface {
	// SavePDUData 當主要儲存失敗時，將PDU數據保存到備份儲存
	SavePDUData(filename string, data []models.PDUData) error

	// SaveDeviceData 當主要儲存失敗時，將空調機架數據保存到備份儲存
	SaveDeviceData(filename string, data []models.DeviceData) error

	// GetPendingPDUData 獲取待處理的PDU數據
	GetPendingPDUData() (map[string][]models.PDUData, error)

	// GetPendingDeviceData 獲取待處理的空調機架數據
	GetPendingDeviceData() (map[string][]models.DeviceData, error)

	// MarkPDUDataProcessed 標記PDU數據已處理
	MarkPDUDataProcessed(filename string) error

	// MarkDeviceDataProcessed 標記空調機架數據已處理
	MarkDeviceDataProcessed(filename string) error
}

// RetryPolicy 定義重試策略接口
type RetryPolicy interface {
	// ShouldRetry 判斷是否應該重試
	ShouldRetry(attempt int, err error) bool

	// NextRetryDelay 返回下次重試的延遲時間(秒)
	NextRetryDelay(attempt int) int
}
