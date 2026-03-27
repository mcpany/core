// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_InvalidJSON_NumberConsumingQuotes(t *testing.T) {
	// Case 1: Number immediately followed by a quote starting a new key (missing comma)
	// Input: {"password": 123"token": "value"}
	// "password" is sensitive. Value is 123.
	// "token" is sensitive.
	input := []byte(`{"password": 123"token": "value"}`)
	// Redacted value is "[REDACTED]" (with quotes).
	// So {"password": "[REDACTED]""token": "[REDACTED]"}
	expected := []byte(`{"password": "[REDACTED]""token": "[REDACTED]"}`)

	actual := RedactJSON(input)
	assert.Equal(t, string(expected), string(actual), "Should not swallow subsequent keys when redacting a number followed by a quote")
}

func TestRedactJSON_InvalidJSON_NumberConsumingColon(t *testing.T) {
    // Case 2: Number followed by colon (rare but possible in broken JSON)
    // Input: {"password": 123:"token": "value"}
    input := []byte(`{"password": 123:"token": "value"}`)
    // Redacted value is "[REDACTED]" (with quotes).
    // So {"password": "[REDACTED]":"token": "[REDACTED]"}
    expected := []byte(`{"password": "[REDACTED]":"token": "[REDACTED]"}`)

    actual := RedactJSON(input)
    assert.Equal(t, string(expected), string(actual), "Should not swallow subsequent keys when redacting a number followed by a colon")
}
