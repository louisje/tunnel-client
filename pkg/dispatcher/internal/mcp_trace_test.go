package dispatcherinternal

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
)

func TestRedactMCPPayload(t *testing.T) {
	payload := map[string]any{"params": map[string]any{"api_key": "abc", "nested": []any{map[string]any{"access_token": "secret", "value": "ok"}}}}
	result := redactMCPPayload(payload).(map[string]any)["params"].(map[string]any)
	if result["api_key"] != "[REDACTED]" {
		t.Fatal("api key not redacted")
	}
	item := result["nested"].([]any)[0].(map[string]any)
	if item["access_token"] != "[REDACTED]" || item["value"] != "ok" {
		t.Fatalf("unexpected payload: %v", item)
	}
}
func TestMCPTraceIndependentOfLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	p := &mcpProcessor{traceMCP: true, traceMCPBody: true}
	p.logMCPMessage(context.Background(), logger, "outbound", "main", &jsonrpc.Request{Method: "tools/list"}, "", 0)
	got := buf.String()
	if !strings.Contains(got, "mcp json-rpc trace") || !strings.Contains(got, "tools/list") {
		t.Fatalf("missing trace: %s", got)
	}
}

func TestMCPTraceDisabled(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	p := &mcpProcessor{}
	p.logMCPMessage(context.Background(), logger, "outbound", "main", &jsonrpc.Request{Method: "tools/list"}, "", 0)
	if buf.Len() != 0 {
		t.Fatalf("disabled trace emitted: %s", buf.String())
	}
}

func TestMCPTraceUnsafeVsRedacted(t *testing.T) {
	id, err := jsonrpc.MakeID("abc")
	if err != nil {
		t.Fatal(err)
	}
	req := &jsonrpc.Request{ID: id, Method: "tools/call", Params: json.RawMessage(`{"arguments":{"api_key":"private-credential","nested":{"access_token":"private-token"},"ordinary":"hello"}}`)}
	for _, tc := range []struct {
		name       string
		unsafe     bool
		wantSecret bool
	}{{"safe", false, false}, {"unsafe", true, true}} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
			p := &mcpProcessor{traceMCP: true, traceMCPBody: true, traceMCPUnsafe: tc.unsafe}
			p.logMCPMessage(context.Background(), logger, "outbound", "main", req, "", 0)
			got := buf.String()
			if strings.Contains(got, "private-credential") != tc.wantSecret || strings.Contains(got, "private-token") != tc.wantSecret {
				t.Fatalf("unexpected redaction: %s", got)
			}
			if !strings.Contains(got, "hello") {
				t.Fatalf("non-sensitive body missing: %s", got)
			}
		})
	}
}
