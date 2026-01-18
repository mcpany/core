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
