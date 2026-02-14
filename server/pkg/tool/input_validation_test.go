// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

func TestCommandTool_InputValidation(t *testing.T) {
	t.Parallel()

	// Test Case 1: Regex Pattern Validation
	t.Run("regex pattern validation failure", func(t *testing.T) {
		t.Parallel()
		schema := configv1.ParameterSchema_builder{
			Name: proto.String("input"),
			Pattern: proto.String("^[a-z]+$"),
		}.Build()

		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"{{input}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: schema}.Build(),
			},
		}.Build()

		cmdTool := newCommandTool("echo", callDef)

		// Invalid input (numbers not allowed)
		inputData := map[string]interface{}{"input": "123"}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err = cmdTool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
		assert.Contains(t, err.Error(), "does not match pattern")
	})

	t.Run("regex pattern validation success", func(t *testing.T) {
		t.Parallel()
		schema := configv1.ParameterSchema_builder{
			Name: proto.String("input"),
			Pattern: proto.String("^[a-z]+$"),
		}.Build()

		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"{{input}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: schema}.Build(),
			},
		}.Build()

		cmdTool := newCommandTool("echo", callDef)

		// Valid input
		inputData := map[string]interface{}{"input": "abc"}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	// Test Case 2: Max Length Validation
	t.Run("max length validation failure", func(t *testing.T) {
		t.Parallel()
		schema := configv1.ParameterSchema_builder{
			Name: proto.String("input"),
			MaxLength: proto.Int32(3),
		}.Build()

		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"{{input}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: schema}.Build(),
			},
		}.Build()

		cmdTool := newCommandTool("echo", callDef)

		// Invalid input (too long)
		inputData := map[string]interface{}{"input": "abcd"}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err = cmdTool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
		assert.Contains(t, err.Error(), "greater than maximum")
	})

	// Test Case 3: Enum Validation
	t.Run("enum validation failure", func(t *testing.T) {
		t.Parallel()
		schema := configv1.ParameterSchema_builder{
			Name: proto.String("mode"),
			Enum: []string{"fast", "slow"},
		}.Build()

		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"--mode={{mode}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{Schema: schema}.Build(),
			},
		}.Build()

		cmdTool := newCommandTool("echo", callDef)

		// Invalid input (not in enum)
		inputData := map[string]interface{}{"mode": "medium"}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		_, err = cmdTool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
		assert.Contains(t, err.Error(), "not in allowed enum list")
	})
}
