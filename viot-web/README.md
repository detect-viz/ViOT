# ViOT 控制台綁定畫面

本專案為一個使用 Material Web Components + HTMX + Golang 開發的 PDU 綁定視覺化操作平台，適用於管理機房中多個 PDU 設備的配置與狀態。

## 專案結構

結構分為：

- 頁面元件位於 `components/` 下，每個元件皆含 html / css / js。
- `demo/` 資料夾用來模擬 API 回傳 JSON。
- 所有組件透過 HTMX 載入，具備畫面分離與重用性。

```bash
.
├── components # 所有 Web Component 拆分在此
│ ├── app-header # 頁面上方 Header + Search bar
│ ├── app-navigation-rail # 側邊欄導航（支援收合）
│ ├── dc-selector # DC1 / DC2 選擇元件（segmented-button）
│ ├── room-tabs # Room 切換的 Tabs（R0 ~ R7）
│ ├── pdu-list # 左側 PDU 待綁定列表
│ ├── pdu-sidebar # 顯示 PDU 詳細資訊與手動綁定按鈕
│ ├── pdu-grid # 中央機櫃 Grid 顯示（根據 room 呈現）
│ └── pdu-dialog # 點選機櫃彈出的綁定或取消綁定視窗
├── pages
│ ├── pdu-deploy.html # 綁定主頁面
│ ├── page1.html # 其他保留頁面
│ ├── page2.html
│ └── page3.html
├── css/style.css # 共用樣式（含主題變數）
├── js/app.js # 全域初始化邏輯
├── demo # JSON 模擬資料
│ ├── datacenters.json
│ ├── rooms.json
│ ├── pdu-list.json
│ ├── pdu-pending.json
│ ├── pdu-172.7.34.5.json # 單一 IP 的詳細資料
├── main.go # 提供靜態頁面與模擬 API 的 Go 伺服器
├── index.html # 預設首頁（可重導向到 /pdu-deploy.html）
└── README.md

```

---

## 主版型說明

| 區塊              | 元件                    | API 載入方式            |
| ----------------- | ----------------------- | ----------------------- |
| `app-header`      | `components/app-header` | HTMX beforeend 載入     |
| `navigation-rail` | 包在 header 裡          | 同上                    |
| `dc-selector`     | 左側 DC 篩選器          | `/api/datacenters`      |
| `pdu-list`        | 左側未綁定 IP 列表      | `/api/pdu-pending-list` |
| `pdu-sidebar`     | 左側 PDU 詳細資訊卡     | `/api/pdu/:id`          |
| `room-tabs`       | 上方 Room 切換 Tab      | `/api/rooms?dc=DC1`     |
| `pdu-grid`        | 右側 Grid 顯示櫃位      | `/api/pdu-list`         |
| `pdu-dialog`      | 綁定／取消綁定對話框    | 互動式 `/api/bind` 等   |

---

## 技術棧

- **UI Framework**：Material Web 3 Components (esm.run)
- **互動引擎**：HTMX + 原生 JS
- **後端模擬 API**：Golang + Gin
- **資料格式**：JSON 靜態檔模擬實際後端格式

---

## API 模擬邏輯

| 路徑                    | 用途                                             |
| ----------------------- | ------------------------------------------------ |
| `/api/datacenters`      | 取得所有資料中心（DC）及其待綁定數量             |
| `/api/rooms`            | 取得所有 Room 列表（含所屬 DC）                  |
| `/api/pdu-pending-list` | 取得左側待綁定 IP 列表                           |
| `/api/pdu-list`         | 取得當前 DC + Room 的所有 PDU 格子（含綁定狀態） |
| `/api/pdu/:ip`          | 取得單一 PDU 詳細資料                            |
| `/api/bind`             | 模擬綁定 API（POST）                             |
| `/api/unbind`           | 模擬解除綁定 API（POST）                         |

---

## 頁面互動邏輯

- 點選 DC → 切換資料中心並更新所有相關元件資料
- 點選 Tabs（R0~R7）→ 切換 Room Grid
- 點選格子（PDU）→ 根據綁定狀態開啟綁定或取消視窗
- 點選手動綁定 → 顯示表單 + Side 選擇，送出呼叫 /api/bind
