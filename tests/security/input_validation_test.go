package security_test

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func newCommandTool(command string, callDef *configv1.CommandLineCallDefinition) tool.Tool {
	if callDef == nil {
		callDef = configv1.CommandLineCallDefinition_builder{}.Build()
	}
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String(command),
	}).Build()

	properties := make(map[string]*structpb.Value)
	for _, param := range callDef.GetParameters() {
		properties[param.GetSchema().GetName()] = structpb.NewStructValue(&structpb.Struct{})
	}

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: properties,
			}),
		},
	}

	return tool.NewCommandTool(
		v1.Tool_builder{InputSchema: inputSchema}.Build(),
		service,
		callDef,
		nil,
		"call-id",
	)
}

func TestCommandTool_InputValidation(t *testing.T) {
	t.Parallel()

	t.Run("fails when pattern does not match", func(t *testing.T) {
		t.Parallel()
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{
						Name:    proto.String("safe_arg"),
						Type:    configv1.ParameterType_STRING.Enum(),
						Pattern: proto.String("^[a-z0-9]+$"),
					}.Build(),
				}.Build(),
			},
		}.Build()
		cmdTool := newCommandTool("echo", callDef)

		// Malicious input
		inputData := map[string]interface{}{"safe_arg": "invalid-chars!"}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err = cmdTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `does not match pattern`)
	})

	t.Run("succeeds when pattern matches", func(t *testing.T) {
		t.Parallel()
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{
						Name:    proto.String("safe_arg"),
						Type:    configv1.ParameterType_STRING.Enum(),
						Pattern: proto.String("^[a-z0-9]+$"),
					}.Build(),
				}.Build(),
			},
		}.Build()
		cmdTool := newCommandTool("echo", callDef)

		// Safe input
		inputData := map[string]interface{}{"safe_arg": "safe123"}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err = cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)
	})

	t.Run("fails when numeric constraint violated", func(t *testing.T) {
		t.Parallel()
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{
						Name:    proto.String("count"),
						Type:    configv1.ParameterType_INTEGER.Enum(),
						Maximum: proto.Float64(10),
					}.Build(),
				}.Build(),
			},
		}.Build()
		cmdTool := newCommandTool("echo", callDef)

		// Invalid input
		inputData := map[string]interface{}{"count": 11}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err = cmdTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `exceeds maximum`)
	})
}
