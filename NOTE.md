# NOTE

Tunnel Client 尚待追蹤的實驗與限制；已完成的實作以 Git 紀錄為準。

## MCP Server → ChatGPT 反向通訊（暫停）

- [x] 合併 Upstream 後保留反向 MCP 通訊研究與 Idle Notification 探測，相關基礎提交：`208137b`。
- [x] 修正 Idle Notification 探測日誌措辭：Control Plane HTTP POST 完成不代表 ChatGPT 已收到或處理通知；現有探測流程可能將 HTTP 404 視為成功，因此不能以成功日誌確認投遞。
- [x] 尚未接入正式通訊流程的 `idleNotificationQueue` 原型與單元測試，保存於獨立分支 `experiment/mcp-idle-notification-queue`（提交 `a3db66a`），不合併至 `master`。原型涵蓋容量限制、Session Generation 隔離、只接受 Notification。
- [ ] 待 ChatGPT／Control Plane 對 Server-initiated MCP Notification 的支援與回饋機制更明確後，重新驗證 Idle Notification 的端對端投遞；必要時再評估 Queue 接入與 Session 生命週期。
- [ ] 分別驗證 Server-initiated Request（例如 `sampling/createMessage`）與 Notification；不要將兩者視為同一種可排隊訊息。

## 與 Terminal Toolkit MCP Events 的關係

- Terminal Toolkit 的 `--experiment-mcp-events` 是獨立的 Webhook 訂閱／事件投遞路徑，不依賴 Tunnel Client 的 Idle Notification Queue。
- 2026-11-16：Terminal Toolkit 新訂閱的 Webhook 收到 HTTP 200，且 Standard Webhooks 簽章交叉驗證通過，但 ChatGPT Work 對話未顯示事件。HTTP 200 不等於端對端處理成功，尚無法確認問題位於 OpenAI 端。
- 現階段不因 MCP Events 實驗而擴充 Tunnel Client 的反向通訊架構；等待後續重新測試。
