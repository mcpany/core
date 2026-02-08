// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"strings"
	"testing"
)

func BenchmarkPaginateRecursive(b *testing.B) {
	// Create a large string (approx 1MB)
	largeString := strings.Repeat("This is a test string to simulate large payload content. ", 18000)

	data := map[string]any{
		"a": largeString,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data["a"] = largeString
		// Paginate last page (approx). Total chars ~ 1,008,000. Page size 1000.
		// Page 1000 -> start = 999,000.
		paginateRecursive(data, 1000, 1000)
	}
}
