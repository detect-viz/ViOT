package processor

// ProcessorConfig 處理器配置
type ProcessorConfig struct {
	Enabled    bool                   `json:"enabled"`
	Type       string                 `json:"type"`
	Interval   int                    `json:"interval"`
	BatchSize  int                    `json:"batch_size"`
	RetryCount int                    `json:"retry_count"`
	Options    map[string]interface{} `json:"options"`
}
