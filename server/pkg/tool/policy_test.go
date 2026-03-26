// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func actionPtr(a configv1.ExportPolicy_Action) *configv1.ExportPolicy_Action {
	return &a
}

func TestShouldExport(t *testing.T) {
	t.Parallel()
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
			policy:   configv1.ExportPolicy_builder{}.Build(),
			want:     true,
		},
		{
			name:     "Default Action Export",
			toolName: "any",
			policy: configv1.ExportPolicy_builder{
				DefaultAction: configv1.ExportPolicy_EXPORT.Enum(),
			}.Build(),
			want: true,
		},
		{
			name:     "Default Action Unexport",
			toolName: "any",
			policy: configv1.ExportPolicy_builder{
				DefaultAction: configv1.ExportPolicy_UNEXPORT.Enum(),
			}.Build(),
			want: false,
		},
		{
			name:     "Rule Match Export",
			toolName: "allowed_tool",
			policy: configv1.ExportPolicy_builder{
				DefaultAction: configv1.ExportPolicy_UNEXPORT.Enum(),
				Rules: []*configv1.ExportRule{
					configv1.ExportRule_builder{
						NameRegex: proto.String("^allowed_.*"),
						Action:    configv1.ExportPolicy_EXPORT.Enum(),
					}.Build(),
				},
			}.Build(),
			want: true,
		},
		{
			name:     "Rule Match Unexport",
			toolName: "hidden_tool",
			policy: configv1.ExportPolicy_builder{
				DefaultAction: configv1.ExportPolicy_EXPORT.Enum(),
				Rules: []*configv1.ExportRule{
					configv1.ExportRule_builder{
						NameRegex: proto.String("^hidden_.*"),
						Action:    configv1.ExportPolicy_UNEXPORT.Enum(),
					}.Build(),
				},
			}.Build(),
			want: false,
		},
		{
			name:     "Rule No Match Fallthrough",
			toolName: "other_tool",
			policy: configv1.ExportPolicy_builder{
				DefaultAction: configv1.ExportPolicy_UNEXPORT.Enum(),
				Rules: []*configv1.ExportRule{
					configv1.ExportRule_builder{
						NameRegex: proto.String("^allowed_.*"),
						Action:    configv1.ExportPolicy_EXPORT.Enum(),
					}.Build(),
				},
			}.Build(),
			want: false,
		},
		{
			name:     "Invalid Regex",
			toolName: "any",
			policy: configv1.ExportPolicy_builder{
				Rules: []*configv1.ExportRule{
					configv1.ExportRule_builder{
						NameRegex: proto.String("["), // Invalid regex
						Action:    configv1.ExportPolicy_UNEXPORT.Enum(),
					}.Build(),
				},
			}.Build(),
			want: true, // Should continue and use default (true)
		},
		{
			name:     "Empty Regex",
			toolName: "any",
			policy: configv1.ExportPolicy_builder{
				Rules: []*configv1.ExportRule{
					configv1.ExportRule_builder{
						NameRegex: proto.String(""),
						Action:    configv1.ExportPolicy_UNEXPORT.Enum(),
					}.Build(),
				},
			}.Build(),
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
	t.Parallel()
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
				configv1.CallPolicy_builder{DefaultAction: configv1.CallPolicy_DENY.Enum()}.Build(),
			},
			toolName: "any",
			want:     false,
		},
		{
			name: "Default Allow",
			policies: []*configv1.CallPolicy{
				configv1.CallPolicy_builder{DefaultAction: configv1.CallPolicy_ALLOW.Enum()}.Build(),
			},
			toolName: "any",
			want:     true,
		},
		{
			name: "Name Regex Deny",
			policies: []*configv1.CallPolicy{
				configv1.CallPolicy_builder{
					DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
					Rules: []*configv1.CallPolicyRule{
						configv1.CallPolicyRule_builder{
							NameRegex: proto.String("^dangerous_.*"),
							Action:    configv1.CallPolicy_DENY.Enum(),
						}.Build(),
					},
				}.Build(),
			},
			toolName: "dangerous_tool",
			want:     false,
		},
		{
			name: "Argument Regex Deny",
			policies: []*configv1.CallPolicy{
				configv1.CallPolicy_builder{
					DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
					Rules: []*configv1.CallPolicyRule{
						configv1.CallPolicyRule_builder{
							ArgumentRegex: proto.String(`.*"secret".*`),
							Action:        configv1.CallPolicy_DENY.Enum(),
						}.Build(),
					},
				}.Build(),
			},
			toolName:  "any",
			arguments: json.RawMessage(`{"key": "secret"}`),
			want:      false,
		},
		{
			name: "Argument Regex Static Check (Nil Args)",
			policies: []*configv1.CallPolicy{
				configv1.CallPolicy_builder{
					DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
					Rules: []*configv1.CallPolicyRule{
						configv1.CallPolicyRule_builder{
							ArgumentRegex: proto.String(`.*`),
							Action:        configv1.CallPolicy_DENY.Enum(),
						}.Build(),
					},
				}.Build(),
			},
			toolName:  "any",
			arguments: nil,
			want:      true, // Rule with arg regex ignored when args are nil
		},
		{
			name: "CallID Regex Deny",
			policies: []*configv1.CallPolicy{
				configv1.CallPolicy_builder{
					DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
					Rules: []*configv1.CallPolicyRule{
						configv1.CallPolicyRule_builder{
							CallIdRegex: proto.String("call_123"),
							Action:      configv1.CallPolicy_DENY.Enum(),
						}.Build(),
					},
				}.Build(),
			},
			toolName: "any",
			callID:   "call_123",
			want:     false,
		},
		{
			name: "Multiple Policies (One blocks)",
			policies: []*configv1.CallPolicy{
				configv1.CallPolicy_builder{DefaultAction: configv1.CallPolicy_ALLOW.Enum()}.Build(),
				configv1.CallPolicy_builder{DefaultAction: configv1.CallPolicy_DENY.Enum()}.Build(),
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
	t.Parallel()
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
				configv1.CallPolicy_builder{DefaultAction: configv1.CallPolicy_DENY.Enum()}.Build(),
			},
			toolName: "any",
			want:     false,
		},
		{
			name: "Default Allow",
			policies: []*configv1.CallPolicy{
				configv1.CallPolicy_builder{DefaultAction: configv1.CallPolicy_ALLOW.Enum()}.Build(),
			},
			toolName: "any",
			want:     true,
		},
		{
			name: "Name Regex Deny",
			policies: []*configv1.CallPolicy{
				configv1.CallPolicy_builder{
					DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
					Rules: []*configv1.CallPolicyRule{
						configv1.CallPolicyRule_builder{
							NameRegex: proto.String("^dangerous_.*"),
							Action:    configv1.CallPolicy_DENY.Enum(),
						}.Build(),
					},
				}.Build(),
			},
			toolName: "dangerous_tool",
			want:     false,
		},
		{
			name: "Argument Regex Deny",
			policies: []*configv1.CallPolicy{
				configv1.CallPolicy_builder{
					DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
					Rules: []*configv1.CallPolicyRule{
						configv1.CallPolicyRule_builder{
							ArgumentRegex: proto.String(".*secret.*"),
							Action:        configv1.CallPolicy_DENY.Enum(),
						}.Build(),
					},
				}.Build(),
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
	t.Parallel()
	policies := []*configv1.CallPolicy{
		configv1.CallPolicy_builder{
			Rules: []*configv1.CallPolicyRule{
				configv1.CallPolicyRule_builder{NameRegex: proto.String("[")}.Build(),
			},
		}.Build(),
	}
	_, err := CompileCallPolicies(policies)
	assert.Error(t, err)
}
