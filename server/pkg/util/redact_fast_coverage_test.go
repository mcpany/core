package util

import (
	"strings"
	"testing"
)

func TestRedactFast_Coverage(t *testing.T) {
	// 1. skipLiteral at end of input
	t.Run("skipLiteral_EndOfInput", func(t *testing.T) {
		input := []byte("true")
		end := skipLiteral(input, 0)
		if end != 4 {
			t.Errorf("expected 4, got %d", end)
		}
	})

	t.Run("skipLiteral_Middle", func(t *testing.T) {
		input := []byte("true,")
		end := skipLiteral(input, 0)
		if end != 4 {
			t.Errorf("expected 4, got %d", end)
		}
	})

	// 2. scanEscapedKeyForSensitive with large key
	t.Run("scanEscapedKeyForSensitive_Large", func(t *testing.T) {
		oldLimit := maxUnescapeLimit
		maxUnescapeLimit = 100 // Set low limit
		defer func() { maxUnescapeLimit = oldLimit }()

		var sb strings.Builder
		for i := 0; i < 200; i++ {
			sb.WriteString("a")
		}
		sb.WriteString("\\u0070assword") // password

		keyContent := []byte(sb.String())

		if !isKeySensitive(keyContent) {
			t.Errorf("expected key to be sensitive")
		}
	})

	t.Run("scanEscapedKeyForSensitive_HexEdgeCases", func(t *testing.T) {
		// Test hex parsing logic in scanEscapedKeyForSensitive
		oldLimit := maxUnescapeLimit
		maxUnescapeLimit = 1 // Force streaming path
		defer func() { maxUnescapeLimit = oldLimit }()

		tests := []struct {
			key      string
			sensitive bool
		}{
			{`\u0070assword`, true}, // Valid hex
			{`\u007`, false},        // Short hex (truncated)
			{`\uZZZZ`, false},       // Invalid hex
			{`\n`, false},           // \n
			{`\r`, false},           // \r
			{`\t`, false},           // \t
			{`\b`, false},           // \b
			{`\f`, false},           // \f
			{`\"`, false},           // \"
			{`\\`, false},           // \\
			{`\/`, false},           // \/
			{`\x`, false},           // Unknown escape
			{`\u000apassword`, true}, // \n then password
			{`\u`, false},            // \u at end
            {`\u0`, false},           // \u0 at end
		}

		for _, tt := range tests {
			if got := isKeySensitive([]byte(tt.key)); got != tt.sensitive {
				t.Errorf("isKeySensitive(%q) = %v, want %v", tt.key, got, tt.sensitive)
			}
		}
	})

	// 3. skipJSONValue coverage
	t.Run("skipJSONValue_Types", func(t *testing.T) {
		tests := []struct {
			input string
			start int
			want  int
		}{
			{`"str"`, 0, 5},
			{`123`, 0, 3},
			{`{"a":1}`, 0, 7},
			{`[1,2]`, 0, 5},
			{`true`, 0, 4},
			{`null`, 0, 4},
			{`false`, 0, 5},
			{`123,`, 0, 3}, // Number until comma
			{`foo`, 3, 3}, // Out of bounds start
		}
		for _, tt := range tests {
			got := skipJSONValue([]byte(tt.input), tt.start)
			if got != tt.want {
				t.Errorf("skipJSONValue(%q) = %d, want %d", tt.input, got, tt.want)
			}
		}
	})

	// Test skipObject and skipArray at end of input
	t.Run("skipStructs_EndOfInput", func(t *testing.T) {
		input := []byte(`{`) // unclosed
		got := skipObject(input, 0)
		if got != 1 {
			t.Errorf("skipObject: expected 1, got %d", got)
		}

		input = []byte(`[`) // unclosed
		got = skipArray(input, 0)
		if got != 1 {
			t.Errorf("skipArray: expected 1, got %d", got)
		}
	})
}
