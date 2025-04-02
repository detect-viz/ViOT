# Telegraf 配置檔案生成器

## 功能說明
此腳本用於自動生成不同PDU設備的Telegraf監控配置檔案。根據提供的PDU設備清單(pdu.csv)，自動為每台設備生成對應的Telegraf配置檔案。

## 前置需求
- Bash 環境
- 需要有對應的PDU設備模板檔案
- 需要有包含PDU設備資訊的CSV檔案(pdu.csv)

## CSV檔案格式
pdu.csv 檔案需包含以下欄位（需包含標題行）：
```csv
manufacturer,model,protocol,device_ip,port
```

## 支援的PDU型號

目前支援以下PDU型號配置：

1. Delta PDUE428 (SNMP協議)
2. Delta PDU1315 (Modbus協議)
3. Delta PDU4425 (Modbus協議)
4. Vertiv Geist (SNMP協議)

## 使用方式

1. 確保pdu.csv檔案存在且格式正確
2. 確保template目錄下有對應的模板檔案
3. 執行腳本：

```bash
./generate_telegraf_conf.sh
```

## 輸出結果

- 腳本會在 `conf`目錄下生成配置檔案
- 配置檔案命名格式：`pdu_製造商_協議-IP:埠號.conf`

## 模板檔案要求

需要在template目錄下準備以下模板檔案：

- `template/pdu_Delta_PDUE428_snmp.conf`
- `template/pdu_Delta_modbus.conf`
- `template/pdu_Vertiv_GEIST_snmp.conf`

模板檔案中可使用的變數：

- ${IP}: 設備IP位址
- ${PORT}: 設備埠號
- ${Model}: PDU型號

## 錯誤處理

- 當遇到不支援的PDU型號組合時，會顯示警告訊息並繼續處理下一筆資料
- 當模板檔案不存在時，會顯示警告訊息並繼續處理下一筆資料
