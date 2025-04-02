package services

import (
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"viot-scanner/global"
	"viot-scanner/models"
	"viot-scanner/utils"
)

// 掃描設備並收集結果
func ScanDevice(snmpConnInfo models.SnmpConnInfo) ([]models.ScanInstanceInfo, error) {
	var results []models.ScanInstanceInfo
	var scanErrors []error

	// 設置超時控制
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				scanErrors = append(scanErrors, fmt.Errorf("掃描過程發生錯誤 [%s]: %v", snmpConnInfo.IP, r))
			}
		}()

		// SNMP 掃描
		snmpScanResult, err := MatchSnmpInstances(snmpConnInfo)
		if snmpScanResult.Model == models.DEFAULT_UNKNOWN_VALUE {
			log.Printf("⛔[%s:%d]未找到符合關鍵字的未知設備", snmpConnInfo.IP, snmpConnInfo.Port)
		}
		results = append(results, snmpScanResult)
		if err == nil {
			// Modbus 探測
			for _, d := range global.Dependencies {

				if snmpScanResult.TargetName == d.SourceTarget {
					modbusResults := MatchModbusInstances(snmpScanResult.IP)
					if len(modbusResults) > 0 {
						log.Printf("🔸[MatchModbusInstances]找到 %d 個設備", len(modbusResults))
						results = append(results, modbusResults...)
					} else {
						scanErrors = append(scanErrors, fmt.Errorf("modbus 探測未找到設備 [%s]，目標設備: %s", snmpConnInfo.IP, d.SourceTarget))
					}
				}
			}
		}
	}()

	// 等待完成或超時
	select {
	case <-done:
		if len(scanErrors) > 0 {
			// 如果有錯誤，返回所有錯誤信息
			return results, fmt.Errorf("掃描過程發生 %d 個錯誤: %v", len(scanErrors), scanErrors)
		}
		return results, nil
	case <-time.After(time.Duration(global.Config.Scanner.FPingTimeout) * time.Second):
		return results, fmt.Errorf("掃描超時 [%s]，請檢查網絡連接或調整超時設置", snmpConnInfo.IP)
	}

}

// 收集掃描結果
func CollectResults(resultChan <-chan models.ScanInstanceInfo) []models.ScanInstanceInfo {
	var results []models.ScanInstanceInfo
	for result := range resultChan {
		results = append(results, result)
	}
	log.Printf("🔸[CollectResults]掃描到 %d 個設備", len(results))
	return results
}

// 分類掃描結果
func ClassifyResults(results []models.ScanInstanceInfo) ([]models.ScanInstanceInfo, []models.ScanInstanceInfo) {
	var pduResult, deviceResult []models.ScanInstanceInfo
	for _, result := range results {
		if result.InstanceType == models.InstanceTypePDU {
			pduResult = append(pduResult, result)
		} else {
			deviceResult = append(deviceResult, result)
		}
	}
	log.Printf("🔸[ClassifyResults]分類結果: PDU %d 個, 設備 %d 個", len(pduResult), len(deviceResult))
	return pduResult, deviceResult
}

// 格式化時間為 RFC3339 格式
func formatTime(t time.Time) string {
	return t.Format(models.TIME_FORMAT_RFC3339)
}

