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

	config := func() *configv1.McpAnyServerConfig {
		secret := configv1.SecretValue_builder{
			PlainText:       proto.String("invalid-key"),
			ValidationRegex: proto.String("^sk-[a-zA-Z0-9]{10}$"),
		}.Build()

		stdio := configv1.McpStdioConnection_builder{
			Command: proto.String("ls"),
			Env: map[string]*configv1.SecretValue{
				"TEST_KEY": secret,
			},
		}.Build()

		mcp := configv1.McpUpstreamService_builder{
			StdioConnection: stdio,
		}.Build()

		svc := configv1.UpstreamServiceConfig_builder{
			Name:       proto.String("test-plaintext-secret"),
			McpService: mcp,
		}.Build()

		return configv1.McpAnyServerConfig_builder{
			UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
		}.Build()
	}()

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

	config := func() *configv1.McpAnyServerConfig {
		secret := configv1.SecretValue_builder{
			EnvironmentVariable: proto.String("TEST_ENV_KEY"),
			ValidationRegex:     proto.String("^sk-[a-zA-Z0-9]{10}$"),
		}.Build()

		stdio := configv1.McpStdioConnection_builder{
			Command: proto.String("ls"),
			Env: map[string]*configv1.SecretValue{
				"TEST_KEY": secret,
			},
		}.Build()

		mcp := configv1.McpUpstreamService_builder{
			StdioConnection: stdio,
		}.Build()

		svc := configv1.UpstreamServiceConfig_builder{
			Name:       proto.String("test-env-secret"),
			McpService: mcp,
		}.Build()

		return configv1.McpAnyServerConfig_builder{
			UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
		}.Build()
	}()

	errs := Validate(context.Background(), config, Server)

	assert.NotEmpty(t, errs, "Validation errors expected for invalid Env var")
	if len(errs) > 0 {
		assert.Contains(t, errs[0].Err.Error(), "secret value does not match validation regex")
	}
}

func TestEmptyPlainTextSecretValidation(t *testing.T) {
	cleanup := mockExecLookPath()
	defer cleanup()

	config := func() *configv1.McpAnyServerConfig {
		secret := configv1.SecretValue_builder{
			PlainText:       proto.String(""),
			ValidationRegex: proto.String("^.+$"),
		}.Build()

		stdio := configv1.McpStdioConnection_builder{
			Command: proto.String("ls"),
			Env: map[string]*configv1.SecretValue{
				"TEST_KEY": secret,
			},
		}.Build()

		mcp := configv1.McpUpstreamService_builder{
			StdioConnection: stdio,
		}.Build()

		svc := configv1.UpstreamServiceConfig_builder{
			Name:       proto.String("test-empty-plaintext"),
			McpService: mcp,
		}.Build()

		return configv1.McpAnyServerConfig_builder{
			UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
		}.Build()
	}()

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

	config := func() *configv1.McpAnyServerConfig {
		secret := configv1.SecretValue_builder{
			EnvironmentVariable: proto.String("TEST_WHITESPACE_KEY"),
			ValidationRegex:     proto.String("^valid-key$"),
		}.Build()

		stdio := configv1.McpStdioConnection_builder{
			Command: proto.String("ls"),
			Env: map[string]*configv1.SecretValue{
				"TEST_KEY": secret,
			},
		}.Build()

		mcp := configv1.McpUpstreamService_builder{
			StdioConnection: stdio,
		}.Build()

		svc := configv1.UpstreamServiceConfig_builder{
			Name:       proto.String("test-whitespace"),
			McpService: mcp,
		}.Build()

		return configv1.McpAnyServerConfig_builder{
			UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
		}.Build()
	}()

	errs := Validate(context.Background(), config, Server)

	// Should be empty if we trim whitespace
	assert.Empty(t, errs, "Validation errors not expected for env var with whitespace")
}

func TestWhitespaceInPlainText_WithRegex(t *testing.T) {
	cleanup := mockExecLookPath()
	defer cleanup()

	config := func() *configv1.McpAnyServerConfig {
		secret := configv1.SecretValue_builder{
			PlainText:       proto.String("  valid-key  "),
			ValidationRegex: proto.String("^valid-key$"),
		}.Build()

		stdio := configv1.McpStdioConnection_builder{
			Command: proto.String("ls"),
			Env: map[string]*configv1.SecretValue{
				"TEST_KEY": secret,
			},
		}.Build()

		mcp := configv1.McpUpstreamService_builder{
			StdioConnection: stdio,
		}.Build()

		svc := configv1.UpstreamServiceConfig_builder{
			Name:       proto.String("test-whitespace-plain"),
			McpService: mcp,
		}.Build()

		return configv1.McpAnyServerConfig_builder{
			UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
		}.Build()
	}()

	errs := Validate(context.Background(), config, Server)

	// Should be empty if we trim whitespace
	assert.Empty(t, errs, "Validation errors not expected for plain text with whitespace")
}
