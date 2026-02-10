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

func TestLocalCommandTool_DoesNotLeakSecretInStderr_JSONProtocol(t *testing.T) {
	// This test ensures that secrets in stderr are redacted when using JSON protocol
	// and the command fails to produce valid JSON stdout.

	secretKey := "API_KEY"
	secretValue := "SuperSecretToken123"

	tool := v1.Tool_builder{
		Name: proto.String("test-secret-leak"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command:               proto.String("sh"),
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
		Env: map[string]*configv1.SecretValue{
			secretKey: configv1.SecretValue_builder{PlainText: proto.String(secretValue)}.Build(),
		},
	}.Build()

	// Command that prints invalid JSON to stdout (triggering the error path)
	// and prints the secret to stderr.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo 'not json'; echo \"Error: Invalid input. Secret is $" + secretKey + "\" >&2"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-leak")

	req := &ExecutionRequest{
		ToolName: "test-secret-leak",
		ToolInputs: []byte("{}"),
	}

	result, err := localTool.Execute(context.Background(), req)

	// We expect an error because stdout is "not json"
	assert.Error(t, err)
	assert.Nil(t, result)

	// The error message should NOT contain the secret.
	if err != nil {
		assert.NotContains(t, err.Error(), secretValue, "Error message should NOT leak the secret")
		assert.Contains(t, err.Error(), "[REDACTED]", "Error message should contain redaction placeholder")
	}
}
