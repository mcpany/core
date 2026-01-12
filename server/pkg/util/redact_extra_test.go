// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		offset   int
		expected bool
	}{
		{
			name:     "simple key",
			input:    `"key": "value"`,
			offset:   1, // Start inside the key
			expected: true,
		},
		{
			name:     "key with whitespace",
			input:    `"key" : "value"`,
			offset:   1,
			expected: true,
		},
		{
			name:     "not a key (no colon)",
			input:    `"key" "value"`,
			offset:   1,
			expected: false,
		},
		{
			name:     "not a key (end of input)",
			input:    `"key"`,
			offset:   1,
			expected: false,
		},
		{
			name:     "escape sequence conservative",
			input:    `"key\"": "value"`,
			offset:   1,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isKey([]byte(tt.input), tt.offset))
		})
	}
}

func TestScanForSensitiveKeys_Extra(t *testing.T) {
    // Cover the case where len(input) < 128
    // Cover validateKeyContext = true

    input := []byte(`{"password": "123"}`)
    assert.True(t, scanForSensitiveKeys(input, true))

    input2 := []byte(`{"user": "alice"}`)
    assert.False(t, scanForSensitiveKeys(input2, true))

    // Case where second char check fails
    // "pass" vs "paaa"
    // "password" starts with 'p', next is 'a'. mask has 'a'.
    // "pb..." next is 'b'. if mask doesn't have 'b', it returns false.
    // 'p' keys: "password", "passwd", "private_key", "proxy-authorization"
    // next chars: 'a', 'r'.
    // So 'b' is not allowed.
    input3 := []byte("pbssword")
    assert.False(t, scanForSensitiveKeys(input3, false))
}

func TestIsKeyColon(t *testing.T) {
    assert.True(t, isKeyColon([]byte(`: value`), 0))
    assert.True(t, isKeyColon([]byte(`   : value`), 0))
    assert.False(t, isKeyColon([]byte(` value`), 0))
    assert.False(t, isKeyColon([]byte(``), 0))
}

func TestSkipString(t *testing.T) {
    // input must start with quote
    input := []byte(`"hello" world`)
    end := skipString(input, 0)
    assert.Equal(t, 7, end) // "hello" is 7 chars

    inputEsc := []byte(`"he\"llo" world`)
    endEsc := skipString(inputEsc, 0)
    assert.Equal(t, 9, endEsc)

    inputUnclosed := []byte(`"hello`)
    endUnclosed := skipString(inputUnclosed, 0)
    assert.Equal(t, 6, endUnclosed)
}

func TestMatchFoldRest(t *testing.T) {
    // Test matchFoldRest manually
    // key is lowercase
    key := []byte("password")

    // Exact match
    assert.True(t, matchFoldRest([]byte("password"), key))
    // Upper case match
    assert.True(t, matchFoldRest([]byte("PASSWORD"), key))
    // Mixed case
    assert.True(t, matchFoldRest([]byte("PaSsWoRd"), key))
    // Mismatch
    assert.False(t, matchFoldRest([]byte("passwork"), key))
    // Short input
    assert.False(t, matchFoldRest([]byte("pass"), key))
}
