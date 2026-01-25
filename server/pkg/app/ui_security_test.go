// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUIHandler_Security(t *testing.T) {
	// Setup dependencies
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	busProvider, _ := bus.NewProvider(nil)
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, promptManager, resourceManager, authManager)
	mcpSrv, _ := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)

	t.Run("Refuse to serve source code", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		app := NewApplication()
		app.fs = fs
		app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

		// Create a "ui" directory that looks like source code
		fs.MkdirAll("ui", 0755)
		afero.WriteFile(fs, "ui/package.json", []byte("{}"), 0644)
		afero.WriteFile(fs, "ui/index.html", []byte("<html></html>"), 0644)

		runCtx, runCancel := context.WithCancel(ctx)
		defer runCancel()

		errChan := make(chan error, 1)
		go func() {
			errChan <- app.runServerMode(runCtx, mcpSrv, busProvider, "127.0.0.1:0", "127.0.0.1:0", 1*time.Second, nil, middleware.NewCachingMiddleware(toolManager), nil, nil, serviceRegistry, nil, "", "", "")
		}()

		require.Eventually(t, func() bool { return app.BoundHTTPPort.Load() != 0 }, 5*time.Second, 100*time.Millisecond)
		port := app.BoundHTTPPort.Load()
		baseURL := "http://127.0.0.1:" + strconv.Itoa(int(port))

		// Request /ui/index.html
		resp, err := http.Get(baseURL + "/ui/index.html")
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should NOT be 200 OK. It might be 404 (Not Found) or 400 (Bad Request from fallback handler)
		// but crucially it should not serve the file.
		assert.NotEqual(t, http.StatusOK, resp.StatusCode, "Should not serve source file")

		body, _ := io.ReadAll(resp.Body)
		assert.NotContains(t, string(body), "<html></html>", "Should not contain file content")
	})

	t.Run("Block dotfiles", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		app := NewApplication()
		app.fs = fs
		app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

		// Create a valid "ui/dist" directory
		fs.MkdirAll("ui/dist", 0755)
		afero.WriteFile(fs, "ui/dist/index.html", []byte("<html></html>"), 0644)
		afero.WriteFile(fs, "ui/dist/.env", []byte("SECRET=true"), 0644)
		fs.MkdirAll("ui/dist/.git", 0755)
		afero.WriteFile(fs, "ui/dist/.git/HEAD", []byte("ref: refs/heads/main"), 0644)

		runCtx, runCancel := context.WithCancel(ctx)
		defer runCancel()

		errChan := make(chan error, 1)
		go func() {
			errChan <- app.runServerMode(runCtx, mcpSrv, busProvider, "127.0.0.1:0", "127.0.0.1:0", 1*time.Second, nil, middleware.NewCachingMiddleware(toolManager), nil, nil, serviceRegistry, nil, "", "", "")
		}()

		require.Eventually(t, func() bool { return app.BoundHTTPPort.Load() != 0 }, 5*time.Second, 100*time.Millisecond)
		port := app.BoundHTTPPort.Load()
		baseURL := "http://127.0.0.1:" + strconv.Itoa(int(port))

		// Request valid file
		resp, err := http.Get(baseURL + "/ui/index.html")
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Request .env
		resp, err = http.Get(baseURL + "/ui/.env")
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

		// Request .git/HEAD
		resp, err = http.Get(baseURL + "/ui/.git/HEAD")
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}
