// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"testing"
)

func BenchmarkUintTokenization(b *testing.B) {
	t := NewSimpleTokenizer()

	// Create a large slice of uints
	uints := make([]uint, 1000)
	for i := 0; i < 1000; i++ {
		uints[i] = uint(i * 12345)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CountTokensInValue(t, uints)
	}
}

func BenchmarkInt32Tokenization(b *testing.B) {
	t := NewSimpleTokenizer()

	ints := make([]int32, 1000)
	for i := 0; i < 1000; i++ {
		ints[i] = int32(i * 1234)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CountTokensInValue(t, ints)
	}
}
