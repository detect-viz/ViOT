package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// Field 表示日誌欄位
type Field struct {
	Key   string
	Value interface{}
}

// String 創建字串型日誌欄位
func String(key string, value string) zap.Field {
	return zap.String(key, value)
}

// Int 創建整數型日誌欄位
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Int64 創建一個 int64 欄位
func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

// Float64 創建一個 float64 欄位
func Float64(key string, value float64) zap.Field {
	return zap.Float64(key, value)
}

// Bool 創建一個布爾欄位
func Bool(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}

// Any 創建任意類型日誌欄位
func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// Strings 創建字串陣列型日誌欄位
func Strings(key string, values []string) zap.Field {
	return zap.Strings(key, values)
}

// Err 創建一個錯誤欄位
func Err(err error) zap.Field {
	return zap.Error(err)
}

// FormatFields 格式化日誌欄位
func FormatFields(fields ...Field) string {
	result := ""
	for i, field := range fields {
		if i > 0 {
			result += " "
		}
		result += fmt.Sprintf("%s=%v", field.Key, field.Value)
	}
	return result
}
