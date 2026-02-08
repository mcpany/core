package util

import "testing"

func TestRedact_EdgeCases(t *testing.T) {
	// Test boundary conditions in scanForSensitiveKeys -> checkPotentialMatch
	t.Run("WordBoundaries", func(t *testing.T) {
		tests := []struct {
			input string
			want  bool
		}{
			{"author", false},      // auth is prefix, but 'o' continues word
			{"authority", false},
			{"authentication", true}, // 'authentication' is in the sensitive list
			{"AUTH", true},
			{"AUTHOR", false},      // continuation
			{"AuthToken", true},    // CamelCase boundary
			{"AUTH_TOKEN", true},   // underscore boundary
			{"my_auth", true},      // found inside
			{"token", true},
			{"tokens", true},
			{"tokenization", false}, // token is prefix, 'i' continues
		}
		for _, tt := range tests {
			got := scanForSensitiveKeys([]byte(tt.input), false)
			if got != tt.want {
				t.Errorf("scanForSensitiveKeys(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	})


	// Test scanJSONForSensitiveKeys
	t.Run("scanJSONForSensitiveKeys", func(t *testing.T) {
		tests := []struct {
			input string
			want  bool
		}{
			{`{"token": "val"}`, true},
			// "nottoken" contains "token", and since we don't check previous character (suffix match allowed),
			// it is considered sensitive.
			{`{"nottoken": "val"}`, true},
			{`{"key": "token"}`, false}, // token in value, not key
			{`"token"`, false},          // just string, no colon
			{`"token":`, true},          // key
			{`{"token" /* comment */ : "val"}`, true}, // key with comment before colon
		}
		for _, tt := range tests {
			got := scanJSONForSensitiveKeys([]byte(tt.input))
			if got != tt.want {
				t.Errorf("scanJSONForSensitiveKeys(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	})
}
