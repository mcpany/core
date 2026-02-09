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
		{"url", "did you mean \"address\"? (Common alias)"},
		{"URL", "did you mean \"address\"? (Common alias)"}, // Case insensitive
		{"uri", "did you mean \"address\"? (Common alias)"},
		{"endpoint", "did you mean \"address\"? (Common alias)"},
		{"endpoints", "did you mean \"address\"? (Common alias)"},
		{"host", "did you mean \"address\"? (Common alias)"},
		{"cmd", "did you mean \"command\"? (Common alias)"},
		{"args", "did you mean \"arguments\"? (Common alias)"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := suggestFix(tt.input, root)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSuggestFix_Fuzzy(t *testing.T) {
	// HttpUpstreamService has 'address'
	root := configv1.HttpUpstreamService_builder{}.Build()

	tests := []struct {
		input    string
		expected string
	}{
		{"adres", "did you mean \"address\"?"},
		{"addres", "did you mean \"address\"?"},
		{"xyz", ""}, // No match
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := suggestFix(tt.input, root)
			assert.Equal(t, tt.expected, got)
		})
	}
}
