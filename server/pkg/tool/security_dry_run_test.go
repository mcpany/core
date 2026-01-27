// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestCommandTool_Security_DryRun_Redaction(t *testing.T) {
	// Setup a secret environment variable
	secretVal := "super-secret-api-key-12345"
	err := os.Setenv("TEST_SECRET_API_KEY", secretVal)
	require.NoError(t, err)
	defer os.Unsetenv("TEST_SECRET_API_KEY")

	// Define a command tool that uses this secret
	callDef := &configv1.CommandLineCallDefinition{
		Parameters: []*configv1.CommandLineParameterMapping{
			{
				Schema: &configv1.ParameterSchema{Name: proto.String("api_key")},
				Secret: &configv1.SecretValue{
					Value: &configv1.SecretValue_EnvironmentVariable{
						EnvironmentVariable: "TEST_SECRET_API_KEY",
					},
				},
			},
			{
				Schema: &configv1.ParameterSchema{Name: proto.String("args")},
			},
		},
	}

	// Create the tool
	cmdTool := newCommandTool("echo", callDef)

	// Prepare input
	inputData := map[string]interface{}{"args": []string{"some-arg"}}
	inputs, err := json.Marshal(inputData)
	require.NoError(t, err)

	// Execute in DryRun mode
	req := &tool.ExecutionRequest{
		ToolInputs: inputs,
		DryRun:     true,
	}

	result, err := cmdTool.Execute(context.Background(), req)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok)
	require.True(t, resultMap["dry_run"].(bool))

	requestMap, ok := resultMap["request"].(map[string]interface{})
	require.True(t, ok)

	// Check env variables for leakage
	envVars, ok := requestMap["env"].([]string)
	require.True(t, ok)

	foundSecret := false
	for _, env := range envVars {
		if strings.Contains(env, secretVal) {
			foundSecret = true
			break
		}
	}

	// Ideally this should fail initially (foundSecret == true)
	// After fix, it should be false
	if foundSecret {
		t.Logf("VULNERABILITY FOUND: Secret value %q leaked in dry run env vars: %v", secretVal, envVars)
		// We assert failure here to confirm the vulnerability exists
		assert.Fail(t, "Secret leaked in dry run")
	}

	// Verify that the secret was indeed redacted
	foundRedacted := false
	expected := "api_key=[REDACTED]"
	for _, env := range envVars {
		if env == expected {
			foundRedacted = true
			break
		}
	}
	assert.True(t, foundRedacted, "Expected to find redacted secret %q in env vars, but got %v", expected, envVars)
}
