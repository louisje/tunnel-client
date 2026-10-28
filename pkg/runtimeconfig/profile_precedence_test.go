package runtimeconfig

import "testing"

func TestProfileValuesOverrideEnvironment(t *testing.T) {
	ambient := map[string]string{"LOG_LEVEL": "debug", "ONLY_ENV": "fallback"}
	lookup := func(key string) (string, bool) { v, ok := ambient[key]; return v, ok }
	cases := []struct {
		name   string
		values *fileConfigValues
		want   string
	}{
		{"named profile", &fileConfigValues{ProfileName: "terminal-mcp", Env: map[string]string{"LOG_LEVEL": "info"}}, "info"},
		{"profile file", &fileConfigValues{ProfileFile: true, Env: map[string]string{"LOG_LEVEL": "warn"}}, "warn"},
		{"regular config", &fileConfigValues{Env: map[string]string{"LOG_LEVEL": "info"}}, "debug"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			merged := lookupEnvWithFileValues(lookup, tc.values)
			if got, _ := merged("LOG_LEVEL"); got != tc.want {
				t.Fatalf("level=%q want %q", got, tc.want)
			}
			if got, _ := merged("ONLY_ENV"); got != "fallback" {
				t.Fatalf("fallback=%q", got)
			}
		})
	}
}
