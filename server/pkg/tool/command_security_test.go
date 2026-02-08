package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
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

	// This should return an error as 'nice' is now considered a shell command
	// and the payload contains blocked characters (semicolon).
	_, err := cmdTool.Execute(context.Background(), req)

    assert.Error(t, err, "Expected error due to shell injection detection")
    if err != nil {
        assert.Contains(t, err.Error(), "shell injection detected", "Error message should mention shell injection")
    }
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

	// Payload that contains dangerous function calls
	payload := "system('ls')"

	// Execute
	req := &ExecutionRequest{
		ToolName: "gnuplot_bypass",
		ToolInputs: []byte(`{"script": "` + payload + `"}`),
	}

	// This should return an error as 'gnuplot' is now considered a shell command
	// and the payload contains blocked characters or function calls.
	_, err := cmdTool.Execute(context.Background(), req)

    assert.Error(t, err, "Expected error due to shell injection detection")
    if err != nil {
        // gnuplot is not handled by isInterpreter, so it falls back to checkUnquotedInjection
        // which blocks '(', ')', ';', etc.
        assert.Contains(t, err.Error(), "shell injection detected", "Error message should mention shell injection")
    }
}
