// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestLocalCommandTool_SecretInArgs_Override(t *testing.T) {
	// Setup environment variable for secret
	os.Setenv("TEST_SECRET", "super-secret-value")
	defer os.Unsetenv("TEST_SECRET")

	// Define a tool that uses a secret in args
	toolDef := v1.Tool_builder{
		Name: proto.String("secret-tool"),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"properties": structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"secret_arg": structpb.NewStructValue(&structpb.Struct{}),
					},
				}),
			},
		},
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Local:   proto.Bool(true),
	}.Build()

	// Call definition with a secret parameter used in args
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{secret_arg}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("secret_arg")}.Build(),
				Secret: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("TEST_SECRET"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	// Attacker tries to inject their own value for secret_arg
	req := &ExecutionRequest{
		ToolName: "secret-tool",
		Arguments: map[string]interface{}{
			"secret_arg": "hacker-value",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	args, ok := resMap["args"].([]string)
	assert.True(t, ok)

	if len(args) > 0 {
		// Ensure that the secret value is used, effectively ignoring/overwriting the user input
		assert.Equal(t, "super-secret-value", args[0], "Expected secret value in args, but got user input")
	}
}

func TestLocalCommandTool_SecretInEnv(t *testing.T) {
	// Setup environment variable for secret
	os.Setenv("TEST_SECRET_ENV", "secret-env-value")
	defer os.Unsetenv("TEST_SECRET_ENV")

	// Define a tool that prints an environment variable
	// We use `sh` to echo the env var
	toolDef := v1.Tool_builder{
		Name: proto.String("secret-env-tool"),
		InputSchema: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"properties": structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"secret_var": structpb.NewStructValue(&structpb.Struct{}),
					},
				}),
			},
		},
	}.Build()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
		Local:   proto.Bool(true),
	}.Build()

	// Call definition with a secret parameter mapped to an environment variable name
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "echo $secret_var"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("secret_var")}.Build(),
				Secret: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("TEST_SECRET_ENV"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	localTool := NewLocalCommandTool(toolDef, service, callDef, nil, "call-id")

	// Attacker tries to inject their own value for secret_var
	req := &ExecutionRequest{
		ToolName: "secret-env-tool",
		Arguments: map[string]interface{}{
			"secret_var": "hacker-value",
		},
	}
	req.ToolInputs, _ = json.Marshal(req.Arguments)

	result, err := localTool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resMap, ok := result.(map[string]interface{})
	assert.True(t, ok)

	stdout, ok := resMap["stdout"].(string)
	assert.True(t, ok)

	// The output should be redacted because the secret was used and outputted.
	// If the user input "hacker-value" was used instead, it would NOT be redacted (as it's not in the secret list).
	assert.Equal(t, redactedPlaceholder, strings.TrimSpace(stdout), "Expected secret value (redacted) in env, but got something else")
}
