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

func TestLocalCommandTool_SSHInjection_Prevention(t *testing.T) {
	t.Parallel()
	// This test verifies that arguments to 'ssh' command are checked for shell injection.

	tool := v1.Tool_builder{
		Name:        proto.String("test-tool-ssh"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ssh"), // Now considered a shell command
		Local:   proto.Bool(true),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("cmd")}.Build()}.Build(),
		},
		Args: []string{"user@host", "echo {{cmd}}"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Case 1: Safe input
	reqSafe := &ExecutionRequest{
		ToolName: "test-tool-ssh",
		Arguments: map[string]interface{}{
			"cmd": "hello",
		},
	}
	reqSafe.ToolInputs, _ = json.Marshal(reqSafe.Arguments)

	_, err := localTool.Execute(context.Background(), reqSafe)
	// Execute will fail because ssh is not installed or network fails,
	// but we only care if it failed due to injection check.
	// We expect NO injection error.
	if err != nil {
		assert.NotContains(t, err.Error(), "shell injection detected")
	}

	// Case 2: Injection Attack
	reqAttack := &ExecutionRequest{
		ToolName: "test-tool-ssh",
		Arguments: map[string]interface{}{
			"cmd": "hello; rm -rf /",
		},
	}
	reqAttack.ToolInputs, _ = json.Marshal(reqAttack.Arguments)

	_, err = localTool.Execute(context.Background(), reqAttack)

	// Should fail with injection error
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected")
	}
}
