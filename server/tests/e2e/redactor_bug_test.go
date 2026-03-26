// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// TestRedactor_Bug_CommentInKey ...
// Summary: TestRedactor_Bug_CommentInKey
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// A credit card number as a key (unlikely but possible).
	// If the redactor bugs out on comments, it will treat this KEY as a VALUE and redact it.
	cc := "1234-5678-9012-3456"
	input := `{"` + cc + `" /* comment */ : "value"}`

	// Config enabling default redactors
	cfg := configv1.DLPConfig_builder{
		Enabled: proto.Bool(true),
	}.Build()

	r := middleware.NewRedactor(cfg, nil)

	redacted, err := r.RedactJSON([]byte(input))
	assert.NoError(t, err)

	// We expect the key to remain UNCHANGED.
	// If the bug exists, it will be redacted.
	expected := input // Should match input exactly because "value" is safe.
	assert.Equal(t, expected, string(redacted))

	// Ensure the key is NOT redacted
	assert.Contains(t, string(redacted), cc)
}

// TestRedactor_Bug_PlainText ...
// Summary: TestRedactor_Bug_PlainText
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// This previously triggered a bug where RedactJSON would modify plain text
	// if it contained quotes and colons.
	input := `This is plain text with "token": "mysecret" embedded.`

	// Config enabling default redactors
	cfg := configv1.DLPConfig_builder{
		Enabled: proto.Bool(true),
	}.Build()

	r := middleware.NewRedactor(cfg, nil)

	redacted, err := r.RedactJSON([]byte(input))
	assert.NoError(t, err)

	// Should verify it is NOT redacted
	assert.Equal(t, input, string(redacted))
}

// TestRedactor_Bug_StringInComment_CorruptsStructure ...
// Summary: TestRedactor_Bug_StringInComment_CorruptsStructure
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// PII in a comment, but the comment contains an unclosed quote from the walker's perspective.
	// Walker sees: "user@example.com */ "
	// It replaces it with "***REDACTED***"
	// Result: {"key": /* "***REDACTED***"value"}
	// The comment closer */ is eaten.
	input := `{"key": /* "user@example.com */ "value"}`

	cfg := configv1.DLPConfig_builder{
		Enabled: proto.Bool(true),
	}.Build()

	r := middleware.NewRedactor(cfg, nil)

	redacted, err := r.RedactJSON([]byte(input))
	assert.NoError(t, err)

	expected := input
	assert.Equal(t, expected, string(redacted))
}
