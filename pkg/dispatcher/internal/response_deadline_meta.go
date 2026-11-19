package dispatcherinternal

import (
	"encoding/json"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
)

const responseDeadlineMetaKey = "terminal-toolkit/responseDeadline"

// forwardResponseDeadline preserves existing request metadata and adds an
// optional, transport-independent hint. Invalid/ambiguous metadata is left intact.
func forwardResponseDeadline(req *jsonrpc.Request, deadline time.Time, polledAt time.Time) *jsonrpc.Request {
	if req == nil || deadline.IsZero() || len(req.Params) == 0 {
		return req
	}
	params, ok := decodeTraceObject(req.Params)
	if !ok {
		return req
	}
	meta := map[string]json.RawMessage{}
	if raw, exists := params["_meta"]; exists {
		meta, ok = decodeTraceObject(raw)
		if !ok {
			return req
		}
	}
	if _, exists := meta[responseDeadlineMetaKey]; exists {
		return req
	}
	remaining := time.Until(deadline).Milliseconds()
	if remaining < 0 {
		remaining = 0
	}
	timeout := deadline.Sub(polledAt).Milliseconds()
	if timeout < 0 {
		timeout = 0
	}
	hint, err := json.Marshal(struct {
		TimeoutMs   int64  `json:"timeoutMs"`
		RemainingMs int64  `json:"remainingMs"`
		Deadline    string `json:"deadline"`
	}{timeout, remaining, deadline.UTC().Format(time.RFC3339Nano)})
	if err != nil {
		return req
	}
	meta[responseDeadlineMetaKey] = hint
	encodedMeta, err := json.Marshal(meta)
	if err != nil {
		return req
	}
	params["_meta"] = encodedMeta
	encodedParams, err := json.Marshal(params)
	if err != nil {
		return req
	}
	copied := *req
	copied.Params = encodedParams
	return &copied
}
