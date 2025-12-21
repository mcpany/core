// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"strings"
	"testing"
)

func BenchmarkRedactJSON(b *testing.B) {
	// Small clean JSON
	cleanSmall := `{"name": "test", "value": 123, "description": "just a test"}`

	// Large clean JSON
	cleanLargeMap := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		cleanLargeMap["key"+string(rune(i))] = "value" + string(rune(i))
	}
	cleanLargeBytes, _ := json.Marshal(cleanLargeMap)
	cleanLarge := string(cleanLargeBytes)

	// Small dirty JSON
	dirtySmall := `{"name": "test", "api_key": "secret", "description": "just a test"}`

	// Large dirty JSON
	dirtyLargeMap := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		if i == 50 {
			dirtyLargeMap["api_key"] = "secret"
		} else {
			dirtyLargeMap["key"+string(rune(i))] = "value" + string(rune(i))
		}
	}
	dirtyLargeBytes, _ := json.Marshal(dirtyLargeMap)
	dirtyLarge := string(dirtyLargeBytes)

	b.Run("CleanSmall", func(b *testing.B) {
		input := []byte(cleanSmall)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			RedactJSON(input)
		}
	})

	b.Run("CleanLarge", func(b *testing.B) {
		input := []byte(cleanLarge)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			RedactJSON(input)
		}
	})

	b.Run("DirtySmall", func(b *testing.B) {
		input := []byte(dirtySmall)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			RedactJSON(input)
		}
	})

	b.Run("DirtyLarge", func(b *testing.B) {
		input := []byte(dirtyLarge)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			RedactJSON(input)
		}
	})

	b.Run("CleanWithEscapes", func(b *testing.B) {
		input := []byte(`{"name": "test", "path": "C:\\Windows\\System32"}`)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			RedactJSON(input)
		}
	})
}

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

func BenchmarkSanitizeOperationID(b *testing.B) {
	b.Run("Clean", func(b *testing.B) {
		input := "validOperationID-123_test"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			SanitizeOperationID(input)
		}
	})

	b.Run("Dirty", func(b *testing.B) {
		input := "invalid op ID with spaces & symbols!"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			SanitizeOperationID(input)
		}
	})

	b.Run("Mixed", func(b *testing.B) {
		// A mix of valid and invalid characters, multiple segments
		input := "op" + strings.Repeat(" ", 5) + "ID" + strings.Repeat("!", 5) + "123"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			SanitizeOperationID(input)
		}
	})
}
