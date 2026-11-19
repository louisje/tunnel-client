package dispatcherinternal

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
)

func TestForwardResponseDeadlinePreservesMeta(t *testing.T) {
	polled := time.Now()
	req := &jsonrpc.Request{Method: "tools/call", Params: json.RawMessage(`{"name":"shell","_meta":{"openai/subject":"opaque"}}`)}
	out := forwardResponseDeadline(req, polled.Add(30*time.Second), polled)
	if out == req {
		t.Fatal("expected copy")
	}
	var params struct {
		Meta map[string]json.RawMessage `json:"_meta"`
	}
	if err := json.Unmarshal(out.Params, &params); err != nil {
		t.Fatal(err)
	}
	if string(params.Meta["openai/subject"]) != `"opaque"` {
		t.Fatal("lost original metadata")
	}
	var hint struct {
		TimeoutMs   int64 `json:"timeoutMs"`
		RemainingMs int64 `json:"remainingMs"`
	}
	if err := json.Unmarshal(params.Meta[responseDeadlineMetaKey], &hint); err != nil {
		t.Fatal(err)
	}
	if hint.TimeoutMs != 30000 || hint.RemainingMs < 0 || hint.RemainingMs > 30000 {
		t.Fatalf("invalid hint: %+v", hint)
	}
	if string(req.Params) != `{"name":"shell","_meta":{"openai/subject":"opaque"}}` {
		t.Fatal("mutated original")
	}
}
func TestForwardResponseDeadlineSkipsInvalidParams(t *testing.T) {
	req := &jsonrpc.Request{Params: json.RawMessage(`{"_meta":null}`)}
	if forwardResponseDeadline(req, time.Now().Add(time.Second), time.Now()) != req {
		t.Fatal("must preserve invalid metadata")
	}
}
