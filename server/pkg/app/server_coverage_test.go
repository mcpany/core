// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheckWithContext_InvalidAddr(t *testing.T) {
	// \n is invalid in URL
	err := HealthCheckWithContext(context.Background(), io.Discard, "invalid\naddr")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
}

func TestRun_WithListenAddress(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configContent := `
global_settings:
  mcp_listen_address: "localhost:0"
upstream_services: []
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	assert.NoError(t, err)

	app := NewApplication()
	errChan := make(chan error, 1)

	go func() {
		// Pass empty port strings so it relies on config
		errChan <- app.Run(ctx, fs, false, "", "", []string{"/config.yaml"}, "", 5*time.Second)
	}()

	// Wait for start (approximated by sleep or checking errChan not closed immediately)
	time.Sleep(100 * time.Millisecond)
	cancel()
	err = <-errChan
	assert.NoError(t, err)
}

