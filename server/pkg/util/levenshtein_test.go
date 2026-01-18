// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import "testing"

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s, t string
		want int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"kitten", "sitting", 3},
		{"flaw", "lawn", 2},
		{"gumbo", "gambol", 2},
		{"book", "back", 2},
		{"foo", "foobar", 3},
	}

	for _, tt := range tests {
		got := LevenshteinDistance(tt.s, tt.t)
		if got != tt.want {
			t.Errorf("LevenshteinDistance(%q, %q) = %d; want %d", tt.s, tt.t, got, tt.want)
		}
	}
}

func TestFindClosestMatch(t *testing.T) {
	candidates := []string{"get_weather", "list_files", "read_file", "write_file", "delete_file"}

	tests := []struct {
		target      string
		maxDistance int
		wantMatch   string
		wantFound   bool
	}{
		{"get_wether", 3, "get_weather", true},
		{"listfiles", 3, "list_files", true},
		{"readfile", 3, "read_file", true},
		{"completely_wrong", 3, "", false},
		{"get_weather", 3, "get_weather", true}, // Exact match
		{"wite_file", 3, "write_file", true},
		{"dlete_file", 3, "delete_file", true},
		{"get_time", 3, "", false}, // Too far
	}

	for _, tt := range tests {
		match, found := FindClosestMatch(tt.target, candidates, tt.maxDistance)
		if found != tt.wantFound {
			t.Errorf("FindClosestMatch(%q) found = %v; want %v", tt.target, found, tt.wantFound)
		}
		if match != tt.wantMatch {
			t.Errorf("FindClosestMatch(%q) match = %q; want %q", tt.target, match, tt.wantMatch)
		}
	}
}
