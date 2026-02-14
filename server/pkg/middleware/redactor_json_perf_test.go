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

func BenchmarkRedactJSON_CustomPattern_SafeStrings(b *testing.B) {
	// Single custom pattern
	patterns := []string{`secret-\d+`}

	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled:        &enabled,
		CustomPatterns: patterns,
	}.Build()

	logger := slog.New(slog.DiscardHandler)
	r := NewRedactor(cfg, logger)

	// Create a payload with 1000 fields, mostly safe strings, no escapes
	var sb strings.Builder
	sb.WriteString(`{`)
	for i := 0; i < 1000; i++ {
		if i > 0 {
			sb.WriteString(`,`)
		}
		// "safe_field_N": "this is a completely safe string without any escapes or special characters"
		sb.WriteString(fmt.Sprintf(`"safe_field_%d": "this is a completely safe string without any escapes or special characters which is quite long to simulate realistic log messages"`, i))
	}
	// Add one secret at the end
	sb.WriteString(`, "secret_field": "secret-12345"`)
	sb.WriteString(`}`)
	input := []byte(sb.String())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.RedactJSON(input)
	}
}

func BenchmarkRedactJSON_NoCustomPattern_SafeStrings(b *testing.B) {
	// No custom patterns
	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled: &enabled,
	}.Build()

	logger := slog.New(slog.DiscardHandler)
	r := NewRedactor(cfg, logger)

	// Same payload
	var sb strings.Builder
	sb.WriteString(`{`)
	for i := 0; i < 1000; i++ {
		if i > 0 {
			sb.WriteString(`,`)
		}
		sb.WriteString(fmt.Sprintf(`"safe_field_%d": "this is a completely safe string without any escapes or special characters which is quite long to simulate realistic log messages"`, i))
	}
	sb.WriteString(`, "secret_field": "secret-12345"`) // Won't be redacted as no custom pattern matches it (unless it matches defaults like CC/SSN/Email which it doesn't)
	sb.WriteString(`}`)
	input := []byte(sb.String())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.RedactJSON(input)
	}
}
