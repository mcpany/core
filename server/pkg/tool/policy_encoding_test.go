// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestPolicyEncodingBypass(t *testing.T) {
	// Define a policy that denies "badword" in arguments
	policy := configv1.CallPolicy_builder{
		DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
		Rules: []*configv1.CallPolicyRule{
			configv1.CallPolicyRule_builder{
				ArgumentRegex: proto.String(".*badword.*"),
				Action:        configv1.CallPolicy_DENY.Enum(),
			}.Build(),
		},
	}.Build()

	compiled, err := CompileCallPolicies([]*configv1.CallPolicy{policy})
	require.NoError(t, err)

	tests := []struct {
		name        string
		input       string
		shouldAllow bool
	}{
		{
			name:        "Normal badword",
			input:       `{"arg": "badword"}`,
			shouldAllow: false, // Blocked
		},
		{
			name:        "Encoded badword (Fixed)",
			input:       `{"arg": "b\u0061dword"}`, // \u0061 is 'a'
			shouldAllow: false,                     // Now BLOCKED (Vulnerability Fixed)
		},
		{
			name:        "Safe word",
			input:       `{"arg": "goodword"}`,
			shouldAllow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := EvaluateCompiledCallPolicy(compiled, "tool", "call1", []byte(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.shouldAllow, allowed, "Policy evaluation mismatch for input: %s", tt.input)
		})
	}
}
