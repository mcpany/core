package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestNiceBypass(t *testing.T) {
	// Setup a LocalCommandTool that uses "nice"
	toolDef := v1.Tool_builder{
		Name: proto.String("nice_bypass"),
	}.Build()

	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("nice"),
	}).Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"sh", "-c", "{{cmd}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("cmd"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	// Initialize the tool
	cmdTool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	// Payload that contains shell metacharacters which SHOULD be blocked if checks were applied
	payload := "echo hacked; echo pwned"

	// Execute
	req := &ExecutionRequest{
		ToolName: "nice_bypass",
		ToolInputs: []byte(`{"cmd": "` + payload + `"}`),
	}

	// This should return an error if security checks are working.
	_, err := cmdTool.Execute(context.Background(), req)

	require.Error(t, err, "Expected error due to shell injection")
	assert.Contains(t, err.Error(), "shell injection detected", "Error should indicate shell injection detection")
}

func TestGnuplotBypass(t *testing.T) {
	// Setup a LocalCommandTool that uses "gnuplot"
	toolDef := v1.Tool_builder{
		Name: proto.String("gnuplot_bypass"),
	}.Build()

	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("gnuplot"),
	}).Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("script"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	// Initialize the tool
	cmdTool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	// Payload: system call
	// gnuplot uses system("cmd")
	// We need to escape quotes for JSON
	payload := `system(\"id\")`

	// Execute
	req := &ExecutionRequest{
		ToolName: "gnuplot_bypass",
		ToolInputs: []byte(`{"script": "` + payload + `"}`),
	}

	// This should return an error if security checks are working.
	// Since gnuplot is now in isShellCommand list, it should check for shell/interpreter injection.
	// checkUnquotedInjection blocks '(', ')', '"'.

	_, err := cmdTool.Execute(context.Background(), req)

	require.Error(t, err, "Expected error due to shell/interpreter injection")
	assert.Contains(t, err.Error(), "shell injection detected", "Error should indicate injection detection")
}
