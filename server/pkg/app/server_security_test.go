// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_APIKeyAuthentication_QueryParamDisabled(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := NewApplication()
	errChan := make(chan error, 1)

	// Set the API key
	viper.Set("api-key", "test-api-key")
	defer viper.Set("api-key", "")

	// Get the address from the listener
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	addr := l.Addr().String()
	_ = l.Close()

	go func() {
		errChan <- app.Run(ctx, fs, false, addr, "", nil, viper.GetString("api-key"), 5*time.Second)
	}()

	// Wait for the server to be ready
	waitForServerReady(t, addr)

	// Make a request with API key in query param - SHOULD FAIL
	req, err := http.NewRequest("GET", "http://"+addr+"/api/v1/topology?api_key=test-api-key", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()

	// Expect 401 Unauthorized because query param is disabled
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "API Key in query param should be ignored")

	// Make a request with API key in Header - SHOULD SUCCEED
	req, err = http.NewRequest("GET", "http://"+addr+"/api/v1/topology", nil)
	require.NoError(t, err)
	req.Header.Set("X-API-Key", "test-api-key")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "API Key in header should be accepted")

	cancel()
	err = <-errChan
	assert.NoError(t, err)
}
