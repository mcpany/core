package health

import (
	"testing"
)

func TestCheckAuth(t *testing.T) {
	allKeys := []string{
		"ANTHROPIC_API_KEY", "OPENAI_API_KEY", "GEMINI_API_KEY",
		"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
		"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
	}

	tests := []struct {
		name       string
		envVars    map[string]string
		wantChecks map[string]CheckResult
	}{
		{
			name:    "No Env Vars",
			envVars: map[string]string{},
			wantChecks: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "missing", Message: "Environment variable not set"},
				"OPENAI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"GEMINI_API_KEY":    {Status: "missing", Message: "Environment variable not set"},
				"oauth_GOOGLE":      {Status: "info", Message: "Not configured"},
				"oauth_GITHUB":      {Status: "info", Message: "Not configured"},
			},
		},
		{
			name: "API Keys Present",
			envVars: map[string]string{
				"ANTHROPIC_API_KEY": "sk-ant-1234567890",
			},
			wantChecks: map[string]CheckResult{
				"ANTHROPIC_API_KEY": {Status: "ok", Message: "Present (...7890)"},
			},
		},
		{
			name: "Short API Key",
			envVars: map[string]string{
				"OPENAI_API_KEY": "123",
			},
			wantChecks: map[string]CheckResult{
				"OPENAI_API_KEY": {Status: "ok", Message: "Present"},
			},
		},
		{
			name: "OAuth Complete",
			envVars: map[string]string{
				"GOOGLE_CLIENT_ID":     "foo",
				"GOOGLE_CLIENT_SECRET": "bar",
			},
			wantChecks: map[string]CheckResult{
				"oauth_GOOGLE": {Status: "ok", Message: "Configured"},
			},
		},
		{
			name: "OAuth Partial (Missing Secret)",
			envVars: map[string]string{
				"GITHUB_CLIENT_ID": "foo",
			},
			wantChecks: map[string]CheckResult{
				"oauth_GITHUB": {Status: "warning", Message: "Partial configuration (missing ID or Secret)"},
			},
		},
		{
			name: "OAuth Partial (Missing ID)",
			envVars: map[string]string{
				"GITHUB_CLIENT_SECRET": "bar",
			},
			wantChecks: map[string]CheckResult{
				"oauth_GITHUB": {Status: "warning", Message: "Partial configuration (missing ID or Secret)"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all keys first to ensure clean state
			for _, k := range allKeys {
				t.Setenv(k, "")
			}
			// Set specific keys for this test case
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			got := CheckAuth()

			for k, want := range tt.wantChecks {
				if gotCheck, ok := got[k]; !ok {
					t.Errorf("CheckAuth() missing check for %s", k)
				} else {
					if gotCheck.Status != want.Status {
						t.Errorf("CheckAuth() %s status = %v, want %v", k, gotCheck.Status, want.Status)
					}
					if gotCheck.Message != want.Message {
						t.Errorf("CheckAuth() %s message = %v, want %v", k, gotCheck.Message, want.Message)
					}
				}
			}
		})
	}
}
