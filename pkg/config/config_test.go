// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/mcpany/core/pkg/logging"
	v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLogLevel_InvalidLevelWarning(t *testing.T) {
	logging.ForTestsOnlyResetLogger()
	var buf bytes.Buffer
	logging.Init(slog.LevelInfo, &buf)

	settings := &Settings{
		logLevel: "invalid-level",
		debug:    false,
	}

	logLevel := settings.LogLevel()

	assert.Equal(t, v1.GlobalSettings_LOG_LEVEL_INFO, logLevel)
	logs := buf.String()
	assert.Contains(t, logs, "Invalid log level specified: 'invalid-level'. Defaulting to INFO.")
}

func TestBindFlags(t *testing.T) {
	cmd := &cobra.Command{}
	BindFlags(cmd)

	// Check a fiew flags to ensure they are bound
	assert.NotNil(t, cmd.PersistentFlags().Lookup("mcp-listen-address"))
	assert.NotNil(t, cmd.Flags().Lookup("grpc-port"))
	assert.NotNil(t, cmd.Flags().Lookup("stdio"))

	// Check that the values are correctly bound to viper
	cmd.Flags().Set("grpc-port", "8081")
	assert.Equal(t, "8081", viper.GetString("grpc-port"))

	cmd.Flags().Set("stdio", "true")
	assert.True(t, viper.GetBool("stdio"))
}
