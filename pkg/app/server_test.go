/*
 * Copyright 2025 Author(s) of MCPXY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package app

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/mcpxy/core/pkg/mcpserver"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	t.Run("with nil fs", func(t *testing.T) {
		fs := setup(nil)
		assert.NotNil(t, fs, "setup should return a valid fs even if input is nil")
		_, ok := fs.(*afero.OsFs)
		assert.True(t, ok, "setup should default to OsFs")
	})

	t.Run("with existing fs", func(t *testing.T) {
		memFs := afero.NewMemMapFs()
		fs := setup(memFs)
		assert.Equal(t, memFs, fs, "setup should return the provided fs")
	})
}

func TestRun_ServerMode(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create a dummy config file
	configContent := `
upstream_services:
  - name: "test-http-service"
    http_service:
      address: "http://localhost:8080"
      calls:
        - operation_id: "echo"
          endpoint_path: "/echo"
          method: "HTTP_METHOD_POST"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	errChan := make(chan error, 1)
	go func() {
		// Use ephemeral ports by passing "0"
		// The test will hang if we use a real port that's not available.
		// We expect the Run function to exit gracefully when the context is canceled.
		errChan <- Run(ctx, fs, false, "0", "0", []string{"/config.yaml"})
	}()

	select {
	case err := <-errChan:
		// We expect a context cancellation error, which is normal for a graceful shutdown in this test setup.
		// The key is that the function returned, which it wouldn't if it were stuck.
		assert.Error(t, err)
	case <-ctx.Done():
		// This is the expected path: the context times out, causing the server to shut down.
		// We'll briefly wait to see if an error comes through on the channel.
		select {
		case err := <-errChan:
			assert.Error(t, err)
		default:
			// Success - the server ran until the context was canceled.
		}
	}
}

func TestRun_ConfigLoadError(t *testing.T) {
	fs := afero.NewMemMapFs()
	// Create a malformed config file
	err := afero.WriteFile(fs, "/config.yaml", []byte("malformed yaml:"), 0o644)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = Run(ctx, fs, false, "0", "0", []string{"/config.yaml"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load services from config")
}

func TestRun_StdioMode(t *testing.T) {
	originalRunStdioMode := runStdioMode
	defer func() { runStdioMode = originalRunStdioMode }()

	var stdioModeCalled bool
	runStdioMode = func(ctx context.Context, mcpSrv *mcpserver.Server) error {
		stdioModeCalled = true
		return nil
	}

	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	Run(ctx, fs, true, "0", "0", nil)

	assert.True(t, stdioModeCalled, "runStdioMode should have been called")
}

func TestRun_ServerStartupErrors(t *testing.T) {
	t.Run("http_server_fail", func(t *testing.T) {
		// Find a free port and occupy it
		l, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)
		defer l.Close()
		port := l.Addr().(*net.TCPAddr).Port

		fs := afero.NewMemMapFs()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Attempt to run the server on the occupied port
		err = Run(ctx, fs, false, fmt.Sprintf("%d", port), "0", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start a server")
	})

	t.Run("grpc_server_fail", func(t *testing.T) {
		// Find a free port and occupy it
		l, err := net.Listen("tcp", "localhost:0")
		require.NoError(t, err)
		defer l.Close()
		port := l.Addr().(*net.TCPAddr).Port

		fs := afero.NewMemMapFs()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Attempt to run the server on the occupied port
		err = Run(ctx, fs, false, "0", fmt.Sprintf("%d", port), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to start a server")
	})
}
