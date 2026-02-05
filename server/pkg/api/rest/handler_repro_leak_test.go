// Copyright 2026 Author(s) of MCP Any
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

func TestValidateConfigHandler_EnvVarEnumeration(t *testing.T) {
	// Setup: Set a secret environment variable
	secretName := "TEST_EXISTING_SECRET"
	secretValue := "dummy"
	os.Setenv(secretName, secretValue)
	defer os.Unsetenv(secretName)

	missingSecretName := "TEST_MISSING_SECRET"
	os.Unsetenv(missingSecretName)

	// Helper to send request and return errors
	validate := func(envVarName string) []string {
		config := fmt.Sprintf(`
upstream_services:
  - name: probe-service
    http_service:
      address: http://localhost:8080
    upstream_auth:
      api_key:
        param_name: X-API-Key
        value:
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
	errorsExisting := validate(secretName)
	if len(errorsExisting) > 0 {
		// Log errors for debugging
		t.Logf("Existing secret errors: %v", errorsExisting)
	}

	// 2. Probe Missing Secret
	errorsMissing := validate(missingSecretName)

	leakDetected := false
	for _, err := range errorsMissing {
		if strings.Contains(err, fmt.Sprintf("environment variable %q is not set", missingSecretName)) {
			leakDetected = true
		}
	}

	// Assertion: We should NOT detect a leak. Both cases should behave similarly regarding secret existence.
	if leakDetected {
		t.Errorf("Leak detected: Validation revealed that %s is missing.", missingSecretName)
	}
}
