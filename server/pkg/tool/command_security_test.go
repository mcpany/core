package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// TestShellWrapperBypass ensures that wrapper commands like "nice" are treated as shells
// and subject to strict injection checks.
func TestShellWrapperBypass(t *testing.T) {
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

	// Payload that contains shell metacharacters which SHOULD be blocked
	payload := "echo hacked; echo pwned"

	// Execute
	req := &ExecutionRequest{
		ToolName: "nice_bypass",
		ToolInputs: []byte(`{"cmd": "` + payload + `"}`),
	}

	// This should return an error if security checks are working.
	_, err := cmdTool.Execute(context.Background(), req)

	assert.Error(t, err, "Expected error for nice wrapper command injection")
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected", "nice command should be treated as shell and block injection")
	}
}

// TestGnuplotInjection ensures that gnuplot is treated as an interpreter/shell
// and subject to strict injection checks.
func TestGnuplotInjection(t *testing.T) {
	// Setup a LocalCommandTool that uses "gnuplot"
	toolDef := v1.Tool_builder{
		Name: proto.String("gnuplot_injection"),
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

	// Payload that attempts to execute a command
	payload := "system('ls')"

	// Execute
	req := &ExecutionRequest{
		ToolName: "gnuplot_injection",
		ToolInputs: []byte(`{"script": "` + payload + `"}`),
	}

	// This should return an error if security checks are working.
	_, err := cmdTool.Execute(context.Background(), req)

	assert.Error(t, err, "Expected error for gnuplot injection")
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected", "gnuplot command should be treated as shell/interpreter and block injection")
	}
}
