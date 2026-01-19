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

func TestRedactor_Bug_CommentInKey(t *testing.T) {
	// A credit card number as a key (unlikely but possible).
	// If the redactor bugs out on comments, it will treat this KEY as a VALUE and redact it.
	cc := "1234-5678-9012-3456"
	input := `{"` + cc + `" /* comment */ : "value"}`

	// Config enabling default redactors
	cfg := &configv1.DLPConfig{
		Enabled: proto.Bool(true),
	}

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
