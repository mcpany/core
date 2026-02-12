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

func TestLocalCommandTool_PythonInjection_GetAttr_Security(t *testing.T) {
	// This test asserts that Python execution BLOCKS bypassing restrictions
	// using getattr() to access dangerous functions like os.system.
	// This test specifically uses obfuscation (string concatenation) to avoid
	// triggering the "system" keyword check, relying on "getattr" blocking.

	tool := v1.Tool_builder{
		Name: proto.String("python-getattr-test"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `python3 -c "print('{{input}}')"`
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('{{input}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("input")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: '); import os; s="sy"+"s"+"tem"; getattr(os, s)("echo PWNED"); print('
	// This payload constructs "system" dynamically to bypass simple keyword checks.
	payload := "'); import os; s=\"sy\"+\"s\"+\"tem\"; getattr(os, s)(\"echo PWNED\"); print('"

	req := &ExecutionRequest{
		ToolName: "python-getattr-test",
		Arguments: map[string]interface{}{
			"input": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)

	// Expect failure due to injection detection (specifically getattr)
	if err == nil {
		resultMap, ok := result.(map[string]interface{})
		output := ""
		if ok {
			output = resultMap["combined_output"].(string)
		}
		t.Errorf("Security check failed: Expected error but got success. Output: %s", output)
	} else {
		// We expect "getattr" to be flagged, or just "interpreter injection detected"
		assert.Contains(t, err.Error(), "interpreter injection detected", "Error should mention injection detection")
	}
}
