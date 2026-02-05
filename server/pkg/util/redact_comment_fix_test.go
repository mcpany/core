// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_NumberWithComment(t *testing.T) {
	// "secret" is a sensitive key in redact.go
	input := []byte(`{
		"secret": 123// this is a comment
	}`)
	expected := []byte(`{
		"secret": "[REDACTED]"// this is a comment
	}`)

	output := RedactJSON(input)

	// Normalize newlines/formatting if necessary, but here we expect exact replacement of value
	assert.Equal(t, string(expected), string(output))
}

func TestRedactJSON_NumberWithBlockComment(t *testing.T) {
	input := []byte(`{
		"secret": 123/* block comment */
	}`)
	expected := []byte(`{
		"secret": "[REDACTED]"/* block comment */
	}`)

	output := RedactJSON(input)
	assert.Equal(t, string(expected), string(output))
}
