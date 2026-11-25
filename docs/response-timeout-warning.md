# Response timeout warning

## Scope and contract

This fork-specific extension complements the control-plane [response deadline](protocol.md). It does not change the upstream timeout or extend the deadline.

For stdio MCP `tools/call` with an actual control-plane response deadline and an original budget greater than 10 seconds, the dispatcher schedules a best-effort notification 10 seconds before the deadline:

```json
{"jsonrpc":"2.0","method":"notifications/tunnel/response_timeout_warning","params":{"requestId":"example","remainingMs":10000}}
```

The request ID retains its JSON-RPC type. Remaining time reflects the actual deadline. No warning is sent for absent deadlines or budgets of 10 seconds or less. Normal completion or cancellation stops the timer. The write is bounded and uses the original active stdio connection; no retry or deadline extension is implied.

A server may acknowledge with `notifications/tunnel/response_timeout_ack`. Tunnel Client logs the acknowledgement, does not forward it to the control plane, and does not wait for it. An ACK is not a tool response or proof of completion.

Terminal Toolkit can use this warning to detach shell/python executions and return an execution ID, or return current waitExecution metadata in versions supporting that behavior. Unsupported tools or servers may ignore it. This is not a standard MCP notification.

## Deadline enforcement

The existing deadline remains authoritative. On expiry, late responses are discarded and the pending request is retired where safe; the warning does not replace cancellation or prove that server work stopped. The optional metadata hint `_meta["terminal-toolkit/responseDeadline"]` communicates timeoutMs, remainingMs and deadline without changing the upstream contract.

## Implementation and verification

- Dispatcher: `pkg/dispatcher/internal/processor.go`, `response_timeout_warning.go`, `response_deadline_meta.go`.
- Stdio transport: `pkg/mcpclient/serialized_forwarding_transport.go`.
- Targeted Go tests (including race checks) and Terminal Toolkit headless/proxy tests passed for the initial implementation. Basic deployed connectivity was verified; actual upstream warning timing was not deliberately triggered. Current operational follow-up belongs in [NOTE.md](../NOTE.md).
