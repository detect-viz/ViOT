package models

type SnmpSession struct {
	Conn     SnmpConn
	Requests []SNMPRequest
}
type SnmpConn struct {
	IP        string
	Port      int
	Version   int
	Community string
}
type SNMPRequest struct {
	Name string
	OID  string
}
