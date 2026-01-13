// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
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
