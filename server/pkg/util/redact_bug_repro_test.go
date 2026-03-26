// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRedactJSON_Bug_CommentPrecededBySlash ...
// Summary: TestRedactJSON_Bug_CommentPrecededBySlash
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestRedactJSON_Bug_BlockCommentPrecededBySlash ...
// Summary: TestRedactJSON_Bug_BlockCommentPrecededBySlash
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestRedactJSON_MultipleSlashesBeforeComment ...
// Summary: TestRedactJSON_MultipleSlashesBeforeComment
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestRedactJSON_ComplexCommentContent ...
// Summary: TestRedactJSON_ComplexCommentContent
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestRedactJSON_SlashAtEndOfSegment ...
// Summary: TestRedactJSON_SlashAtEndOfSegment
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestRedactJSON_MixedComments ...
// Summary: TestRedactJSON_MixedComments
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
