// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestCompiledCallPolicy_CEL(t *testing.T) {
	tests := []struct {
		name          string
		policyJSON    string
		toolName      string
		callID        string
		arguments     map[string]interface{}
		expectedAllow bool
	}{
		{
			name: "CEL Match Allow",
			policyJSON: `{
				"default_action": "DENY",
				"rules": [
					{
						"action": "ALLOW",
						"cel": "tool_name == 'allowed_tool'"
					}
				]
			}`,
			toolName:      "allowed_tool",
			expectedAllow: true,
		},
		{
			name: "CEL No Match Deny",
			policyJSON: `{
				"default_action": "DENY",
				"rules": [
					{
						"action": "ALLOW",
						"cel": "tool_name == 'allowed_tool'"
					}
				]
			}`,
			toolName:      "other_tool",
			expectedAllow: false,
		},
		{
			name: "CEL Argument Match",
			policyJSON: `{
				"default_action": "DENY",
				"rules": [
					{
						"action": "ALLOW",
						"cel": "arguments.safe == true"
					}
				]
			}`,
			toolName:      "any_tool",
			arguments:     map[string]interface{}{"safe": true},
			expectedAllow: true,
		},
		{
			name: "CEL Argument Mismatch",
			policyJSON: `{
				"default_action": "DENY",
				"rules": [
					{
						"action": "ALLOW",
						"cel": "arguments.safe == true"
					}
				]
			}`,
			toolName:      "any_tool",
			arguments:     map[string]interface{}{"safe": false},
			expectedAllow: false,
		},
		{
			name: "CEL Complex Logic",
			policyJSON: `{
				"default_action": "DENY",
				"rules": [
					{
						"action": "ALLOW",
						"cel": "tool_name.startsWith('read_') && arguments.path.startsWith('/tmp')"
					}
				]
			}`,
			toolName:      "read_file",
			arguments:     map[string]interface{}{"path": "/tmp/test.txt"},
			expectedAllow: true,
		},
		{
			name: "CEL Complex Logic Fail",
			policyJSON: `{
				"default_action": "DENY",
				"rules": [
					{
						"action": "ALLOW",
						"cel": "tool_name.startsWith('read_') && arguments.path.startsWith('/tmp')"
					}
				]
			}`,
			toolName:      "read_file",
			arguments:     map[string]interface{}{"path": "/etc/passwd"},
			expectedAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &configv1.CallPolicy{}
			err := protojson.Unmarshal([]byte(tt.policyJSON), policy)
			require.NoError(t, err, "Failed to unmarshal policy JSON")

			compiled, err := CompileCallPolicies([]*configv1.CallPolicy{policy})
			require.NoError(t, err)

			argsBytes, err := json.Marshal(tt.arguments)
			require.NoError(t, err)

			allowed, err := EvaluateCompiledCallPolicy(compiled, tt.toolName, tt.callID, argsBytes)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedAllow, allowed)
		})
	}
}

func TestCompiledCallPolicy_InvalidCEL(t *testing.T) {
	policyJSON := `{
		"rules": [
			{
				"cel": "invalid syntax ???"
			}
		]
	}`
	policy := &configv1.CallPolicy{}
	err := protojson.Unmarshal([]byte(policyJSON), policy)
	require.NoError(t, err)

	_, err = CompileCallPolicies([]*configv1.CallPolicy{policy})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CEL expression")
}
