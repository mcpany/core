// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestCommandLineService_MissingSecretEnv_Validation(t *testing.T) {
	ctx := context.Background()

	// Mock execLookPath to allow "echo" to pass command validation
	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		return "/usr/bin/" + file, nil
	}

	service := &configv1.UpstreamServiceConfig{
		Name: proto.String("broken-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
			CommandLineService: &configv1.CommandLineUpstreamService{
				Command: proto.String("echo"),
				Env: map[string]*configv1.SecretValue{
					"MY_SECRET": {
						Value: &configv1.SecretValue_EnvironmentVariable{
							EnvironmentVariable: "MISSING_ENV_VAR_12345",
						},
					},
				},
			},
		},
	}

	err := ValidateOrError(ctx, service)
	assert.Error(t, err, "Validation should FAIL because the env var is missing")
	// The error might be wrapped, so we check the string
	assert.Contains(t, err.Error(), "MISSING_ENV_VAR_12345", "Error message should mention the missing variable")
	assert.Contains(t, err.Error(), "environment variable", "Error message should mention it is an environment variable issue")
}
