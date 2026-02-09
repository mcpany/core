// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tokenizer

import (
	"math"
	"testing"
)

// TestSimpleTokenizeInt64_Comprehensive exercises all branches of simpleTokenizeInt64.
// It covers positive and negative integers of all magnitudes.
func TestSimpleTokenizeInt64_Comprehensive(t *testing.T) {
	// SimpleTokenizer uses len(text) / 4.
	// If count < 1, it returns 1.

	tests := []struct {
		val  int64
		want int
	}{
		// Small integers (< 7 digits + sign) -> 1 token
		{0, 1},
		{1, 1},
		{9, 1},
		{10, 1},
		{99, 1},
		{100, 1},
		{999, 1},
		{1000, 1},
		{9999, 1},
		{10000, 1},
		{99999, 1},
		{100000, 1},
		{999999, 1},
		{-1, 1},
		{-9, 1},
		{-10, 1},
		{-99, 1},
		{-100, 1},
		{-999, 1},
		{-1000, 1},
		{-9999, 1},
		{-10000, 1},
		{-99999, 1},
		{-999999, 1},

		// Larger integers - checking magnitudes
		// 10^6 (7 digits) -> 7/4 = 1.75 -> 1
		{1000000, 1},
		{-1000000, 2}, // -1000000 (8 chars) -> 8/4 = 2

		// 10^7 (8 digits) -> 2 tokens
		{10000000, 2},
		{-10000000, 2}, // 9 chars -> 2

		// 10^8 (9 digits) -> 2 tokens
		{100000000, 2},
		{-100000000, 2}, // 10 chars -> 2

		// 10^9 (10 digits) -> 2 tokens
		{1000000000, 2},
		{-1000000000, 2}, // 11 chars -> 2

		// 10^10 (11 digits) -> 2 tokens
		{10000000000, 2},
		{-10000000000, 3}, // 12 chars -> 3

		// 10^11 (12 digits) -> 3 tokens
		{100000000000, 3},
		{-100000000000, 3}, // 13 chars -> 3

		// 10^12 (13 digits) -> 3 tokens
		{1000000000000, 3},
		{-1000000000000, 3}, // 14 chars -> 3

		// 10^13 (14 digits) -> 3 tokens
		{10000000000000, 3},
		{-10000000000000, 3}, // 15 chars -> 3

		// 10^14 (15 digits) -> 3 tokens
		{100000000000000, 3},
		{-100000000000000, 4}, // 16 chars -> 4

		// 10^15 (16 digits) -> 4 tokens
		{1000000000000000, 4},
		{-1000000000000000, 4}, // 17 chars -> 4

		// 10^16 (17 digits) -> 4 tokens
		{10000000000000000, 4},
		{-10000000000000000, 4}, // 18 chars -> 4

		// 10^17 (18 digits) -> 4 tokens
		{100000000000000000, 4},
		{-100000000000000000, 4}, // 19 chars -> 4

		// 10^18 (19 digits) -> 4 tokens
		{1000000000000000000, 4},
		{-1000000000000000000, 5}, // 20 chars -> 5

		// Max Int64 (19 digits)
		{math.MaxInt64, 4}, // 9223372036854775807 (19 chars) -> 4 tokens

		// Min Int64
		{math.MinInt64, 5}, // -9223372036854775808 (20 chars) -> 5 tokens
	}

	for _, tt := range tests {
		got := simpleTokenizeInt64(tt.val)
		if got != tt.want {
			t.Errorf("simpleTokenizeInt64(%d) = %d, want %d", tt.val, got, tt.want)
		}
	}
}

