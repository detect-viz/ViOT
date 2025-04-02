package models

type ModbusSession struct {
	Conn     ModbusConn
	Requests []ModbusRequest
}
type ModbusConn struct {
	IP      string
	Port    int
	Mode    string
	SlaveID byte
}
type ModbusRequest struct {
	RegisterType string
	RegisterName string
	Address      uint16
	Length       uint16
}
