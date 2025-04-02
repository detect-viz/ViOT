package health

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// HealthStatus 表示系統健康狀態
type HealthStatus struct {
	Status        string    `json:"status"`         // 系統狀態：healthy, degraded, unhealthy
	Timestamp     time.Time `json:"timestamp"`      // 檢查時間
	CPUUsage      float64   `json:"cpu_usage"`      // CPU 使用率
	MemoryUsage   float64   `json:"memory_usage"`   // 內存使用率
	DiskUsage     float64   `json:"disk_usage"`     // 磁盤使用率
	NetworkStatus string    `json:"network_status"` // 網絡狀態
	LogFileSize   int64     `json:"log_file_size"`  // 日誌文件大小
	Errors        []string  `json:"errors"`         // 錯誤信息
}

// DeviceHealth 表示設備健康狀態
type DeviceHealth struct {
	IP           string    `json:"ip"`            // 設備 IP
	Status       string    `json:"status"`        // 設備狀態：online, offline, degraded
	LastSeen     time.Time `json:"last_seen"`     // 最後在線時間
	ResponseTime int64     `json:"response_time"` // 響應時間（毫秒）
	Errors       []string  `json:"errors"`        // 錯誤信息
}

// ServiceHealth 表示服務健康狀態
type ServiceHealth struct {
	Name         string    `json:"name"`          // 服務名稱
	Status       string    `json:"status"`        // 服務狀態：online, degraded, offline
	LastSeen     time.Time `json:"last_seen"`     // 最後在線時間
	ResponseTime int64     `json:"response_time"` // 響應時間（毫秒）
	Errors       []string  `json:"errors"`        // 錯誤信息
}

// HealthChecker 健康檢查器
type HealthChecker struct {
	deviceStatus  map[string]*DeviceHealth
	serviceStatus map[string]*ServiceHealth
	lastCheck     time.Time
	logPath       string
}

// NewHealthChecker 創建新的健康檢查器
func NewHealthChecker(logPath string) *HealthChecker {
	return &HealthChecker{
		deviceStatus:  make(map[string]*DeviceHealth),
		serviceStatus: make(map[string]*ServiceHealth),
		lastCheck:     time.Now(),
		logPath:       logPath,
	}
}

// CheckSystemHealth 檢查系統健康狀態
func (hc *HealthChecker) CheckSystemHealth() (*HealthStatus, error) {
	status := &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
	}

	// 檢查 CPU 使用率
	cpuUsage, err := hc.getCPUUsage()
	if err != nil {
		status.Errors = append(status.Errors, fmt.Sprintf("CPU 使用率檢查失敗: %v", err))
	} else {
		status.CPUUsage = cpuUsage
		if cpuUsage > 80 {
			status.Errors = append(status.Errors, "CPU 使用率過高")
		}
	}

	// 檢查內存使用率
	memoryUsage, err := hc.getMemoryUsage()
	if err != nil {
		status.Errors = append(status.Errors, fmt.Sprintf("內存使用率檢查失敗: %v", err))
	} else {
		status.MemoryUsage = memoryUsage
		if memoryUsage > 80 {
			status.Errors = append(status.Errors, "內存使用率過高")
		}
	}

	// 檢查磁盤使用率
	diskUsage, err := hc.getDiskUsage()
	if err != nil {
		status.Errors = append(status.Errors, fmt.Sprintf("磁盤使用率檢查失敗: %v", err))
	} else {
		status.DiskUsage = diskUsage
		if diskUsage > 80 {
			status.Errors = append(status.Errors, "磁盤使用率過高")
		}
	}

	// 檢查網絡狀態
	networkStatus, err := hc.checkNetworkStatus()
	if err != nil {
		status.Errors = append(status.Errors, fmt.Sprintf("網絡狀態檢查失敗: %v", err))
	} else {
		status.NetworkStatus = networkStatus
		if networkStatus != "healthy" {
			status.Errors = append(status.Errors, "網絡連接異常")
		}
	}

	// 檢查日誌文件大小
	logSize, err := hc.getLogFileSize()
	if err != nil {
		status.Errors = append(status.Errors, fmt.Sprintf("日誌文件大小檢查失敗: %v", err))
	} else {
		status.LogFileSize = logSize
		if logSize > 100*1024*1024 { // 100MB
			status.Errors = append(status.Errors, "日誌文件過大")
		}
	}

	// 更新系統狀態
	if len(status.Errors) > 0 {
		status.Status = "degraded"
	}

	return status, nil
}

