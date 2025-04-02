package interfaces

import "viot/models"

// OutputInterface 輸出接口
type OutputInterface interface {
	Write(data []models.DeviceData) error
}
