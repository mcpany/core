// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestValidateConfigHandler_BasicAuth_EnvVarEnumeration(t *testing.T) {
	// Setup: Set a secret environment variable
	secretName := "TEST_EXISTING_SECRET_BASIC_AUTH"
	secretValue := "dummy"
	os.Setenv(secretName, secretValue)
	defer os.Unsetenv(secretName)

	missingSecretName := "TEST_MISSING_SECRET_BASIC_AUTH"
	os.Unsetenv(missingSecretName)

	// Helper to send request and return errors
	validate := func(envVarName string) []string {
		config := fmt.Sprintf(`
upstream_services:
  - name: probe-service
    http_service:
      address: http://localhost:8080
    upstream_auth:
      basic_auth:
        username: user
        password:
          environment_variable: %s
`, envVarName)

		reqBody := map[string]string{"content": config}
		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/config/validate", bytes.NewBuffer(jsonBody))
		w := httptest.NewRecorder()

		ValidateConfigHandler(w, req)

		var resp ValidateConfigResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		return resp.Errors
	}

	// 1. Probe Existing Secret
	// This should succeed (valid configuration, secret exists) OR fail on schema validation if my config is wrong.
	// But crucially, it should NOT return an error about secret resolution failure.
	errorsExisting := validate(secretName)
	if len(errorsExisting) > 0 {
		t.Logf("Existing secret errors: %v", errorsExisting)
	}

	// 2. Probe Missing Secret
	// This should behave identically to Existing Secret (either succeed or fail schema/other validation).
	// It should NOT reveal that the secret is missing.
	errorsMissing := validate(missingSecretName)

	leakDetected := false
	for _, err := range errorsMissing {
		// If we see "environment variable ... is not set", it's a leak from validateSecretValue (which should be skipped)
		if strings.Contains(err, fmt.Sprintf("environment variable %q is not set", missingSecretName)) {
			leakDetected = true
            t.Logf("Leak detected: validateSecretValue revealed missing env var")
		}
        // If we see "failed to resolve basic auth password secret", it's a leak from util.ResolveSecret
        if strings.Contains(err, "failed to resolve basic auth password secret") {
             leakDetected = true
             t.Logf("Leak detected: util.ResolveSecret revealed missing env var")
        }
	}

	if leakDetected {
		t.Errorf("VULNERABILITY CONFIRMED: Leak detected via Basic Auth password validation.")
	}
}
