// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"testing"
)

func BenchmarkLevenshteinDistance(b *testing.B) {
	s1 := "kittens"
	s2 := "sitting"
	b.Run("Short", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			LevenshteinDistance(s1, s2)
		}
	})

	long1 := strings.Repeat("a", 100)
	long2 := strings.Repeat("b", 100)
	b.Run("Medium", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			LevenshteinDistance(long1, long2)
		}
	})

	veryLong1 := strings.Repeat("a", 1000)
	veryLong2 := strings.Repeat("b", 1000)
	b.Run("Long", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			LevenshteinDistance(veryLong1, veryLong2)
		}
	})
}
