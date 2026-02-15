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

func TestRedactJSON_MultipleSlashesBeforeComment(t *testing.T) {
	input := []byte(`{
		"a": 1 / 2 / 3, // "password": "secret"
		"b": "value"
	}`)
	expected := []byte(`{
		"a": 1 / 2 / 3, // "password": "secret"
		"b": "value"
	}`)
	result := RedactJSON(input)
	assert.Equal(t, string(expected), string(result))
}

func TestRedactJSON_ComplexCommentContent(t *testing.T) {
	input := []byte(`{
		"a": 1, // comment with "quotes" and / slashes and * stars
		"password": "real_secret"
	}`)
	expected := []byte(`{
		"a": 1, // comment with "quotes" and / slashes and * stars
		"password": "[REDACTED]"
	}`)
	result := RedactJSON(input)
	assert.Equal(t, string(expected), string(result))
}

func TestRedactJSON_SlashAtEndOfSegment(t *testing.T) {
	// Invalid JSON ending with slash, but followed by valid key.
	// The parser splits by quotes.
	// Segment between "a": 1 and "b": 2 contains /.
	input := []byte(`{
		"a": 1 /,
		"b": 2
	}`)
	expected := []byte(`{
		"a": 1 /,
		"b": 2
	}`)
	result := RedactJSON(input)
	assert.Equal(t, string(expected), string(result))
}

func TestRedactJSON_MixedComments(t *testing.T) {
	input := []byte(`{
		"a": 1 /* block */ / 2 // line
		, "password": "secret"
	}`)
	expected := []byte(`{
		"a": 1 /* block */ / 2 // line
		, "password": "[REDACTED]"
	}`)
	result := RedactJSON(input)
	assert.Equal(t, string(expected), string(result))
}
