package validation

import (
	"testing"
)

func TestIsValidURL_ControlChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid URL", "http://example.com", true},
		{"Valid URL with path", "http://example.com/path", true},
		{"URL with newline", "http://example.com\n", false},
		{"URL with tab", "http://example.com\t", false},
		{"URL with null byte", "http://example.com\x00", false},
		{"URL with carriage return", "http://example.com\r", false},
		{"URL with backspace", "http://example.com\b", false},
		{"URL with delete char", "http://example.com\x7f", false},
		{"URL with vertical tab", "http://example.com\v", false},
		{"URL with form feed", "http://example.com\f", false},
		{"URL with escape", "http://example.com\x1b", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsValidURL(tc.input); got != tc.expected {
				t.Errorf("IsValidURL(%q) = %v; want %v", tc.input, got, tc.expected)
			}
		})
	}
}
