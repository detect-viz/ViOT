package models

import "time"

// OutputResult 輸出結果
type OutputResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// OutputStatus 輸出狀態
type OutputStatus struct {
	Running        bool      `json:"running"`
	LastOutput     time.Time `json:"last_output"`
	LastBackup     time.Time `json:"last_backup"`
	ProcessedCount int64     `json:"processed_count"`
	ErrorCount     int64     `json:"error_count"`
	LastError      string    `json:"last_error,omitempty"`
}

// OutputInstance 輸出實例
type OutputInstance struct {
	Name      string       `json:"name"`
	Type      string       `json:"type"`
	Config    OutputConfig `json:"config"`
	Status    OutputStatus `json:"status"`
	StartTime time.Time    `json:"start_time"`
}
