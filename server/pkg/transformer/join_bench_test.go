// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package transformer

import (
	"testing"
)

func BenchmarkJoinStrings(b *testing.B) {
	t := NewTransformer()
	listSize := 1000
	items := make([]string, listSize)
	for i := 0; i < listSize; i++ {
		items[i] = "item"
	}
	templateStr := `{{join "," .items}}`
	data := map[string]any{"items": items}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := t.Transform(templateStr, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJoinInts(b *testing.B) {
	t := NewTransformer()
	listSize := 1000
	items := make([]int, listSize)
	for i := 0; i < listSize; i++ {
		items[i] = i
	}
	templateStr := `{{join "," .items}}`
	data := map[string]any{"items": items}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := t.Transform(templateStr, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
