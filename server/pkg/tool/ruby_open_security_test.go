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

func TestRubyOpenInjection(t *testing.T) {
	// This test reproduces a Command Injection vulnerability where Ruby's open()
	// function executes commands if the argument starts with '|'.
	// This relies on the fact that pipe character '|' is NOT blocked in double-quoted strings.

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ruby"),
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-e", "print open(\"{{url}}\").read"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("url"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := pb.Tool_builder{
		Name: proto.String("ruby_open"),
	}.Build()

	tool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	// Payload starting with pipe
	payload := "|echo RCE_SUCCESS"

	req := &ExecutionRequest{
		ToolName: "ruby_open",
		ToolInputs: []byte(`{"url": "` + payload + `"}`),
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

		t.Logf("Stdout: %s", stdout)

		if assert.Contains(t, stdout, "RCE_SUCCESS") {
			t.Fatal("VULNERABILITY CONFIRMED: Ruby open() injection successful")
		}
	}
}
