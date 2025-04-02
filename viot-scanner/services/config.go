package services

import (
	"fmt"
	"os"
	"slices"
	"viot-scanner/global"
	"viot-scanner/models"

	"gopkg.in/yaml.v2"
)

func LoadConfig(filename string) error {
	cfg, err := loadScannerConfig(filename)
	if err != nil {
		return err
	}
	err = loadTargetSettings(cfg.Scanner.TargetFile)
	if err != nil {
		return err
	}
	err = loadIPRangeConfig(cfg.Scanner.IPRangeFile)
	if err != nil {
		return err
	}
	return nil
}

// 載入掃描器設定檔
func loadScannerConfig(filename string) (*models.Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("讀取配置文件失敗: %v", err)
	}

	var config models.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失敗: %v", err)
	}

	// 初始化全局配置
	global.Init(&config)

	return &config, nil
}

// 載入目標設定檔
func loadTargetSettings(filename string) error {
	var targets models.TargetSettings
	// 讀取目標設定
	targetData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("讀取目標設定檔失敗: %v", err)
	}

	if err := yaml.Unmarshal(targetData, &targets); err != nil {
		return fmt.Errorf("解析目標設定檔失敗: %v", err)
	}

	// 更新全局目標設定
	global.TargetSettings = targets
	loadSNMPInstanceLib(targets)
	loadModbusInstanceLib(global.ModbusSettings, targets)
	return nil
}

// 載入 IP 範圍設定檔
func loadIPRangeConfig(filename string) error {
	var ip_ranges []models.IPRange

	// 讀取目標設定
	ipRangeData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("讀取目標設定檔失敗: %v", err)
	}

	if err := yaml.Unmarshal(ipRangeData, &ip_ranges); err != nil {
		return fmt.Errorf("解析目標設定檔失敗: %v", err)
	}

	// 更新全局 IP 範圍配置
	global.IPRangeConfigs = ip_ranges

	return nil
}

// FindTarget 根據名稱查找目標
func FindTargetByName(targetName string) models.Target {
	for _, target := range global.TargetSettings.Targets {
		if target.Name == targetName {
			return target
		}
	}
	return models.Target{}
}

// GetSNMPInstanceFieldRequest 獲取 SNMP 實例欄位請求
func getSNMPInstanceFieldRequest(targetName string) []models.SNMPRequest {
	switch targetName {
	case models.PDUE428_NAME:
		return []models.SNMPRequest{
			{
				Name: models.FIELD_MODEL_NAME,
				OID:  models.SNMP_OID_DELTA_PDU_MODEL,
			},
			{
				Name: models.FIELD_SERIAL_NUMBER_NAME,
				OID:  models.SNMP_OID_DELTA_PDU_SERIAL_NUMBER,
			},
			{
				Name: models.FIELD_VERSION_NAME,
				OID:  models.SNMP_OID_DELTA_PDU_VERSION,
			},
			{
				Name: models.FIELD_UPTIME_NAME,
				OID:  models.SNMP_OID_UPTIME,
			},
		}
	case models.PDU6PS56_NAME:
		return []models.SNMPRequest{
			{
				Name: models.FIELD_MODEL_NAME,
				OID:  models.SNMP_OID_VERTIV_6PS56_MODEL,
			},
			{
				Name: models.FIELD_SERIAL_NUMBER_NAME,
				OID:  models.SNMP_OID_VERTIV_6PS56_SERIAL_NUMBER,
			},
			{
				Name: models.FIELD_VERSION_NAME,
				OID:  models.SNMP_OID_VERTIV_6PS56_VERSION,
			},
			{
				Name: models.FIELD_UPTIME_NAME,
				OID:  models.SNMP_OID_UPTIME,
			},
		}
	case models.SWITCH5130_NAME,
		models.SWITCH5945_NAME,
		models.IOT_CLIENT_NAME:
		return []models.SNMPRequest{
			{
				Name: models.FIELD_SNMP_ENGINE_ID_NAME,
				OID:  models.SNMP_OID_SNMP_ENGINE_ID,
			},
			{
				Name: models.FIELD_SYS_DESCR_NAME,
				OID:  models.SNMP_OID_SYS_DESCR,
			},
			{
				Name: models.FIELD_SYS_NAME,
				OID:  models.SNMP_OID_SYS_NAME,
			},
			{
				Name: models.FIELD_UPTIME_NAME,
				OID:  models.SNMP_OID_UPTIME,
			},
		}

	case models.UNKNOWN_NAME:
		return []models.SNMPRequest{
			{
				Name: models.FIELD_SYS_NAME,
				OID:  models.SNMP_OID_SYS_NAME,
			},
			{
				Name: models.FIELD_SYS_DESCR_NAME,
				OID:  models.SNMP_OID_SYS_DESCR,
			},
			{
				Name: models.FIELD_UPTIME_NAME,
				OID:  models.SNMP_OID_UPTIME,
			},
		}
	default:
		return nil
	}
}

