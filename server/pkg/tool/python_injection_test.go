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

func TestLocalCommandTool_Execute_PythonInjection(t *testing.T) {
	// This test demonstrates that python is treated as a shell command,
	// preventing code injection via argument substitution.

	toolDef := &v1.Tool{
		Name: proto.String("python_tool"),
	}

	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("python3"),
	}

	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-c", "print('{{msg}}')"},
		Parameters: []*configv1.CommandLineParameterMapping{
			{
				Schema: &configv1.ParameterSchema{
					Name: proto.String("msg"),
				},
			},
		},
	}

	ct := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call-id")

	// Malicious input trying to break out of python string
	// msg = '); print("INJECTED"); print('
	// Resulting code: print(''); print("INJECTED"); print('')

	injectionPayload := "'); print(\"INJECTED\"); print('"
	jsonInput, _ := json.Marshal(map[string]string{"msg": injectionPayload})

	req := &ExecutionRequest{
		ToolName: "python_tool",
		ToolInputs: jsonInput,
	}

	// Execute
	_, err := ct.Execute(context.Background(), req)

	// We expect the execution to fail because "shell injection detected"
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected")
	}
}
