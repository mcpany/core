// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"log/slog"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// BenchmarkRedactor_RedactJSON ...
// Summary: BenchmarkRedactor_RedactJSON
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled:        &enabled,
		CustomPatterns: []string{`secret-\d+`},
	}.Build()
	r := NewRedactor(cfg, slog.Default())

	input := []byte(`{
		"email": "user@example.com",
		"name": "John Doe",
		"id": 12345,
		"details": {
			"address": "123 Main St",
			"notes": "This is a secret-123 message",
			"phone": "555-0199"
		},
		"history": [
			{"action": "login", "timestamp": "2023-01-01T12:00:00Z"},
			{"action": "view", "item": "secret-999"}
		]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.RedactJSON(input)
	}
}

// BenchmarkRedactor_RedactJSON_NoCustomPatterns ...
// Summary: BenchmarkRedactor_RedactJSON_NoCustomPatterns
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled: &enabled,
		// No custom patterns
	}.Build()
	r := NewRedactor(cfg, slog.Default())

	input := []byte(`{
		"email": "user@example.com",
		"name": "John Doe",
		"id": 12345,
		"details": {
			"address": "123 Main St",
			"notes": "This is a secret-123 message",
			"phone": "555-0199"
		},
		"history": [
			{"action": "login", "timestamp": "2023-01-01T12:00:00Z"},
			{"action": "view", "item": "secret-999"}
		]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.RedactJSON(input)
	}
}

// BenchmarkRedactor_RedactJSON_NoCustomPatterns_Large ...
// Summary: BenchmarkRedactor_RedactJSON_NoCustomPatterns_Large
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled: &enabled,
	}.Build()
	r := NewRedactor(cfg, slog.Default())

	// Construct a larger JSON payload
	input := []byte(`{
		"users": [
			{"id": 1, "email": "user1@example.com", "data": "secret-1"},
			{"id": 2, "email": "user2@example.com", "data": "secret-2"},
			{"id": 3, "email": "user3@example.com", "data": "secret-3"},
			{"id": 4, "email": "user4@example.com", "data": "secret-4"},
			{"id": 5, "email": "user5@example.com", "data": "secret-5"},
			{"id": 6, "email": "user6@example.com", "data": "secret-6"},
			{"id": 7, "email": "user7@example.com", "data": "secret-7"},
			{"id": 8, "email": "user8@example.com", "data": "secret-8"},
			{"id": 9, "email": "user9@example.com", "data": "secret-9"},
			{"id": 10, "email": "user10@example.com", "data": "secret-10"}
		]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.RedactJSON(input)
	}
}

// BenchmarkRedactor_RedactJSON_AllSafe ...
// Summary: BenchmarkRedactor_RedactJSON_AllSafe
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled: &enabled,
	}.Build()
	r := NewRedactor(cfg, slog.Default())

	input := []byte(`{
		"status": "ok",
		"type": "message",
		"content": "some safe content",
		"details": {
			"key": "value",
			"description": "just some text"
		}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.RedactJSON(input)
	}
}

// BenchmarkRedactor_RedactJSON_Large ...
// Summary: BenchmarkRedactor_RedactJSON_Large
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled:        &enabled,
		CustomPatterns: []string{`secret-\d+`},
	}.Build()
	r := NewRedactor(cfg, slog.Default())

	// Construct a larger JSON payload
	input := []byte(`{
		"users": [
			{"id": 1, "email": "user1@example.com", "data": "secret-1"},
			{"id": 2, "email": "user2@example.com", "data": "secret-2"},
			{"id": 3, "email": "user3@example.com", "data": "secret-3"},
			{"id": 4, "email": "user4@example.com", "data": "secret-4"},
			{"id": 5, "email": "user5@example.com", "data": "secret-5"},
			{"id": 6, "email": "user6@example.com", "data": "secret-6"},
			{"id": 7, "email": "user7@example.com", "data": "secret-7"},
			{"id": 8, "email": "user8@example.com", "data": "secret-8"},
			{"id": 9, "email": "user9@example.com", "data": "secret-9"},
			{"id": 10, "email": "user10@example.com", "data": "secret-10"}
		]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.RedactJSON(input)
	}
}
