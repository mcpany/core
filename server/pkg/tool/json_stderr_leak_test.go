// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_JSONProtocol_DoesNotLeakSecretsInStderr(t *testing.T) {
	// 1. Setup
	secretKey := "SERVICE_SECRET_KEY"
	secretValue := "SuperSecretServiceValue123"

	// Define tool
	tool := v1.Tool_builder{
		Name: proto.String("test-json-leak"),
	}.Build()

	// Define service with secret env var
	service := configv1.CommandLineUpstreamService_builder{
		Command:               proto.String("sh"),
		Local:                 proto.Bool(true),
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
		Env: map[string]*configv1.SecretValue{
			secretKey: configv1.SecretValue_builder{
				PlainText: proto.String(secretValue),
			}.Build(),
		},
	}.Build()

	// Define call
	// Command: sh -c 'echo "Invalid JSON"; echo "Secret is $SERVICE_SECRET_KEY" >&2'
	// This will fail JSON decoding, and print secret to stderr.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo \"Invalid JSON\"; echo \"Secret is $" + secretKey + "\" >&2"},
	}.Build()

	// Create LocalCommandTool
	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-leak")

	req := &ExecutionRequest{
		ToolName: "test-json-leak",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// 2. Execute
	_, err := localTool.Execute(context.Background(), req)

	// 3. Assert
	assert.Error(t, err)
	// The error message should contain stderr
	// We assert that it does NOT contain the secret value
	assert.NotContains(t, err.Error(), secretValue, "Error message should NOT contain the secret value from stderr")
}

func TestCommandTool_JSONProtocol_DoesNotLeakSecretsInStderr(t *testing.T) {
	// 1. Setup
	secretKey := "SERVICE_SECRET_KEY_2"
	secretValue := "SuperSecretServiceValue456"

	// Define tool
	tool := v1.Tool_builder{
		Name: proto.String("test-json-leak-2"),
	}.Build()

	// Define service with secret env var
	service := configv1.CommandLineUpstreamService_builder{
		Command:               proto.String("sh"),
		Local:                 proto.Bool(true),
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
		Env: map[string]*configv1.SecretValue{
			secretKey: configv1.SecretValue_builder{
				PlainText: proto.String(secretValue),
			}.Build(),
		},
	}.Build()

	// Define call
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo \"Invalid JSON\"; echo \"Secret is $" + secretKey + "\" >&2"},
	}.Build()

	// Create CommandTool
	cmdTool := NewCommandTool(tool, service, callDef, nil, "call-id-leak-2")

	req := &ExecutionRequest{
		ToolName: "test-json-leak-2",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// 2. Execute
	_, err := cmdTool.Execute(context.Background(), req)

	// 3. Assert
	assert.Error(t, err)
	// The error message should contain stderr
	// We assert that it does NOT contain the secret value
	assert.NotContains(t, err.Error(), secretValue, "Error message should NOT contain the secret value from stderr")
}