// getCPUUsage 獲取 CPU 使用率
func (hc *HealthChecker) getCPUUsage() (float64, error) {
	if runtime.GOOS == "linux" {
		// 讀取 /proc/stat
		data, err := os.ReadFile("/proc/stat")
		if err != nil {
			return 0, err
		}

		// 解析 CPU 時間
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "cpu ") {
				fields := strings.Fields(line)
				if len(fields) < 8 {
					return 0, fmt.Errorf("invalid cpu stat format")
				}

				// 計算總時間和空閒時間
				total := 0
				for i := 1; i < 8; i++ {
					val, err := strconv.Atoi(fields[i])
					if err != nil {
						return 0, err
					}
					total += val
				}

				idle, err := strconv.Atoi(fields[4])
				if err != nil {
					return 0, err
				}

				// 計算使用率
				usage := 100.0 * (1.0 - float64(idle)/float64(total))
				return usage, nil
			}
		}
		return 0, fmt.Errorf("cpu stat not found")
	}

	// 在其他系統上使用 runtime 包
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return 0, nil // 簡化版本
}

// getMemoryUsage 獲取內存使用率
func (hc *HealthChecker) getMemoryUsage() (float64, error) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return float64(mem.Alloc) / float64(mem.Sys) * 100, nil
}

// getDiskUsage 獲取磁盤使用率
func (hc *HealthChecker) getDiskUsage() (float64, error) {
	if runtime.GOOS == "linux" {
		// 讀取 /proc/mounts 獲取根分區
		data, err := ioutil.ReadFile("/proc/mounts")
		if err != nil {
			return 0, err
		}

		// 解析根分區
		lines := strings.Split(string(data), "\n")
		var rootDevice string
		for _, line := range lines {
			if strings.Contains(line, " / ") {
				fields := strings.Fields(line)
				if len(fields) >= 1 {
					rootDevice = fields[0]
					break
				}
			}
		}

		if rootDevice == "" {
			return 0, fmt.Errorf("root device not found")
		}

		// 使用 df 命令獲取使用率
		cmd := fmt.Sprintf("df %s", rootDevice)
		output, err := exec.Command("sh", "-c", cmd).Output()
		if err != nil {
			return 0, err
		}

		// 解析 df 輸出
		lines = strings.Split(string(output), "\n")
		if len(lines) < 2 {
			return 0, fmt.Errorf("invalid df output")
		}

		fields := strings.Fields(lines[1])
		if len(fields) < 5 {
			return 0, fmt.Errorf("invalid df fields")
		}

		usage, err := strconv.ParseFloat(strings.TrimSuffix(fields[4], "%"), 64)
		if err != nil {
			return 0, err
		}

		return usage, nil
	}

	return 0, fmt.Errorf("unsupported operating system")
}

// checkNetworkStatus 檢查網絡狀態
func (hc *HealthChecker) checkNetworkStatus() (string, error) {
	// 檢查網絡連接
	interfaces, err := net.Interfaces()
	if err != nil {
		return "unhealthy", err
	}

	hasActiveInterface := false
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			hasActiveInterface = true
			break
		}
	}

	if !hasActiveInterface {
		return "unhealthy", fmt.Errorf("no active network interface found")
	}

	return "healthy", nil
}

