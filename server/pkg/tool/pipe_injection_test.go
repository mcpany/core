package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_PipeInjection_ShAwk(t *testing.T) {
	// t.Parallel()

	// Define a tool that uses 'sh'.
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"script": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}
	tool := v1.Tool_builder{
		Name:        proto.String("sh-awk-tool"),
		InputSchema: inputSchema,
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
	}.Build()

	// Configured to run `sh -c "awk '{{script}}' /dev/null"`
	// Note usage of double quotes for sh argument, and single quotes for awk script.
	// This puts {{script}} in Single Quoted context for MCP Any.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "awk '{{script}}' /dev/null"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	// Payload uses awk pipe to execute command
	// BEGIN { print "pwned" | "cat" }
	payload := "BEGIN { print \"pwned\" | \"cat\" }"

	req := &ExecutionRequest{
		ToolName: "sh-awk-tool",
		Arguments: map[string]interface{}{
			"script": payload,
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// After hardening, this should fail validation because 'sh' is an interpreter and single quoted arg contains '|'.
	result, err := localTool.Execute(context.Background(), req)

	// Assert that we get an error about shell injection
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shell injection detected")
	assert.Nil(t, result)
}
