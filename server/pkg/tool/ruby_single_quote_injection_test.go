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

func TestRubySingleQuoteInjection(t *testing.T) {
	// This test attempts to reproduce an RCE vulnerability where Ruby open("|cmd")
    // can be injected into a SINGLE-quoted argument string.

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
	}.Build()

	// Ruby script: print open('{{input}}').read
    // Note: The template uses single quotes around the placeholder.
    // The shell sees: ruby -e "print open('{{input}}').read" (if passed as one arg)
    // Or if args are split: "-e", "print open('{{input}}').read"
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "print open('{{input}}').read"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("input"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := pb.Tool_builder{
		Name: proto.String("ruby_open_single"),
	}.Build()

	// Use LocalCommandTool to execute locally
	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Payload: |echo RCE_SUCCESS
	payload := "|echo RCE_SUCCESS"

	req := &ExecutionRequest{
		ToolName: "ruby_open_single",
		ToolInputs: []byte(`{"input": "` + payload + `"}`),
	}

	result, err := tool.Execute(context.Background(), req)

	if err != nil {
		t.Logf("Execution blocked (good): %v", err)
		assert.Contains(t, err.Error(), "injection detected")
	} else {
		resMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		stdout, ok := resMap["stdout"].(string)
		require.True(t, ok)

        // If execution succeeded, it means the command ran.
        // We check if "RCE_SUCCESS" is in stdout.

		t.Logf("Stdout: %s", stdout)

		if assert.Contains(t, stdout, "RCE_SUCCESS") {
			t.Fatal("VULNERABILITY CONFIRMED: Ruby open injection successful in single quotes")
		}
	}
}
