// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexliesenfeld/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockChecker struct {
	res        health.CheckerResult
	stopCalled bool
}

func (m *mockChecker) Check(ctx context.Context) health.CheckerResult {
	return m.res
}

func (m *mockChecker) Stop() {
	m.stopCalled = true
}

func (m *mockChecker) Start() {
}

func (m *mockChecker) GetRunningPeriodicCheckCount() int {
    return 0
}

func (m *mockChecker) IsStarted() bool {
    return true
}

func TestUpstream_CheckHealth(t *testing.T) {
	ctx := context.Background()

	t.Run("nil checker", func(t *testing.T) {
		u := &Upstream{}
		err := u.CheckHealth(ctx)
		assert.NoError(t, err)
	})

	t.Run("health up", func(t *testing.T) {
		checker := &mockChecker{
			res: health.CheckerResult{Status: health.StatusUp},
		}
		u := &Upstream{checker: checker}
		err := u.CheckHealth(ctx)
		assert.NoError(t, err)
	})

	t.Run("health down", func(t *testing.T) {
		checker := &mockChecker{
			res: health.CheckerResult{Status: health.StatusDown},
		}
		u := &Upstream{checker: checker}
		err := u.CheckHealth(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "health check failed")
	})
}

func TestUpstream_Shutdown_Health(t *testing.T) {
	ctx := context.Background()

	t.Run("nil checker", func(t *testing.T) {
		u := &Upstream{}
		err := u.Shutdown(ctx)
		assert.NoError(t, err)
	})

	t.Run("with checker stop", func(t *testing.T) {
		checker := &mockChecker{}
		u := &Upstream{checker: checker}
		err := u.Shutdown(ctx)
		assert.NoError(t, err)
		assert.True(t, checker.stopCalled)
	})

	t.Run("cleanup bundle directory", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "mcp-bundle-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir) // Ensure cleanup even if test fails

		serviceID := "test-service-id"
		bundlePath := filepath.Join(tempDir, serviceID)
		err = os.Mkdir(bundlePath, 0755)
		require.NoError(t, err)

		u := &Upstream{
			serviceID:     serviceID,
			BundleBaseDir: tempDir,
		}

		err = u.Shutdown(ctx)
		assert.NoError(t, err)

		// Verify directory is deleted
		_, err = os.Stat(bundlePath)
		assert.True(t, os.IsNotExist(err), "bundle directory should have been removed")
	})

	t.Run("untrack service id without dir", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "mcp-bundle-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		u := &Upstream{
			serviceID:     "missing-service",
			BundleBaseDir: tempDir,
		}

		err = u.Shutdown(ctx)
		assert.NoError(t, err) // Should not error if directory does not exist
	})
}
