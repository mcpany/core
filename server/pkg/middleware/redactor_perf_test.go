// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"fmt"
	"log/slog"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

func BenchmarkRedactor_RedactString(b *testing.B) {
	// Create a config with 50 custom patterns
	var patterns []string
	for i := 0; i < 50; i++ {
		patterns = append(patterns, fmt.Sprintf("secret-%d", i))
	}
	// Add some more complex patterns
	patterns = append(patterns, `api_key_[a-zA-Z0-9]{32}`)
	patterns = append(patterns, `token:[a-f0-9]{64}`)

	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled:        &enabled,
		CustomPatterns: patterns,
	}.Build()

	r := NewRedactor(cfg, slog.Default())

	input := `
		Here is some text with secrets.
		secret-10 is here.
		secret-42 is also here.
		And an api_key_abcdef1234567890abcdef1234567890 over there.
		But token:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef too.
		Just some random text to make it longer.
		Repeating: secret-5, secret-20, secret-33.
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.RedactString(input)
	}
}
