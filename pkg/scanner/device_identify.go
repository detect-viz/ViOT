package scanner

import (
	"fmt"
	"viot/models"
)

// identifyDevice 識別設備
func (s *scanManagerImpl) identifyDevice(ip string, config *models.ScanSetting) (*models.DeviceInfo, error) {
	// 創建設備信息
	deviceInfo := &models.DeviceInfo{
		IP: ip,
	}

	// 填充區域信息
	zone := s.findZoneForIP(ip)
	if zone != nil {
		deviceInfo.Factory = zone.Factory
		deviceInfo.Phase = zone.Phase
		deviceInfo.Datacenter = zone.Datacenter
		deviceInfo.Room = zone.Room
	}

	// 嘗試SNMP識別
	if scanner, ok := s.scanners["snmp"]; ok {
		for _, port := range config.SNMPPorts {
			result, err := scanner.Scan(s.ctx, ip, port)
			if err == nil && result != nil {
				// 複製相關信息
				*deviceInfo = *result
				// 確保保留區域信息
				if zone != nil {
					deviceInfo.Factory = zone.Factory
					deviceInfo.Phase = zone.Phase
					deviceInfo.Datacenter = zone.Datacenter
					deviceInfo.Room = zone.Room
				}
				return deviceInfo, nil
			}
		}
	}

	// 嘗試Modbus識別
	if scanner, ok := s.scanners["modbus"]; ok {
		for _, port := range config.ModbusPorts {
			result, err := scanner.Scan(s.ctx, ip, port)
			if err == nil && result != nil {
				// 複製相關信息
				*deviceInfo = *result
				// 確保保留區域信息
				if zone != nil {
					deviceInfo.Factory = zone.Factory
					deviceInfo.Phase = zone.Phase
					deviceInfo.Datacenter = zone.Datacenter
					deviceInfo.Room = zone.Room
				}
				return deviceInfo, nil
			}
		}
	}

	return nil, fmt.Errorf("無法識別設備")
}
