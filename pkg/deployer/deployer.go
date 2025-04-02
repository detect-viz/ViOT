package deployer

import (
	"context"
	"fmt"

	"viot/logger"
	"viot/models"
)

// Deployer 部署器工廠
type Deployer struct {
	logger logger.Logger
}

// NewDeployer 創建部署器工廠
func NewDeployer(logger logger.Logger) *Deployer {
	return &Deployer{
		logger: logger,
	}
}

// Deploy 部署設備
func (d *Deployer) Deploy(name string, config *models.DeployConfig) error {
	// 根據配置創建具體的部署器
	var deployer interface {
		Deploy(ctx context.Context, device models.Device) (*models.DeployResult, error)
	}

	switch {
	case config.Telegraf.ConfigPath != "":
		// 轉換配置類型
		telegrafConfig := models.TelegrafDeployConfig{
			TemplateDir:     config.Telegraf.TemplatesPath,
			OutputDir:       config.Telegraf.DeployPath,
			TelegrafBin:     config.Telegraf.BinaryPath,
			RestartCommand:  config.Telegraf.RestartCommand,
			ValidateCommand: config.Telegraf.VerifyCommand,
		}
		deployer = NewTelegrafDeployer(telegrafConfig, d.logger)
	default:
		return fmt.Errorf("不支持的部署配置類型: %s", name)
	}

	// 創建設備信息
	device := models.Device{
		IP:           name,
		Protocol:     "telegraf", // 默認為 telegraf
		Manufacturer: "unknown",  // 默認值
		Model:        "unknown",  // 默認值
		Type:         "unknown",  // 默認值
	}

	// 執行部署
	result, err := deployer.Deploy(context.Background(), device)
	if err != nil {
		return fmt.Errorf("部署失敗: %v", err)
	}

	if !result.Success {
		return fmt.Errorf("部署失敗: %s", result.Error)
	}

	return nil
}
