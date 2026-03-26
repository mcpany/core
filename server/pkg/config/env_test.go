// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestEnvVarConfigPath(t *testing.T) {
	// Reset viper
	viper.Reset()

	cmd := &cobra.Command{}
	BindRootFlags(cmd)

	// Set environment variable
	os.Setenv("MCPANY_CONFIG_PATH", "env-config.yaml")
	defer os.Unsetenv("MCPANY_CONFIG_PATH")

	settings := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}
	fs := afero.NewMemMapFs()
	// Create files to avoid load errors, though we mostly care about configPaths parsing
	_ = afero.WriteFile(fs, "env-config.yaml", []byte("upstream_services: []"), 0644)

	_ = settings.Load(cmd, fs)

	val := settings.ConfigPaths()

	// We expect the value to be []string{"env-config.yaml"}
	assert.Equal(t, []string{"env-config.yaml"}, val)
}

func TestEnvVarConfigPathMultiple(t *testing.T) {
	// Reset viper
	viper.Reset()

	cmd := &cobra.Command{}
	BindRootFlags(cmd)

	// Set environment variable with multiple paths
	os.Setenv("MCPANY_CONFIG_PATH", "env-config1.yaml,env-config2.yaml")
	defer os.Unsetenv("MCPANY_CONFIG_PATH")

	settings := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "env-config1.yaml", []byte("upstream_services: []"), 0644)
	_ = afero.WriteFile(fs, "env-config2.yaml", []byte("upstream_services: []"), 0644)

	_ = settings.Load(cmd, fs)

	val := settings.ConfigPaths()

	// We expect the value to be []string{"env-config1.yaml", "env-config2.yaml"}
	assert.Equal(t, []string{"env-config1.yaml", "env-config2.yaml"}, val)
}

func TestEnvVarProfilesMultiple(t *testing.T) {
	// Reset viper
	viper.Reset()

	cmd := &cobra.Command{}
	BindFlags(cmd)

	// Set environment variable with multiple profiles
	os.Setenv("MCPANY_PROFILES", "profile1,profile2")
	defer os.Unsetenv("MCPANY_PROFILES")

	settings := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}
	fs := afero.NewMemMapFs()

	_ = settings.Load(cmd, fs)

	val := settings.Profiles()

	// We expect the value to be []string{"profile1", "profile2"}
	assert.Equal(t, []string{"profile1", "profile2"}, val)
}

func TestEnvVarConfigPathWithSpaces(t *testing.T) {
	// Reset viper
	viper.Reset()

	cmd := &cobra.Command{}
	BindRootFlags(cmd)

	// Set environment variable with multiple paths and spaces
	// Note: pflag/viper might parse "env-config1.yaml, env-config2.yaml" weirdly.
	// If it splits by space, it might be ["env-config1.yaml,", "env-config2.yaml"]
	os.Setenv("MCPANY_CONFIG_PATH", "env-config1.yaml, env-config2.yaml")
	defer os.Unsetenv("MCPANY_CONFIG_PATH")

	settings := &Settings{
		proto: configv1.GlobalSettings_builder{}.Build(),
	}
	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "env-config1.yaml", []byte("upstream_services: []"), 0644)
	_ = afero.WriteFile(fs, "env-config2.yaml", []byte("upstream_services: []"), 0644)

	_ = settings.Load(cmd, fs)

	val := settings.ConfigPaths()

	// We expect the value to be clean paths
	assert.Equal(t, []string{"env-config1.yaml", "env-config2.yaml"}, val)
}
