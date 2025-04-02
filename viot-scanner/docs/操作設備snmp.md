## Device Model Type

snmpwalk -v2c -c public 10.1.249.34 .1.3.6.1.4.1.2021.50.3.101.1
"CD-SNMP-MIB::ucdavis.50.3.101.1 = STRING: "'WCM-421'

## 串口 1 連線狀態，"0":空閒，"1":占用中

snmpwalk -v2c -c public 10.1.249.34 .1.3.6.1.4.1.2021.50.11.101.1

## 串口 2 連線狀態，"0":空閒，"1":占用中

snmpwalk -v2c -c public 10.1.249.34 .1.3.6.1.4.1.2021.50.12.101.1

## snmpget:串口服務狀態(連線數)

snmpget -v2c -c public 10.1.249.34 .1.3.6.1.4.1.2021.50.10

## snmpset:寫入"1"重啟串口服務(強制中止連線)

snmpset -v2c -c public 10.1.249.34 .1.3.6.1.4.1.2021.50.10 i 1

## snmpEngineID

root@zbox:~# snmpwalk -v2c -c public 10.1.249.34 .1.3.6.1.6.3.10.2.1.1.0
SNMP-FRAMEWORK-MIB::snmpEngineID.0 = Hex-STRING: 80 00 1F 88 03 40 D6 3C BA 2E 7C

## MAC Address

root@zbox:~# arp -n 10.1.249.37
Address HWtype HWaddress Flags Mask Iface
10.1.249.37 ether 00:18:23:68:4a:9d C enp6s18

## Moxa AP

root@zbox:/etc/telegraf# snmpwalk -v 2c -c public 172.5.63.210 | grep STRING
SNMPv2-MIB::sysDescr.0 = STRING: AWK-1161A
SNMPv2-MIB::sysName.0 = STRING: moxa-awk-1161a

## SWITCH

root@zbox:~# snmpwalk -v 1 -c public 10.1.249.40 .1.3.6.1.2.1.47.1.1.1.1.2.2
ENTITY-MIB::entPhysicalDescr.2 = STRING: HPE 5130-48G-PoE+-4SFP+ (370W) EI JG937A Software Version 7.1.045

## IOT CLIENT

root@zbox:~# snmpwalk -v 1 -c public 10.1.249.34 .1.3.6.1.4.1.2021.50.3.101.1
"CD-SNMP-MIB::ucdavis.50.3.101.1 = STRING: "'WCM-421'

root@zbox:~# snmpwalk -v 1 -c public 10.1.249.34 UCD-SNMP-MIB::ucdavis.50.3.101.1
"CD-SNMP-MIB::ucdavis.50.3.101.1 = STRING: "'WCM-421'

