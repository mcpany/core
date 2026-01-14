// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_StartupFailure_WhenServiceFailsToRegister(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	configOpenAPI := `
upstream_services:
 - name: "broken-openapi"
   openapi_service:
     address: "http://localhost:9999"
     spec_url: "http://localhost:9999/swagger.json"
`
	err := afero.WriteFile(fs, "/config_openapi.yaml", []byte(configOpenAPI), 0o644)
	require.NoError(t, err)

	app := NewApplication()

	// We run the app. It should return error because of startup failure.
	err = app.Run(ctx, fs, false, "localhost:0", "", []string{"/config_openapi.yaml"}, "", 5*time.Second)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "startup failed")
	assert.Contains(t, err.Error(), "broken-openapi")
}

func TestRun_StartupSuccess_WhenServiceRegisters(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Short timeout because we expect success quickly
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// HTTP service is lazy, so it always registers successfully (tools added = 0)
	configHTTP := `
upstream_services:
 - name: "working-http"
   http_service:
     address: "http://localhost:8080"
`
	err := afero.WriteFile(fs, "/config_http.yaml", []byte(configHTTP), 0o644)
	require.NoError(t, err)

	app := NewApplication()

	errChan := make(chan error, 1)
	go func() {
		// Run should block until context cancel if successful
		errChan <- app.Run(ctx, fs, false, "localhost:0", "", []string{"/config_http.yaml"}, "", 1*time.Second)
	}()

	// Wait for startup to complete (Run blocks) or fail
	select {
	case err := <-errChan:
		// If it returns early with error, that's bad (unless context cancelled)
		// But we expect it to wait for context cancel.
		// If startup failed, it returns non-nil error.
		// If startup succeeded, it waits, then returns nil on cancel.
		// But here we are waiting for result.
		// If it's nil, it means it shutdown gracefully.
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		// If still running after 2s, we assume startup succeeded and it's running.
		// Trigger shutdown.
		cancel()
		err := <-errChan
		assert.NoError(t, err)
	}
}
