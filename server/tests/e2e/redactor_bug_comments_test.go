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

func TestRedactor_Bug_CommentCorruption(t *testing.T) {
	// A comment that contains a "string" which, if processed, would look like a PII.
	// But it is inside a comment, so it should be ignored.
	// More importantly, we want to ensure the comment structure is not destroyed.

	// Input with a string inside a block comment that looks like a string.
	// "email@example.com" triggers PII redaction if treated as a string.
	input := `{"key": "value" /* "email@example.com" */}`

	cfg := &configv1.DLPConfig{
		Enabled: proto.Bool(true),
	}

	r := middleware.NewRedactor(cfg, nil)

	redacted, err := r.RedactJSON([]byte(input))
	assert.NoError(t, err)

	// Expected: input is unchanged.
	// If bug exists: "email@example.com" is redacted to "***REDACTED***".
	assert.Equal(t, input, string(redacted))
}

func TestRedactor_Bug_CommentStructure(t *testing.T) {
	// Input that causes structure corruption if bug exists.
	// /* "oops */ "real string"
	// "oops */ " is treated as string.
	// We use "email@example.com" to trigger redaction.

	// Construct input: /* "email@example.com */ "real"
	input := `/* "email@example.com */ "real"`

	cfg := &configv1.DLPConfig{
		Enabled: proto.Bool(true),
	}

	r := middleware.NewRedactor(cfg, nil)

	redacted, err := r.RedactJSON([]byte(input))
	assert.NoError(t, err)

	// Expected: input is unchanged (comments skipped).
	// If bug exists: /* "***REDACTED***real"
	assert.Equal(t, input, string(redacted))
}
