// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s1, s2 string
		want   int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"abc", "ac", 1},
		{"abc", "def", 3},
		{"kitten", "sitting", 3},
		{"rosettacode", "raisethysword", 8},
		{"get_weather", "get_wether", 1},
		{"get_weather", "get_weath", 2},
		{"get_weather", "getweather", 1},
	}

	for _, tt := range tests {
		got := LevenshteinDistance(tt.s1, tt.s2)
		if got != tt.want {
			t.Errorf("LevenshteinDistance(%q, %q) = %d; want %d", tt.s1, tt.s2, got, tt.want)
		}
	}
}

func TestLevenshteinDistance_NonASCII(t *testing.T) {
	tests := []struct {
		s1, s2 string
		want   int
	}{
		// Non-ASCII
		{"café", "cafe", 1}, // é vs e
		{"こんにちは", "こんちは", 1}, // Deletion
		{"😊", "😊", 0},
		{"😊", "😢", 1},
		{"abc", "abç", 1},
	}

	for _, tt := range tests {
		got := LevenshteinDistance(tt.s1, tt.s2)
		if got != tt.want {
			t.Errorf("LevenshteinDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, got, tt.want)
		}
	}
}

func TestLevenshteinDistance_LargeASCII(t *testing.T) {
	// Trigger the non-stack path (m+1 > 64)
	s1 := "this is a very long string to trigger the heap allocation path in levenshteinASCII because it is longer than sixty four characters"
	s2 := "this is a very long string to trigger the heap allocation path in levenshteinASCII because it is longer than sixty four characters." // one char diff

	got := LevenshteinDistance(s1, s2)
	if got != 1 {
		t.Errorf("LevenshteinDistance long string = %d, want 1", got)
	}
}

func TestLevenshteinDistance_LargeUnicode(t *testing.T) {
	// Trigger the non-stack path (m+1 > 256) for Unicode
	// Japanese 'a' is 3 bytes. 300 chars > 256.
	s1 := strings.Repeat("あ", 300)
	s2 := strings.Repeat("あ", 299) + "い" // Substitution at the end

	got := LevenshteinDistance(s1, s2)
	if got != 1 {
		t.Errorf("LevenshteinDistance large unicode = %d, want 1", got)
	}
}

func TestLevenshteinDistance_Swap(t *testing.T) {
	s1 := "short"
	s2 := "longer_string"
	// s2 is longer than s1, should work
	got := LevenshteinDistance(s1, s2)
	// longer_string (13) vs short (5).
	// l, o, n, g, e, r, _, s, t, r, i, n, g
	// s, h, o, r, t
	// "short" -> "shor" (del t) -> "sho" (del r) ...
	// easier: distance is at least 13-5 = 8.
	if got < 8 {
		t.Errorf("LevenshteinDistance(%q, %q) = %d, expected >= 8", s1, s2, got)
	}
}

func TestLevenshteinDistanceWithLimit(t *testing.T) {
	tests := []struct {
		s1, s2 string
		limit  int
		want   int // If > limit, any value > limit is acceptable, but our implementation returns limit + 1
	}{
		{"abc", "abc", 1, 0},
		{"abc", "abd", 1, 1},
		{"abc", "abe", 1, 1}, // substitution 'c'->'e'
		{"abc", "def", 1, 2}, // distance is 3, limit 1 -> returns 2 (limit+1)
		{"abc", "abcd", 1, 1},
		{"abc", "abcde", 1, 2}, // distance 2, limit 1 -> returns 2
		{"kitten", "sitting", 3, 3},
		{"kitten", "sitting", 2, 3}, // distance 3, limit 2 -> returns 3
		{"rosettacode", "raisethysword", 5, 6}, // distance 8, limit 5 -> returns 6

		// ASCII optimization check (length diff > limit)
		{"abc", "abcdef", 2, 3}, // diff 3, limit 2 -> returns 3

		// Unicode
		{"café", "cafe", 1, 1},
		{"café", "cafe", 0, 1}, // distance 1, limit 0 -> returns 1
		{"😊", "😢", 0, 1},
	}

	for _, tt := range tests {
		got := LevenshteinDistanceWithLimit(tt.s1, tt.s2, tt.limit)
		if tt.want > tt.limit {
			if got <= tt.limit {
				t.Errorf("LevenshteinDistanceWithLimit(%q, %q, %d) = %d; want > %d", tt.s1, tt.s2, tt.limit, got, tt.limit)
			}
		} else {
			if got != tt.want {
				t.Errorf("LevenshteinDistanceWithLimit(%q, %q, %d) = %d; want %d", tt.s1, tt.s2, tt.limit, got, tt.want)
			}
		}
	}
}

func TestLevenshteinDistanceWithLimit_StackBoundary(t *testing.T) {
	// Test boundary conditions for stack allocation (256 runes)
	// Case 1: 256 runes (fits in stack)
	s256a := strings.Repeat("a", 256)
	s256b := strings.Repeat("a", 255) + "b"
	if got := LevenshteinDistanceWithLimit(s256a, s256b, 5); got != 1 {
		t.Errorf("LevenshteinDistanceWithLimit 256 boundary failed: got %d, want 1", got)
	}

	// Case 2: 257 runes (falls back to heap)
	s257a := strings.Repeat("a", 257)
	s257b := strings.Repeat("a", 256) + "b"
	if got := LevenshteinDistanceWithLimit(s257a, s257b, 5); got != 1 {
		t.Errorf("LevenshteinDistanceWithLimit 257 boundary failed: got %d, want 1", got)
	}

	// Case 3: Mixed unicode at boundary
	// "世" is 3 bytes. 256 * 3 = 768 bytes, but 256 runes.
	u256a := strings.Repeat("世", 256)
	u256b := strings.Repeat("世", 255) + "界"
	if got := LevenshteinDistanceWithLimit(u256a, u256b, 5); got != 1 {
		t.Errorf("LevenshteinDistanceWithLimit unicode 256 boundary failed: got %d, want 1", got)
	}
}
