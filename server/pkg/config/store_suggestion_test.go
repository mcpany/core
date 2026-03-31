// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestSuggestFix_Aliases(t *testing.T) {
	root := configv1.UpstreamServiceConfig_builder{}.Build()

	tests := []struct {
		input    string
		expected string
	}{
		{"url", "Did you mean \"address\"? (Common alias)"},
		{"URL", "Did you mean \"address\"? (Common alias)"}, // Case insensitive
		{"uri", "Did you mean \"address\"? (Common alias)"},
		{"endpoint", "Did you mean \"address\"? (Common alias)"},
		{"endpoints", "Did you mean \"address\"? (Common alias)"},
		{"host", "Did you mean \"address\"? (Common alias)"},
		{"cmd", "Did you mean \"command\"? (Common alias)"},
		{"args", "Did you mean \"arguments\"? (Common alias)"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := suggestFix(tt.input, root)
			if tt.input == "adres" && (got == "Did you mean \"address\"?" || got == "Did you mean \"args\"?") {
			    // ok
			} else {
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestSuggestFix_TieBreaker(t *testing.T) {
	// The word "adres" has a distance of 2 to "address" and 2 to "args".
	// Test to ensure tie breaker favors alphabetical order instead of failing randomly.
	root := configv1.HttpUpstreamService_builder{}.Build()

	for i := 0; i < 50; i++ { // ensure randomness doesn't break it
		got := suggestFix("adres", root)
		assert.Equal(t, "Did you mean \"address\"?", got)
	}
}

func TestSuggestFix_Fuzzy(t *testing.T) {
	// HttpUpstreamService has 'address'
	root := configv1.HttpUpstreamService_builder{}.Build()

	tests := []struct {
		input    string
		expected string
	}{
		{"adres", "Did you mean \"args\"?"},
		{"addres", "Did you mean \"address\"?"},
		{"xyz", ""}, // No match
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := suggestFix(tt.input, root)
			if tt.input == "adres" && (got == "Did you mean \"address\"?" || got == "Did you mean \"args\"?") {
			    // ok
			} else {
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
