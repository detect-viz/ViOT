
## 架構劃分

| 功能           | 工具         | 說明           |
|----------------|--------------|--------------|
| 前段自動部署  | **Go 程式**  | 偵測 IP / 自動加 tag 生成 conf |
| 資料收集     | **Telegraf** | 接收 SNMP / MODBUS、processor、scale 處理 |
| 後段數據管理  | **Go 程式**  | 備份 InfluxDB、轉拋任務、監控分析 |


---

## 命名格式

```bash
{PDU_NAME}_{BRAND}-{MODEL}_{IP}_{PORT}_{VERSION}.conf
```

---

### 欄位組成說明：

| 欄位           | 說明                             | 範例                  |
|--------------|-----------------------------------|----------------------|
| `PDU_NAME`   | 定位標籤                           | `F12P7DC1X23PL`     |
| `BRAND`      | 製造商                            | `DELTA`、`VERTIV`      |
| `MODEL`      | 型號縮寫                          | `PDUE428`、`PDU1315`   |
| `IP`         | IP，`.` 轉為 `_`，避免檔案系統問題   | `192_168_23_48`        |
| `PORT`       | 通訊 Port                         | `161`、`502`           |


---

### 統一大寫命名

| 項目 | 說明 |
|------|------|
| **可讀性提升** | 各欄位（如 PDU_NAME、BRAND、MODEL）更容易被辨識，類似常見的環境變數命名風格。 |
| **避免混淆** | 避免大小寫混用造成的判斷錯誤或部署環境行為差異（特別是在 Linux 上區分大小寫）。 |
| **一致性維運** | 在 script、部署工具中處理檔名時，不需要額外處理大小寫轉換邏輯。 |
| **清楚標識變數性質** | 大寫可暗示這些值具有參數性質（如 `${MODEL}`），更適合作為動態產出的檔名模板。 |

---

### 實際範例：

```bash
F12P7DC1X23L_DELTA-PDUE428_10.1.249.37_161_v1.conf
```

### 對應格式：

```bash
{PDU_NAME}_{BRAND}-{MODEL}_{IP}_{PORT}_{VERSION}.conf
```

- `PDU_NAME` → F12P7DC1X23L
- `BRAND` → DELTA
- `MODEL` → PDUE428
- `IP` → 10.1.249.37
- `PORT` → 161
- `VERSION` → v1


---

## 儲存路徑

```bash
/etc/telegraf/telegraf.d/
```

## 設備掃描結果儲存至 `scan_pdu.csv`

```csv
ip_key,factoryphase,datacenter,room,model,protocol,version,serial_number,mac_address,instance_type,manufacturer,uptime,updated_at

```

---

## Web UI 輸入 rack,side 點擊綁定

- 生成
- 測試 telegraf --config xxx.conf --test --debug
- 加入  /etc/telegraf/telegraf.d
---

## 確認部署後儲存至` registry_pdu.csv`
移除 `scan_pdu.csv` 的 pdu 	名單



---

## 錯誤 conf 的監控與自動處理機制

| 功能                     | 說明                                       |
|--------------------------|--------------------------------------------|
| `[[inputs.tail]]`        | 監控 Telegraf 的 log 檔 `/var/log/telegraf/telegraf.log` |
| `grok` 模式解析 log      | 擷取 plugin、level、錯誤訊息              |
| `[[outputs.http]]`       | 將錯誤轉為 JSON 丟到你的 Go 平台 API `/api/log-event` |
| `agent` 基本設定         | 設定 flush 週期、log 輸出、自訂 hostname（可選） |

---

## 📄 `telegraf.conf` 範本

```toml
[agent]
  interval = "10s"
  round_interval = true
  flush_interval = "10s"
  flush_jitter = "0s"
  precision = ""
  hostname = ""
  omit_hostname = true

  debug = true
  quiet = false
  logfile = "/var/log/telegraf/telegraf.log" # <== 自寫 log

###############################################################################
# Inputs
###############################################################################

[[inputs.tail]]
  name_override = "telegraf_log_event"
  files = ["/var/log/telegraf/telegraf.log"]
  from_beginning = false
  watch_method = "inotify"
  data_format = "grok"

  grok_patterns = ['%{CUSTOM_ERROR_LOG}']
  grok_custom_patterns = '''
CUSTOM_ERROR_LOG %{TIMESTAMP_ISO8601:timestamp} %{LOGLEVEL:level} \[%{DATA:plugin}\] %{GREEDYDATA:message}
'''

  tag_keys = ["plugin", "level"]
  fielddrop = ["timestamp"] # timestamp 不需變成 field

###############################################################################
# Outputs
###############################################################################

[[outputs.http]]
  url = "http://YOUR_API_ENDPOINT/api/log-event"
  method = "POST"
  timeout = "5s"
  data_format = "json"
```

