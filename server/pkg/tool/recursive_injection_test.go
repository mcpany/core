package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestRecursiveInjection(t *testing.T) {
	// This test demonstrates that recursive substitution allows a value to become a placeholder
	// which is then substituted in a subsequent iteration.

	// Create service definition
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Local: proto.Bool(true),
	}.Build()

	// Create call definition with two parameters 'a' and 'b'
	// and args that use both placeholders.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{a}}", "{{b}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("a"),
				}.Build(),
			}.Build(),
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("b"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolProto := v1.Tool_builder{
		Name: proto.String("recursive_test"),
	}.Build()

	tool := NewLocalCommandTool(
		toolProto,
		service,
		callDef,
		nil,
		"call_id",
	)

	// Inputs: a="{{b}}", b="INJECTED"
	inputs := map[string]interface{}{
		"a": "{{b}}",
		"b": "INJECTED",
	}
	inputsBytes, _ := json.Marshal(inputs)

	req := &ExecutionRequest{
		ToolName:   "recursive_test",
		ToolInputs: inputsBytes,
	}

	// We verify that the input containing "{{" is blocked immediately.
	_, err := tool.Execute(context.Background(), req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "recursive injection attempt detected")
}
