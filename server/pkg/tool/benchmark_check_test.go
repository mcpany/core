// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"strings"
	"testing"
)

func BenchmarkCheckUnquotedKeywords(b *testing.B) {
	// Create a large input string with many safe words
	input := strings.Repeat("hello world safe word test process run execution ", 1000)
	keywords := []string{"system", "exec", "popen", "eval"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := checkUnquotedKeywords(input, keywords); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCheckContextualKeywords(b *testing.B) {
	// Create a large input string with many safe words and context delimiters
	input := strings.Repeat("hello.world safe[0] test(arg) process.run execution ", 1000)
	keywords := []string{"system", "exec", "popen", "eval"}
	delimiters := []rune{'(', '=', ':', ' ', '\'', '"'}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := checkContextualKeywords(input, keywords, delimiters); err != nil {
			b.Fatal(err)
		}
	}
}
