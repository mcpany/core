package middleware

import (
	"encoding/json"
	"testing"

	"log/slog"
)

func TestLazyLogPayload(t *testing.T) {
	// Payload with sensitive data
	payload := map[string]interface{}{
		"method": "login",
		"params": map[string]interface{}{
			"username": "user",
			"password": "secret_password",
		},
	}

	lazy := LazyLogPayload{Value: payload}
	logValue := lazy.LogValue()

	if logValue.Kind() != slog.KindString {
		t.Errorf("Expected String kind, got %v", logValue.Kind())
	}

	jsonStr := logValue.String()
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &decoded); err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	params, ok := decoded["params"].(map[string]interface{})
	if !ok {
		t.Fatalf("params not found or invalid type")
	}

	if params["password"] != "[REDACTED]" {
		t.Errorf("Expected password to be redacted, got %v", params["password"])
	}
}
