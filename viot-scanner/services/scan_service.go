package services

import (
	"fmt"
	"log"
	"sync"
	"time"
	"viot-scanner/global"
	"viot-scanner/models"
	"viot-scanner/utils"
)

// RunScanRoom 執行掃描房間
func RunScanRoom() (map[string]models.ScanActiveIPInfo, error) {
	allScanActiveIPInfo := make(map[string]models.ScanActiveIPInfo)
	ipRanges := global.IPRangeConfigs
	scanner := global.Config.Scanner
	var wg sync.WaitGroup
	var mu sync.Mutex // 添加互斥鎖保護 map 寫入
	errChan := make(chan error, len(ipRanges))

	// 使用信號量控制並發數
	semaphore := make(chan struct{}, scanner.MaxConcurrent)

	for _, ipRange := range ipRanges {
		wg.Add(1)
		semaphore <- struct{}{} // 獲取信號量

		go func(room models.IPRange) {
			defer wg.Done()
			defer func() { <-semaphore }() // 釋放信號量

			// 設置超時控制
			done := make(chan struct{})
			go func() {
				defer close(done)
				defer func() {
					if r := recover(); r != nil {
						errChan <- fmt.Errorf("掃描房間時發生錯誤 [%s]: %v", room.Room, r)
					}
				}()

				// 計算 CIDR
				cidr, err := utils.CalculateCIDR(room.StartIP, room.EndIP)
				if err != nil {
					errChan <- fmt.Errorf("計算 CIDR 失敗 [%s-%s]: %v", room.StartIP, room.EndIP, err)
					return
				}

				// 檢查 IP 存活狀態
				activeIPs, err := utils.ExecuteFping(cidr, scanner.FPingRetries, scanner.FPingTimeout)
				if err != nil {
					errChan <- fmt.Errorf("執行 fping 失敗 [%s]: %v", room.Room, err)
					return
				}

				if len(activeIPs) == 0 {
					errChan <- fmt.Errorf("房間 %s 沒有存活的 IP 地址", room.Room)
					return
				}

				// 使用互斥鎖保護 map 寫入
				mu.Lock()
				for _, ip := range activeIPs {
					allScanActiveIPInfo[ip] = models.ScanActiveIPInfo{
						Factory:    room.Factory,
						Phase:      room.Phase,
						Datacenter: room.Datacenter,
						Room:       room.Room,
					}
				}
				mu.Unlock()
			}()

			// 等待完成或超時
			select {
			case <-done:
				return
			case <-time.After(time.Duration(scanner.FPingTimeout) * time.Second):
				errChan <- fmt.Errorf("掃描房間超時 [%s]", room.Room)
			}
		}(ipRange)
	}

	// 等待所有 goroutine 完成
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 收集錯誤
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
		log.Printf("掃描錯誤: %v", err)
	}

	// 如果有錯誤，返回第一個錯誤
	if len(errors) > 0 {
		return allScanActiveIPInfo, errors[0]
	}

	return allScanActiveIPInfo, nil
}
