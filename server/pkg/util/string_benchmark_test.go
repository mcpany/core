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

	// Mixed case: one string is non-ASCII, one is ASCII.
	// This tests the optimization where we avoid converting the ASCII string to []rune.
	mixed1 := "weäther"
	mixed2 := "weather"
	b.Run("Mixed", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			LevenshteinDistance(mixed1, mixed2)
		}
	})

	// Non-ASCII case: both strings are non-ASCII.
	// This tests the stack allocation optimization for small non-ASCII strings.
	nonAscii1 := "こんにちは"
	nonAscii2 := "こんばんは"
	b.Run("NonASCII", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			LevenshteinDistance(nonAscii1, nonAscii2)
		}
	})
}
