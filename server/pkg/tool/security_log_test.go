package tool

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSecurity_PrettyPrint_RedactsSecrets(t *testing.T) {
	t.Parallel()
	// Create inputs with sensitive data
	inputs := map[string]interface{}{
		"api_key": "SUPER_SECRET_KEY_12345",
		"password": "my_password",
		"username": "user",
        "some_secret_token": "token123",
        "nested": map[string]interface{}{
            "client_secret": "secret_in_nested",
        },
	}
	inputBytes, _ := json.Marshal(inputs)

	// Use prettyPrint which is used for logging
	output := prettyPrint(inputBytes, "application/json")

	// Check that secrets are redacted
    if strings.Contains(output, "SUPER_SECRET_KEY_12345") {
        t.Errorf("Secret 'api_key' was not redacted")
    }
    if !strings.Contains(output, "[REDACTED]") {
        t.Errorf("Expected redaction marker [REDACTED]")
    }

    // Check specific fields
    if strings.Contains(output, "my_password") {
        t.Errorf("Secret 'password' was not redacted")
    }
    if strings.Contains(output, "token123") {
        t.Errorf("Secret 'some_secret_token' was not redacted")
    }
    if strings.Contains(output, "secret_in_nested") {
        t.Errorf("Secret 'client_secret' in nested map was not redacted")
    }

    // Check that non-sensitive data is preserved
    if !strings.Contains(output, "user") {
        t.Errorf("Non-sensitive data 'username' was lost")
    }
}
