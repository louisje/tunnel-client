package dispatcherinternal

import (
	"context"
	"encoding/json"
	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/openai/tunnel-client/pkg/types"
	"log/slog"
	"regexp"
	"strings"
	"time"
)

func (p *mcpProcessor) logMCPMessage(ctx context.Context, logger *slog.Logger, direction string, channel types.Channel, msg jsonrpc.Message, method string, duration time.Duration) {
	if !p.traceMCP {
		return
	}
	attrs := []slog.Attr{slog.String("direction", direction), slog.String("channel", channel.String())}
	attrs = append(attrs, messageSummaryAttrs(msg)...)
	if method != "" {
		attrs = append(attrs, slog.String("rpc_method", method))
	}
	if direction == "inbound" && method != "" {
		attrs = append(attrs, slog.Duration("elapsed", duration))
	}
	if p.traceMCPBody {
		raw, err := jsonrpc.EncodeMessage(msg)
		if err != nil {
			attrs = append(attrs, slog.String("payload_error", err.Error()))
		} else if p.traceMCPUnsafe {
			attrs = append(attrs, slog.String("payload", string(raw)))
		} else {
			var value any
			if json.Unmarshal(raw, &value) == nil {
				attrs = append(attrs, slog.Any("payload", redactMCPPayload(value)))
			}
		}
	}
	// Bypass the minimum log level without modifying other application logging.
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "mcp json-rpc trace", 0)
	record.AddAttrs(attrs...)
	_ = logger.Handler().Handle(ctx, record)
}

var inlineSecret = regexp.MustCompile("(?i)(bearer\\s+)[A-Za-z0-9._~+/-]+|(?:sk-proj-|sk-|ghp_|github_pat_)[A-Za-z0-9_\\-]+")

func redactMCPString(s string) string {
	return inlineSecret.ReplaceAllStringFunc(s, func(match string) string {
		if len(match) >= 7 && strings.EqualFold(match[:7], "Bearer ") {
			return match[:7] + "[REDACTED]"
		}
		return "[REDACTED]"
	})
}
func redactMCPPayload(v any) any {
	switch x := v.(type) {
	case map[string]any:
		for key, val := range x {
			k := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "-", ""), "_", ""))
			if strings.Contains(k, "token") || strings.Contains(k, "secret") || strings.Contains(k, "password") || strings.Contains(k, "authorization") || strings.Contains(k, "apikey") || strings.Contains(k, "credential") || strings.Contains(k, "privatekey") || k == "cookie" || k == "setcookie" {
				x[key] = "[REDACTED]"
			} else {
				x[key] = redactMCPPayload(val)
			}
		}
	case string:
		return redactMCPString(x)
	case []any:
		for i, val := range x {
			x[i] = redactMCPPayload(val)
		}
	}
	return v
}
