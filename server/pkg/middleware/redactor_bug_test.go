// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestRedactor_Bug_CommentParsingBypass(t *testing.T) {
	// Security Bug: If WalkJSONStrings fails to detect a comment because of a preceding slash (e.g. division),
	// and that comment contains a quote, it might consume subsequent real values as part of a fake string.
	// This would cause the Redactor to MISS sensitive data in those real values.

	enabled := true
	cfg := &configv1.DLPConfig{
		Enabled: &enabled,
	}
	r := NewRedactor(cfg, nil)

	// Input: {"calc": 10 / 2 /* " */, "email": "sensitive@example.com"}
	// If bug exists:
	// 1. It sees 10 / 2.
	// 2. It misses the comment start /* because of the preceding /.
	// 3. It thinks " inside the comment is the start of a string.
	// 4. It scans until the next quote, which is at the end of "sensitive@example.com".
	// 5. It treats ` */, "email": "sensitive@example.com` as the string content.
	// 6. It does NOT see "sensitive@example.com" as a value to be redacted.
	// 7. Result: "sensitive@example.com" remains in the output.

	input := []byte(`{"calc": 10 / 2 /* " */, "email": "sensitive@example.com"}`)

	redacted, err := r.RedactJSON(input)
	assert.NoError(t, err)

	output := string(redacted)

	// We expect the email to be redacted.
	assert.NotContains(t, output, "sensitive@example.com", "PII should be redacted")
	assert.Contains(t, output, "***REDACTED***", "Should contain redaction marker")
}
