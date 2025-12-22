package tool

import (
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func actionPtr(a configv1.ExportPolicy_Action) *configv1.ExportPolicy_Action {
	return &a
}

func TestShouldExport(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		policy   *configv1.ExportPolicy
		want     bool
	}{
		{
			name:     "Nil Policy",
			toolName: "any",
			policy:   nil,
			want:     true,
		},
		{
			name:     "Default Action Unspecified",
			toolName: "any",
			policy:   &configv1.ExportPolicy{},
			want:     true,
		},
		{
			name:     "Default Action Export",
			toolName: "any",
			policy: &configv1.ExportPolicy{
				DefaultAction: actionPtr(configv1.ExportPolicy_EXPORT),
			},
			want: true,
		},
		{
			name:     "Default Action Unexport",
			toolName: "any",
			policy: &configv1.ExportPolicy{
				DefaultAction: actionPtr(configv1.ExportPolicy_UNEXPORT),
			},
			want: false,
		},
		{
			name:     "Rule Match Export",
			toolName: "allowed_tool",
			policy: &configv1.ExportPolicy{
				DefaultAction: actionPtr(configv1.ExportPolicy_UNEXPORT),
				Rules: []*configv1.ExportRule{
					{
						NameRegex: strPtr("^allowed_.*"),
						Action:    actionPtr(configv1.ExportPolicy_EXPORT),
					},
				},
			},
			want: true,
		},
		{
			name:     "Rule Match Unexport",
			toolName: "hidden_tool",
			policy: &configv1.ExportPolicy{
				DefaultAction: actionPtr(configv1.ExportPolicy_EXPORT),
				Rules: []*configv1.ExportRule{
					{
						NameRegex: strPtr("^hidden_.*"),
						Action:    actionPtr(configv1.ExportPolicy_UNEXPORT),
					},
				},
			},
			want: false,
		},
		{
			name:     "Rule No Match Fallthrough",
			toolName: "other_tool",
			policy: &configv1.ExportPolicy{
				DefaultAction: actionPtr(configv1.ExportPolicy_UNEXPORT),
				Rules: []*configv1.ExportRule{
					{
						NameRegex: strPtr("^allowed_.*"),
						Action:    actionPtr(configv1.ExportPolicy_EXPORT),
					},
				},
			},
			want: false,
		},
		{
			name:     "Invalid Regex",
			toolName: "any",
			policy: &configv1.ExportPolicy{
				Rules: []*configv1.ExportRule{
					{
						NameRegex: strPtr("["), // Invalid regex
						Action:    actionPtr(configv1.ExportPolicy_UNEXPORT),
					},
				},
			},
			want: true, // Should continue and use default (true)
		},
		{
			name:     "Empty Regex",
			toolName: "any",
			policy: &configv1.ExportPolicy{
				Rules: []*configv1.ExportRule{
					{
						NameRegex: strPtr(""),
						Action:    actionPtr(configv1.ExportPolicy_UNEXPORT),
					},
				},
			},
			want: true, // Should skipped empty regex
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldExport(tt.toolName, tt.policy)
			assert.Equal(t, tt.want, got)
		})
	}
}

func strPtr(s string) *string { return &s }

func callActionPtr(a configv1.CallPolicy_Action) *configv1.CallPolicy_Action {
	return &a
}

