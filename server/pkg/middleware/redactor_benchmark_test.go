// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// BenchmarkRedactJSON compares the current implementation with an optimized one.
func BenchmarkRedactJSON(b *testing.B) {
	raw := []byte(`{"status": "ok", "message": "everything is fine", "data": {"id": "123", "value": "some safe string"}}`)
	enabled := true
	config := &configv1.DLPConfig{
		Enabled: &enabled,
	}
	r := NewRedactor(config, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.RedactJSON(raw)
	}
}

func BenchmarkRedactJSON_WithPII(b *testing.B) {
	raw := []byte(`{"email": "test@example.com", "id": "123"}`)
	enabled := true
	config := &configv1.DLPConfig{
		Enabled: &enabled,
	}
	r := NewRedactor(config, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.RedactJSON(raw)
	}
}
