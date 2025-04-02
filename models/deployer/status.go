package deployer

import "time"

// DeployResult 部署結果
type DeployResult struct {
	Success    bool   `json:"success"`
	DeviceInfo Device `json:"device_info"`
	ConfigPath string `json:"config_path"`
	Error      string `json:"error,omitempty"`
}

// DeployerStatus 部署器狀態
type DeployerStatus struct {
	Running       bool      `json:"running"`
	LastDeploy    time.Time `json:"last_deploy"`
	DeployedCount int64     `json:"deployed_count"`
	ErrorCount    int64     `json:"error_count"`
	LastError     string    `json:"last_error,omitempty"`
}

// DeployerInstance 部署器實例
type DeployerInstance struct {
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Config    DeployerConfig `json:"config"`
	Status    DeployerStatus `json:"status"`
	StartTime time.Time      `json:"start_time"`
}
