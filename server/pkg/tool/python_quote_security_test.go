package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestPythonDoubleQuoteInjection(t *testing.T) {
	// This test reproduces an RCE vulnerability where Python code can be injected
	// into a double-quoted argument string.

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("python3"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "print('Hello, {{name}}!')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("name"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := pb.Tool_builder{
		Name: proto.String("python_hello"),
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Payload using newlines:
	// name = "user')\nimport os\nos.system('echo RCE_SUCCESS')\n#"
	// Result:
	// print('Hello, user')
	// import os
	// os.system('echo RCE_SUCCESS')
	// #!')

	req := &ExecutionRequest{
		ToolName: "python_hello",
		ToolInputs: []byte(`{"name": "user')\nimport os\nos.system('echo RCE_SUCCESS')\n#"}`),
	}

	result, err := tool.Execute(context.Background(), req)

	// If the vulnerability exists, err should be nil (execution successful)
	// and the output should contain "RCE_SUCCESS"
	if err != nil {
		t.Logf("Execution blocked (good): %v", err)
		// If it was blocked by our security check, it should be a specific error
		assert.Contains(t, err.Error(), "injection detected")
	} else {
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		stdout, ok := resMap["stdout"].(string)
		require.True(t, ok)

		t.Logf("Stdout: %s", stdout)

		if assert.Contains(t, stdout, "RCE_SUCCESS") {
			t.Fatal("VULNERABILITY CONFIRMED: Python code injection successful")
		}
	}
}
