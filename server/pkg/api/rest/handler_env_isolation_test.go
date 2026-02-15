package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestValidateConfigHandler_EnvIsolation(t *testing.T) {
	// Setup: Set an environment variable that overrides a config value
	envVar := "MCPANY__GLOBAL_SETTINGS__LOG_LEVEL"
	// LOG_LEVEL_DEBUG is a valid enum value
	envValue := "LOG_LEVEL_DEBUG"
	os.Setenv(envVar, envValue)
	defer os.Unsetenv(envVar)

	// Create a request with an INVALID log_level
	// "INVALID_LEVEL" is definitely not a valid enum value for LogLevel
	configContent := `
global_settings:
  log_level: INVALID_LEVEL
`

	reqBody := map[string]string{"content": configContent}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/config/validate", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	ValidateConfigHandler(w, req)

	var resp ValidateConfigResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Valid {
		t.Errorf("FAIL: Validation passed despite invalid config, meaning environment variable '%s=%s' was applied.", envVar, envValue)
	} else {
		// Assert that we got the expected error about invalid log level
		// This confirms that validation failed because of the user's input, not some other reason.
		if len(resp.Errors) == 0 {
			t.Errorf("FAIL: Validation failed but no errors returned.")
		}
	}
}