// 對支援 SNMP 的 IP 做 OID 查詢並嘗試識別設備
func MatchSnmpInstances(snmpConnInfo models.SnmpConnInfo) (models.ScanInstanceInfo, error) {

	snmpInstanceLib := global.SNMPInstanceLib

	// 遍歷每個 SNMP 目標收集資訊
	var snmpMatchResult []string
	for _, matchSession := range snmpInstanceLib.CustomModelSessions {
		var results []string
		matchSession.Conn.IP = snmpConnInfo.IP
		matchSession.Conn.Port = snmpConnInfo.Port
		matchSession.Conn.Community = snmpConnInfo.Community
		matchSession.Conn.Version = snmpConnInfo.Version
		// 嘗試匹配
		result, err := GetSnmpValue(matchSession)
		if err != nil || result == nil {
			continue
		}
		for _, value := range result {
			if value == nil || value.(string) == "" || value.(string) == " " {
				continue
			}
			// 使用正則表達式檢查是否包含英文或數字
			matched, _ := regexp.Compile(`[a-zA-Z0-9]`)
			if !matched.MatchString(value.(string)) {
				continue
			}
			results = append(results, value.(string))
		}
		snmpMatchResult = append(snmpMatchResult, strings.Join(results, " "))
	}

	// 檢查每個 SNMP 目標收集資訊
	instanceInfo, detailSession := GetSNMPTargetByMatchModelValue(snmpConnInfo.IP, snmpConnInfo.Port, models.PROTOCOL_SNMP, snmpMatchResult)

	// 嘗試獲取 Detail 數據
	detailSession.Conn.IP = snmpConnInfo.IP
	detailSession.Conn.Port = snmpConnInfo.Port
	detailSession.Conn.Community = snmpConnInfo.Community
	detailSession.Conn.Version = snmpConnInfo.Version

	snmpDetailResult, err := GetSnmpValue(detailSession)
	if err != nil {
		log.Printf("===> [Detail SNMP] 查詢失敗 [%s:%d]: %v", snmpConnInfo.IP, detailSession.Conn.Port, err)
		//continue
	}

	// 取得 Mac Address
	mac, err := utils.GetMacByArp(snmpConnInfo.IP)
	if err != nil {
		log.Printf("===> [MacAddress] 查詢失敗 [%s]: %v", snmpConnInfo.IP, err)
		//continue
	}
	instanceInfo.MacAddress = mac

	// 設置詳細資訊
	instanceInfo.Version = getStringValue(snmpDetailResult, models.FIELD_VERSION_NAME)
	instanceInfo.SerialNumber = getStringValue(snmpDetailResult, models.FIELD_SERIAL_NUMBER_NAME)
	instanceInfo.SnmpEngineID = getStringValue(snmpDetailResult, models.FIELD_SNMP_ENGINE_ID_NAME)
	instanceInfo.SysName = getStringValue(snmpDetailResult, models.FIELD_SYS_NAME)
	instanceInfo.SysDescr = getStringValue(snmpDetailResult, models.FIELD_SYS_DESCR_NAME)
	instanceInfo.Uptime = parseUptimeToSeconds(getStringValue(snmpDetailResult, models.FIELD_UPTIME_NAME))
	return instanceInfo, nil

}

// FindTarget 根據名稱查找目標
func GetSNMPTargetByMatchModelValue(ip string, port int, protocol string, matchResult []string) (models.ScanInstanceInfo, models.SnmpSession) {
	snmpInstanceLib := global.SNMPInstanceLib
	var detailSession models.SnmpSession
	var instanceInfo models.ScanInstanceInfo
	var isMatchInstance bool
	// 遍歷每個匹配的值 取得對應的 InstanceInfo & 詳細資訊 Session
	for _, matchValue := range matchResult {
		if matchValue == "<nil>" || matchValue == "" || isMatchInstance {
			continue
		}
		for i, target := range snmpInstanceLib.Targets {
			if target.Protocol != models.PROTOCOL_SNMP {
				continue
			}

			if strings.Contains(matchValue, target.SNMP.MatchValueContains) {

				detailSession = snmpInstanceLib.DetailFieldSessions[i]
				instanceInfo = models.ScanInstanceInfo{
					TargetName:   target.Name,
					IP:           ip,
					Port:         port,
					Protocol:     protocol,
					Model:        target.Model,
					Manufacturer: target.Manufacturer,
					InstanceType: target.InstanceType,
				}
				isMatchInstance = true
				log.Printf("✅[%s:%d]找到符合關鍵字的設備: %s", instanceInfo.IP, instanceInfo.Port, target.Model)
			}
		}
	}

	// 如果未知設備使用未知會話
	if !isMatchInstance {
		detailSession = global.SNMPInstanceLib.UnknownSession
		instanceInfo = models.ScanInstanceInfo{
			TargetName:   models.UNKNOWN_NAME,
			Port:         global.SNMPSettings.Port,
			Protocol:     models.PROTOCOL_SNMP,
			Model:        models.DEFAULT_UNKNOWN_VALUE,
			Manufacturer: models.DEFAULT_UNKNOWN_VALUE,
			InstanceType: models.InstanceTypeUnknown,
		}
	}

	return instanceInfo, detailSession

}

