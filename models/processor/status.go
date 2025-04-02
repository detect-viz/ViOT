package processor

import "time"

// ProcessorStatus 處理器狀態
type ProcessorStatus struct {
	Running        bool      `json:"running"`
	StartTime      time.Time `json:"start_time"`
	LastProcessed  time.Time `json:"last_processed"`
	ProcessedCount int64     `json:"processed_count"`
	ErrorCount     int64     `json:"error_count"`
	LastError      string    `json:"last_error,omitempty"`
	Stats          map[string]interface{}
}
