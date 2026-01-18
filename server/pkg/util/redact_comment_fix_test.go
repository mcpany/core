package util

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_Comments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
		missing  []string
	}{
		{
			name: "Object with line comment containing brace",
			input: `{
				"secret": {
					"a": 1,
					"b": 2 // closing brace } here
				},
				"public": "visible"
			}`,
			contains: []string{`"secret": "[REDACTED]"`, `"public": "visible"`},
			missing:  []string{"here"},
		},
		{
			name: "Array with line comment containing bracket",
			input: `{
				"secret": [
					1,
					2 // closing bracket ] here
				],
				"public": "visible"
			}`,
			contains: []string{`"secret": "[REDACTED]"`, `"public": "visible"`},
			missing:  []string{"here"},
		},
		{
			name: "Object with block comment containing brace",
			input: `{
				"secret": {
					"a": 1,
					/* multi-line
					   closing brace } here
					*/
					"b": 2
				},
				"public": "visible"
			}`,
			contains: []string{`"secret": "[REDACTED]"`, `"public": "visible"`},
			missing:  []string{"here"},
		},
		{
			name: "Array with block comment containing bracket",
			input: `{
				"secret": [
					1,
					/* multi-line
					   closing bracket ] here
					*/
					2
				],
				"public": "visible"
			}`,
			contains: []string{`"secret": "[REDACTED]"`, `"public": "visible"`},
			missing:  []string{"here"},
		},
        {
            name: "Object with comment containing quote",
            input: `{
                "secret": {
                    "a": 1 // comment with "quote"
                },
                "public": "visible"
            }`,
            contains: []string{`"secret": "[REDACTED]"`, `"public": "visible"`},
            missing: []string{`quote`},
        },
        {
            name: "Object with comment containing unmatched quote",
            input: `{
                "secret": {
                    "a": 1 // comment with " unmatched quote
                },
                "public": "visible"
            }`,
            contains: []string{`"secret": "[REDACTED]"`, `"public": "visible"`},
            missing: []string{`unmatched`},
        },
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := RedactJSON([]byte(tc.input))
			outStr := string(output)
			for _, c := range tc.contains {
				assert.Contains(t, outStr, c)
			}
			for _, m := range tc.missing {
				assert.NotContains(t, outStr, m)
			}
		})
	}
}
