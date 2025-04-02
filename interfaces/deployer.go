package interfaces

import (
	"viot/models"
	"viot/models/deployer"
)

// DeployerInterface 部署器接口
type DeployerInterface interface {
	Deploy(device models.Device) error
	Start() error
	Stop() error
	GetStatus() deployer.DeployerStatus
}
