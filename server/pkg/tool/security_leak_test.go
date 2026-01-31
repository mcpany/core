// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_DoesNotLeakHostEnv(t *testing.T) {
	// Set a sensitive environment variable in the host process
	secretKey := "HOST_SECRET_KEY"
	secretValue := "SuperSecretValue123"
	os.Setenv(secretKey, secretValue)
	defer os.Unsetenv(secretKey)

	tool := v1.Tool_builder{
		Name: proto.String("test-env-leak"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
	}.Build()
	// Try to echo the secret environment variable
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo $" + secretKey},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "test-env-leak",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	// stdout should NOT contain the secret value
	stdout, ok := resultMap["stdout"].(string)
	assert.True(t, ok)

	assert.NotContains(t, stdout, secretValue, "Host environment variable should NOT be leaked")
}

func TestCommandTool_DoesNotLeakHostEnv(t *testing.T) {
	// Set a sensitive environment variable in the host process
	secretKey := "HOST_SECRET_KEY_2"
	secretValue := "SuperSecretValue456"
	os.Setenv(secretKey, secretValue)
	defer os.Unsetenv(secretKey)

	tool := v1.Tool_builder{
		Name: proto.String("test-env-leak-2"),
	}.Build()
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
	}.Build()
	// Try to echo the secret environment variable
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo $" + secretKey},
	}.Build()

	// Create CommandTool instead of LocalCommandTool
	cmdTool := NewCommandTool(tool, service, callDef, nil, "call-id-2")

	req := &ExecutionRequest{
		ToolName: "test-env-leak-2",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := cmdTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	// stdout should NOT contain the secret value
	stdout, ok := resultMap["stdout"].(string)
	assert.True(t, ok)

	assert.NotContains(t, stdout, secretValue, "Host environment variable should NOT be leaked via CommandTool")
}

func TestLocalCommandTool_SecretLeak_Stderr(t *testing.T) {
	// Setup a tool using JSON protocol
	// Command: sh -c 'echo "SECRET=$SECRET_VAL" >&2; echo "invalid json"'

	secretVal := "super_secret_password"

	inputSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"properties": structpb.NewStructValue(&structpb.Struct{
				Fields: map[string]*structpb.Value{},
			}),
		},
	}
	toolProto := v1.Tool_builder{
		Name:        proto.String("leak-tool"),
		InputSchema: inputSchema,
	}.Build()

	secretValProto := configv1.SecretValue_builder{
		PlainText: proto.String(secretVal),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
		Env: map[string]*configv1.SecretValue{
			"SECRET_VAL": secretValProto,
		},
	}.Build()

	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo \"SECRET=$SECRET_VAL\" >&2; echo \"invalid json\""},
	}.Build()

	localTool := NewLocalCommandTool(toolProto, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "leak-tool",
		ToolInputs: []byte("{}"),
	}

	_, err := localTool.Execute(context.Background(), req)

	// Expect error because stdout is not valid JSON
	assert.Error(t, err)

	// The secret should NOT be present in the error message
	assert.NotContains(t, err.Error(), secretVal, "Secret leaked in error message!")
}
