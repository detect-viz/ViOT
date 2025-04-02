**PDU 綁定頁面** 的完整功能說明

---

### 🔧 頁面名稱

`PDU 綁定作業`

---

### 範例操作流程

使用者選擇 DC → Room
中間顯示該房間的 Grid
從左側清單點選一筆 PDU（IP + SN）
點選 Grid 中的某個位置（如 E99L）
彈出綁定 Dialog（顯示 SN、IP、選定櫃位）
點「確認綁定」 → 送出 /api/bind-pdu
UI 即時更新（綠燈）、左側清單移除該筆

### 📌 頁面分區說明

---

#### ✅ Top bar 區

- **右側 Search bar**
- 機櫃快速搜尋 由 [/api/pdu-list] 提供資料本地過濾 允許搜尋欄位或座標（如 D102L）來快速定位 Grid 目標

#### ✅ 左側功能區：

**DC 快速篩選**

- **用途**：快速切換不同資料中心（DC）。
- **元件**：

  - `m3-segmented-button` 切換 DC。
  - 每個 DC 的 `button` 可以顯示紅點 + 數字，代表待綁定數量（badge）。

- **API**：GET [/api/datacenters]

```json
[
  {
    "dc_name": "DC1",
    "unbind": 5
  },
  {
    "dc_name": "DC2",
    "unbind": 0
  }
]
```

---

**待綁定 PDU 清單 + 手動輸入**

- **API**：GET [/api/pdu-pending-list?dc=DC1]

```json
[
  {
    "ip": "172.9.5.6",
    "room": "R1",
    "mfg": "Detla",
    "model": "PDUE428",
    "sn": "Y67589054",
    "update_at": "2025-02-23 09:33:00"
  },
  {
    "ip": "172.9.5.9",
    "room": "R5",
    "mfg": "Detla",
    "model": "PDUE428",
    "sn": "Y67589059",
    "update_at": "2025-02-23 09:33:00"
  }
]
```

- **用途**：顯示「待綁定」的 PDU 裝置清單，從即時掃描的結果中取得。
- **項目格式**：
  - 使用 `m3-list` 搭配 `checkbox`（可勾選一筆）。
  - 勾選後根據 room 切換右側 tab 該 Room 的機櫃區域。
- **互動功能**：
  - 勾選後，會顯示詳細資訊卡片。
  - 卡片最下方有一個「手動輸入綁定」按鈕，
    點擊後顯示輸入櫃位 <md-text-field>，
    `L / R` 選擇（使用 `m3-radio-button`）。
    - POST [/api/validate-rack] → 驗證手動輸入櫃位
      成功 → 允許綁定 POST [/api/bind-pdu]
      失敗 → 提示錯誤訊息

```json
{
  "rack": "R119"
}
// 回應
{ "valid": true, "message": "櫃位可使用" }
{ "valid": false, "message": "櫃位格式錯誤" }

```

POST [/api/bind-pdu]

```json
{
  "ip": "192.168.21.87",
  "rack": "R8"  // 允許手動輸入的櫃位
  "side": "R"
}
// 回應
{ "success": true, "message": "PDU 綁定成功" }

```

---

#### ✅ 右側功能區：

**Room 快速篩選**

- **用途**：快速切換不同 Room 的機櫃區域。

- **元件**：

  - `m3-tabs` 切換 Room（例：R0 ～ R7）。
  - 每個 Room 的 `tab` 可以顯示紅點 + 數字，代表待綁定數量（badge）。

- **API**：GET [/api/rooms?dc=DC1]

```json
[
  {
    "room_name": "R0",
    "unbind": 2
  },
  {
    "room_name": "R1",
    "unbind": 0
  }
]
```

**機櫃編號 Grid**

- **用途**：以網格呈現所有機櫃，每格代表一個 PDU 。

  - 固定 col = 8，顯示字為欄位 `display_name`

- **API**：GET [/api/pdu-list?dc=DC1&room=R3]

```json
[
  {
    "display_name": "D99 L",
    "pdu_name": "F12P7DC1D99PL",
    "is_pdu": 1,
    "monitored": 1
  },
  {
    "display_name": "D99 R",
    "pdu_name": "F12P7DC1D99R",
    "is_pdu": 1,
    "monitored": 0
  },
  {
    "display_name": "G107 L",
    "pdu_name": "",
    "is_pdu": 0,
    "monitored": 0
  },
  {
    "display_name": "G107 R",
    "pdu_name": "",
    "is_pdu": 0,
    "monitored": 0
  }
]
```

- **互動方式**：
  - **顏色代表狀態**：
    - **灰色底**：尚未綁定 `monitored=0`，點擊格子 → 彈窗 **綁定表單**。會觸發 `/api/bind`）
    - **綠色底**：已綁定 `monitored=1`，點擊格子 → 彈窗 **綁定表單(取消)** （會觸發 `/api/unbind`）
    - **透明底**：空機櫃 `is_pdu=0`，以 `disabled` 狀態呈現，不可點擊。

---

**尚未綁定 PDU 綁定彈窗 / 表單**

- **用途**：用於顯示選定 PDU 或 IP 的詳細資訊，並進行綁定操作。
- 沒有勾選 IP，點擊未綁定 PDU 要彈窗提示，請先勾選 IP
- **顯示項目**：
  - IP 地址
  - PDU 的標準名稱（例如：F12P7DC1C102PL）
- **操作按鈕**：
  - `綁定`（會觸發 `/api/bind`）
  - `返回`

---

**已綁定 PDU 綁定彈窗 / 表單**

- **用途**：顯示選定 PDU 詳細資訊。
- **顯示項目**：
  - IP 地址
  - PDU 的標準名稱（例如：F12P7DC1C102PL）
- **操作按鈕**：

  - `取消綁定`（會觸發 `/api/unbind`）
  - `返回`

- **API**：POST [/api/unbind]

```json
{
  "pdu_name": "F12P7DC1C102PL"
}
```

---

### 🔄 背景邏輯（自動）

- 每 5 ～ 10 秒會呼叫 `/api/pdu-list` 自動更新列表。(設變數，集中到 config 管理)
- Grid 也會自動根據狀態變色。(設變數，集中到 config 管理)
- 綁定完成後，即時更新畫面，該 PDU 從候選清單中移除，Grid 標記為已綁定。
