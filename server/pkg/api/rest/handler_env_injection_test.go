package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestValidateConfigHandler_EnvVarInjection(t *testing.T) {
	// Setup: Set an environment variable that overrides a config value to something INVALID.
	// This proves that server env vars are being merged into the user-provided config.
	envVarName := "MCPANY__UPSTREAM_SERVICES__0__HTTP_SERVICE__ADDRESS"
	envVarValue := "invalid-url-from-env" // Invalid URL
	os.Setenv(envVarName, envVarValue)
	defer os.Unsetenv(envVarName)

	// Valid configuration provided by the user
	configContent := `
upstream_services:
  - name: test-service
    http_service:
      address: http://valid.com
`

	reqBody := map[string]string{"content": configContent}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/config/validate", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	ValidateConfigHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp ValidateConfigResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// We expect validation to PASS because env var injection should be disabled.
	if !resp.Valid {
		t.Errorf("Expected validation success, but it failed. Errors: %v", resp.Errors)
	}
}
