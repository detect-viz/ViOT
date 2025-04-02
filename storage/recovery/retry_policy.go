package recovery

import (
	"net"
	"strings"
)

// ExponentialBackoffRetry 實現指數退避重試策略
type ExponentialBackoffRetry struct {
	// 最大重試次數
	MaxRetries int
	// 初始延遲(秒)
	InitialDelay int
	// 最大延遲(秒)
	MaxDelay int
	// 退避因子
	BackoffFactor float64
}

// NewExponentialBackoffRetry 創建一個新的指數退避重試策略
func NewExponentialBackoffRetry(maxRetries, initialDelay, maxDelay int, backoffFactor float64) *ExponentialBackoffRetry {
	return &ExponentialBackoffRetry{
		MaxRetries:    maxRetries,
		InitialDelay:  initialDelay,
		MaxDelay:      maxDelay,
		BackoffFactor: backoffFactor,
	}
}

// DefaultRetryPolicy 返回默認的重試策略
func DefaultRetryPolicy() *ExponentialBackoffRetry {
	return &ExponentialBackoffRetry{
		MaxRetries:    5,
		InitialDelay:  1,   // 1秒
		MaxDelay:      300, // 5分鐘
		BackoffFactor: 2.0, // 每次延遲加倍
	}
}

// ShouldRetry 判斷是否應該重試
func (r *ExponentialBackoffRetry) ShouldRetry(attempt int, err error) bool {
	// 超過最大重試次數
	if attempt >= r.MaxRetries {
		return false
	}

	// 根據錯誤類型判斷是否可重試
	return isRetryableError(err)
}

// NextRetryDelay 返回下次重試的延遲時間(秒)
func (r *ExponentialBackoffRetry) NextRetryDelay(attempt int) int {
	// 計算延遲時間（指數增長）
	delay := float64(r.InitialDelay)
	for i := 0; i < attempt; i++ {
		delay *= r.BackoffFactor
	}

	// 確保不超過最大延遲
	if int(delay) > r.MaxDelay {
		return r.MaxDelay
	}

	return int(delay)
}

// 判斷錯誤是否可以重試
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 網絡連接錯誤通常可以重試
	if netErr, ok := err.(net.Error); ok {
		// 超時錯誤
		if netErr.Timeout() {
			return true
		}
		// 臨時錯誤
		if netErr.Temporary() {
			return true
		}
	}

	// 檢查常見的可重試錯誤信息
	errMsg := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"connection closed",
		"i/o timeout",
		"broken pipe",
		"temporary failure",
		"server unavailable",
		"service unavailable",
		"database is locked",
		"too many connections",
		"no space left on device",
		"too many open files",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}
