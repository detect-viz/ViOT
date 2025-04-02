package models

import "time"

// Protocol constants
const (
	PROTOCOL_SNMP   = "snmp"
	PROTOCOL_MODBUS = "modbus"
)

// Field constants
const (
	FIELD_MODEL_NAME          = "model"
	FIELD_SERIAL_NUMBER_NAME  = "serial_number"
	FIELD_VERSION_NAME        = "version"
	FIELD_MAC_ADDRESS_NAME    = "mac_address"
	FIELD_SNMP_ENGINE_ID_NAME = "snmp_engine_id"
	FIELD_UPTIME_NAME         = "uptime"
	FIELD_SYS_NAME            = "sys_name"
	FIELD_SYS_DESCR_NAME      = "sys_descr"
	FIELD_MANUFACTURER_NAME   = "manufacturer"
)

// Instance type constants
const (
	InstanceTypePDU     = "pdu"     // 電力設備（如Delta PDU）
	InstanceTypeSwitch  = "switch"  // 網路交換器
	InstanceTypeAP      = "ap"      // 無線基地台
	InstanceTypeGateway = "gateway" // IoT 協定轉接器（如 WCM-421）
	InstanceTypeUnknown = "unknown" // 尚未分類的設備
)

// Device model constants
const (
	PDU4445_NAME    = "PDU-DELTA-PDU4445"
	PDU1315_NAME    = "PDU-DELTA-PDU1315"
	PDUE428_NAME    = "PDU-DELTA-PDUE428"
	PDU6PS56_NAME   = "PDU-VERTIV-6PS56"
	SWITCH5130_NAME = "SWITCH-HPE-5130"
	SWITCH5945_NAME = "SWITCH-HPE-5945"
	AP_NAME         = "AP-MOXA-AWK-1161A"
	IOT_CLIENT_NAME = "IOT_CLIENT-LIENEO-WCM-421"
	UNKNOWN_NAME    = "unknown"
)

// Default values
const (
	DEFAULT_UNKNOWN_VALUE = "unknown"
	GlobalScanInterval    = 300 // 5 minutes default
)

// Registry constants
const (
	REGISTRY_PDU_LIST    = "registry_pdu.csv"
	REGISTRY_DEVICE_LIST = "registry_device.csv"
	SCAN_PDU_LIST        = "scan_pdu.csv"
	SCAN_DEVICE_LIST     = "scan_device.csv"
)

// Time format constants
const (
	TIME_FORMAT_RFC3339 = time.RFC3339
)

// SNMP OID constants
const (
	// Basic OIDs
	SNMP_OID_SYS_NAME       = "1.3.6.1.2.1.1.5.0"
	SNMP_OID_SYS_DESCR      = "1.3.6.1.2.1.1.1.0"
	SNMP_OID_UPTIME         = "1.3.6.1.2.1.1.3.0"
	SNMP_OID_SNMP_ENGINE_ID = "1.3.6.1.6.3.10.2.1.1.0"

	// DELTA PDU OIDs
	SNMP_OID_DELTA_PDU_MODEL         = "1.3.6.1.4.1.2254.2.32.1.6.1.3.1"
	SNMP_OID_DELTA_PDU_SERIAL_NUMBER = "1.3.6.1.4.1.2254.2.32.1.6.1.4.1"
	SNMP_OID_DELTA_PDU_VERSION       = "1.3.6.1.4.1.2254.2.32.1.4.0"

	// VERTIV 6PS56 OIDs
	SNMP_OID_VERTIV_6PS56_MODEL         = "1.3.6.1.4.1.21239.5.2.1.8.0"
	SNMP_OID_VERTIV_6PS56_SERIAL_NUMBER = "1.3.6.1.4.1.21239.5.2.1.10.0"
	SNMP_OID_VERTIV_6PS56_VERSION       = "1.3.6.1.4.1.21239.5.2.1.2.0"

	// IOT_CLIENT-LIENEO-WCM-421 OIDs
	SNMP_OID_IOT_CLIENT_WCM_421_MODEL             = "1.3.6.1.4.1.2021.50.3.101.1"
	SNMP_OID_IOT_CLIENT_WCM_421_STATUS_PORT_1     = "1.3.6.1.4.1.2021.50.11.101.1"
	SNMP_OID_IOT_CLIENT_WCM_421_STATUS_PORT_2     = "1.3.6.1.4.1.2021.50.12.101.1"
	SNMP_OID_IOT_CLIENT_WCM_421_GET_CONNECT_COUNT = "1.3.6.1.4.1.2021.50.10"
	SNMP_OID_IOT_CLIENT_WCM_421_SET_RESTART       = "1.3.6.1.4.1.2021.50.10 i 1"
)

// Modbus constants
const (
	MODBUS_DELTA_PDU_MODEL_ADDRESS        = uint16(1024)
	MODBUS_DELTA_PDU_SERIAL_ADDRESS       = uint16(768)
	MODBUS_DELTA_PDU_MODEL_REGISTER_TYPE  = "input"
	MODBUS_DELTA_PDU_SERIAL_REGISTER_TYPE = "holding"
	MODBUS_DELTA_PDU_SERIAL_LENGTH        = 8
	MODBUS_DELTA_PDU_MODEL_LENGTH         = 8
)

// CSV Header constants
const (
	CSV_PDU_HEADER    = "IP,Protocol,Model,Version,SerialNumber,MacAddress,InstanceType,Manufacturer,Uptime,UptimeSeconds,UpdatedAt"
	CSV_DEVICE_HEADER = "IP,Protocol,Model,Version,SerialNumber,MacAddress,InstanceType,Manufacturer,SysName,SysDescr,Uptime,UptimeSeconds,UpdatedAt"
)
