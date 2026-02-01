// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestEvaluateCompiledCallPolicy_Bypass(t *testing.T) {
	t.Parallel()

	// Policy: Deny if arguments contain "dangerous" OR ">"
	policies := []*configv1.CallPolicy{
		configv1.CallPolicy_builder{
			DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
			Rules: []*configv1.CallPolicyRule{
				configv1.CallPolicyRule_builder{
					ArgumentRegex: proto.String(`.*"dangerous".*`),
					Action:        configv1.CallPolicy_DENY.Enum(),
				}.Build(),
				configv1.CallPolicyRule_builder{
					ArgumentRegex: proto.String(`.*>.*`),
					Action:        configv1.CallPolicy_DENY.Enum(),
				}.Build(),
			},
		}.Build(),
	}

	compiled, err := CompileCallPolicies(policies)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		arguments []byte
		want      bool // true = allowed, false = denied
	}{
		{
			name:      "Blocked Standard",
			arguments: json.RawMessage(`{"key": "dangerous"}`),
			want:      false,
		},
		{
			name:      "Bypass Unicode Escape",
			// "dangerous" -> "\u0064angerous"
			// The regex looks for literal "dangerous", so it won't match "\u0064angerous"
			// But canonicalization should fix this, so it matches "dangerous".
			arguments: json.RawMessage(`{"key": "\u0064angerous"}`),
			want:      false, // Should be blocked!
		},
		{
			name:      "Blocked Symbol >",
			arguments: json.RawMessage(`{"cmd": "ls > file"}`),
			want:      false, // Should be blocked!
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluateCompiledCallPolicy(compiled, "any", "", tt.arguments)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got, "Policy check result mismatch")
		})
	}
}
