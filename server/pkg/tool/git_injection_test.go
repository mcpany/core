package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestGitInjection(t *testing.T) {
	// Define a tool that uses 'git'
	tool := v1.Tool_builder{
		Name: proto.String("git-tool"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("git"),
		Local:   proto.Bool(true),
	}.Build()

	param := configv1.CommandLineParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("arg"),
		}.Build(),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args:       []string{"{{arg}}"},
		Parameters: []*configv1.CommandLineParameterMapping{param},
	}.Build()

    policies := []*configv1.CallPolicy{}

	localTool := NewLocalCommandTool(tool, service, callDef, policies, "call-id")

	// Try to inject a flag
	req := &ExecutionRequest{
		ToolName: "git-tool",
		ToolInputs: []byte(`{"arg": "-c"}`),
	}

	result, err := localTool.Execute(context.Background(), req)

	// Expect error due to argument injection
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "argument injection detected")

    // Try valid input
    reqValid := &ExecutionRequest{
		ToolName: "git-tool",
		ToolInputs: []byte(`{"arg": "status"}`),
	}
    // "git status" might fail if not in a repo, but we just check if it executed or validation failed.
    // If validation fails, err is "parameter 'arg': argument injection detected..."
    // If execution fails, err is "failed to execute command..."

    res, err := localTool.Execute(context.Background(), reqValid)
    // We expect execution to proceed (even if git fails with status code)
    // Actually Execute returns result map with return_code.
    // So err should be nil (execution successful, even if exit code is non-zero)
    // Wait, Execute returns error if command start fails or other errors.
    // If exit code != 0, it returns result with "status": "error".

    assert.NoError(t, err)
    if resMap, ok := res.(map[string]interface{}); ok {
        // Assert it ran git
        assert.Equal(t, "git", resMap["command"])
    }
}
