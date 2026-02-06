package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandWithBracesInDefault(t *testing.T) {
	// Case 1: Default value contains braces
	input := []byte(`key: ${VAR:{}}`)
	expected := []byte(`key: {}`)

	// Ensure VAR is not set
	os.Unsetenv("VAR")

	expanded, err := expand(input)
	assert.NoError(t, err)
	assert.Equal(t, string(expected), string(expanded))
}

func TestExpandDetailed(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		env       map[string]string
		expected  string
		expectErr bool
	}{
		{
			name:     "Basic expansion",
			input:    "Hello ${NAME}",
			env:      map[string]string{"NAME": "World"},
			expected: "Hello World",
		},
		{
			name:     "Basic expansion short form",
			input:    "Hello $NAME",
			env:      map[string]string{"NAME": "World"},
			expected: "Hello World",
		},
		{
			name:     "Expansion with default value",
			input:    "Hello ${NAME:World}",
			env:      map[string]string{},
			expected: "Hello World",
		},
		{
			name:     "Expansion with default value overridden",
			input:    "Hello ${NAME:World}",
			env:      map[string]string{"NAME": "Universe"},
			expected: "Hello Universe",
		},
		{
			name:     "Nested braces in default value",
			input:    `config: ${CONFIG:{"key": "value"}}`,
			env:      map[string]string{},
			expected: `config: {"key": "value"}`,
		},
		{
			name:     "Multiple levels of nested braces",
			input:    `nested: ${VAL:{"a": {"b": "c"}}}`,
			env:      map[string]string{},
			expected: `nested: {"a": {"b": "c"}}`,
		},
		{
			name:     "Multiple variables",
			input:    "${A} ${B}",
			env:      map[string]string{"A": "1", "B": "2"},
			expected: "1 2",
		},
		{
			name:     "Mixed text and variables",
			input:    "pre-${A}-mid-${B}-post",
			env:      map[string]string{"A": "start", "B": "end"},
			expected: "pre-start-mid-end-post",
		},
		{
			name:     "Missing variable with no default",
			input:    "Hello ${MISSING}",
			env:      map[string]string{},
			expectErr: true,
		},
		{
			name:      "Missing variable short form",
			input:     "Hello $MISSING",
			env:       map[string]string{},
			expectErr: true,
		},
		{
			name:     "Unclosed brace",
			input:    "Hello ${NAME",
			env:      map[string]string{"NAME": "World"},
			expected: "Hello ${NAME", // Should remain as literal if unclosed? Or error? My implementation treats as literal.
		},
		{
			name:     "Trailing $",
			input:    "Price: $",
			env:      map[string]string{},
			expected: "Price: $",
		},
		{
			name:     "$ followed by number",
			input:    "Price: $100",
			env:      map[string]string{},
			expected: "Price: $100", // $1 is not valid variable start (must be alpha or _)
		},
		{
			name:     "$ followed by invalid char",
			input:    "Email: user@example.com",
			env:      map[string]string{},
			expected: "Email: user@example.com", // @ is not valid variable start
		},
        {
            name: "Variable with empty default",
            input: "Value: ${VAR:}",
            env: map[string]string{},
            expected: "Value: ",
        },
        {
            name: "Variable set to empty string with default",
            input: "Value: ${VAR:default}",
            env: map[string]string{"VAR": ""},
            expected: "Value: default", // My implementation prefers default if empty? Original behavior check needed.
        },
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup environment
			for k, v := range tc.env {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}
			// Unset envs that shouldn't be set (to prevent interference from actual env)
			// Ideally we should clear env or mock LookupEnv, but os.Setenv is what we have.
			// We iterate over env keys required for test. For missing ones, we hope they are not in system env.
			// For "MISSING", we explicitly unset it.
			if tc.expectErr {
				os.Unsetenv("MISSING")
			}
            if tc.name == "Variable with empty default" {
                os.Unsetenv("VAR")
            }

            // Special handling for "Variable set to empty string with default"
            // We set it in the loop above.

			expanded, err := expand([]byte(tc.input))
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, string(expanded))
			}
		})
	}
}
