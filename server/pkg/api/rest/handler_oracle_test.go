package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfigHandler_OracleAttack(t *testing.T) {
	// Setup: Set a secret environment variable
	os.Setenv("TEST_ORACLE_SECRET", "supersecret")
	defer os.Unsetenv("TEST_ORACLE_SECRET")

	// Helper function to send request
	validate := func(configContent string) (ValidateConfigResponse, int) {
		reqBody := ValidateConfigRequest{
			Content: configContent,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()

		ValidateConfigHandler(w, req)

		var resp ValidateConfigResponse
		_ = json.NewDecoder(w.Body).Decode(&resp)
		return resp, w.Code
	}

	// Case 1: Secret exists
	// We use a minimal valid config with basic auth using the secret
	configWithSecret := `
upstream_services:
  - name: "service-with-secret"
    http_service:
      address: "http://localhost:8080"
    upstream_auth:
      basic_auth:
        username: "user"
        password:
          environment_variable: "TEST_ORACLE_SECRET"
`
	resp1, code1 := validate(configWithSecret)
	assert.Equal(t, http.StatusOK, code1)
    // If the secret exists, validation passes (no empty password error)
    // We expect no errors related to the password.
    for _, err := range resp1.Errors {
        assert.False(t, strings.Contains(err, "basic auth 'password' is empty"), "Should not report empty password for existing secret")
    }

	// Case 2: Secret does not exist
	configMissingSecret := `
upstream_services:
  - name: "service-missing-secret"
    http_service:
      address: "http://localhost:8080"
    upstream_auth:
      basic_auth:
        username: "user"
        password:
          environment_variable: "TEST_ORACLE_MISSING"
`
	resp2, code2 := validate(configMissingSecret)
	assert.Equal(t, http.StatusOK, code2)

    // If the secret is missing, validation SHOULD PASS now because we are mocking the secret resolution
    // to prevent oracle attacks.
    foundError := false
    for _, err := range resp2.Errors {
        if strings.Contains(err, "basic auth 'password' is empty") || strings.Contains(err, "failed to resolve") {
            foundError = true
            break
        }
    }
    assert.False(t, foundError, "Should NOT report error for missing secret in validation mode (Oracle Prevention). Got: %v", resp2.Errors)

    // We verified that both Case 1 (Exists) and Case 2 (Missing) return NO errors regarding the secret.
    // Thus, an attacker cannot distinguish if an env var exists or not.
}
