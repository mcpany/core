// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactJSON_Bug_CommentPrecededBySlash(t *testing.T) {
	// The bug: If a non-comment slash appears before a comment, the comment is not detected,
	// and content inside the comment (like quotes) is processed as JSON.

	// Case 1: "password" inside a comment should NOT be redacted.
	// But because of the preceding '/', the comment check fails, and "password" is seen as a key.
	input := []byte(`{
		"a": 10 / 2, // "password": "secret"
		"b": "value"
	}`)

	// Expected: No change, because "password" is in a comment.
	expected := []byte(`{
		"a": 10 / 2, // "password": "secret"
		"b": "value"
	}`)

	result := RedactJSON(input)
	assert.Equal(t, string(expected), string(result))
}

func TestRedactJSON_Bug_BlockCommentPrecededBySlash(t *testing.T) {
	input := []byte(`{
		"a": 10 / 2, /* "password": "secret" */
		"b": "value"
	}`)

	expected := []byte(`{
		"a": 10 / 2, /* "password": "secret" */
		"b": "value"
	}`)

	result := RedactJSON(input)
	assert.Equal(t, string(expected), string(result))
}

func TestRedactJSON_Bug_MultipleSlashes(t *testing.T) {
	input := []byte(`{
		"a": 10 / 2 / 5, // "password": "secret"
		"b": "value"
	}`)

	expected := []byte(`{
		"a": 10 / 2 / 5, // "password": "secret"
		"b": "value"
	}`)

	result := RedactJSON(input)
	assert.Equal(t, string(expected), string(result))
}

func TestRedactJSON_Bug_SlashAtEnd(t *testing.T) {
	// Slash just before the quote starts (conceptually)
	// Input: {"a": /, "password": "secret"}
	// It's invalid JSON, but we check if "password" is still redacted.
	// Since / is not a comment start (unless followed by / or *), "password" should be detected as key.
	input := []byte(`{ "a": /, "password": "secret" }`)
	expected := []byte(`{ "a": /, "password": "[REDACTED]" }`)

	result := RedactJSON(input)
	assert.Equal(t, string(expected), string(result))
}

func TestRedactJSON_Bug_CommentWithSlashes(t *testing.T) {
	input := []byte(`{
		"a": 1, // comment / with / slashes "password": "secret"
		"b": "value"
	}`)
	expected := input // No change

	result := RedactJSON(input)
	assert.Equal(t, string(expected), string(result))
}

func TestRedactJSON_Bug_SlashFollowedByQuote(t *testing.T) {
	// {"key": /"password": "secret"}
	// Slash is followed by quote. Not a comment.
	input := []byte(`{ "key": /"password": "secret" }`)
	expected := []byte(`{ "key": /"password": "[REDACTED]" }`)

	result := RedactJSON(input)
	assert.Equal(t, string(expected), string(result))
}
