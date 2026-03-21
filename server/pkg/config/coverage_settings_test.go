// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestSettings_LogFormat_Fallback(t *testing.T) {
	viper.Reset()
	viper.Set("log-format", "INVALID_FORMAT")

	s2 := &Settings{}
	assert.Equal(t, configv1.GlobalSettings_LOG_FORMAT_TEXT, s2.LogFormat())
}

func TestSettings_LogLevel_Invalid(t *testing.T) {
	s := &Settings{
		logLevel: "INVALID_LEVEL",
	}
	// This should log a warning, but return INFO
	assert.Equal(t, configv1.GlobalSettings_LOG_LEVEL_INFO, s.LogLevel())
}

func TestGetStringSlice_Coverage(t *testing.T) {
	viper.Reset()

	// Case 1: String with comma (Env var simulation)
	viper.Set("env-list", "a, b, c")
	res := getStringSlice("env-list")
	assert.Equal(t, []string{"a", "b", "c"}, res)

	// Case 2: Slice with comma inside items
	viper.Set("slice-list", []string{"a,b", " c "})
	res2 := getStringSlice("slice-list")
	assert.Equal(t, []string{"a", "b", "c"}, res2)

	// Case 3: Empty items
	viper.Set("empty-items", "a,,c")
	res3 := getStringSlice("empty-items")
	assert.Equal(t, []string{"a", "c"}, res3)
}
