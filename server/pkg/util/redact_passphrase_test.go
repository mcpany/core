package util

import (
	"encoding/json"
	"testing"
)

func TestRedactSensitiveKeys(t *testing.T) {
	data := map[string]string{
		"passphrase":  "secret_value_1",
		"passphrases": "secret_value_2",
		"ssh_key":     "secret_value_3",
		"bearer_token": "secret_token",
	}
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}

	redacted := RedactJSON(b)

	// We verify by parsing the redacted JSON
	var m map[string]string
	if err := json.Unmarshal(redacted, &m); err != nil {
		t.Fatalf("Failed to unmarshal redacted JSON: %v", err)
	}

	keysToCheck := []string{"passphrase", "passphrases", "ssh_key", "bearer_token"}

	for _, k := range keysToCheck {
		if val, ok := m[k]; !ok {
			t.Errorf("Key %s missing from redacted JSON", k)
		} else if val != "[REDACTED]" {
			t.Errorf("Key %s was NOT redacted. Got: %s", k, val)
		}
	}
}
