package deployer

import (
	"context"
	"os/exec"
	"path/filepath"
	"text/template"

	"viot/logger"
	"viot/models"
)

// TelegrafDeployer 部署 Telegraf 配置
type TelegrafDeployer struct {
	logger         logger.Logger
	config         models.TelegrafDeployConfig
	templateEngine *template.Template
}

// NewTelegrafDeployer 創建 Telegraf 部署器
func NewTelegrafDeployer(config models.TelegrafDeployConfig, log logger.Logger) *TelegrafDeployer {
	return &TelegrafDeployer{
		config:         config,
		logger:         log.Named("telegraf-deployer"),
		templateEngine: newTemplateEngine(config.TemplateDir),
	}
}

// newTemplateEngine 創建模板引擎
func newTemplateEngine(templateDir string) *template.Template {
	// 初始化模板引擎
	return template.New("telegraf")
}

// Deploy 實現部署
func (d *TelegrafDeployer) Deploy(ctx context.Context, device models.Device) (*models.DeployResult, error) {
	// 1. 根據設備信息生成配置
	configPath, err := d.GenerateConfig(device)
	if err != nil {
		return &models.DeployResult{
			Success:    false,
			DeviceInfo: device,
			Error:      err.Error(),
		}, err
	}

	// 2. 驗證配置
	if err := d.Validate(configPath); err != nil {
		return &models.DeployResult{
			Success:    false,
			DeviceInfo: device,
			ConfigPath: configPath,
			Error:      err.Error(),
		}, err
	}

	// 3. 重啟 Telegraf 以應用新配置
	if err := d.restartTelegraf(); err != nil {
		return &models.DeployResult{
			Success:    false,
			DeviceInfo: device,
			ConfigPath: configPath,
			Error:      err.Error(),
		}, err
	}

	return &models.DeployResult{
		Success:    true,
		DeviceInfo: device,
		ConfigPath: configPath,
	}, nil
}

// Validate 驗證配置
func (d *TelegrafDeployer) Validate(configPath string) error {
	// 執行驗證命令
	cmd := exec.Command(d.config.TelegrafBin, "--test", "--config", configPath)
	return cmd.Run()
}

// GenerateConfig 生成配置
func (d *TelegrafDeployer) GenerateConfig(device models.Device) (string, error) {
	// 1. 選擇合適的模板
	_ = d.selectTemplate(device)

	// 2. 使用模板生成配置
	outputPath := filepath.Join(d.config.OutputDir, device.IP+".conf")
	// 使用模板引擎生成配置文件

	return outputPath, nil
}

// selectTemplate 選擇適合設備的模板
func (d *TelegrafDeployer) selectTemplate(device models.Device) string {
	// 根據設備的製造商、型號和協議選擇模板
	templatePrefix := device.Protocol
	if device.Manufacturer != "" && device.Model != "" {
		return templatePrefix + "_" + device.Manufacturer + "_" + device.Model + ".tmpl"
	}
	return templatePrefix + "_default.tmpl"
}

// restartTelegraf 重啟 Telegraf
func (d *TelegrafDeployer) restartTelegraf() error {
	// 執行重啟命令
	cmd := exec.Command("bash", "-c", d.config.RestartCommand)
	return cmd.Run()
}
