// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

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

func TestCommandTool_ShellInjection_Prevention(t *testing.T) {
	// Setup a CommandTool configured to use "awk"
	// This test verifies that shell injection protections are active for CommandTool.

	// Define the tool service
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
	}).Build()

	// Define the call definition allowing "args"
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build(),
			}.Build(),
		},
	}.Build()

	// Allow "args" in input schema
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"args": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}

	toolProto := v1.Tool_builder{
		Name:        proto.String("rce-test-awk"),
		InputSchema: inputSchema,
	}.Build()

	cmdTool := tool.NewCommandTool(
		toolProto,
		service,
		callDef,
		nil,
		"call-id",
	)

	// Payload: awk 'BEGIN { system("echo INJECTED") }'
	// We pass the program as the first argument.

	inputData := map[string]interface{}{
		"args": []string{"BEGIN { system(\"echo INJECTED\") }"},
	}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)

	req := &tool.ExecutionRequest{ToolInputs: inputs}

	// Execute
	_, err = cmdTool.Execute(context.Background(), req)

	// Expect an error due to injection detection
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "injection detected")
		t.Log("Security Fix Verified: Injection blocked")
	} else {
		t.Error("Security Fix Failed: Injection was not blocked")
	}
}

func TestCommandTool_EnvInjection_Prevention(t *testing.T) {
	// Setup a CommandTool configured to use "awk"
	// This test verifies that shell injection protections are active for environment variables loaded via configuration.

	// Define the tool service
	service := (&configv1.CommandLineUpstreamService_builder{
		Command: proto.String("awk"),
		Env: map[string]*configv1.SecretValue{
			"PYTHONPATH": (&configv1.SecretValue_builder{
				PlainText: proto.String("should_be_skipped"),
			}).Build(),
			"BASH_ENV": (&configv1.SecretValue_builder{
				PlainText: proto.String("should_be_skipped"),
			}).Build(),
			"SAFE_VAR": (&configv1.SecretValue_builder{
				PlainText: proto.String("should_be_included"),
			}).Build(),
		},
	}).Build()

	// Define the call definition allowing "args"
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("args")}.Build(),
			}.Build(),
		},
	}.Build()

	// Allow "args" in input schema
	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{
					"args": structpb.NewStructValue(&structpb.Struct{}),
				},
			}),
		},
	}

	toolProto := v1.Tool_builder{
		Name:        proto.String("env-test-awk"),
		InputSchema: inputSchema,
	}.Build()

	// Use LocalCommandTool to run directly on the host (where we can intercept env variables)
	cmdTool := tool.NewLocalCommandTool(
		toolProto,
		service,
		callDef,
		nil,
		"call-id",
	)

	// Since we can't easily capture the environment variables from the result of the `echo` command without changing test runner,
	// we will run a dummy script that echoes the variables or just use the dry-run feature.

	inputData := map[string]interface{}{
	}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)

	req := &tool.ExecutionRequest{ToolInputs: inputs, DryRun: true}

	// Execute
	res, err := cmdTool.Execute(context.Background(), req)
	require.NoError(t, err)

	resMap, ok := res.(map[string]interface{})
	require.True(t, ok)

	dryRunMap, ok := resMap["request"].(map[string]interface{})
	require.True(t, ok)

	envList, ok := dryRunMap["env"].([]string)
	require.True(t, ok)

	foundPythonPath := false
	foundBashEnv := false
	foundSafeVar := false

	for _, e := range envList {
		if e == "PYTHONPATH=should_be_skipped" {
			foundPythonPath = true
		}
		if e == "BASH_ENV=should_be_skipped" {
			foundBashEnv = true
		}
		if e == "SAFE_VAR=should_be_included" {
			foundSafeVar = true
		}
	}

	assert.False(t, foundPythonPath, "Security Fix Failed: PYTHONPATH was not blocked")
	assert.False(t, foundBashEnv, "Security Fix Failed: BASH_ENV was not blocked")
	assert.True(t, foundSafeVar, "SAFE_VAR was unexpectedly blocked")
}
