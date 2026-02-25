package validation

import (
	"testing"
)

func TestIsLoopbackShorthand(t *testing.T) {
	tests := []struct {
		val      string
		expected bool
	}{
		{"127.1", true},
		{"127.0.1", true},
		{"127.0.0.1", true}, // Even though parseable, it matches pattern
		{"127.255", true},
		{"127.0.0.255", true},
		{"127.", false}, // Too short
		{"127.a", false},
		{"127.txt", false},
		{"10.1", false},
		{"0", false},
		{"127", false},
		{"127..", true}, // Technically invalid but matches pattern. Should be blocked anyway.
	}

	for _, tt := range tests {
		t.Run(tt.val, func(t *testing.T) {
			result := IsLoopbackShorthand(tt.val)
			if result != tt.expected {
				t.Errorf("IsLoopbackShorthand(%q) = %v, expected %v", tt.val, result, tt.expected)
			}
		})
	}
}
