package scanner

import (
	"fmt"
	"net"
	"strings"
)

// ExpandIPRange 展開IP範圍為具體IP地址列表
func ExpandIPRange(ipRange string) ([]string, error) {
	var ips []string

	// 檢查是否是CIDR格式
	if strings.Contains(ipRange, "/") {
		ip, ipnet, err := net.ParseCIDR(ipRange)
		if err != nil {
			return nil, err
		}

		// 從CIDR中獲取所有IP
		for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
			ips = append(ips, ip.String())
		}

		// 移除網絡地址和廣播地址
		if len(ips) > 2 {
			ips = ips[1 : len(ips)-1]
		}
	} else if strings.Contains(ipRange, "-") {
		// 處理範圍格式: 192.168.1.1-192.168.1.10
		parts := strings.Split(ipRange, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("無效的IP範圍格式: %s", ipRange)
		}

		startIP := net.ParseIP(strings.TrimSpace(parts[0]))
		endIP := net.ParseIP(strings.TrimSpace(parts[1]))

		if startIP == nil || endIP == nil {
			return nil, fmt.Errorf("無效的IP地址: %s", ipRange)
		}

		// 確保兩個IP是IPV4
		startIP = startIP.To4()
		endIP = endIP.To4()

		if startIP == nil || endIP == nil {
			return nil, fmt.Errorf("IP範圍必須是IPv4: %s", ipRange)
		}

		// 確保開始IP <= 結束IP
		if bytes4ToUint32(startIP) > bytes4ToUint32(endIP) {
			return nil, fmt.Errorf("開始IP必須小於或等於結束IP: %s", ipRange)
		}

		// 生成IP列表
		for ip := cloneIP(startIP); bytes4ToUint32(ip) <= bytes4ToUint32(endIP); incIP(ip) {
			ips = append(ips, ip.String())
		}
	} else {
		// 單一IP
		ip := net.ParseIP(strings.TrimSpace(ipRange))
		if ip == nil {
			return nil, fmt.Errorf("無效的IP地址: %s", ipRange)
		}
		ips = append(ips, ip.String())
	}

	return ips, nil
}

// incIP 增加IP地址
func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// cloneIP 複製IP地址
func cloneIP(ip net.IP) net.IP {
	clone := make(net.IP, len(ip))
	copy(clone, ip)
	return clone
}

// bytes4ToUint32 將IPv4地址轉換為uint32
func bytes4ToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}
