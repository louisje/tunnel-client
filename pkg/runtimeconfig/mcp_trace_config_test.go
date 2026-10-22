package runtimeconfig

import "testing"

func TestMCPTraceConfigPrecedence(t *testing.T) {
	path := writeTempConfigFile(t, "control_plane:\n  tunnel_id: tunnel_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n  api_key: fake\nmcp:\n  server_urls:\n    - channel: main\n      url: https://mcp.example/mcp\n  trace: true\n  trace-body: true\n  trace-unsafe: true\n")
	cases := []struct {
		name                string
		args                []string
		env                 map[string]string
		trace, body, unsafe bool
	}{
		{"yaml", nil, nil, true, true, true},
		{"env", nil, map[string]string{"MCP_TRACE": "false", "MCP_TRACE_BODY": "false", "MCP_TRACE_UNSAFE": "false"}, false, false, false},
		{"flags", []string{"--mcp.trace=false", "--mcp.trace-body=true", "--mcp.trace-unsafe=false"}, map[string]string{"MCP_TRACE_BODY": "false", "MCP_TRACE_UNSAFE": "true"}, true, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := LoadRuntimeForTest(append([]string{"--config", path}, tc.args...), lookupEnvMap(tc.env))
			if err != nil {
				t.Fatal(err)
			}
			if cfg.MCP.Trace != tc.trace || cfg.MCP.TraceBody != tc.body || cfg.MCP.TraceUnsafe != tc.unsafe {
				t.Fatalf("trace=%v body=%v unsafe=%v", cfg.MCP.Trace, cfg.MCP.TraceBody, cfg.MCP.TraceUnsafe)
			}
		})
	}
}
