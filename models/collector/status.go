package collector

import "time"

// CollectorStatus 收集器狀態
type CollectorStatus struct {
	Running       bool      `json:"running"`
	LastCollected time.Time `json:"last_collected"`
	DataPoints    int64     `json:"data_points"`
	ErrorCount    int64     `json:"error_count"`
	SuccessRate   float64   `json:"success_rate"`
}
