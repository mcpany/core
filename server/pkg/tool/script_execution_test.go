package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
)

func TestShellInjection_ScriptExecution(t *testing.T) {
	// Case: Script execution (Vulnerable if not detected as shell command)
	t.Run("script_execution_should_be_protected", func(t *testing.T) {
		cmd := "./myscript.sh" // Not in the static list, but is a script
		tool := createTestScriptTool(cmd)
		req := &ExecutionRequest{
			ToolName: "test-script",
			ToolInputs: []byte(`{"input": "'; echo 'pwned'; '"}`),
		}

		_, err := tool.Execute(context.Background(), req)

		// If vulnerable, it will try to execute and fail with "executable not found"
		// (because myscript.sh doesn't exist).
		// If secure, it should fail with "shell injection detected".

		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "shell injection detected", "script execution should be protected")
		}
	})

	t.Run("bat_script_execution_should_be_protected", func(t *testing.T) {
		cmd := "deploy.bat"
		tool := createTestScriptTool(cmd)
		req := &ExecutionRequest{
			ToolName: "test-script",
			ToolInputs: []byte(`{"input": "& calc.exe"}`),
		}

		_, err := tool.Execute(context.Background(), req)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "shell injection detected", "bat script execution should be protected")
		}
	})
}

func createTestScriptTool(command string) Tool {
	toolDef := v1.Tool_builder{Name: stringPtr("test-tool")}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: &command,
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{input}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: stringPtr("input")}.Build(),
			}.Build(),
		},
	}.Build()
	return NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")
}

func stringPtr(s string) *string {
	return &s
}
