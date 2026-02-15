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

func TestLocalCommandTool_Execute_EnvInjection(t *testing.T) {
	// This test demonstrates that environment variables are not sanitized for shell injection,
	// allowing arbitrary code execution when the command is a shell script that expands the variable.

	tool := v1.Tool_builder{
		Name: proto.String("test-tool-env-injection"),
	}.Build()

	// Service configured to run a shell command that echoes an environment variable.
	// This simulates a common pattern where a script uses an environment variable.
	// e.g. bash -c "echo $USER_INPUT"
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("bash"),
		Local:   proto.Bool(true),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo $USER_INPUT"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("USER_INPUT"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload: "hello; echo INJECTED"
	// Expected behavior (Vulnerable): The shell expands $USER_INPUT to "hello; echo INJECTED", executing "echo INJECTED".
	// Expected behavior (Secure): The input is blocked or sanitized.

	req := &ExecutionRequest{
		ToolName: "test-tool-env-injection",
		Arguments: map[string]interface{}{
			"USER_INPUT": "hello; echo INJECTED",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)

	// Secure behavior: The injection attempt should be detected, and the environment variable
	// should be omitted. The execution should succeed (because the input is valid otherwise),
	// but the output should NOT contain the injected payload.
	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	stdout := resultMap["stdout"].(string)

	assert.NotContains(t, stdout, "INJECTED", "Shell injection payload executed!")
}