// GetModbusInstanceFieldRequest 獲取 Modbus 實例欄位請求
func getModbusInstanceFieldRequest(targetName string) []models.ModbusRequest {
	switch targetName {
	case models.PDU1315_NAME, models.PDU4445_NAME:
		return []models.ModbusRequest{
			{ // model split to version
				RegisterType: models.MODBUS_DELTA_PDU_MODEL_REGISTER_TYPE,
				RegisterName: models.FIELD_MODEL_NAME,
				Address:      models.MODBUS_DELTA_PDU_MODEL_ADDRESS,
				Length:       models.MODBUS_DELTA_PDU_MODEL_LENGTH,
			},
			{
				RegisterType: models.MODBUS_DELTA_PDU_SERIAL_REGISTER_TYPE,
				RegisterName: models.FIELD_SERIAL_NUMBER_NAME,
				Address:      models.MODBUS_DELTA_PDU_SERIAL_ADDRESS,
				Length:       models.MODBUS_DELTA_PDU_SERIAL_LENGTH,
			},
		}
	default:
		return nil
	}
}

func loadSNMPInstanceLib(targetSettings models.TargetSettings) {

	snmpInstanceLib := models.SNMPInstanceLib{}
	snmpInstanceLib.TargetNames = []string{}

	for _, targetConfig := range targetSettings.Targets {
		if targetConfig.Protocol != models.PROTOCOL_SNMP {
			continue
		}

		// 創建 Match SNMP
		matchSession := models.SnmpSession{

			Requests: []models.SNMPRequest{
				{
					Name: targetConfig.Name,
					OID:  targetConfig.SNMP.MatchOID,
				},
			},
		}
		// 創建 Detail SNMP
		oids := getSNMPInstanceFieldRequest(targetConfig.Name)
		detailSession := models.SnmpSession{
			Requests: oids,
		}
		snmpInstanceLib.CustomModelSessions = append(snmpInstanceLib.CustomModelSessions, matchSession)

		snmpInstanceLib.DetailFieldSessions = append(snmpInstanceLib.DetailFieldSessions, detailSession)
		snmpInstanceLib.Targets = append(snmpInstanceLib.Targets, targetConfig)
		snmpInstanceLib.TargetNames = append(snmpInstanceLib.TargetNames, targetConfig.Name)
	}

	// 創建 Unknown Detail SNMP
	baseConn := models.SnmpConn{
		Port:      global.SNMPSettings.Port,
		Version:   global.SNMPSettings.Version,
		Community: global.SNMPSettings.Community,
	}
	oids := getSNMPInstanceFieldRequest(models.UNKNOWN_NAME)
	detailSession := models.SnmpSession{
		Conn:     baseConn,
		Requests: oids,
	}
	snmpInstanceLib.UnknownSession = detailSession

	global.SNMPInstanceLib = snmpInstanceLib
}

func loadModbusInstanceLib(config models.ModbusSettings, targetSettings models.TargetSettings) {

	modbusInstanceLib := models.ModbusInstanceLib{}
	modbusInstanceLib.TargetNames = []string{}
	for _, port := range config.Port {
		for _, slaveID := range config.SlaveID {
			baseConn := models.ModbusConn{
				Port:    port,
				Mode:    config.Mode,
				SlaveID: byte(slaveID),
			}

			var mfg_targets []string
			for _, targetConfig := range targetSettings.Targets {
				mfg_targets = append(mfg_targets, targetConfig.Manufacturer)
				if targetConfig.Protocol != models.PROTOCOL_MODBUS {
					continue
				}
				if !slices.Contains(mfg_targets, targetConfig.Manufacturer) {
					continue
				}

				// 創建 Match Modbus
				matchSession := models.ModbusSession{
					Conn: baseConn,
					Requests: []models.ModbusRequest{
						{
							RegisterType: targetConfig.Modbus.MatchRegisterType,
							RegisterName: targetConfig.Name,
							Address:      uint16(targetConfig.Modbus.MatchAddress),
							Length:       uint16(targetConfig.Modbus.MatchLength),
						},
					},
				}
				// 創建 Detail Modbus
				oids := getModbusInstanceFieldRequest(targetConfig.Name)
				detailSession := models.ModbusSession{
					Conn:     baseConn,
					Requests: oids,
				}
				modbusInstanceLib.MatchSession = matchSession
				modbusInstanceLib.DetailSession = detailSession
				modbusInstanceLib.Targets = append(modbusInstanceLib.Targets, targetConfig)
				modbusInstanceLib.TargetNames = append(modbusInstanceLib.TargetNames, targetConfig.Name)
			}
		}
	}
	global.ModbusInstanceLib = modbusInstanceLib
}
