package tool_test

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestCommandTool_ExtraCoverage(t *testing.T) {
	t.Parallel()

	// 1. Path traversal in parameters
	t.Run("path traversal in parameters", func(t *testing.T) {
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("path")}.Build(),
				}.Build(),
			},
		}.Build()
		cmdTool := newCommandTool("ls", callDef)
		inputData := map[string]interface{}{"path": "../secret"}
		inputs, _ := json.Marshal(inputData)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := cmdTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal attempt detected")
	})

	// 2. Absolute path in parameters (local execution)
	t.Run("absolute path in parameters", func(t *testing.T) {
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("path")}.Build(),
				}.Build(),
			},
		}.Build()
		cmdTool := newCommandTool("ls", callDef)
		inputData := map[string]interface{}{"path": "/etc/passwd"}
		inputs, _ := json.Marshal(inputData)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := cmdTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "absolute path detected")
	})

	// 3. Argument injection in parameters
	t.Run("argument injection in parameters", func(t *testing.T) {
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"{{arg}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("arg")}.Build(),
				}.Build(),
			},
		}.Build()
		cmdTool := newCommandTool("echo", callDef)
		inputData := map[string]interface{}{"arg": "-rf"}
		inputs, _ := json.Marshal(inputData)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := cmdTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "argument injection detected")
	})

	// 4. 'args' parameter not in schema
	t.Run("args parameter not in schema", func(t *testing.T) {
		// Define tool WITHOUT 'args' in schema
		// newCommandTool creates schema based on params.
		callDef := configv1.CommandLineCallDefinition_builder{}.Build()
		cmdTool := newCommandTool("echo", callDef)

		inputData := map[string]interface{}{"args": []string{"hello"}}
		inputs, _ := json.Marshal(inputData)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := cmdTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "'args' parameter is not allowed")
	})

	// 5. 'args' parameter invalid type (not array of strings)
	t.Run("args parameter invalid type", func(t *testing.T) {
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build(),
				}.Build(),
			},
		}.Build()
		cmdTool := newCommandTool("echo", callDef)
		inputData := map[string]interface{}{"args": []interface{}{123}}
		inputs, _ := json.Marshal(inputData)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := cmdTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "non-string value in 'args' array")
	})

	// 6. Path traversal in 'args'
	t.Run("path traversal in args", func(t *testing.T) {
		callDef := configv1.CommandLineCallDefinition_builder{
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build(),
				}.Build(),
			},
		}.Build()
		cmdTool := newCommandTool("ls", callDef)
		inputData := map[string]interface{}{"args": []string{"../secret"}}
		inputs, _ := json.Marshal(inputData)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err := cmdTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal attempt detected")
	})
}
