package models

import (
	"time"

	"viot/models/collector"
	"viot/models/deployer"
	"viot/models/processor"
	"viot/models/scanner"
	"viot/models/webservice"
)

// Config 主系統配置
type Config struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Scheduler      SchedulerConfig
	ProcessMonitor ProcessMonitorConfig
	Logging        LoggingConfig
	InfluxDB       InfluxDBConfig
	Output         OutputConfig
	Services       ServiceController `json:"services"`
}

// ServiceController 獨立模組控制器配置
type ServiceController struct {
	Web struct {
		Enabled                     bool   `json:"enabled"`
		ConfigPath                  string `json:"config_path"`
		webservice.WebserviceConfig `json:"webservice"`
	}
	Processor struct {
		Enabled                   bool `json:"enabled"`
		Interval                  int  `json:"interval"`
		processor.ProcessorConfig `json:"processor"`
	} `json:"processor"`
	Collector struct {
		Enabled                   bool   `json:"enabled"`
		ConfigPath                string `json:"config_path"`
		collector.CollectorConfig `json:"collector"`
	}
	Scanner struct {
		Enabled               bool   `json:"enabled"`
		ConfigPath            string `json:"config_path"`
		scanner.ScannerConfig `json:"scanner"`
	}
	Deployer struct {
		Enabled               bool   `json:"enabled"`
		ConfigPath            string `json:"config_path"`
		deployer.DeployConfig `json:"deployer"`
	}
}

// OutputConfig 輸出配置
type OutputConfig struct {
	Enabled    bool   `json:"enabled"`
	Type       string `json:"type"`
	Interval   int    `json:"interval"`
	BatchSize  int    `json:"batch_size"`
	RetryCount int    `json:"retry_count"`
}

// SchedulerConfig 調度器配置
type SchedulerConfig struct {
	Enabled  bool `json:"enabled"`
	Interval int  `json:"interval"`
}

// ProcessMonitorConfig 進程監控配置
type ProcessMonitorConfig struct {
	Enabled  bool `json:"enabled"`
	Interval int  `json:"interval"`
}

// LoggingConfig 日誌配置
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
	Compress   bool   `json:"compress"`
}

// InfluxDBConfig InfluxDB配置
type InfluxDBConfig struct {
	URL          string       `json:"url"`
	Token        string       `json:"token"`
	Org          string       `json:"org"`
	Bucket       string       `json:"bucket"`
	Interval     int          `json:"interval"`
	DataPolicy   dataPolicy   `json:"data_policy"`
	BackupPolicy backupPolicy `json:"backup_policy"`
}

// DataPolicy 數據策略配置
type dataPolicy struct {
	FailWritePath     string `yaml:"fail_write_path"`
	LowSpaceThreshold int64  `yaml:"low_space_threshold"`
}

// BackupPolicy 備份策略配置
type backupPolicy struct {
	Path      string        `yaml:"path"`
	Interval  time.Duration `yaml:"interval"`
	Retention time.Duration `yaml:"retention"`
}
