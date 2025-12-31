// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive

import (
	"bytes"
	"encoding/json"
	"testing"
)

func BenchmarkRedactJSON(b *testing.B) {
	// Small clean JSON
	cleanSmall := `{"name": "test", "value": 123, "description": "just a test"}`

	// Large clean JSON
	cleanLargeMap := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		cleanLargeMap["key" + string(rune(i))] = "value" + string(rune(i))
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
			dirtyLargeMap["key" + string(rune(i))] = "value" + string(rune(i))
		}
	}
	dirtyLargeBytes, _ := json.Marshal(dirtyLargeMap)
	dirtyLarge := string(dirtyLargeBytes)

	// Array benchmarks
	cleanArray := `[{"name": "test", "value": 123}, {"name": "test2", "value": 456}]`
	// Dirty Array - actually contains sensitive key in one of the objects
	dirtyArray := `[{"name": "test", "api_key": "secret"}, {"name": "test2", "value": 456}]`

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

	b.Run("CleanArray", func(b *testing.B) {
		input := []byte(cleanArray)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			RedactJSON(input)
		}
	})

	b.Run("DirtyArray", func(b *testing.B) {
		input := []byte(dirtyArray)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			RedactJSON(input)
		}
	})
}

func BenchmarkBytesContainsFold(b *testing.B) {
	// Quadratic case: frequent matches for first char, but no second char match.
	// Input: "a a a a ..." (50KB)
	size := 50000
	inputQuad := bytes.Repeat([]byte("a "), size)
	// Key: "abc" (starts with 'a')
	key := []byte("abc")

	b.Run("Quadratic", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bytesContainsFold(inputQuad, key)
		}
	})

	// Linear case: normal text, match at end.
	inputLin := bytes.Repeat([]byte(" "), size)
	inputLin = append(inputLin, []byte("abc")...)

	b.Run("Linear", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			bytesContainsFold(inputLin, key)
		}
	})
}
