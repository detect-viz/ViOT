package services

import (
	"encoding/json"
	"net/http"
	"time"
	"viot-scanner/global"
	"viot-scanner/health"
)

// UpdateScanIntervalRequest 更新掃描間隔的請求
type UpdateScanIntervalRequest struct {
	Room     string `json:"room"`
	Interval int    `json:"interval"` // 掃描間隔（秒）
	Duration int    `json:"duration"` // 持續時間（分鐘）
}

// UpdateScanIntervalResponse 更新掃描間隔的響應
type UpdateScanIntervalResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetScanIntervalRequest 獲取掃描間隔的請求
type GetScanIntervalRequest struct {
	Room string `json:"room"`
}

// GetScanIntervalResponse 獲取掃描間隔的響應
type GetScanIntervalResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Interval int    `json:"interval"`
}

// HealthCheckResponse 健康檢查響應
type HealthCheckResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Status  string `json:"status"`
	Report  string `json:"report"`
}

// StartAPIServer 啟動 API 服務器
func StartAPIServer(scanner *global.Scanner, port string, healthChecker *health.HealthChecker) error {
	// 更新掃描間隔的 API
	http.HandleFunc("/api/scan/interval", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "只支援 POST 方法", http.StatusMethodNotAllowed)
			return
		}

		var req UpdateScanIntervalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "無效的請求格式", http.StatusBadRequest)
			return
		}

		// 驗證請求
		if req.Room == "" || req.Interval <= 0 || req.Duration <= 0 {
			http.Error(w, "無效的參數", http.StatusBadRequest)
			return
		}

		// 更新掃描間隔
		duration := time.Duration(req.Duration) * time.Minute
		scanner.AddTemporaryScanTask(req.Room, duration)
		scanner.UpdateRoomSchedule(req.Room, req.Interval)

		json.NewEncoder(w).Encode(UpdateScanIntervalResponse{
			Success: true,
			Message: "成功更新掃描間隔",
		})
	})

	// 獲取掃描間隔的 API
	http.HandleFunc("/api/scan/interval/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "只支援 POST 方法", http.StatusMethodNotAllowed)
			return
		}

		var req GetScanIntervalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "無效的請求格式", http.StatusBadRequest)
			return
		}

		// 驗證請求
		if req.Room == "" {
			http.Error(w, "無效的參數", http.StatusBadRequest)
			return
		}

		interval := scanner.GetRoomScanInterval(req.Room)
		json.NewEncoder(w).Encode(GetScanIntervalResponse{
			Success:  true,
			Message:  "成功獲取掃描間隔",
			Interval: interval,
		})
	})

	// 健康檢查的 API
	http.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "只支援 GET 方法", http.StatusMethodNotAllowed)
			return
		}

		report := healthChecker.GetHealthReport()
		status := "healthy"
		if report == "" {
			status = "unknown"
		}

		json.NewEncoder(w).Encode(HealthCheckResponse{
			Success: true,
			Message: "服務健康狀態",
			Status:  status,
			Report:  report,
		})
	})

	return http.ListenAndServe(":"+port, nil)
}
