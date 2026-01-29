package util //nolint:revive

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

func BenchmarkReplaceURLPath(b *testing.B) {
	path := "/api/v1/users/{{userId}}/posts/{{postId}}"
	params := map[string]interface{}{
		"userId": "123",
		"postId": "456",
	}

	b.Run("Standard", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			ReplaceURLPath(path, params, nil)
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
