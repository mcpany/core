// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestBasicAuthSecurity_NoSecretResolution ensures that Basic Auth validation
// does NOT resolve secrets (read files or make network requests) when validation mode is enabled.
// This prevents Information Disclosure (file existence) and Blind SSRF.
func TestBasicAuthSecurity_NoSecretResolution(t *testing.T) {
	// 1. Construct a config with basic_auth pointing to a non-existent file.
	// If the server tries to read it, it will fail with "no such file or directory" or "not allowed".
	// If the server respects SkipSecretValidationKey (as set in ValidateConfigHandler),
	// it should skip reading the file and validation should pass (or fail on other things, but not file read).
	configContent := `
users:
  - id: "user1"
    authentication:
      basic_auth:
        username: "admin"
        password:
          file_path: "/tmp/nonexistent_file_for_ssrf_test_12345"
`

	reqBody := ValidateConfigRequest{
		Content: configContent,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	// 2. Call the handler
	ValidateConfigHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %d", resp.StatusCode)
	}

	var validationResp ValidateConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&validationResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// 3. Assert that we do NOT get a secret resolution error.
	// If we get "failed to resolve basic auth password secret", it means the server tried to read the file.
	for _, errStr := range validationResp.Errors {
		if strings.Contains(errStr, "failed to resolve basic auth password secret") {
			t.Fatalf("VULNERABILITY CONFIRMED: Server attempted to resolve the secret during validation: %s", errStr)
		}
	}
}
