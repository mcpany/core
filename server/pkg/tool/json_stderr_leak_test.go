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

// TestLocalCommandTool_DoesNotLeakSecretsInStderr_JSONProtocol verifies that
// when using the JSON protocol, if the command fails to output valid JSON,
// the stderr (which is included in the error message) does not leak secrets.
func TestLocalCommandTool_DoesNotLeakSecretsInStderr_JSONProtocol(t *testing.T) {
	// Define a secret environment variable
	secretKey := "SERVICE_SECRET_KEY"
	secretValue := "SuperSecretValueJSONLeak"

	tool := v1.Tool_builder{
		Name: proto.String("test-json-leak"),
	}.Build()

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

	// The command will echo something that is NOT valid JSON to stdout
	// AND echo the secret to stderr.
	// This will trigger the JSON decode error path in Execute.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo 'This is not JSON'; echo 'Secret is '" + "$" + secretKey + " >&2"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id-leak")

	req := &ExecutionRequest{
		ToolName: "test-json-leak",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute the tool
	_, err := localTool.Execute(context.Background(), req)

	// We expect an error because stdout is not valid JSON
	assert.Error(t, err)

	// The error message should contain Stderr output because of the implementation
	errMsg := err.Error()
	assert.Contains(t, errMsg, "failed to execute JSON CLI command")

	// CRITICAL: The error message MUST NOT contain the secret value
	// If this assertion fails, it means the secret was leaked in the error message.
	if !assert.NotContains(t, errMsg, secretValue, "Service secret environment variable leaked in error message!") {
		t.Logf("Leaked error message: %s", errMsg)
	} else {
		t.Log("Secret was successfully redacted or not present.")
	}
}
