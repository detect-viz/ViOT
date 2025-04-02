package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"math/bits"
	"net"
	"os/exec"
	"strings"
)

// CalculateCIDR 計算能涵蓋 startIP 到 endIP 的最小 CIDR 區塊
func CalculateCIDR(startIP, endIP string) (string, error) {
	start := net.ParseIP(startIP).To4()
	end := net.ParseIP(endIP).To4()

	if start == nil || end == nil {
		return "", fmt.Errorf("無效的 IPv4 位址: %s 或 %s", startIP, endIP)
	}

	startInt := ipToUint32(start)
	endInt := ipToUint32(end)

	if startInt > endInt {
		return "", fmt.Errorf("起始 IP 大於結束 IP: %s > %s", startIP, endIP)
	}

	// 計算需要幾個共同前綴位元
	diff := startInt ^ endInt
	maskLen := 32 - bits.Len32(diff)

	network := startInt & (^uint32(0) << (32 - maskLen))
	networkIP := uint32ToIP(network)

	return fmt.Sprintf("%s/%d", networkIP.String(), maskLen), nil
}

// ExpandRange 將起始與結束 IP 擴展成完整 IP 清單
func ExpandRange(startIP, endIP string) ([]string, error) {
	start := net.ParseIP(startIP).To4()
	end := net.ParseIP(endIP).To4()

	if start == nil || end == nil {
		return nil, fmt.Errorf("無效 IP 位址: %s 或 %s", startIP, endIP)
	}

	startInt := ipToUint32(start)
	endInt := ipToUint32(end)

	if startInt > endInt {
		return nil, fmt.Errorf("起始 IP 大於結束 IP: %s > %s", startIP, endIP)
	}

	var ips []string
	for i := startInt; i <= endInt; i++ {
		ips = append(ips, uint32ToIP(i).String())
	}
	return ips, nil
}

// ExecuteFping 執行 fping 命令並返回存活的 IP 列表
func ExecuteFping(cidr string, retries, timeout int) ([]string, error) {
	// 構建 fping 命令
	args := []string{
		"-C", fmt.Sprintf("%d", retries),
		"-t", fmt.Sprintf("%d", timeout),
		"-r", "0",
		"-A",
		"-q",
		"-g", cidr,
	}

	// 執行命令
	cmd := exec.Command("fping", args...)

	// 記錄原始輸出
	log.Printf("fping command: %s", cmd.String())

	output, _ := cmd.CombinedOutput()
	//fmt.Printf("err: %v\n", err)
	//var aliveHosts []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	var activeIPs []string
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) != 2 || strings.Contains(parts[1], "-") {
			continue
		}

		ip := strings.TrimSpace(parts[0])
		activeIPs = append(activeIPs, ip)
	}
	log.Printf("===> [fping] 找到 %d 個存活的 IP", len(activeIPs))
	return activeIPs, nil
}

// HasAliveIPsInCIDR 判斷 CIDR 中是否有存活 IP
func HasAliveIPsInCIDR(cidr string, retries, timeout int) (bool, []string, error) {
	aliveIPs, err := ExecuteFping(cidr, retries, timeout)
	if err != nil {
		return false, nil, err
	}
	return len(aliveIPs) > 0, aliveIPs, nil
}

// HasAliveIPsInRange 判斷起始與結束 IP 範圍中是否有存活 IP
func HasAliveIPsInRange(startIP, endIP string, retries, timeout int) (bool, []string, error) {
	// 取得 CIDR
	cidr, err := CalculateCIDR(startIP, endIP)
	if err != nil {
		return false, nil, fmt.Errorf("CIDR 計算失敗 (%s - %s): %v", startIP, endIP, err)
	}
	log.Printf("範圍 %s - %s 對應 CIDR: %s", startIP, endIP, cidr)

	// 執行 fping
	aliveIPs, err := ExecuteFping(cidr, retries, timeout)
	if err != nil {
		return false, nil, fmt.Errorf("fping 掃描失敗: %v", err)
	}

	// 擴展出目標範圍內所有 IP
	targetRange, err := ExpandRange(startIP, endIP)
	if err != nil {
		return false, nil, fmt.Errorf("IP 範圍展開失敗: %v", err)
	}
	log.Printf("預期應包含 IP 數量: %d", len(targetRange))

	targetSet := make(map[string]struct{})
	for _, ip := range targetRange {
		targetSet[ip] = struct{}{}
	}

	var matched []string
	for _, ip := range aliveIPs {
		if _, ok := targetSet[ip]; ok {
			log.Printf("✅ 命中 R0 範圍內存活 IP: %s", ip)
			matched = append(matched, ip)
		} else {
			log.Printf("⚠️ IP %s 不屬於目標範圍，略過", ip)
		}
	}

	if len(matched) == 0 {
		log.Printf("❌ 無任何存活 IP 屬於 %s - %s 範圍內", startIP, endIP)
	} else {
		log.Printf("✅ 共找到 %d 個存活 IP 屬於目標範圍", len(matched))
	}

	return len(matched) > 0, matched, nil
}

// ipToUint32 將 IP 地址轉換為 uint32
func ipToUint32(ip net.IP) uint32 {
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// uint32ToIP 將 uint32 轉換為 IP 地址
func uint32ToIP(n uint32) net.IP {
	return net.IPv4(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
}
