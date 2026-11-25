# MCP experiments

## MCP Server → ChatGPT 反向通訊（暫停）

- [x] 合併 Upstream 後保留反向 MCP 通訊研究與 Idle Notification 探測，相關基礎提交：`208137b`。
- [x] 修正 Idle Notification 探測日誌措辭：Control Plane HTTP POST 完成不代表 ChatGPT 已收到或處理通知；現有探測流程可能將 HTTP 404 視為成功，因此不能以成功日誌確認投遞。
- [x] 尚未接入正式通訊流程的 `idleNotificationQueue` 原型與單元測試，保存於獨立分支 `experiment/mcp-idle-notification-queue`（提交 `a3db66a`），不合併至 `master`。原型涵蓋容量限制、Session Generation 隔離、只接受 Notification。
- Server-initiated Request 與 Notification 是不同協定語意，不應視為同一種可排隊訊息。

## 與 Terminal Toolkit MCP Events 的關係

- Terminal Toolkit 的 `--experiment-mcp-events` 是獨立的 Webhook 訂閱／事件投遞路徑，不依賴 Tunnel Client 的 Idle Notification Queue。
- 2026-11-16：Terminal Toolkit 新訂閱的 Webhook 收到 HTTP 200，且 Standard Webhooks 簽章交叉驗證通過，但 ChatGPT Work 對話未顯示事件。HTTP 200 不等於端對端處理成功，尚無法確認問題位於 OpenAI 端。
- 此實驗不足以支持擴充 Tunnel Client 反向通訊架構。後續研究見 [ROADMAP.md](../ROADMAP.md)。

