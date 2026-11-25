# ROADMAP

記錄未來規劃、待辦、研究及暫緩事項，不代表已承諾時程。隨開發進程即時滾動更新，主動淘汰已完成、失效或被取代的事項；歷史由 Git 保存。開始實作時標記已啟動，當前進度與交接以 [NOTE.md](NOTE.md) 為準。維護方式見 [Development workflow](docs/development-workflow.md)。

## Backlog

- **mTLS certificate automation**: operator tooling for certificate issuance and rotation.
- **Control-plane allowlisting**: evaluate IP-based controls or tighter network restrictions.

## Research

- **Readiness**: assess remaining gaps against the existing [health and readiness model](docs/health.md); do not duplicate implemented connectivity and startup checks.
- **Tracing**: assess remaining end-to-end OpenTelemetry span coverage across poll, MCP and response; existing trace-context propagation is documented in [configuration](docs/configuration.md).

## Deferred

- **MCP reverse communication / progress forwarding**: revisit when ChatGPT / Control Plane support and feedback are clearer. Validate server-initiated notifications and requests (e.g. sampling) separately; evaluate queue integration and session lifecycle only after delivery is demonstrated. See [MCP experiments](docs/mcp-experiments.md).
