package services

import (
	"fmt"
	"time"
	"viot-scanner/global"
	"viot-scanner/models"

	"github.com/gosnmp/gosnmp"
)

// NewSnmpClient 創建 SNMP 客戶端
func NewSnmpClient(ip string, port int, community string, version gosnmp.SnmpVersion) (*gosnmp.GoSNMP, error) {
	// fmt.Printf("正在連接到 SNMP 設備 %s:%d...\n", ip, port)

	snmp := &gosnmp.GoSNMP{
		Target:    ip,
		Port:      uint16(port),
		Community: community,
		Version:   version,
		Timeout:   time.Duration(global.Config.SNMP.Timeout) * time.Second,
	}

	// 確保在返回錯誤時不會留下未關閉的連接
	err := snmp.Connect()
	if err != nil {
		return nil, fmt.Errorf("SNMP 連接失敗: %v", err)
	}

	return snmp, nil
}

func TestSnmpConnect(ip string, port int, community string) (bool, int) {
	// 如果 port 為 0，使用預設的 SNMP 端口 161
	if port == 0 {
		port = 161
	}

	// 先嘗試 SNMP v2c
	snmp, err := NewSnmpClient(ip, port, community, gosnmp.Version2c)
	if err == nil {
		// 嘗試獲取 sysDescr OID 來驗證連接
		oid := "1.3.6.1.2.1.1.1.0" // sysDescr
		result, err := snmp.Get([]string{oid})
		snmp.Conn.Close()
		if err == nil && len(result.Variables) > 0 && result.Variables[0].Value != nil {
			return true, 2
		}
	}

	// 如果 v2c 失敗，嘗試 v1
	snmp, err = NewSnmpClient(ip, port, community, gosnmp.Version1)
	if err == nil {
		// 嘗試獲取 sysDescr OID 來驗證連接
		oid := "1.3.6.1.2.1.1.1.0" // sysDescr
		result, err := snmp.Get([]string{oid})
		snmp.Conn.Close()
		if err == nil && len(result.Variables) > 0 && result.Variables[0].Value != nil {
			return true, 1
		}
	}

	return false, 0
}

// GetSnmpValue 使用 SNMP 獲取設備信息
func GetSnmpValue(snmpSession models.SnmpSession) (map[string]interface{}, error) {
	var version gosnmp.SnmpVersion
	switch snmpSession.Conn.Version {
	case 1:
		version = gosnmp.Version1
	case 2:
		version = gosnmp.Version2c
	case 3:
		version = gosnmp.Version3
	}

	snmp, err := NewSnmpClient(snmpSession.Conn.IP, snmpSession.Conn.Port, snmpSession.Conn.Community, version)
	if err != nil {
		return nil, fmt.Errorf("❌ 創建 SNMP 客戶端失敗 [IP:%s, Port:%d, Community:%s]: %v", snmpSession.Conn.IP, snmpSession.Conn.Port, snmpSession.Conn.Community, err)
	}

	// 連接設備
	err = snmp.Connect()
	if err != nil {
		return nil, fmt.Errorf("SNMP 連接失敗: %v", err)
	}
	defer snmp.Conn.Close()

	snmpOids := make([]string, 0, len(snmpSession.Requests))
	for _, request := range snmpSession.Requests {
		snmpOids = append(snmpOids, request.OID)
	}

	result, err := snmp.Get(snmpOids)
	if err != nil {
		return nil, fmt.Errorf("SNMP 獲取失敗: %v", err)
	}

	// 解析結果
	info := make(map[string]interface{})

	for _, request := range snmpSession.Requests {
		oidIndex := 0
		for i, oid := range snmpOids {
			if oid == request.OID {
				oidIndex = i
				break
			}
		}

		info[request.Name] = result.Variables[oidIndex].Value
	}
	return info, nil
}
