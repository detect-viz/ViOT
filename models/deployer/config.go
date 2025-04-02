package deployer

import "time"

// DeployConfig 部署配置
type DeployConfig struct {
	VerifyAfterDeploy  bool           `yaml:"verify_after_deploy"`
	RollbackOnFailure  bool           `yaml:"rollback_on_failure"`
	ConcurrentDeploys  int            `yaml:"concurrent_deploys"`
	DeploymentTimeout  string         `yaml:"deployment_timeout"`
	RestartServices    bool           `yaml:"restart_services"`
	BackupBeforeDeploy bool           `yaml:"backup_before_deploy"`
	NotifyOnDeploy     bool           `yaml:"notify_on_deploy"`
	Telegraf           TelegrafConfig `yaml:"telegraf"`
}

// TelegrafConfig Telegraf 配置
type TelegrafConfig struct {
	// HTTP 服務器配置
	ListenAddr   string        `json:"listen_addr" yaml:"listen_addr"`
	Endpoint     string        `json:"endpoint" yaml:"endpoint"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout" yaml:"idle_timeout"`

	// 數據處理配置
	BatchSize    int           `json:"batch_size" yaml:"batch_size"`
	FlushTimeout time.Duration `json:"flush_timeout" yaml:"flush_timeout"`

	// 緩衝區配置
	BufferSize int           `json:"buffer_size" yaml:"buffer_size"`
	MaxRetries int           `json:"max_retries" yaml:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay" yaml:"retry_delay"`

	// 部署配置
	ConfigPath     string `json:"config_path" yaml:"config_path"`
	DeployPath     string `json:"deploy_path" yaml:"deploy_path"`
	TemplatesPath  string `json:"templates_path" yaml:"templates_path"`
	BackupPath     string `json:"backup_path" yaml:"backup_path"`
	BinaryPath     string `json:"binary_path" yaml:"binary_path"`
	VerifyCommand  string `json:"verify_command" yaml:"verify_command"`
	RestartCommand string `json:"restart_command" yaml:"restart_command"`
}

// TelegrafDeployConfig Telegraf 部署配置
type TelegrafDeployConfig struct {
	TemplateDir     string `yaml:"template_dir"`     // 模板目錄
	OutputDir       string `yaml:"output_dir"`       // 輸出目錄
	TelegrafBin     string `yaml:"telegraf_bin"`     // Telegraf 可執行文件路徑
	RestartCommand  string `yaml:"restart_command"`  // 重啟 Telegraf 的命令
	ValidateCommand string `yaml:"validate_command"` // 驗證配置的命令
}
