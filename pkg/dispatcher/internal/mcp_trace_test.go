package dispatcherinternal

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

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

func TestMCPTraceEmbeddedSecrets(t *testing.T) {
	for _, input := range []string{
		"Authorization: Bearer abc123xyz",
		"token sk-proj-exampleSecret123",
		"github_pat_testingSecret123",
	} {
		if got := redactMCPString(input); strings.Contains(got, "Secret123") || strings.Contains(got, "abc123xyz") {
			t.Fatalf("secret leaked: %s", got)
		}
	}
}

func TestMCPTraceMessageKinds(t *testing.T) {
	id, err := jsonrpc.MakeID("request-17")
	if err != nil {
		t.Fatal(err)
	}
	messages := []struct {
		name    string
		message jsonrpc.Message
		expect  []string
	}{
		{"call", &jsonrpc.Request{ID: id, Method: "tools/call", Params: json.RawMessage(`{"name":"example"}`)}, []string{"outbound", "tools/call", "request-17"}},
		{"notification", &jsonrpc.Request{Method: "notifications/progress", Params: json.RawMessage(`{"progress":1}`)}, []string{"inbound", "notifications/progress"}},
		{"response", &jsonrpc.Response{ID: id, Result: json.RawMessage(`{"ok":true}`)}, []string{"inbound", "request-17", "rpc_method=tools/call", "elapsed="}},
		{"error", &jsonrpc.Response{ID: id, Error: &jsonrpc.Error{Code: -32603, Message: "synthetic failure"}}, []string{"inbound", "request-17", "synthetic failure"}},
	}
	for _, tc := range messages {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&buf, nil))
			p := &mcpProcessor{traceMCP: true, traceMCPBody: true}
			direction, method := "inbound", "tools/call"
			if tc.name == "call" {
				direction = "outbound"
				method = ""
			}
			p.logMCPMessage(context.Background(), logger, direction, "main", tc.message, method, time.Millisecond)
			for _, part := range tc.expect {
				if !strings.Contains(buf.String(), part) {
					t.Fatalf("missing %q in %s", part, buf.String())
				}
			}
		})
	}
}
