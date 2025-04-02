package services

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"time"
	"viot-scanner/global"
	"viot-scanner/models"
	"viot-scanner/utils"
)

var (
	pending_pdu_header = []string{
		"ip_key",
		"model",
		"ip",
		"protocol",
		"version",
		"serial_number",
		"mac_address",
		"instance_type",
		"manufacturer",
		"uptime",
		"updated_at",
	}
	pending_device_header = []string{
		"ip_key",
		"ip",
		"protocol",
		"model",
		"mac_address",
		"instance_type",
		"manufacturer",
		"uptime",
		"updated_at",
	}
)

func FilteredResults(results []models.ScanInstanceInfo) []models.ScanInstanceInfo {
	var filteredResults []models.ScanInstanceInfo

	// 1. 過濾掉 model 為 unknown 的結果
	for _, result := range results {
		if result.Model == models.DEFAULT_UNKNOWN_VALUE {
			continue
		}
		// 生成 ip_key
		ipKey := fmt.Sprintf("%s:%d", result.IP, result.Port)

		switch result.Protocol {
		case models.PROTOCOL_MODBUS:
			ipKey = fmt.Sprintf("%s:%d:%d", result.IP, result.Port, result.SlaveID)
		case models.PROTOCOL_SNMP:
			ipKey = fmt.Sprintf("%s:%d", result.IP, result.Port)
		}

		// 2. 過濾掉已掃描的結果
		if !slices.Contains(global.ScanIPKeyList, ipKey) {
			result.IPKey = ipKey
			filteredResults = append(filteredResults, result)
		}

		// 3. 過濾掉已註冊的設備
		if _, ok := global.ScanResults[ipKey]; ok {
			continue
		}
	}
	return filteredResults
}

// 儲存結果
func SaveResults(results []models.ScanInstanceInfo) ([]models.ScanInstanceInfo, error) {
	filteredResults := FilteredResults(results)

	// 分類
	pduResult, deviceResult := ClassifyResults(filteredResults)
	if len(pduResult) == 0 && len(deviceResult) == 0 {
		log.Printf("⚠️ 沒有找到任何新設備")
		return nil, nil
	}

	var saveErrors []error

	// 檢查並創建 PDU 結果文件
	if len(pduResult) > 0 {
		if err := saveToCSV(pduResult, true); err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("儲存 PDU 結果失敗: %v", err))
		} else {
			log.Printf("成功儲存 %d 個 PDU 設備結果", len(pduResult))
		}
	}

	// 檢查並創建設備結果文件
	if len(deviceResult) > 0 {
		if err := saveToCSV(deviceResult, false); err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("儲存設備結果失敗: %v", err))
		} else {
			log.Printf("成功儲存 %d 個設備結果", len(deviceResult))
		}
	}

	// 如果有任何保存錯誤，返回錯誤
	if len(saveErrors) > 0 {
		return nil, fmt.Errorf("保存結果時發生 %d 個錯誤: %v", len(saveErrors), saveErrors)
	}

	return filteredResults, nil
}

// saveToCSV 將結果保存到 CSV 文件
func saveToCSV(results []models.ScanInstanceInfo, isPDU bool) error {
	var filename string
	var header []string
	if isPDU {
		filename = models.SCAN_PDU_LIST
		header = pending_pdu_header
	} else {
		filename = models.SCAN_DEVICE_LIST
		header = pending_device_header
	}
	// 檢查文件是否存在
	fileExists := true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fileExists = false
	}

	// 打開或創建文件
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("開啟文件失敗: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 如果文件不存在，寫入標題行
	if !fileExists {
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("寫入標題行失敗: %v", err)
		}
	}

	// 寫入數據
	currentTime := formatTime(time.Now())
	for _, result := range results {
		var record []string

		//ip_key,model,ip,protocol,version,serial_number,mac_address,instance_type,manufacturer,uptime,updated_at
		if isPDU {
			record = []string{
				result.IPKey,                         // ip_key
				result.Model,                         // model
				result.IP,                            // ip
				result.Protocol,                      // protocol
				result.Version,                       // version
				result.SerialNumber,                  // serial_number
				result.MacAddress,                    // mac_address
				result.InstanceType,                  // instance_type
				result.Manufacturer,                  // manufacturer
				strconv.FormatInt(result.Uptime, 10), // uptime
				currentTime,                          // updated_at
			}
		} else {
			record = []string{
				result.IPKey,                         // ip_key
				result.IP,                            // ip
				result.Protocol,                      // protocol
				result.Model,                         // model
				result.MacAddress,                    // mac_address
				result.InstanceType,                  // instance_type
				result.Manufacturer,                  // manufacturer
				strconv.FormatInt(result.Uptime, 10), // uptime
				currentTime,                          // updated_at
			}
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("寫入數據失敗 [%s]: %v", result.IP, err)
		}
	}

	return nil
}

// FilterRegisteredDevices 過濾掉已註冊的設備
func FilterRegisteredDevices(results []models.ScanInstanceInfo) []models.ScanInstanceInfo {
	// 讀取已註冊的設備列表
	registry_pdu_list := utils.ReadCSVByColumn(models.REGISTRY_PDU_LIST, "ip_key")
	registry_device_list := utils.ReadCSVByColumn(models.REGISTRY_DEVICE_LIST, "ip_key")

	// 合併已註冊的設備列表
	registered_devices := append(registry_pdu_list, registry_device_list...)

	// 過濾結果
	var filtered_results []models.ScanInstanceInfo
	for _, result := range results {
		// 生成 ip_key
		ipKey := fmt.Sprintf("%s:%d", result.IP, result.Port)
		switch result.Protocol {
		case models.PROTOCOL_MODBUS:
			ipKey = fmt.Sprintf("%s:%d:%d", result.IP, result.Port, result.SlaveID)
		case models.PROTOCOL_SNMP:
			ipKey = fmt.Sprintf("%s:%d", result.IP, result.Port)
		}

		// 如果設備不在已註冊列表中，則保留
		if !slices.Contains(registered_devices, ipKey) {
			filtered_results = append(filtered_results, result)
		}
	}

	if len(filtered_results) < len(results) {
		log.Printf("🔸[FilterRegisteredDevices]過濾掉 %d 個已註冊設備", len(results)-len(filtered_results))
	}

	return filtered_results
}