root@zbox:~# snmpwalk -v 1 -c public 10.1.249.34 .1.3.6.1.4.1.2021.50
UCD-SNMP-MIB::ucdavis.50.1.1.1 = INTEGER: 1
UCD-SNMP-MIB::ucdavis.50.1.2.1 = STRING: "brand"
UCD-SNMP-MIB::ucdavis.50.1.3.1 = STRING: "/bin/echo lieneo"
UCD-SNMP-MIB::ucdavis.50.1.100.1 = INTEGER: 0
UCD-SNMP-MIB::ucdavis.50.1.101.1 = STRING: "lieneo"
UCD-SNMP-MIB::ucdavis.50.1.102.1 = INTEGER: 0
UCD-SNMP-MIB::ucdavis.50.1.103.1 = ""
UCD-SNMP-MIB::ucdavis.50.2.1.1 = INTEGER: 1
UCD-SNMP-MIB::ucdavis.50.2.2.1 = STRING: "wrtver"
UCD-SNMP-MIB::ucdavis.50.2.3.1 = STRING: "/bin/cat /etc/openwrt_version"
UCD-SNMP-MIB::ucdavis.50.2.100.1 = INTEGER: 0
UCD-SNMP-MIB::ucdavis.50.2.101.1 = STRING: "r24012-d8dd03c46f"
UCD-SNMP-MIB::ucdavis.50.2.102.1 = INTEGER: 0
UCD-SNMP-MIB::ucdavis.50.2.103.1 = ""
UCD-SNMP-MIB::ucdavis.50.3.1.1 = INTEGER: 1
UCD-SNMP-MIB::ucdavis.50.3.2.1 = STRING: "model_type"
UCD-SNMP-MIB::ucdavis.50.3.3.1 = STRING: "/bin/sh -c \"grep MODEL_TYP /etc/iiot_gate_info | cut -d= -f2 | tr -d \\\\\\ \""
UCD-SNMP-MIB::ucdavis.50.3.100.1 = INTEGER: 0
"CD-SNMP-MIB::ucdavis.50.3.101.1 = STRING: "'WCM-421'
UCD-SNMP-MIB::ucdavis.50.3.102.1 = INTEGER: 0
UCD-SNMP-MIB::ucdavis.50.3.103.1 = ""
UCD-SNMP-MIB::ucdavis.50.4.1.1 = INTEGER: 1
UCD-SNMP-MIB::ucdavis.50.4.2.1 = STRING: "hardware_version"
UCD-SNMP-MIB::ucdavis.50.4.3.1 = STRING: "/bin/sh -c \"grep HW_VER /etc/iiot_gate_info | sed \\\"s/HW_VER=//\\\" \""
UCD-SNMP-MIB::ucdavis.50.4.100.1 = INTEGER: 0
"CD-SNMP-MIB::ucdavis.50.4.101.1 = STRING: "'r1.0'
UCD-SNMP-MIB::ucdavis.50.4.102.1 = INTEGER: 0
UCD-SNMP-MIB::ucdavis.50.4.103.1 = ""
UCD-SNMP-MIB::ucdavis.50.5.1.1 = INTEGER: 1
UCD-SNMP-MIB::ucdavis.50.5.2.1 = STRING: "os_version"
UCD-SNMP-MIB::ucdavis.50.5.3.1 = STRING: "/bin/sh -c \"grep OS_VER /etc/iiot_gate_info | sed \\\"s/OS_VER=//\\\" \""
UCD-SNMP-MIB::ucdavis.50.5.100.1 = INTEGER: 0
"CD-SNMP-MIB::ucdavis.50.5.101.1 = STRING: "'b20241113'
UCD-SNMP-MIB::ucdavis.50.5.102.1 = INTEGER: 0
UCD-SNMP-MIB::ucdavis.50.5.103.1 = ""
UCD-SNMP-MIB::ucdavis.50.11.1.1 = INTEGER: 1
UCD-SNMP-MIB::ucdavis.50.11.2.1 = STRING: "checkPort7000"
UCD-SNMP-MIB::ucdavis.50.11.3.1 = STRING: "/bin/sh /root/snmpsh/checkPort.sh 7000"
UCD-SNMP-MIB::ucdavis.50.11.100.1 = INTEGER: 0
UCD-SNMP-MIB::ucdavis.50.11.101.1 = STRING: "1"
UCD-SNMP-MIB::ucdavis.50.11.102.1 = INTEGER: 0
UCD-SNMP-MIB::ucdavis.50.11.103.1 = ""
UCD-SNMP-MIB::ucdavis.50.12.1.1 = INTEGER: 1
UCD-SNMP-MIB::ucdavis.50.12.2.1 = STRING: "checkPort7001"
UCD-SNMP-MIB::ucdavis.50.12.3.1 = STRING: "/bin/sh /root/snmpsh/checkPort.sh 7001"
UCD-SNMP-MIB::ucdavis.50.12.100.1 = INTEGER: 0
UCD-SNMP-MIB::ucdavis.50.12.101.1 = STRING: "0"
UCD-SNMP-MIB::ucdavis.50.12.102.1 = INTEGER: 0
UCD-SNMP-MIB::ucdavis.50.12.103.1 = ""

## VERTIV

root@zbox:/usr/share/snmp/mibs# snmptranslate -On VERTIV-V5-MIB::productSerialNumber.0
.1.3.6.1.4.1.21239.5.2.1.10.0

root@zbox:/usr/share/snmp/mibs# snmptranslate -On VERTIV-V5-MIB::productModelNumber.0
.1.3.6.1.4.1.21239.5.2.1.8.0

root@zbox:/usr/share/snmp/mibs# snmptranslate -On
VERTIV-V5-MIB::productVersion.0
.1.3.6.1.4.1.21239.5.2.1.2.0



/opt/homebrew/bin/fping -C 2 -r 0 -t 1500 -q -g 10.1.249.0 10.1.249.254