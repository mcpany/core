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

func TestPythonAssertAllowed(t *testing.T) {
	// Setup: Define a tool that executes python code via -c
	toolDef := v1.Tool_builder{
		Name: proto.String("python-eval"),
	}.Build()

	serviceConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "{{code}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("code"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	// Create the tool
	tool := NewLocalCommandTool(toolDef, serviceConfig, callDef, nil, "test-call-id")

	// Payload: Try to execute assert which should be allowed in Python
	payload := "assert 1 == 1"

	inputs := map[string]interface{}{
		"code": payload,
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "python-eval",
		ToolInputs: inputBytes,
		DryRun:     true, // DryRun runs validation but skips execution
	}

	// Execute
	_, err := tool.Execute(context.Background(), req)

	// Assert: No error expected
	assert.NoError(t, err, "Expected no error because assert should be allowed in Python")
}
