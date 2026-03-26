// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

func BenchmarkRedactString_ManyPatterns(b *testing.B) {
	// Create 50 custom patterns simulating different secrets
	patterns := make([]string, 50)
	for i := 0; i < 50; i++ {
		// Pattern matches "secret-N-<digits>"
		patterns[i] = fmt.Sprintf(`secret-%d-\d+`, i)
	}

	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled:        &enabled,
		CustomPatterns: patterns,
	}.Build()

	// Suppress logging during benchmark
	logger := slog.New(slog.DiscardHandler)
	r := NewRedactor(cfg, logger)

	// Create a payload (approx 4KB) with distributed matches
	var sb strings.Builder
	sb.WriteString(`{"description": "This is a large payload with sensitive data interspersed throughout. "`)
	for i := 0; i < 100; i++ {
		// Inject a secret that matches one of the patterns
		secret := fmt.Sprintf("secret-%d-%d", i%50, 1000+i)
		sb.WriteString(fmt.Sprintf(`, "field_%d": "Contains sensitive info: %s which should be redacted"`, i, secret))
	}
	sb.WriteString(`}`)
	input := sb.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.RedactString(input)
	}
}
