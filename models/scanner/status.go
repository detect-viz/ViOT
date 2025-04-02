package scanner

import "time"

// ScanStatus 掃描狀態
type ScanStatus struct {
	Running       bool
	StartTime     time.Time
	EndTime       time.Time
	TotalIPs      int
	ScannedIPs    int
	FoundDevices  int
	Progress      float64
	RemainingTime int
}

// ScanResult 掃描結果
type ScanResult struct {
	IP            string
	Protocol      string
	Port          int
	DeviceConfig  *DeviceConfig
	CollectedData map[string]string
	MAC           string
	ScanTime      time.Time
}

// ScanSetting 掃描設置
type ScanSetting struct {
	IPRanges    []string
	SNMPPorts   []int
	ModbusPorts []int
	Interval    int
	Duration    int
}