func TestEvaluateCallPolicy(t *testing.T) {
	tests := []struct {
		name      string
		policies  []*configv1.CallPolicy
		toolName  string
		callID    string
		arguments []byte
		want      bool
	}{
		{
			name:     "No Policies",
			policies: nil,
			toolName: "any",
			want:     true,
		},
		{
			name: "Default Deny",
			policies: []*configv1.CallPolicy{
				{DefaultAction: callActionPtr(configv1.CallPolicy_DENY)},
			},
			toolName: "any",
			want:     false,
		},
		{
			name: "Default Allow",
			policies: []*configv1.CallPolicy{
				{DefaultAction: callActionPtr(configv1.CallPolicy_ALLOW)},
			},
			toolName: "any",
			want:     true,
		},
		{
			name: "Name Regex Deny",
			policies: []*configv1.CallPolicy{
				{
					DefaultAction: callActionPtr(configv1.CallPolicy_ALLOW),
					Rules: []*configv1.CallPolicyRule{
						{
							NameRegex: strPtr("^dangerous_.*"),
							Action:    callActionPtr(configv1.CallPolicy_DENY),
						},
					},
				},
			},
			toolName: "dangerous_tool",
			want:     false,
		},
		{
			name: "Argument Regex Deny",
			policies: []*configv1.CallPolicy{
				{
					DefaultAction: callActionPtr(configv1.CallPolicy_ALLOW),
					Rules: []*configv1.CallPolicyRule{
						{
							ArgumentRegex: strPtr(`.*"secret".*`),
							Action:        callActionPtr(configv1.CallPolicy_DENY),
						},
					},
				},
			},
			toolName:  "any",
			arguments: json.RawMessage(`{"key": "secret"}`),
			want:      false,
		},
		{
			name: "Argument Regex Static Check (Nil Args)",
			policies: []*configv1.CallPolicy{
				{
					DefaultAction: callActionPtr(configv1.CallPolicy_ALLOW),
					Rules: []*configv1.CallPolicyRule{
						{
							ArgumentRegex: strPtr(`.*`),
							Action:        callActionPtr(configv1.CallPolicy_DENY),
						},
					},
				},
			},
			toolName:  "any",
			arguments: nil,
			want:      true, // Rule with arg regex ignored when args are nil
		},
		{
			name: "CallID Regex Deny",
			policies: []*configv1.CallPolicy{
				{
					DefaultAction: callActionPtr(configv1.CallPolicy_ALLOW),
					Rules: []*configv1.CallPolicyRule{
						{
							CallIdRegex: strPtr("call_123"),
							Action:      callActionPtr(configv1.CallPolicy_DENY),
						},
					},
				},
			},
			toolName: "any",
			callID:   "call_123",
			want:     false,
		},
		{
			name: "Multiple Policies (One blocks)",
			policies: []*configv1.CallPolicy{
				{DefaultAction: callActionPtr(configv1.CallPolicy_ALLOW)},
				{DefaultAction: callActionPtr(configv1.CallPolicy_DENY)},
			},
			toolName: "any",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateCallPolicy(tt.policies, tt.toolName, tt.callID, tt.arguments)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEvaluateCompiledCallPolicy(t *testing.T) {
	tests := []struct {
		name      string
		policies  []*configv1.CallPolicy
		toolName  string
		callID    string
		arguments []byte
		want      bool
	}{
		{
			name:     "No Policies",
			policies: nil,
			toolName: "any",
			want:     true,
		},
		{
			name: "Default Deny",
			policies: []*configv1.CallPolicy{
				{DefaultAction: callActionPtr(configv1.CallPolicy_DENY)},
			},
			toolName: "any",
			want:     false,
		},
		{
			name: "Default Allow",
			policies: []*configv1.CallPolicy{
				{DefaultAction: callActionPtr(configv1.CallPolicy_ALLOW)},
			},
			toolName: "any",
			want:     true,
		},
		{
			name: "Name Regex Deny",
			policies: []*configv1.CallPolicy{
				{
					DefaultAction: callActionPtr(configv1.CallPolicy_ALLOW),
					Rules: []*configv1.CallPolicyRule{
						{
							NameRegex: strPtr("^dangerous_.*"),
							Action:    callActionPtr(configv1.CallPolicy_DENY),
						},
					},
				},
			},
			toolName: "dangerous_tool",
			want:     false,
		},
		{
			name: "Argument Regex Deny",
			policies: []*configv1.CallPolicy{
				{
					DefaultAction: callActionPtr(configv1.CallPolicy_ALLOW),
					Rules: []*configv1.CallPolicyRule{
						{
							ArgumentRegex: strPtr(".*secret.*"),
							Action:        callActionPtr(configv1.CallPolicy_DENY),
						},
					},
				},
			},
			toolName:  "any",
			arguments: json.RawMessage("{\"key\": \"secret\"}"),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compiled, err := CompileCallPolicies(tt.policies)
			assert.NoError(t, err)
			got, err := EvaluateCompiledCallPolicy(compiled, tt.toolName, tt.callID, tt.arguments)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompileCallPolicies_InvalidRegex(t *testing.T) {
	policies := []*configv1.CallPolicy{
		{
			Rules: []*configv1.CallPolicyRule{
				{NameRegex: strPtr("[")},
			},
		},
	}
	_, err := CompileCallPolicies(policies)
	assert.Error(t, err)
}
