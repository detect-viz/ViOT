package collector

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"viot/models"
)

// IPRange 表示 IP 掃描範圍
type IPRange struct {
	StartIP  string
	EndIP    string
	Protocol string
	Ports    []int
	Enabled  bool
	Interval time.Duration
}

// CollectorConfig 收集器配置
type CollectorConfig struct {
	// 基本配置
	Enabled  bool
	Interval time.Duration

	// IP掃描配置
	IPRangeFile    string
	IPRanges       []IPRange
	ScanInterval   time.Duration
	ScanTimeout    time.Duration
	ScanConcurrent int

	// 部署配置
	AutoDeploy       bool
	PendingListPath  string
	RegistryPath     string
	DeployConfigPath string

	// 工作目錄配置
	WorkDir string
	LogDir  string
	DataDir string

	// 協議配置
	Protocols struct {
		SNMP struct {
			Community string
			Version   string
			Port      int
			Timeout   int
			Retries   int
		}
		Modbus struct {
			Host    string
			Port    int
			SlaveID byte
			Timeout int
		}
	}
}

// LoadIPRanges 從CSV文件加載IP範圍
func LoadIPRanges(filename string) ([]models.IPRange, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("無法打開IP範圍文件: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// 跳過標題行
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("讀取CSV標題失敗: %v", err)
	}

	var ranges []models.IPRange
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("讀取CSV記錄失敗: %v", err)
		}

		// 解析端口
		ports := make([]int, 0)
		if record[3] != "" {
			portStr := record[3]
			port, err := strconv.Atoi(portStr)
			if err == nil {
				ports = append(ports, port)
			}
		}

		// 解析時間間隔
		interval := 5 * time.Minute
		if record[5] != "" {
			if d, err := time.ParseDuration(record[5]); err == nil {
				interval = d
			}
		}

		ipRange := models.IPRange{
			StartIP:  record[0],
			EndIP:    record[1],
			Protocol: record[2],
			Ports:    ports,
			Enabled:  record[4] == "true",
			Interval: interval,
		}
		ranges = append(ranges, ipRange)
	}

	return ranges, nil
}

// LoadCollectorConfig 加載收集器配置
func LoadCollectorConfig(configPath string) (*models.CollectorConfig, error) {
	config := &models.CollectorConfig{
		Enabled:        true,
		Interval:       time.Minute,
		ScanInterval:   time.Hour,
		ScanTimeout:    time.Minute,
		ScanConcurrent: 10,
		AutoDeploy:     false,
		WorkDir:        ".",
		LogDir:         "logs",
		DataDir:        "data",
	}

	// 設置文件路徑
	config.IPRangeFile = filepath.Join(config.WorkDir, "ip_range.csv")
	config.PendingListPath = filepath.Join(config.DataDir, "pending_list.csv")
	config.RegistryPath = filepath.Join(config.DataDir, "registry.csv")
	config.DeployConfigPath = filepath.Join(config.WorkDir, "deployer.yaml")

	// 加載 IP 範圍
	ranges, err := LoadIPRanges(config.IPRangeFile)
	if err != nil {
		return nil, fmt.Errorf("加載IP範圍失敗: %v", err)
	}
	config.IPRanges = ranges

	return config, nil
}
