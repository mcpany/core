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
		// ASCII tests
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

		// Unicode/Rune tests (triggers levenshteinRunes)
		{"世界", "世界", 0},       // Identical
		{"世界", "世", 1},        // Deletion
		{"世", "世界", 1},        // Insertion
		{"世界", "世 界", 1},      // Space insertion
		{"hello", "héllo", 1},   // Substitution (e vs é)
		{"café", "coffee", 4},   // Mixed
		{"😊", "😢", 1},           // Emojis (substitution)
		{"😊", "😊😊", 1},          // Emoji insertion
		{"こんにちは", "こんちには", 2}, // Transposition-like (actually 2 subs or del+ins)

		// Mixed Empty/Unicode tests (triggers early returns in levenshteinRunes)
		{"", "世界", 2},
		{"世界", "", 2},
	}

	for _, tt := range tests {
		got := LevenshteinDistance(tt.s1, tt.s2)
		if got != tt.want {
			t.Errorf("LevenshteinDistance(%q, %q) = %d; want %d", tt.s1, tt.s2, got, tt.want)
		}
	}
}
