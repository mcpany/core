// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"strings"
	"testing"
	"github.com/spf13/afero"
)

func TestLoadServices_ValidationFailure(t *testing.T) {
	// Create a new memory filesystem
	fs := afero.NewMemMapFs()

	// Case 1: Missing Secret Env Var (Runtime validation)
	brokenConfigSecrets := `
upstream_services:
  - name: "broken_env"
    command_line_service:
      command: "echo"
    upstream_auth:
      api_key:
        param_name: "key"
        value:
          environment_variable: "NON_EXISTENT_VAR_12345"
`
	if err := afero.WriteFile(fs, "broken_secrets.yaml", []byte(brokenConfigSecrets), 0644); err != nil {
		t.Fatalf("Failed to write mock config file: %v", err)
	}

	// Case 2: Missing Expansion Env Var (Load time validation)
	brokenConfigExpand := `
upstream_services:
  - name: "${MISSING_NAME_VAR}"
    command_line_service:
      command: "echo"
`
	if err := afero.WriteFile(fs, "broken_expand.yaml", []byte(brokenConfigExpand), 0644); err != nil {
		t.Fatalf("Failed to write mock config file: %v", err)
	}

	// Test Case 1
	t.Run("MissingSecretEnvVar", func(t *testing.T) {
		store := NewFileStore(fs, []string{"broken_secrets.yaml"})
		os.Unsetenv("NON_EXISTENT_VAR_12345")
		_, err := LoadServices(context.Background(), store, "server")
		if err == nil {
			t.Fatal("Expected failure")
		}
		expected := "environment variable \"NON_EXISTENT_VAR_12345\" is not set"
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("Expected %q, got: %v", expected, err)
		}
	})

	// Test Case 2
	t.Run("MissingExpansionEnvVar", func(t *testing.T) {
		store := NewFileStore(fs, []string{"broken_expand.yaml"})
		os.Unsetenv("MISSING_NAME_VAR")
		_, err := LoadServices(context.Background(), store, "server")
		if err == nil {
			t.Fatal("Expected failure")
		}
		// The current store.go returns "missing environment variables: [MISSING_NAME_VAR]"
		// We also want to verify the helpful hint is present.
		expected := "missing environment variables: [MISSING_NAME_VAR]"
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("Expected %q, got: %v", expected, err)
		}

		expectedHint := "Please set them in your environment or provide a default value"
		if !strings.Contains(err.Error(), expectedHint) {
			t.Errorf("Expected hint %q, got: %v", expectedHint, err)
		}
	})
}
