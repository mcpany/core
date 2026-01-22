package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestSecurityPathValidationRepro(t *testing.T) {
	// Setup allowed paths to empty to ensure strict checking
	validation.SetAllowedPaths([]string{})

	// Create a dummy config with a volume mount that is likely to fail
	// "/etc/passwd" is commonly not allowed if not in CWD or allowed paths
	cmdService := &configv1.CommandLineUpstreamService{
		Command: proto.String("echo"),
		ContainerEnvironment: &configv1.ContainerEnvironment{
			Image: proto.String("ubuntu:latest"),
			Volumes: map[string]string{
				"/etc/passwd": "/tmp/passwd",
			},
		},
	}

	err := validateCommandLineService(cmdService)
	assert.Error(t, err)

	// Since IsAllowedPath error is just a wrapped error in validateCommandLineService, it might not be a top-level ActionableError immediately unless wrapped.
	// But let's check the error message text which we improved.
	assert.Contains(t, err.Error(), "Add \"/etc/passwd\" to 'allowed_file_paths' in 'global_settings'")
}

func TestValidationRegexErrorRepro(t *testing.T) {
	// Case 1: Regex validation failure
	secret := &configv1.SecretValue{
		Value: &configv1.SecretValue_PlainText{
			PlainText: "invalid-value",
		},
		ValidationRegex: proto.String("^[0-9]+$"),
	}

	err := validateSecretValue(secret)
	assert.Error(t, err)

	// Should be actionable now
	ae, ok := err.(*ActionableError)
	assert.True(t, ok, "Expected ActionableError")
	if ok {
		assert.Contains(t, ae.Suggestion, "Ensure the secret value matches the regex pattern")
	}
}
