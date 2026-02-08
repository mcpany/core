package config

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

func TestValidatorActionableErrors(t *testing.T) {
	// 1. Test missing env var
	secretEnv := configv1.SecretValue_builder{
		EnvironmentVariable: proto.String("MISSING_ENV_VAR_TEST"),
	}.Build()
	err := validateSecretValue(context.Background(), secretEnv)
	if err == nil {
		t.Fatal("Expected error for missing env var")
	}
	if !strings.Contains(err.Error(), "-> Fix: Set the environment variable") {
		t.Errorf("Expected actionable suggestion for env var, got: %v", err)
	}

	// 2. Test missing command
	err = validateCommandExists(context.Background(), "non_existent_command_test", "")
	if err == nil {
		t.Fatal("Expected error for missing command")
	}
	if !strings.Contains(err.Error(), "-> Fix: Ensure") {
		t.Errorf("Expected actionable suggestion for command, got: %v", err)
	}

	// 3. Test missing file
	err = validateFileExists(context.Background(), "/non/existent/file/path/test", "")
	if err == nil {
		t.Fatal("Expected error for missing file")
	}
	if !strings.Contains(err.Error(), "-> Fix: Check if the file exists") {
		t.Errorf("Expected actionable suggestion for file, got: %v", err)
	}

	// 4. Test invalid URL
	addr := "htp://invalid-scheme.com"
	svc := configv1.HttpUpstreamService_builder{
		Address: proto.String(addr),
	}.Build()
	err = validateHTTPService(svc)
	if err == nil {
		t.Fatal("Expected error for invalid URL")
	}
	if !strings.Contains(err.Error(), "-> Fix: Use 'http' or 'https'") {
		t.Errorf("Expected actionable suggestion for URL, got: %v", err)
	}
}