// 對 Modbus Client（如 WCM-421）進行探測，取得其連接的 PDU 型號
func MatchModbusInstances(ip string) []models.ScanInstanceInfo {
	var results []models.ScanInstanceInfo
	modbusConnInfos := []models.ModbusConnInfo{}

	modbusInstanceLib := global.ModbusInstanceLib
	matchSession := modbusInstanceLib.MatchSession
	detailSession := modbusInstanceLib.DetailSession
	matchSession.Conn.IP = ip
	count := 0
	for _, port := range global.ModbusSettings.Port {
		for _, slaveID := range global.ModbusSettings.SlaveID {
			count++
			matchSession.Conn.Port = port
			matchSession.Conn.SlaveID = byte(slaveID)

			// 執行 Modbus 查詢
			modbusResult, err := GetModbusValue(matchSession)
			if err != nil {

				continue
			}

			// 檢查結果
			modelValue := ""
			for _, value := range modbusResult {
				if value != nil {
					modelValue = value.(string)

					// 匹配設備
					for _, target := range global.ModbusInstanceLib.Targets {
						if strings.Contains(modelValue, target.Modbus.MatchValueContains) {
							log.Printf("🛜 [%s:%d:%d]找到符合關鍵字的設備: %s ", ip, matchSession.Conn.Port, matchSession.Conn.SlaveID, modelValue)
							modbusConnInfos = append(modbusConnInfos, models.ModbusConnInfo{
								IP:           ip,
								Port:         port,
								SlaveID:      slaveID,
								Model:        target.Model,
								Manufacturer: target.Manufacturer,
								InstanceType: target.InstanceType,
							})
						}
						break
					}
				} else {
					// log.Printf("⛔[%s:%d:%d]查詢結果中沒有 model 資訊 ", ip, matchSession.Conn.Port, matchSession.Conn.SlaveID)
					continue
				}
			}

			if modelValue == "" {
				continue
			}

		}

		if len(modbusConnInfos) == 0 {
			continue
		}

		for _, modbusConnInfo := range modbusConnInfos {
			result := models.ScanInstanceInfo{
				IP:           ip,
				Port:         modbusConnInfo.Port,
				SlaveID:      modbusConnInfo.SlaveID,
				Protocol:     models.PROTOCOL_MODBUS,
				Model:        modbusConnInfo.Model,
				Manufacturer: modbusConnInfo.Manufacturer,
				InstanceType: modbusConnInfo.InstanceType,
			}
			// 嘗試獲取 Detail 數據
			detailSession.Conn.IP = ip
			detailSession.Conn.Port = modbusConnInfo.Port
			detailSession.Conn.SlaveID = byte(modbusConnInfo.SlaveID)
			modbusDetailResult, err := GetModbusValue(detailSession)
			if err != nil {
				log.Printf("Modbus  [%s:%d:%d]: %v", ip, matchSession.Conn.Port, matchSession.Conn.SlaveID, err)
				continue
			}

			// 設置詳細資訊
			result.Version = getStringValue(modbusDetailResult, models.FIELD_VERSION_NAME)
			result.SerialNumber = getStringValue(modbusDetailResult, models.FIELD_SERIAL_NUMBER_NAME)
			result.Uptime = parseUptimeToSeconds(getStringValue(modbusDetailResult, models.FIELD_UPTIME_NAME))
			results = append(results, result)
		}
	}

	return results
}

// 獲取字串值
func getStringValue(result map[string]interface{}, key string) string {
	if value, ok := result[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// 解析運行時間為秒數
func parseUptimeToSeconds(uptimeStr string) int64 {
	if uptimeStr == "" {
		return 0
	}

	// 移除括號
	uptimeStr = strings.Trim(uptimeStr, "()")

	// 分割時間單位
	parts := strings.Split(uptimeStr, " ")
	if len(parts) != 2 {
		return 0
	}

	// 解析數值
	value, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0
	}

	// 根據單位轉換為秒
	switch parts[1] {
	case "days":
		return value * 24 * 60 * 60
	case "hours":
		return value * 60 * 60
	case "minutes":
		return value * 60
	case "seconds":
		return value
	default:
		return 0
	}
}

// 格式化實例資訊
func formatInstances(result models.ScanInstanceInfo) models.ScanInstanceInfo {
	// 設置預設值
	if result.Model == "" {
		result.Model = models.DEFAULT_UNKNOWN_VALUE
	}
	if result.Version == "" {
		result.Version = models.DEFAULT_UNKNOWN_VALUE
	}
	if result.SerialNumber == "" {
		result.SerialNumber = models.DEFAULT_UNKNOWN_VALUE
	}
	if result.MacAddress == "" {
		result.MacAddress = models.DEFAULT_UNKNOWN_VALUE
	}
	if result.InstanceType == "" {
		result.InstanceType = models.InstanceTypeUnknown
	}
	if result.Manufacturer == "" {
		result.Manufacturer = models.DEFAULT_UNKNOWN_VALUE
	}
	if result.SysName == "" {
		result.SysName = models.DEFAULT_UNKNOWN_VALUE
	}
	if result.SysDescr == "" {
		result.SysDescr = models.DEFAULT_UNKNOWN_VALUE
	}
	if result.Uptime == 0 {
		result.Uptime = 0
	}

	return result
}

// 過濾掉 pending_pdu_list 或 pending_device_list 中的 IP
func FilterScanList(results []models.SnmpConnInfo) []models.SnmpConnInfo {
	var filteredResults []models.SnmpConnInfo
	for _, result := range results {
		if slices.Contains(global.ScanIPKeyList, result.IP) {
			continue
		}
		filteredResults = append(filteredResults, result)
	}
	return filteredResults
}
