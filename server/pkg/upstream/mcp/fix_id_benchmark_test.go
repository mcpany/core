// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"
)

// Define a struct that matches the "broken struct" behavior
// The original code says: // Expect {value:1}
// And it prints with %+v
type brokenID struct {
	value int // unexported
}

// Current implementation logic copy for benchmarking the "before" state
// We can't easily benchmark the private function in the same package without modifying it first if we want to compare "Old" vs "New" in the same file
// unless we just modify the file and run benchmark before and after.
// But to be scientific, I'll include the "Old" implementation here.

func fixIDOld(id interface{}) interface{} {
	if id == nil {
		return nil
	}
	// Check if it's already simple type
	switch v := id.(type) {
	case string, int, int64, float64:
		return v
	}

	// If it's the broken struct, print it and parse
	s := fmt.Sprintf("%+v", id)
	// Expect {value:1}
	re := regexp.MustCompile(`value:(\d+)`)
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		if i, err := strconv.Atoi(matches[1]); err == nil {
			return i
		}
	}
	reStr := regexp.MustCompile(`value:([^}]+)`)
	matchesStr := reStr.FindStringSubmatch(s)
	if len(matchesStr) > 1 {
		return matchesStr[1]
	}

	return id
}

// Ensure fixIDOld behaves as expected
func TestFixIDOld(t *testing.T) {
	id := brokenID{value: 123}
	res := fixIDOld(id)
	if res != 123 {
		t.Errorf("Expected 123, got %v", res)
	}
}

func BenchmarkFixID_RegexPath(b *testing.B) {
	id := brokenID{value: 123}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fixIDOld(id)
	}
}

// Benchmark the optimized version (simulated)
var (
	benchReInt = regexp.MustCompile(`value:(\d+)`)
	benchReStr = regexp.MustCompile(`value:([^}]+)`)
)

func fixIDOptimized(id interface{}) interface{} {
	if id == nil {
		return nil
	}
	// Check if it's already simple type
	switch v := id.(type) {
	case string, int, int64, float64:
		return v
	}

	// If it's the broken struct, print it and parse
	s := fmt.Sprintf("%+v", id)
	// Expect {value:1}
	matches := benchReInt.FindStringSubmatch(s)
	if len(matches) > 1 {
		if i, err := strconv.Atoi(matches[1]); err == nil {
			return i
		}
	}
	matchesStr := benchReStr.FindStringSubmatch(s)
	if len(matchesStr) > 1 {
		return matchesStr[1]
	}

	return id
}

func BenchmarkFixID_Optimized(b *testing.B) {
	id := brokenID{value: 123}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fixIDOptimized(id)
	}
}
