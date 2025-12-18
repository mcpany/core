// Copyright 2025 Author(s) of MCP Any
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

func TestLocalCommandTool_EnvironmentInjection(t *testing.T) {
	// Setup a tool that runs 'env' to print environment variables
	tool := &v1.Tool{
		Name:        proto.String("env-tool"),
		Description: proto.String("Prints environment"),
	}
	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("env"),
		Local:   proto.Bool(true),
	}

	// Define explicitly allowed parameters
	callDef := &configv1.CommandLineCallDefinition{
		Parameters: []*configv1.CommandLineParameterMapping{
			{
				Schema: &configv1.ParameterSchema{
					Name: proto.String("VALID_VAR"),
				},
			},
		},
	}

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "")

	// Inject a malicious environment variable and a valid one
	maliciousEnvVar := "MALICIOUS_VAR"
	maliciousValue := "hacked"

	req := &ExecutionRequest{
		ToolName: "env-tool",
		Arguments: map[string]interface{}{
			maliciousEnvVar: maliciousValue,
			"VALID_VAR": "valid_value",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	stdout, ok := resultMap["stdout"].(string)
	assert.True(t, ok)

	// Check if the malicious variable is NOT present in the output
	assert.NotContains(t, stdout, maliciousEnvVar+"="+maliciousValue, "Malicious environment variable should NOT be injected")

	// Check if the valid variable IS present
	assert.Contains(t, stdout, "VALID_VAR=valid_value", "Valid parameter should be passed as environment variable")
}
