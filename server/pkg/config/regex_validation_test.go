// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func mockExecLookPath() func() {
	oldLookPath := execLookPath
	execLookPath = func(file string) (string, error) {
		return "/bin/ls", nil
	}
	return func() { execLookPath = oldLookPath }
}

func TestPlainTextSecretValidation(t *testing.T) {
	cleanup := mockExecLookPath()
	defer cleanup()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("test-plaintext-secret"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{
								Command: proto.String("ls"),
								Env: map[string]*configv1.SecretValue{
									"TEST_KEY": {
										Value: &configv1.SecretValue_PlainText{
											PlainText: "invalid-key",
										},
										ValidationRegex: proto.String("^sk-[a-zA-Z0-9]{10}$"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	errs := Validate(context.Background(), config, Server)

	assert.NotEmpty(t, errs, "Validation errors expected for invalid PlainText")
	if len(errs) > 0 {
		assert.Contains(t, errs[0].Err.Error(), "secret value does not match validation regex")
	}
}

func TestEnvSecretValidation(t *testing.T) {
	cleanup := mockExecLookPath()
	defer cleanup()

	os.Setenv("TEST_ENV_KEY", "invalid-key")
	defer os.Unsetenv("TEST_ENV_KEY")

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("test-env-secret"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{
								Command: proto.String("ls"),
								Env: map[string]*configv1.SecretValue{
									"TEST_KEY": {
										Value: &configv1.SecretValue_EnvironmentVariable{
											EnvironmentVariable: "TEST_ENV_KEY",
										},
										ValidationRegex: proto.String("^sk-[a-zA-Z0-9]{10}$"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	errs := Validate(context.Background(), config, Server)

	assert.NotEmpty(t, errs, "Validation errors expected for invalid Env var")
	if len(errs) > 0 {
		assert.Contains(t, errs[0].Err.Error(), "secret value does not match validation regex")
	}
}

func TestEmptyPlainTextSecretValidation(t *testing.T) {
	cleanup := mockExecLookPath()
	defer cleanup()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("test-empty-plaintext"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{
								Command: proto.String("ls"),
								Env: map[string]*configv1.SecretValue{
									"TEST_KEY": {
										Value: &configv1.SecretValue_PlainText{
											PlainText: "",
										},
										ValidationRegex: proto.String("^.+$"), // Requires at least one char
									},
								},
							},
						},
					},
				},
			},
		},
	}

	errs := Validate(context.Background(), config, Server)

	// This should fail because empty string doesn't match ^.+$
	assert.NotEmpty(t, errs, "Validation errors expected for empty PlainText")
	if len(errs) > 0 {
		assert.Contains(t, errs[0].Err.Error(), "secret value does not match validation regex")
	}
}

func TestWhitespaceInEnvVar_WithRegex(t *testing.T) {
	cleanup := mockExecLookPath()
	defer cleanup()

	// Set env var with whitespace
	os.Setenv("TEST_WHITESPACE_KEY", "  valid-key  ")
	defer os.Unsetenv("TEST_WHITESPACE_KEY")

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("test-whitespace"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{
								Command: proto.String("ls"),
								Env: map[string]*configv1.SecretValue{
									"TEST_KEY": {
										Value: &configv1.SecretValue_EnvironmentVariable{
											EnvironmentVariable: "TEST_WHITESPACE_KEY",
										},
										ValidationRegex: proto.String("^valid-key$"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	errs := Validate(context.Background(), config, Server)

	// Should be empty if we trim whitespace
	assert.Empty(t, errs, "Validation errors not expected for env var with whitespace")
}

func TestWhitespaceInPlainText_WithRegex(t *testing.T) {
	cleanup := mockExecLookPath()
	defer cleanup()

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("test-whitespace-plain"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_StdioConnection{
							StdioConnection: &configv1.McpStdioConnection{
								Command: proto.String("ls"),
								Env: map[string]*configv1.SecretValue{
									"TEST_KEY": {
										Value: &configv1.SecretValue_PlainText{
											PlainText: "  valid-key  ",
										},
										ValidationRegex: proto.String("^valid-key$"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	errs := Validate(context.Background(), config, Server)

	// Should be empty if we trim whitespace
	assert.Empty(t, errs, "Validation errors not expected for plain text with whitespace")
}
