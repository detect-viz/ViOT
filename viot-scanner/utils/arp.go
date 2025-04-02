package utils

import (
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/mdlayher/arp"
)

// GetMacByArp 使用 ARP 協議獲取 MAC 地址
func GetMacByArp(ipStr string) (string, error) {
	// 解析 IP 地址
	addr, err := netip.ParseAddr(ipStr)
	if err != nil {
		return "", fmt.Errorf("無效的 IP 地址: %s", err)
	}

	// 獲取所有網絡接口
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("獲取網絡接口失敗: %v", err)
	}

	// 遍歷每個接口
	for _, iface := range ifaces {
		// 跳過關閉或不支援廣播的接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagBroadcast == 0 {
			continue
		}

		// 創建 ARP 客戶端
		client, err := arp.Dial(&iface)
		if err != nil {
			continue
		}
		defer client.Close()

		// 設置超時
		client.SetDeadline(time.Now().Add(1 * time.Second))

		// 解析 MAC 地址
		hwAddr, err := client.Resolve(addr)
		if err == nil && len(hwAddr) == 6 {
			return hwAddr.String(), nil
		}
	}

	return "", fmt.Errorf("無法獲取 MAC 地址")
}
