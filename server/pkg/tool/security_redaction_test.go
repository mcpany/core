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

func TestCommandTool_RedactsSecretsInOutput(t *testing.T) {
	secretValue := "SuperSecretValueRedactionTest"

	toolDef := &v1.Tool{
		Name: proto.String("test-secret-redaction"),
	}

	service := &configv1.CommandLineUpstreamService{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
	}

	// Define a parameter "MY_SECRET" that comes from a configured secret
	callDef := &configv1.CommandLineCallDefinition{
		Args: []string{"-c", "echo $MY_SECRET"}, // The command will echo the env var
		Parameters: []*configv1.CommandLineParameterMapping{
			{
				Schema: &configv1.ParameterSchema{
					Name: proto.String("MY_SECRET"),
				},
				Secret: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{
						PlainText: secretValue,
					},
				},
			},
		},
	}

	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	req := &ExecutionRequest{
		ToolName:   "test-secret-redaction",
		ToolInputs: []byte("{}"),
	}

	// Execute
	result, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)

	// Verify
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	stdout, ok := resultMap["stdout"].(string)
	assert.True(t, ok)

	// The output should NOT contain the secret value.
	assert.NotContains(t, stdout, secretValue, "Secret value should be redacted from stdout")
	assert.Contains(t, stdout, "[REDACTED]", "Secret value should be replaced with [REDACTED]")
}