// TestCountTokensReflect_Comprehensive exercises reflection logic.
func TestCountTokensReflect_Comprehensive(t *testing.T) {
	tokenizer := NewSimpleTokenizer()

	// 1. Struct with unexported fields
	type MixedVisibility struct {
		Exported   string
		unexported string
		Exported2  int
	}
	s := MixedVisibility{
		Exported:   "abcd", // 1 token
		unexported: "longstringthatwouldchangecount",
		Exported2:  1234, // 1 token
	}
	// Only exported fields should be counted.
	// "abcd"(1) + "1234"(1) = 2
	got, err := CountTokensInValue(tokenizer, s)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got != 2 {
		t.Errorf("MixedVisibility struct: got %d, want 2", got)
	}

	// 2. Map with non-string keys
	m := map[int]string{
		1234: "abcd",
	}
	// Key: 1234 -> 1 token. Value: "abcd" -> 1 token. Total 2.
	got, err = CountTokensInValue(tokenizer, m)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got != 2 {
		t.Errorf("Map with int keys: got %d, want 2", got)
	}

	// 3. Slice of interface{} with mixed types
	slice := []interface{}{
		"abcd", // 1
		1234,   // 1
		true,   // 1 ("true" -> 4 chars / 4 = 1)
		nil,    // 1 ("null" -> 4 chars / 4 = 1)
	}
	// Total 4.
	got, err = CountTokensInValue(tokenizer, slice)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got != 4 {
		t.Errorf("Slice of mixed types: got %d, want 4", got)
	}

	// 4. Nested Pointers
	i := 1234
	p1 := &i
	p2 := &p1
	p3 := &p2
	// Should traverse all pointers to int 1234 -> 1 token
	got, err = CountTokensInValue(tokenizer, p3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got != 1 {
		t.Errorf("Nested pointers: got %d, want 1", got)
	}

	// 5. Struct with Bool field (for optimization coverage)
	type BoolStruct struct {
		Flag bool
	}
	bs := BoolStruct{Flag: true}
	// "true" -> 4 chars -> 1 token
	got, err = CountTokensInValue(tokenizer, bs)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got != 1 {
		t.Errorf("BoolStruct: got %d, want 1", got)
	}

	// 6. Map with Bool values (for reflection optimization coverage)
	// Key: 1 (int) -> 1 token. Value: true (bool) -> 1 token. Total 2.
	mb := map[int]bool{1: true}
	got, err = CountTokensInValue(tokenizer, mb)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got != 2 {
		t.Errorf("Map with bool values: got %d, want 2", got)
	}

	// 7. Map with Int values (fallback in reflection)
	// Key: 1 -> 1. Value: 2 -> 1. Total 2.
	mi := map[int]int{1: 2}
	got, err = CountTokensInValue(tokenizer, mi)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got != 2 {
		t.Errorf("Map with int values: got %d, want 2", got)
	}

	// 8. Custom Slice type (bypasses []string optimization, hits reflection string optimization)
	type MyStringSlice []string
	mss := MyStringSlice{"abcd", "efgh"}
	// 1 + 1 = 2
	got, err = CountTokensInValue(tokenizer, mss)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if got != 2 {
		t.Errorf("MyStringSlice: got %d, want 2", got)
	}
}

// TestRecursiveCycleDetection exercises cycle detection logic in reflection paths.
func TestRecursiveCycleDetection(t *testing.T) {
	tokenizer := NewSimpleTokenizer()

	// 1. Map Cycle
	type RecursiveMap map[string]interface{}
	m := make(RecursiveMap)
	m["self"] = m // Direct cycle

	// CountTokensInValue uses a visited map, so it should detect this.
	_, err := CountTokensInValue(tokenizer, m)
	if err == nil {
		t.Error("Expected error for map cycle, got nil")
	}

	// 2. Slice Cycle
	type RecursiveSlice []interface{}
	s := make(RecursiveSlice, 1)
	s[0] = s // Direct cycle
	_, err = CountTokensInValue(tokenizer, s)
	if err == nil {
		t.Error("Expected error for slice cycle, got nil")
	}

	// 3. Pointer Cycle (Struct)
	type Node struct {
		Next *Node
	}
	n := &Node{}
	n.Next = n
	_, err = CountTokensInValue(tokenizer, n)
	if err == nil {
		t.Error("Expected error for struct pointer cycle, got nil")
	}
}
