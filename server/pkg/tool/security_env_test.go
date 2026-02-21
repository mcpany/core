// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestDangerousEnvVars(t *testing.T) {
	// Setup a tool definition with parameters mapped to environment variables
	tool := v1.Tool_builder{
		Name: proto.String("test-dangerous-env"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("env"), // Use 'env' to print environment
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			// LD_PRELOAD is currently blocked
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("LD_PRELOAD")}.Build()}.Build(),
			// GCONV_PATH (Should be blocked now)
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("GCONV_PATH")}.Build()}.Build(),
			// BASH_FUNC_x%% (Should be blocked now)
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("BASH_FUNC_x%%")}.Build()}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// 1. Test LD_PRELOAD (Should be blocked)
	args := map[string]interface{}{
		"LD_PRELOAD": "malicious.so",
	}
	req := &ExecutionRequest{
		ToolName: "test-dangerous-env",
		Arguments: args,
	}
	inputsJSON, _ := json.Marshal(req.Arguments)
	req.ToolInputs = inputsJSON

	result, err := localTool.Execute(context.Background(), req)
	require.NoError(t, err)

	resultMap := result.(map[string]interface{})
	stdout := resultMap["stdout"].(string)

	assert.NotContains(t, stdout, "LD_PRELOAD", "LD_PRELOAD should be blocked and not present in env output")

	// 2. Test GCONV_PATH (Should be blocked)
	args2 := map[string]interface{}{
		"GCONV_PATH": "malicious",
	}
	req2 := &ExecutionRequest{
		ToolName: "test-dangerous-env",
		Arguments: args2,
	}
	inputsJSON2, _ := json.Marshal(req2.Arguments)
	req2.ToolInputs = inputsJSON2

	result2, err := localTool.Execute(context.Background(), req2)
	require.NoError(t, err)
	resultMap2 := result2.(map[string]interface{})
	stdout2 := resultMap2["stdout"].(string)

	assert.NotContains(t, stdout2, "GCONV_PATH", "GCONV_PATH should be blocked")

	// 3. Test BASH_FUNC_ (Should be blocked)
	args3 := map[string]interface{}{
		"BASH_FUNC_x%%": "() { :; }; id",
	}
	req3 := &ExecutionRequest{
		ToolName: "test-dangerous-env",
		Arguments: args3,
	}
	inputsJSON3, _ := json.Marshal(req3.Arguments)
	req3.ToolInputs = inputsJSON3

	result3, err := localTool.Execute(context.Background(), req3)
	require.NoError(t, err)
	resultMap3 := result3.(map[string]interface{})
	stdout3 := resultMap3["stdout"].(string)

	assert.NotContains(t, stdout3, "BASH_FUNC_x%%", "BASH_FUNC_x%% should be blocked")
}

// Helper for strings.Contains check if needed, but assert.NotContains is better.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
