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

func TestLocalCommandTool_JSONProtocol_LeaksStderr(t *testing.T) {
	secretValue := "SuperSensitiveData123"

	tool := v1.Tool_builder{
		Name: proto.String("test-json-leak"),
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
		CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
		Env: map[string]*configv1.SecretValue{
			"MY_SECRET": configv1.SecretValue_builder{
				PlainText: proto.String(secretValue),
			}.Build(),
		},
	}.Build()

	// Command outputs invalid JSON to stdout (triggering the error path)
	// AND prints the secret to stderr.
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo 'Invalid JSON' >&1; echo $MY_SECRET >&2"},
	}.Build()

	localTool := NewLocalCommandTool(tool, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "test-json-leak",
		Arguments: map[string]interface{}{},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	// Execute should fail because of invalid JSON
	_, err := localTool.Execute(context.Background(), req)
	assert.Error(t, err)

	// The error message should NOT contain the secret (fix verification)
	assert.NotContains(t, err.Error(), secretValue, "Error message should NOT contain secret (leak prevented)")
}
