// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestBindRootFlags(t *testing.T) {
	cmd := &cobra.Command{}
	BindRootFlags(cmd)

	// Check flags exist
	assert.NotNil(t, cmd.PersistentFlags().Lookup("mcp-listen-address"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("config-path"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("metrics-listen-address"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("debug"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("log-level"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("logfile"))

	// Check viper binding (indirectly)
	// BindPFlag returns error if flag not found, which we checked.
	// But BindRootFlags calls os.Exit(1) on error, which is hard to test unless we mock os.Exit or ensure it doesn't fail.
	// Since we are passing a valid command with flags just added, it shouldn't fail.
}

func TestBindServerFlags(t *testing.T) {
	cmd := &cobra.Command{}
	BindServerFlags(cmd)

	assert.NotNil(t, cmd.Flags().Lookup("grpc-port"))
	assert.NotNil(t, cmd.Flags().Lookup("stdio"))
	assert.NotNil(t, cmd.Flags().Lookup("shutdown-timeout"))
	assert.NotNil(t, cmd.Flags().Lookup("api-key"))
	assert.NotNil(t, cmd.Flags().Lookup("profiles"))
	assert.NotNil(t, cmd.Flags().Lookup("db-path"))
}

func TestSettings_ToProto(t *testing.T) {
	viper.Reset()
	s := GlobalSettings()
	proto := s.ToProto()
	assert.NotNil(t, proto)
}
