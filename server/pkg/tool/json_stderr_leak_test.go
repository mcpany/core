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

func TestLocalCommandTool_DoesNotLeakSecretsInJSONStderr(t *testing.T) {
	secretValue := "SuperSecretValue123"

	tool := v1.Tool_builder{
		Name: proto.String("test-json-leak"),
	}.Build()

	// Define a service with a secret environment variable
	// We use builders as per other tests
	commProto := configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
		CommunicationProtocol: &commProto,
		Env: map[string]*configv1.SecretValue{
			"SECRET_VAR": configv1.SecretValue_builder{
				PlainText: proto.String(secretValue),
			}.Build(),
		},
	}.Build()

	// The command echoes invalid JSON to stdout (to trigger the error path)
	// and echoes the secret to stderr.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo 'not json'; echo \"Secret is $SECRET_VAR\" >&2"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "test-json-leak",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute should fail because stdout is not JSON
	result, err := localTool.Execute(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)

	// The error message contains Stderr. We assert that the secret value is NOT in it.
	errMsg := err.Error()
	assert.Contains(t, errMsg, "failed to execute JSON CLI command")

	// Verify leakage (this should fail before fix)
	if assert.NotContains(t, errMsg, secretValue, "Secret value leaked in stderr in error message") {
		// If it doesn't contain the secret, it might be because we didn't run the command correctly?
		// Or maybe the fix is already there? (Unlikely)
	} else {
		t.Logf("Confirmed: Secret leaked in error message: %s", errMsg)
	}
}
