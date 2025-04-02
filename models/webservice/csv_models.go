package webservice

import "time"

// DeviceStatus 設備狀態記錄
type DeviceStatus struct {
	Name      string    `csv:"name"`
	IP        string    `csv:"ip"`
	Monitored bool      `csv:"monitored"`
	Status    string    `csv:"status"`
	Uptime    int64     `csv:"uptime"`
	UpdateAt  time.Time `csv:"update_at"`
}

// PDUStatus PDU狀態記錄
type PDUStatus struct {
	Name      string    `csv:"name"`
	IP        string    `csv:"ip"`
	Monitored bool      `csv:"monitored"`
	Status    string    `csv:"status"`
	Uptime    int64     `csv:"uptime"`
	UpdateAt  time.Time `csv:"update_at"`
}

// Tag 標籤結構
type Tag struct {
	Key   string `csv:"key"`
	Value string `csv:"value"`
}

// DeviceTag 設備標籤記錄
type DeviceTag struct {
	Name     string    `csv:"name"`
	Tags     []Tag     `csv:"tags"`
	UpdateAt time.Time `csv:"update_at"`
}

// PDUTag PDU標籤記錄
type PDUTag struct {
	Name     string    `csv:"name"`
	Tags     []Tag     `csv:"tags"`
	UpdateAt time.Time `csv:"update_at"`
}

// DevicePending 待處理設備記錄
type DevicePending struct {
	IP           string    `csv:"ip"`
	Factory      string    `csv:"factory"`
	Phase        string    `csv:"phase"`
	Datacenter   string    `csv:"datacenter"`
	Room         string    `csv:"room"`
	InstanceType string    `csv:"instance_type"`
	Manufacturer string    `csv:"manufacturer"`
	Model        string    `csv:"model"`
	MacAddress   string    `csv:"mac_address"`
	SnmpEngineID string    `csv:"snmp_engine_id"`
	UpdateAt     time.Time `csv:"update_at"`
}

// PDUPending 待處理PDU記錄
type PDUPending struct {
	IP           string    `csv:"ip"`
	Factory      string    `csv:"factory"`
	Phase        string    `csv:"phase"`
	Datacenter   string    `csv:"datacenter"`
	Room         string    `csv:"room"`
	InstanceType string    `csv:"instance_type"`
	Manufacturer string    `csv:"manufacturer"`
	Model        string    `csv:"model"`
	Version      string    `csv:"version"`
	SerialNumber string    `csv:"serial_number"`
	MacAddress   string    `csv:"mac_address"`
	UpdateAt     time.Time `csv:"update_at"`
}

// DeviceRegistry 設備註冊表記錄
type DeviceRegistry struct {
	Name         string    `csv:"name"`
	Factory      string    `csv:"factory"`
	Phase        string    `csv:"phase"`
	Datacenter   string    `csv:"datacenter"`
	Room         string    `csv:"room"`
	Type         string    `csv:"type"`
	Manufacturer string    `csv:"manufacturer"`
	Model        string    `csv:"model"`
	MAC          string    `csv:"mac"`
	SnmpEngineID string    `csv:"snmp_engine_id"`
	UpdateAt     time.Time `csv:"update_at"`
}

// PDURegistry PDU註冊表記錄
type PDURegistry struct {
	Name         string    `csv:"name"`
	Factory      string    `csv:"factory"`
	Phase        string    `csv:"phase"`
	Datacenter   string    `csv:"datacenter"`
	Room         string    `csv:"room"`
	Type         string    `csv:"type"`
	Manufacturer string    `csv:"manufacturer"`
	Model        string    `csv:"model"`
	MAC          string    `csv:"mac"`
	Version      string    `csv:"version"`
	SerialNumber string    `csv:"serial_number"`
	UpdateAt     time.Time `csv:"update_at"`
}

// Position 位置記錄
type Position struct {
	Factory    string `csv:"factory"`
	Phase      string `csv:"phase"`
	Datacenter string `csv:"datacenter"`
	Room       string `csv:"room"`
	Name       string `csv:"name"`
	IsPDU      bool   `csv:"is_pdu"`
}

// AddTag 添加標籤
func (dt *DeviceTag) AddTag(key, value string) {
	dt.Tags = append(dt.Tags, Tag{Key: key, Value: value})
}

// RemoveTag 移除標籤
func (dt *DeviceTag) RemoveTag(key string) {
	for i := len(dt.Tags) - 1; i >= 0; i-- {
		if dt.Tags[i].Key == key {
			dt.Tags = append(dt.Tags[:i], dt.Tags[i+1:]...)
			break
		}
	}
}

// UpdateTag 更新標籤
func (dt *DeviceTag) UpdateTag(key, value string) {
	for i := range dt.Tags {
		if dt.Tags[i].Key == key {
			dt.Tags[i].Value = value
			break
		}
	}
}

// GetTag 獲取標籤值
func (dt *DeviceTag) GetTag(key string) (string, bool) {
	for _, tag := range dt.Tags {
		if tag.Key == key {
			return tag.Value, true
		}
	}
	return "", false
}

// HasTag 檢查是否存在標籤
func (dt *DeviceTag) HasTag(key string) bool {
	for _, tag := range dt.Tags {
		if tag.Key == key {
			return true
		}
	}
	return false
}

// AddTag 添加標籤
func (pt *PDUTag) AddTag(key, value string) {
	pt.Tags = append(pt.Tags, Tag{Key: key, Value: value})
}

// RemoveTag 移除標籤
func (pt *PDUTag) RemoveTag(key string) {
	for i := len(pt.Tags) - 1; i >= 0; i-- {
		if pt.Tags[i].Key == key {
			pt.Tags = append(pt.Tags[:i], pt.Tags[i+1:]...)
			break
		}
	}
}

// UpdateTag 更新標籤
func (pt *PDUTag) UpdateTag(key, value string) {
	for i := range pt.Tags {
		if pt.Tags[i].Key == key {
			pt.Tags[i].Value = value
			break
		}
	}
}

// GetTag 獲取標籤值
func (pt *PDUTag) GetTag(key string) (string, bool) {
	for _, tag := range pt.Tags {
		if tag.Key == key {
			return tag.Value, true
		}
	}
	return "", false
}

// HasTag 檢查是否存在標籤
func (pt *PDUTag) HasTag(key string) bool {
	for _, tag := range pt.Tags {
		if tag.Key == key {
			return true
		}
	}
	return false
}
