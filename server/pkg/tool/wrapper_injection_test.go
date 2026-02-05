package tool

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestWrapperInjection(t *testing.T) {
	// Setup a tool definition using 'timeout' which executes another command.
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("timeout"),
	}.Build()

    // Parameter mapping
    schema := configv1.ParameterSchema_builder{Name: proto.String("input")}.Build()
    mapping := configv1.CommandLineParameterMapping_builder{
        Schema: schema,
    }.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{
			"1s",
			"bash",
			"-c",
			"echo \"{{input}}\"",
		},
        Parameters: []*configv1.CommandLineParameterMapping{mapping},
	}.Build()

    toolProto := v1.Tool_builder{
        Name: proto.String("vulnerable_tool"),
    }.Build()

	tool := NewLocalCommandTool(
		toolProto,
		service,
		callDef,
		nil,
		"test-call-id",
	)

	// Malicious input trying to inject shell command via subshell
	// We use $(...) which is blocked by checkForShellInjection (blocks $) if validation runs.
	input := "$(echo pwned)"

	req := &ExecutionRequest{
		ToolName: "vulnerable_tool",
		ToolInputs: []byte(`{"input": "` + input + `"}`),
	}

	_, err := tool.Execute(context.Background(), req)

	// We expect a security error now.

    if err == nil {
        t.Fatal("Vulnerability confirmed: No error returned, validation skipped. (FIX FAILED)")
    } else {
        if strings.Contains(err.Error(), "shell injection detected") {
             t.Logf("Security check working: %v", err)
        } else {
             t.Fatalf("Unexpected error: %v. Expected shell injection detection.", err)
        }
    }
}