// getLogFileSize 獲取日誌文件大小
func (hc *HealthChecker) getLogFileSize() (int64, error) {
	if hc.logPath == "" {
		return 0, fmt.Errorf("log path not set")
	}

	info, err := os.Stat(hc.logPath)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

// UpdateDeviceHealth 更新設備健康狀態
func (hc *HealthChecker) UpdateDeviceHealth(ip string, status string, responseTime int64, err error) {
	deviceHealth := &DeviceHealth{
		IP:           ip,
		Status:       status,
		LastSeen:     time.Now(),
		ResponseTime: responseTime,
	}

	if err != nil {
		deviceHealth.Status = "degraded"
		deviceHealth.Errors = append(deviceHealth.Errors, err.Error())
	}

	hc.deviceStatus[ip] = deviceHealth
}

// GetDeviceHealth 獲取設備健康狀態
func (hc *HealthChecker) GetDeviceHealth(ip string) (*DeviceHealth, bool) {
	health, exists := hc.deviceStatus[ip]
	return health, exists
}

// GetAllDeviceHealth 獲取所有設備的健康狀態
func (hc *HealthChecker) GetAllDeviceHealth() map[string]*DeviceHealth {
	return hc.deviceStatus
}

// UpdateServiceHealth 更新服務健康狀態
func (hc *HealthChecker) UpdateServiceHealth(name string, status string, responseTime int64, err error) {
	serviceHealth := &ServiceHealth{
		Name:         name,
		Status:       status,
		LastSeen:     time.Now(),
		ResponseTime: responseTime,
	}

	if err != nil {
		serviceHealth.Status = "degraded"
		serviceHealth.Errors = append(serviceHealth.Errors, err.Error())
	}

	hc.serviceStatus[name] = serviceHealth
}

// GetServiceHealth 獲取服務健康狀態
func (hc *HealthChecker) GetServiceHealth(name string) (*ServiceHealth, bool) {
	health, exists := hc.serviceStatus[name]
	return health, exists
}

// GetAllServiceHealth 獲取所有服務的健康狀態
func (hc *HealthChecker) GetAllServiceHealth() map[string]*ServiceHealth {
	return hc.serviceStatus
}

// GetHealthReport 獲取健康報告
func (hc *HealthChecker) GetHealthReport() string {
	systemHealth, err := hc.CheckSystemHealth()
	if err != nil {
		return fmt.Sprintf("系統健康檢查失敗: %v", err)
	}

	report := fmt.Sprintf("系統健康報告 (%s)\n", systemHealth.Timestamp.Format(time.RFC3339))
	report += fmt.Sprintf("系統狀態: %s\n", systemHealth.Status)
	report += fmt.Sprintf("CPU 使用率: %.2f%%\n", systemHealth.CPUUsage)
	report += fmt.Sprintf("內存使用率: %.2f%%\n", systemHealth.MemoryUsage)
	report += fmt.Sprintf("磁盤使用率: %.2f%%\n", systemHealth.DiskUsage)

	if len(systemHealth.Errors) > 0 {
		report += "\n系統錯誤:\n"
		for _, err := range systemHealth.Errors {
			report += fmt.Sprintf("- %s\n", err)
		}
	}

	report += "\n服務狀態:\n"
	for name, health := range hc.serviceStatus {
		report += fmt.Sprintf("\n服務 %s:\n", name)
		report += fmt.Sprintf("  狀態: %s\n", health.Status)
		report += fmt.Sprintf("  最後在線: %s\n", health.LastSeen.Format(time.RFC3339))
		report += fmt.Sprintf("  響應時間: %dms\n", health.ResponseTime)
		if len(health.Errors) > 0 {
			report += "  錯誤:\n"
			for _, err := range health.Errors {
				report += fmt.Sprintf("  - %s\n", err)
			}
		}
	}

	return report
}

// UpdateHealth 更新服務健康狀態
func (hc *HealthChecker) UpdateHealth(name string, startTime time.Time, err error) {
	responseTime := time.Since(startTime).Milliseconds()
	status := "online"
	if err != nil {
		status = "degraded"
	}
	hc.UpdateServiceHealth(name, status, responseTime, err)
}
