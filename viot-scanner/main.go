package main

import (
	"fmt"
	"log"
	"sync"
	"time"
	"viot-scanner/global"
	"viot-scanner/health"
	"viot-scanner/models"
	"viot-scanner/services"
)

var (
	healthChecker = health.NewHealthChecker("logs/viot-scanner.log")
)

func main() {
	log.Println("🔻🔻🔻🔻🔻 掃描流程開始 🔻🔻🔻🔻🔻")

	startTime := time.Now()

	// 使用 defer 確保在程序結束時更新健康狀態
	defer func() {
		if r := recover(); r != nil {
			log.Printf("程序發生嚴重錯誤: %v", r)
			healthChecker.UpdateHealth("viot-scanner", startTime, fmt.Errorf("程序崩潰: %v", r))
		}
	}()

	// 1. 載入 Scanner 設定檔
	err := services.LoadConfig("config/config.yaml")
	if err != nil {
		healthChecker.UpdateHealth("viot-scanner", startTime, err)
		log.Fatalf("載入設定檔失敗: %v", err)
	}

	// 主循環
	ticker := time.NewTicker(time.Duration(global.GlobalScanner.GetScanInterval("global")) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if err := runScanProcess(startTime, healthChecker); err != nil {
			log.Printf("掃描流程執行失敗: %v", err)
		}
	}
}

// runScanProcess 執行掃描流程
func runScanProcess(startTime time.Time, healthChecker *health.HealthChecker) error {
	// 初始化掃描列表
	global.InitScanList()

	// 2. 掃描 IP 範圍
	activeIPInfos, err := services.RunScanRoom()
	if err != nil {
		healthChecker.UpdateHealth("viot-scanner", startTime, err)
		log.Printf("IP 範圍掃描失敗: %v", err)
		return err
	}

	var activeIPs []string
	for ip := range activeIPInfos {
		activeIPs = append(activeIPs, ip)
	}

	if len(activeIPs) == 0 {
		log.Println("沒有找到存活的 IP 地址")
		healthChecker.UpdateHealth("viot-scanner", startTime, nil)
		return nil
	}

	// 3. 檢查 SNMP 可用性
	snmpAvailableInfos := []models.SnmpConnInfo{}
	for _, ip := range activeIPs {
		snmpAvailable, version := services.TestSnmpConnect(ip, global.SNMPSettings.Port, global.Config.SNMP.Community)
		if snmpAvailable {
			snmpConnInfo := models.SnmpConnInfo{
				IP:        ip,
				Port:      global.SNMPSettings.Port,
				Community: global.Config.SNMP.Community,
				Version:   version,
			}
			snmpAvailableInfos = append(snmpAvailableInfos, snmpConnInfo)
		}
	}

	// 過濾掉 pending_pdu_list 或 pending_device_list 中的 IP
	snmpNewInfos := services.FilterScanList(snmpAvailableInfos)

	if len(snmpNewInfos) == 0 {
		log.Println("沒有找到可用的 SNMP 設備")
		healthChecker.UpdateHealth("viot-scanner", startTime, nil)
		return nil
	} else {
		log.Printf("找到 %d 個可用的 SNMP 設備", len(snmpNewInfos))
	}

	// 4. 並發處理
	var wg sync.WaitGroup
	resultChan := make(chan models.ScanInstanceInfo, len(snmpAvailableInfos)*2)
	semaphore := make(chan struct{}, global.Config.Scanner.MaxConcurrent)
	errChan := make(chan error, len(snmpAvailableInfos))

	// 處理每個 SNMP 可用的 IP
	for _, snmpConnInfo := range snmpAvailableInfos {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(snmpConnInfo models.SnmpConnInfo) {
			defer wg.Done()
			defer func() { <-semaphore }()

			results, err := services.ScanDevice(snmpConnInfo)
			if err != nil {
				errChan <- err
				return
			}

			for _, result := range results {
				resultChan <- result
			}
		}(snmpConnInfo)
	}

	// 等待所有 goroutine 完成
	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	// 收集結果和錯誤
	var scanErrors []error

	// 收集錯誤
	for err := range errChan {
		scanErrors = append(scanErrors, err)
		log.Printf("掃描錯誤: %v", err)
	}

	// 收集結果
	allResults := services.CollectResults(resultChan)

	for i, result := range allResults {
		// 輸出詳細的連接信息
		fmt.Println("\n=== [", i, "] SNMP 信息 [", result.IP, "]===")
		fmt.Printf("ip: %s\n", result.IP)
		fmt.Printf("protocol: %s\n", result.Protocol)
		fmt.Printf("mac_address: %s\n", result.MacAddress)
		fmt.Printf("version: v%s\n", result.Version)
		fmt.Printf("instance_type: %s\n", result.InstanceType)
		fmt.Printf("model: %s\n", result.Model)
		fmt.Printf("serial: %s\n", result.SerialNumber)
		fmt.Printf("uptime: %d\n", result.Uptime)
		fmt.Printf("manufacturer: %s\n", result.Manufacturer)
		fmt.Printf("factory: %s\n", result.Factory)
		fmt.Printf("sys_name: %s\n", result.SysName)
		fmt.Printf("sys_descr: %s\n", result.SysDescr)
		fmt.Printf("snmp_engine_id: %s\n", result.SnmpEngineID)

	}

	// 5. 儲存結果
	saveResults, err := services.SaveResults(allResults)
	if err != nil {
		log.Printf("儲存結果失敗: %v", err)
		scanErrors = append(scanErrors, err)
	}

	// 6. 更新全局結果
	global.UpdateScanResults(saveResults)

	// 更新健康狀態
	services.UpdateHealthStatus(healthChecker, startTime, scanErrors)

	log.Printf("✅ 掃描流程結束，耗時 %s", time.Since(startTime))
	if len(scanErrors) > 0 {
		log.Printf("⚠️ 掃描過程中發生 %d 個錯誤", len(scanErrors))
	}

	// 清理過期任務
	global.GlobalScanner.CleanupExpiredSchedules()

	// 定期執行健康檢查
	services.StartHealthCheck(healthChecker)
	return nil
}
