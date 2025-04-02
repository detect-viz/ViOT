package utils

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

// 檢查存活 IP 是否開啟指定 port
func CheckPort(activeIPs []string, protocol string, ports []int, timeout int, worker int) []string {
	var wg sync.WaitGroup
	resultChan := make(chan string, len(activeIPs))
	semaphore := make(chan struct{}, worker) // 控制併發數量

	for _, ip := range activeIPs {
		wg.Add(1)
		semaphore <- struct{}{} // 獲取信號量

		go func(ip string) {
			defer wg.Done()
			defer func() { <-semaphore }() // 釋放信號量

			for _, port := range ports {
				// 處理 IPv6 地址
				addr := ip
				if net.ParseIP(ip).To4() == nil {
					addr = fmt.Sprintf("[%s]", ip)
				}
				address := addr + ":" + strconv.Itoa(port)
				conn, err := net.DialTimeout(protocol, address, time.Duration(timeout)*time.Second)
				if err == nil {
					resultChan <- ip
					conn.Close()
					break
				}
			}
		}(ip)
	}

	// 等待所有 goroutine 完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集結果
	var enabledIPs []string
	for ip := range resultChan {
		enabledIPs = append(enabledIPs, ip)
	}

	return enabledIPs
}
