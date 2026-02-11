// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"fmt"
	"log/slog"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

func BenchmarkRedactString(b *testing.B) {
	// Create 50 custom patterns simulating different secret types
	var patterns []string
	for i := 0; i < 50; i++ {
		patterns = append(patterns, fmt.Sprintf(`secret-val-%d-\w+`, i))
	}

	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled:        &enabled,
		CustomPatterns: patterns,
	}.Build()

	r := NewRedactor(cfg, slog.Default())

	// Create input string with some matches
	// We make it reasonably long to simulate a JSON payload or similar
	input := "This is a normal string with some secrets: "
	for i := 0; i < 5; i++ {
		// Insert matches for patterns 0, 10, 20, 30, 40
		input += fmt.Sprintf("secret-val-%d-abc123xyz ", i*10)
	}
	// Add padding
	for i := 0; i < 100; i++ {
		input += "safe-data-block "
	}
	input += "and some more text at the end."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.RedactString(input)
	}
}
