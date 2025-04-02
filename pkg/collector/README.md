# 數據收集服務 (Data Collection Service)

## 概述

數據收集服務是一個用於收集各種設備數據的模組化系統。它支持多種輸入插件（SNMP、Modbus、Telegraf）並提供統一的數據收集和處理介面。

## 架構

```
pkg/collector/
├── core/              # 核心組件
│   ├── input_manager.go    # 輸入插件管理器
│   ├── device_input.go     # 設備輸入基類
│   └── interface.go        # 接口定義
├── inputs/            # 輸入插件實現
│   ├── modbus_collector.go # Modbus 收集器
│   ├── snmp_collector.go   # SNMP 收集器
│   └── telegraf_collector.go # Telegraf 收集器
├── wrapper/           # 數據處理包裝器
├── config.go          # 配置加載
├── inputs.yaml        # 配置文件
└── README.md          # 本文檔
```

## 核心組件

### 1. 輸入插件管理器 (InputManager)

- 管理所有輸入插件的生命週期
- 提供統一的註冊和訪問介面
- 處理插件的啟動、停止和狀態監控

### 2. 設備輸入基類 (DeviceInput)

- 提供基本的設備數據收集功能
- 實現通用的設備操作方法
- 處理狀態追踪和錯誤處理

### 3. 輸入插件

#### SNMP 收集器

- 支持 v1/v2c 協議
- 自動發現和配置設備
- 支持多種設備型號（Delta、Vertiv 等）
- 內置預設配置和參數合併

#### Modbus 收集器

- 支持 TCP 模式
- 自動發現和配置設備
- 支持多種寄存器操作

#### Telegraf 收集器

- 接收 Telegraf 推送的數據
- 支持多種數據格式
- 提供數據緩衝和處理

## 配置說明

### 全局配置

```yaml
global:
  scan_interval: 1h # 掃描間隔
  scan_timeout: 1m # 掃描超時
  scan_concurrent: 10 # 並發數
  auto_deploy: false # 自動部署
```

### 輸入插件配置

每個輸入插件都包含以下基本配置：

```yaml
- name: "input_name" # 插件名稱
  type: "protocol_type" # 協議類型
  enabled: true # 是否啟用
  interval: 60s # 收集間隔

  # IP範圍配置
  ip_ranges:
    - start_ip: "x.x.x.x"
      end_ip: "x.x.x.x"
      protocol: "protocol"
      ports: [port]
```

## 使用方法

### 1. 創建輸入插件

```go
// SNMP 收集器
config := &models.SNMPCollectorConfig{
    Community: "public",
    Version:   "2c",
    Port:     161,
}
collector, err := inputs.NewSNMPCollector(config)
```

### 2. 部署設備

```go
device := models.Device{
    IP:   "192.168.1.100",
    Type: "snmp",
    Tags: map[string]string{
        "community": "public",
        "version":   "2c",
        "port":     "161",
    },
}
err := collector.Deploy(device)
```

### 3. 收集數據

```go
ctx := context.Background()
data, err := collector.Collect(ctx, device)
```

## 數據處理流程

1. 設備發現和配置

   - 自動掃描指定 IP 範圍
   - 識別設備類型和協議
   - 應用預設配置

2. 數據收集

   - 定期輪詢設備數據
   - 處理超時和重試
   - 轉換數據格式

3. 數據處理
   - 數據驗證和過濾
   - 添加標籤和元數據
   - 格式化輸出

## 錯誤處理

- 連接錯誤自動重試
- 設備離線狀態追踪
- 詳細的錯誤日誌
- 可配置的錯誤處理策略

## 擴展性

要添加新的輸入插件：

1. 實現必要的接口方法：

   - `Start()`
   - `Stop()`
   - `Collect()`
   - `Deploy()`
   - `Match()`

2. 提供合適的配置結構
3. 實現數據轉換邏輯

## 注意事項

1. SNMP 配置

   - 使用預設值簡化配置
   - 注意版本兼容性
   - 正確設置超時參數

2. 設備發現

   - 合理設置掃描範圍
   - 控制掃描頻率
   - 處理離線設備

3. 數據收集
   - 避免過於頻繁的輪詢
   - 正確處理數據類型轉換
   - 添加必要的標籤信息
