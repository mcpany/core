// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
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
