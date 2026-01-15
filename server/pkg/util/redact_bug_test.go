// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_EscapedKey(t *testing.T) {
	// "auth" is a sensitive key
	// "au\u0074h" is "auth" escaped
	input := []byte(`{"au\u0074h": "sensitive_value"}`)
	expected := []byte(`{"au\u0074h": "[REDACTED]"}`)

	result := RedactJSON(input)
	assert.Equal(t, string(expected), string(result))
}

func TestRedactJSON_EscapedKey_Complex(t *testing.T) {
	// "password" is sensitive
	// "pass\u0077ord"
	input := []byte(`{"pass\u0077ord": "sensitive_value_123"}`)
	expected := []byte(`{"pass\u0077ord": "[REDACTED]"}`)

	result := RedactJSON(input)
	assert.Equal(t, string(expected), string(result))
}

func TestRedactJSON_LargeKeyWithEscapes(t *testing.T) {
	// Create a key larger than 1024 bytes
	// "password" escaped as "p\u0061ssword"
	padding := strings.Repeat("a", 1100)
	escapedPassword := `p\u0061ssword`
	key := padding + escapedPassword
	input := []byte(`{"` + key + `": "secret_value"}`)

	// Verify the key length is > 1024
	if len(key) <= 1024 {
		t.Fatalf("Key length %d is not > 1024", len(key))
	}

	output := RedactJSON(input)

	// Parse output to check value
	var m map[string]interface{}
	// We need to unmarshal to see if value was redacted.
	// Note: unmarshaling the key will decode \u0061 to a.
	err := json.Unmarshal(output, &m)
	assert.NoError(t, err)

	// The key in the map will have 'a' instead of \u0061
	expectedKey := padding + "password"
	val, ok := m[expectedKey]
	if !ok {
		t.Fatalf("Key not found in output")
	}

	assert.Equal(t, "[REDACTED]", val, "Large key with escapes should be redacted")
}

func TestRedactJSON_HugeKeyWithEscapes(t *testing.T) {
	// Create a key larger than 1MB (1024*1024 bytes)
	// This tests the unescape path logic (limit is 2MB)
	// 1MB + padding
	targetLen := 1024*1024 + 100
	escapedPassword := `p\u0061ssword`
	paddingLen := targetLen - len(escapedPassword)
	padding := strings.Repeat("a", paddingLen)

	key := padding + escapedPassword
	// JSON: {"key": "secret"}
	input := []byte(`{"` + key + `": "secret_value"}`)

	if len(key) <= 1024*1024 {
		t.Fatalf("Key length %d is not > 1MB", len(key))
	}

	output := RedactJSON(input)

	// Parse output to check value
	var m map[string]interface{}
	err := json.Unmarshal(output, &m)
	assert.NoError(t, err)

	expectedKey := padding + "password"
	val, ok := m[expectedKey]
	if !ok {
		t.Fatalf("Key not found in output")
	}

	assert.Equal(t, "[REDACTED]", val, "Huge key (1MB) with escapes should be redacted")
}

func TestRedactJSON_SuperHugeKeyWithEscapes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping super huge key test in short mode")
	}
	// Create a key larger than 2MB (limit)
	// This tests the fail-safe logic for extremely large keys
	targetLen := 2*1024*1024 + 100
	escapedPassword := `p\u0061ssword`
	paddingLen := targetLen - len(escapedPassword)
	padding := strings.Repeat("a", paddingLen)

	key := padding + escapedPassword
	input := []byte(`{"` + key + `": "secret_value"}`)

	output := RedactJSON(input)

	// Direct byte check to avoid expensive Unmarshal
	if !bytes.Contains(output, []byte(`"[REDACTED]"`)) {
		t.Errorf("Output does not contain redacted placeholder")
	}
	if bytes.Contains(output, []byte(`"secret_value"`)) {
		t.Errorf("Output contains secret value")
	}
}
