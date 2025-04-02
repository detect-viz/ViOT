package models

type SNMPInstanceLib struct {
	TargetNames         []string      `json:"target_names"`
	Targets             []Target      `json:"targets"`
	CustomModelSessions []SnmpSession `json:"custom_model_sessions"` //step1
	DetailFieldSessions []SnmpSession `json:"detail_field_sessions"` //step2
	UnknownSession      SnmpSession   `json:"unknown_session"`
}

type ModbusInstanceLib struct {
	TargetNames   []string      `json:"target_names"`
	Targets       []Target      `json:"targets"`
	MatchSession  ModbusSession `json:"match_session"`
	DetailSession ModbusSession `json:"detail_session"`
}
