// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"testing"
)

func BenchmarkSanitizeID(b *testing.B) {
	// Common case: valid ID, no sanitization needed, no hash appended
	b.Run("ValidID_NoHash", func(b *testing.B) {
		ids := []string{"valid_service_name", "valid_tool_name"}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			SanitizeID(ids, false, 53, 8)
		}
	})

	// Case needing sanitization but not length truncation
	b.Run("DirtyID_NoHash", func(b *testing.B) {
		ids := []string{"invalid service name", "invalid/tool/name"}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			SanitizeID(ids, false, 53, 8)
		}
	})

	// Case needing truncation (and thus hash)
	b.Run("LongID_NeedsHash", func(b *testing.B) {
		longID := strings.Repeat("a", 100)
		ids := []string{longID}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			SanitizeID(ids, false, 53, 8)
		}
	})

	// Case explicitly requesting hash
	b.Run("ValidID_WithHash", func(b *testing.B) {
		ids := []string{"valid_service_name"}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			SanitizeID(ids, true, 53, 8)
		}
	})
}
