// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import "testing"

func BenchmarkIsSensitiveKey(b *testing.B) {
	keys := []string{
		"name", "id", "description", "created_at", "updated_at",
		"api_key", "password", "auth_token",
		"not_sensitive", "user_id", "email",
		"very_long_key_name_that_should_not_trigger_anything_but_takes_time_to_scan",
		"AuthToken", "AUTHORITY", "authToken",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, k := range keys {
			IsSensitiveKey(k)
		}
	}
}
