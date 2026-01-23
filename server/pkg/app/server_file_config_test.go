// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldEnableFileConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVal      string
		configPaths []string
		expected    bool
	}{
		{
			name:        "Env explicitly true",
			envVal:      "true",
			configPaths: []string{},
			expected:    true,
		},
		{
			name:        "Env explicitly true (case insensitive)",
			envVal:      "TRUE",
			configPaths: []string{},
			expected:    true,
		},
		{
			name:        "Env explicitly 1",
			envVal:      "1",
			configPaths: []string{},
			expected:    true,
		},
		{
			name:        "Env explicitly false",
			envVal:      "false",
			configPaths: []string{"config.yaml"},
			expected:    false,
		},
		{
			name:        "Env explicitly 0",
			envVal:      "0",
			configPaths: []string{"config.yaml"},
			expected:    false,
		},
		{
			name:        "Env explicitly invalid",
			envVal:      "foo",
			configPaths: []string{"config.yaml"},
			expected:    false,
		},
		{
			name:        "Env unset, no config paths",
			envVal:      "",
			configPaths: []string{},
			expected:    false,
		},
		{
			name:        "Env unset, with config paths",
			envVal:      "",
			configPaths: []string{"config.yaml"},
			expected:    true,
		},
		{
			name:        "Env unset, with multiple config paths",
			envVal:      "",
			configPaths: []string{"config1.yaml", "config2.yaml"},
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldEnableFileConfig(tt.envVal, tt.configPaths)
			assert.Equal(t, tt.expected, result)
		})
	}
}
