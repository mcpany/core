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
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_Execute_CommandInjection_Vulnerability(t *testing.T) {
	// Setup: Tool definition allows "args" in InputSchema, but CallDefinition does NOT have it as a parameter.
	// Current behavior: Checks InputSchema, so it allows "args".
	// Desired behavior: Should check CallDefinition, so it should REJECT "args".

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"args": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	toolProto := &v1.Tool{
		Name:        proto.String("vuln-test-tool"),
		InputSchema: inputSchema,
	}
	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}
	// Empty parameters in CallDefinition - "args" is NOT explicitly defined
	callDef := &configv1.CommandLineCallDefinition{
		Parameters: []*configv1.CommandLineParameterMapping{},
	}

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "vuln-test-tool",
		Arguments: map[string]interface{}{
			"args": []interface{}{"INJECTED"},
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute
	result, err := localTool.Execute(context.Background(), req)

	// In the vulnerable state, this succeeds because InputSchema has "args".
	// The fix should fail this test if we assert Unexplained Error, OR we can assert what we want:
	// We want an error.

	// If the vulnerability exists, err is nil.
	// We assert Error to demonstrate the requirement.
	// This test will FAIL initially.
	assert.Error(t, err, "Should reject 'args' when not in CallDefinition parameters")
	if err != nil {
		assert.Contains(t, err.Error(), "parameter is not allowed", "Error message should indicate forbidden parameter")
	} else {
		// Verify injection happened (stdout contains INJECTED)
		resultMap, ok := result.(map[string]interface{})
		if ok {
			t.Logf("Vulnerability reproducible: Command executed with injected args: %v", resultMap["stdout"])
		}
	}
}
