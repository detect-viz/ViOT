package models

import "time"

type ScanActiveIPInfo struct {
	IP         string
	Factory    string
	Phase      string
	Datacenter string
	Room       string
}

type SnmpConnInfo struct {
	IP        string
	Port      int
	Community string
	Version   int
}

type ModbusConnInfo struct {
	IP           string
	Port         int
	SlaveID      int
	Model        string
	Manufacturer string
	InstanceType string
}

// 探測目標設備的基礎結構
type MatchTargetInfo struct {
	TargetName         string
	TargetProtocol     string
	TargetModel        string
	TargetManufacturer string
	TargetInstanceType string
	ScanIP             string
	ScanPort           int
	ScanSlaveID        int
	ScanDetail         map[string]interface{}
}

// 探測目標設備的詳細結構
type ScanInstanceInfo struct {
	TargetName   string `json:"target_name"`
	IPKey        string `json:"ip_key"`
	IP           string `json:"ip"`
	Port         int    `json:"port"`
	SlaveID      int    `json:"slave_id"`
	Protocol     string `json:"protocol"`
	Model        string `json:"model"`
	Version      string `json:"version"`
	SerialNumber string `json:"serial_number"`
	SnmpEngineID string `json:"snmp_engine_id"`
	SysName      string `json:"sys_name"`
	SysDescr     string `json:"sys_descr"`
	Uptime       int64  `json:"uptime"`
	InstanceType string `json:"instance_type"`
	Manufacturer string `json:"manufacturer"`
	MacAddress   string `json:"mac_address"`
	Factory      string `json:"factory"`
	Phase        string `json:"phase"`
	Datacenter   string `json:"datacenter"`
	Room         string `json:"room"`
}

// 設備身份資訊
type DeviceIdentity struct {
	IPKey         string    `json:"ip_key"`
	IP            string    `json:"ip"`
	Protocol      string    `json:"protocol"`
	Factory       string    `json:"factory"`    // from ip_location
	Phase         string    `json:"phase"`      // from ip_location
	Datacenter    string    `json:"datacenter"` // from ip_location
	Room          string    `json:"room"`       // from ip_location
	InstanceType  string    `json:"instance_type"`
	Manufacturer  string    `json:"manufacturer"`
	Model         string    `json:"model"`
	Version       string    `json:"version"`
	SerialNumber  string    `json:"serial_number"`
	MacAddress    string    `json:"mac_address"`
	SnmpEngineID  string    `json:"snmp_engine_id"`
	SysName       string    `json:"sys_name"`
	SysDescr      string    `json:"sys_descr"`
	Uptime        string    `json:"uptime"`
	UptimeSeconds int64     `json:"uptime_seconds"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PDU 設備身份資訊
type PDUIdentity struct {
	IPKey         string    `json:"ip_key"`
	IP            string    `json:"ip"`
	Protocol      string    `json:"protocol"`
	Factory       string    `json:"factory"`
	Phase         string    `json:"phase"`
	Datacenter    string    `json:"datacenter"`
	Room          string    `json:"room"`
	InstanceType  string    `json:"instance_type"`
	Manufacturer  string    `json:"manufacturer"`
	Model         string    `json:"model"`
	Version       string    `json:"version"`
	SerialNumber  string    `json:"serial_number"`
	MacAddress    string    `json:"mac_address"`
	Uptime        string    `json:"uptime"`
	UptimeSeconds int64     `json:"uptime_seconds"`
	UpdatedAt     time.Time `json:"updated_at"`
}
