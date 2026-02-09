// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestLocalCommandTool_LeaksSecretInStderr_WithJsonProtocol(t *testing.T) {
	// 1. Setup
	secretKey := "SERVICE_SECRET_KEY"
	secretValue := "SuperSecretValueJSON"

	// Define the tool
	toolDef := v1.Tool_builder{
		Name: proto.String("test-json-leak"),
	}.Build()

	// Define the service with JSON protocol and a secret env var
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

	// The command will print the secret to stderr and then output invalid JSON to stdout
	// to trigger the error path where stderr is included in the error message.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo 'Secret is ' $" + secretKey + " >&2; echo 'Not JSON'"},
	}.Build()

	localTool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id-json-leak")

	req := &ExecutionRequest{
		ToolName: "test-json-leak",
		ToolInputs: []byte("{}"),
	}

	// 2. Execute
	// Expect an error because stdout is not JSON
	result, err := localTool.Execute(context.Background(), req)

	// 3. Verify
	assert.Error(t, err)
	assert.Nil(t, result)

	// The error message should contain Stderr content
	errMsg := err.Error()
	assert.Contains(t, errMsg, "failed to execute JSON CLI command")

	// Vulnerability Check: The secret value SHOULD BE present in the error message currently.
	// Once fixed, this assertion should be changed to NotContains.
	// For the reproduction test, we assert that it IS contained to prove the vulnerability.
	assert.NotContains(t, errMsg, secretValue, "Error message should NOT contain leaked secret (Vulnerability Fixed)")
	assert.Contains(t, errMsg, "[REDACTED]", "Error message should contain redacted placeholder")
}

func TestCommandTool_LeaksSecretInStderr_WithJsonProtocol(t *testing.T) {
	// 1. Setup
	secretKey := "SERVICE_SECRET_KEY_2"
	secretValue := "SuperSecretValueJSON2"

	// Define the tool
	toolDef := v1.Tool_builder{
		Name: proto.String("test-json-leak-2"),
	}.Build()

	// Define the service with JSON protocol and a secret env var
	service := configv1.CommandLineUpstreamService_builder{
		Command:               proto.String("sh"),
		Local:                 proto.Bool(true), // Use Local=true for CommandTool test too, as it defaults to LocalExecutor if image is empty
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
		Env: map[string]*configv1.SecretValue{
			secretKey: configv1.SecretValue_builder{
				PlainText: proto.String(secretValue),
			}.Build(),
		},
	}.Build()

	// The command will print the secret to stderr and then output invalid JSON to stdout
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo 'Secret is ' $" + secretKey + " >&2; echo 'Not JSON'"},
	}.Build()

	cmdTool := NewCommandTool(toolDef, service, callDef, nil, "call-id-json-leak-2")

	req := &ExecutionRequest{
		ToolName: "test-json-leak-2",
		ToolInputs: []byte("{}"),
	}

	// 2. Execute
	result, err := cmdTool.Execute(context.Background(), req)

	// 3. Verify
	assert.Error(t, err)
	assert.Nil(t, result)

	errMsg := err.Error()
	assert.Contains(t, errMsg, "failed to execute JSON CLI command")

	assert.NotContains(t, errMsg, secretValue, "Error message should NOT contain leaked secret (Vulnerability Fixed)")
	assert.Contains(t, errMsg, "[REDACTED]", "Error message should contain redacted placeholder")
}
