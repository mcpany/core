// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

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

	// Small dirty JSON Array
	dirtySmallArray := `[{"name": "test"}, {"api_key": "secret"}, {"description": "just a test"}]`

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

	b.Run("DirtySmallArray", func(b *testing.B) {
		input := []byte(dirtySmallArray)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			RedactJSON(input)
		}
	})
}
