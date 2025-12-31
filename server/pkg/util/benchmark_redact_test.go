// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive

import (
	"encoding/json"
	"testing"
)

func BenchmarkRedactJSON(b *testing.B) {
	// Small clean JSON
	cleanSmall := `{"name": "test", "value": 123, "description": "just a test"}`

	// Large clean JSON
	cleanLargeMap := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		cleanLargeMap["key"+string(rune(i))] = "value" + string(rune(i))
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
			dirtyLargeMap["key"+string(rune(i))] = "value" + string(rune(i))
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
	// Large string value benchmark
	// This simulates a large log entry or similar that might be passed through
	// where one field is very large but not an object/array.
	// We want to see if we waste time scanning it.
	largeStringMap := map[string]string{
		"short": "val",
		// 1MB string
		"large": string(make([]byte, 1024*1024)),
	}
	// Make it "dirty" so we actually parse the map
	largeStringMap["api_key"] = "secret"

	largeStringBytes, _ := json.Marshal(largeStringMap)

	b.Run("LargeStringValue", func(b *testing.B) {
		input := largeStringBytes
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			RedactJSON(input)
		}
	})
}
