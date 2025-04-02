package global

import (
	"log"
	"slices"
	"sync"

	"viot-scanner/models"
	"viot-scanner/utils"
)

var (
	// 配置相關
	Config         *models.Config        // 主配置
	ModbusSettings models.ModbusSettings // Modbus 設置
	SNMPSettings   models.SNMPSettings   // SNMP 設置
	IPRangeConfigs []models.IPRange      // IP 範圍配置
	TargetSettings models.TargetSettings // 目標設備配置
	Dependencies   []models.Dependencies // 設備依賴關係

	// 掃描器相關
	GlobalScanner     *Scanner                 // 全局掃描器實例
	SNMPInstanceLib   models.SNMPInstanceLib   // SNMP 實例庫
	ModbusInstanceLib models.ModbusInstanceLib // Modbus 實例庫

	// 掃描結果相關
	ScanResultsMutex sync.RWMutex
	ScanResults      map[string]models.ScanInstanceInfo // key = ip_key
	ScanIPKeyList    []string                           // 掃描列表
)

// 初始化全局變量
func Init(config *models.Config) {
	Config = config
	ModbusSettings = config.Modbus
	SNMPSettings = config.SNMP
	IPRangeConfigs = make([]models.IPRange, 0)
	TargetSettings = models.TargetSettings{}
	Dependencies = config.Dependencies

	// 初始化掃描器
	GlobalScanner = NewScanner(config)

	// 初始化掃描結果
	ScanResults = make(map[string]models.ScanInstanceInfo)
}

// 初始化掃描列表
func InitScanList() {
	var scan_list []string
	scan_pdu_file := models.SCAN_PDU_LIST
	scan_device_file := models.SCAN_DEVICE_LIST

	// 讀取 csv "ip_key" 欄位的值
	scan_pdu_list := utils.ReadCSVByColumn(scan_pdu_file, "ip_key")
	scan_device_list := utils.ReadCSVByColumn(scan_device_file, "ip_key")

	scan_list = append(scan_pdu_list, scan_device_list...)
	if len(scan_list) == 0 {
		log.Printf("⚠️ 警告: 沒有找到任何待掃描的設備")
	} else {
		log.Printf("✅ 成功讀取待掃描列表: PDU %d 個, 設備 %d 個", len(scan_pdu_list), len(scan_device_list))
	}
	ScanIPKeyList = scan_list
}

// 更新掃描結果 ScanIPKeyList 和 ScanResults
func UpdateScanResults(results []models.ScanInstanceInfo) {
	ScanResultsMutex.Lock()
	defer ScanResultsMutex.Unlock()
	// 將結果轉換為 map
	updatedScanResults := make(map[string]models.ScanInstanceInfo)
	for _, result := range results {
		ipKey := result.IPKey
		updatedScanResults[ipKey] = result

		if !slices.Contains(ScanIPKeyList, ipKey) {
			ScanIPKeyList = append(ScanIPKeyList, ipKey)
		}
	}
	ScanResults = updatedScanResults
}

// 獲取掃描結果
func GetScanResults() map[string]models.ScanInstanceInfo {
	ScanResultsMutex.RLock()
	defer ScanResultsMutex.RUnlock()
	return ScanResults
}

// 根據 ip_key 獲取掃描結果
func GetScanResultByIPKey(ipKey string) (models.ScanInstanceInfo, bool) {
	ScanResultsMutex.RLock()
	defer ScanResultsMutex.RUnlock()
	result, exists := ScanResults[ipKey]
	return result, exists
}
