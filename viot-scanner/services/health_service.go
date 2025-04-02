package services

import (
	"fmt"
	"log"
	"time"
	"viot-scanner/health"
)

// 更新健康狀態
func UpdateHealthStatus(healthChecker *health.HealthChecker, startTime time.Time, errors []error) {
	if len(errors) > 0 {
		healthChecker.UpdateHealth("viot-scanner", startTime, fmt.Errorf("掃描過程中發生 %d 個錯誤", len(errors)))
	} else {
		healthChecker.UpdateHealth("viot-scanner", startTime, nil)
	}
}

// 定期執行健康檢查
func StartHealthCheck(healthChecker *health.HealthChecker) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			report := healthChecker.GetHealthReport()
			log.Println("健康檢查報告:\n", report)
		}
	}()
}
