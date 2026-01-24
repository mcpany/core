package util

import (
	"testing"
)

func TestRedactDSN_SpaceInPassword(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Space in password",
			input:    "postgres://user:pass word@host:5432/db",
			expected: "postgres://user:[REDACTED]@host:5432/db",
		},
		{
			name:     "At sign in password",
			input:    "postgres://user:p@ss@host:5432/db",
			expected: "postgres://user:[REDACTED]@host:5432/db",
		},
		{
			name:     "Multiple DSNs with special chars",
			input:    "postgres://u:p@ss@h postgres://u2:p2@h2",
			expected: "postgres://u:[REDACTED]@h postgres://u2:[REDACTED]@h2",
		},
        {
            name:     "Space in password followed by text",
            input:    "postgres://user:pass word@host some text",
            expected: "postgres://user:[REDACTED]@host some text",
        },
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RedactDSN(tc.input)
			if got != tc.expected {
				t.Errorf("RedactDSN(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}
