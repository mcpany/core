package app
// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRunner is a mock implementation of the Runner interface
type MockRunner struct {
	mock.Mock
}

func (m *MockRunner) Run(ctx context.Context, fs afero.Fs, stdio bool, jsonrpcPort, grpcPort string, configPaths []string, apiKey string, shutdownTimeout time.Duration) error {
	args := m.Called(ctx, fs, stdio, jsonrpcPort, grpcPort, configPaths, apiKey, shutdownTimeout)
	return args.Error(0)
}

func (m *MockRunner) ReloadConfig(ctx context.Context, fs afero.Fs, configPaths []string) error {
	args := m.Called(ctx, fs, configPaths)
	return args.Error(0)
}

// TestApplication_Run_RespectsConfigPaths tests that config paths are respected
// even if MCPANY_ENABLE_FILE_CONFIG is not set.
// Note: This test verifies the logic within Application.Run by checking if it attempts to load
// from the file store. However, Application.Run creates the store internally.
// We can't easily mock the internal config.NewFileStore call without DI.
// But we can check if it fails with a specific error if the config file is invalid/missing,
// proving it TRIED to load it.
func TestApplication_Run_RespectsConfigPaths(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	app := NewApplication()
	ctx := context.Background()

	// Create a dummy config file
	configPath := "test_config.yaml"
	// Invalid YAML to force an error during LoadServices, which proves it tried to load it.
	// If it ignored it, it would succeed (empty config).
	_ = afero.WriteFile(fs, configPath, []byte("invalid_yaml: :"), 0644)

	// Run
	err := app.Run(ctx, fs, false, "8080", "9090", []string{configPath}, "test-key", 5*time.Second)

	// Assert
	// If it tried to load, it should fail with a config error.
	// If it ignored it, it would likely succeed (or fail later on port binding if mocked poorly).
	// But LoadServices failure happens early.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load services from config")
}

// TestApplication_Run_IgnoresMissingConfigPaths tests that it runs fine if no config is provided
func TestApplication_Run_NoConfig(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	app := NewApplication()
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately to exit Run loop if it starts successfully
	cancel()

	// Run with empty config paths
	// We expect it to start (and then exit due to context cancel) or fail due to port binding.
	// But getting past config loading is the key.
	err := app.Run(ctx, fs, false, "0", "0", nil, "test-key", 1*time.Second)

	// It might error on port binding or something else, but NOT config loading.
	if err != nil {
		assert.NotContains(t, err.Error(), "failed to load services from config")
	}
}
