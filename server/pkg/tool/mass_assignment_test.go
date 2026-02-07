// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_MassAssignment(t *testing.T) {
	// t.Parallel() // Can cause issues if other tests modify global state, but should be fine here.

	tool := v1.Tool_builder{
		Name: proto.String("test-tool-mass-assignment"),
	}.Build()

	// Service defines a command with a placeholder {{secret_flag}}
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"--public={{public}}", "--secret={{secret_flag}}"},
		// Only "public" is exposed in the schema
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("public")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Attacker sends "secret_flag" which is NOT in schema
	req := &ExecutionRequest{
		ToolName: "test-tool-mass-assignment",
		Arguments: map[string]interface{}{
			"public":      "safe",
			"secret_flag": "pwned",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	// If vulnerable, the output will contain "--secret=pwned".
	// Since we fixed it, the input "secret_flag" is filtered out.
	// The placeholder {{secret_flag}} remains in the argument, so the output should contain "--secret={{secret_flag}}".
	assert.NotContains(t, resultMap["stdout"], "pwned")
	assert.Contains(t, resultMap["stdout"], "--secret={{secret_flag}}")
}
