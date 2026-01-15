// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/logging"
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
)

func TestConfigureUIHandler(t *testing.T) {
	// Setup dependencies
	busProvider, _ := bus.NewProvider(nil)
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)
	toolManager := tool.NewManager(busProvider)
	authManager := auth.NewManager()
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, prompt.NewManager(), resource.NewManager(), authManager)
	mcpSrv, _ := mcpserver.NewServer(context.Background(), toolManager, prompt.NewManager(), resource.NewManager(), authManager, serviceRegistry, busProvider, false)

	t.Run("Prioritize ./ui/out", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		fs.MkdirAll("ui/out", 0755)
		fs.MkdirAll("ui/dist", 0755)
		fs.MkdirAll("ui", 0755)

		app := NewApplication()
		app.fs = fs
		app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		logging.ForTestsOnlyResetLogger()
		var buf ThreadSafeBuffer
		logging.Init(slog.LevelInfo, &buf)

		_ = app.runServerMode(ctx, mcpSrv, busProvider, "localhost:0", "localhost:0", 1*time.Second, nil, middleware.NewCachingMiddleware(toolManager), nil, serviceRegistry, nil)

		logs := buf.String()
		assert.NotContains(t, logs, "UI directory ./ui contains package.json")
		assert.NotContains(t, logs, "No UI directory found")
	})

	t.Run("Block ./ui with package.json", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		fs.MkdirAll("ui", 0755)
		afero.WriteFile(fs, "ui/package.json", []byte("{}"), 0644)

		app := NewApplication()
		app.fs = fs
		app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		logging.ForTestsOnlyResetLogger()
		var buf ThreadSafeBuffer
		logging.Init(slog.LevelInfo, &buf)

		_ = app.runServerMode(ctx, mcpSrv, busProvider, "localhost:0", "localhost:0", 1*time.Second, nil, middleware.NewCachingMiddleware(toolManager), nil, serviceRegistry, nil)

		logs := buf.String()
		assert.Contains(t, logs, "UI directory ./ui contains package.json. Refusing to serve")
	})

	t.Run("Allow ./ui without package.json", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		fs.MkdirAll("ui", 0755)
		// No package.json

		app := NewApplication()
		app.fs = fs
		app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		logging.ForTestsOnlyResetLogger()
		var buf ThreadSafeBuffer
		logging.Init(slog.LevelInfo, &buf)

		_ = app.runServerMode(ctx, mcpSrv, busProvider, "localhost:0", "localhost:0", 1*time.Second, nil, middleware.NewCachingMiddleware(toolManager), nil, serviceRegistry, nil)

		logs := buf.String()
		assert.NotContains(t, logs, "Refusing to serve")
		assert.NotContains(t, logs, "No UI directory found")
	})

	t.Run("No UI directory", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		app := NewApplication()
		app.fs = fs
		app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		logging.ForTestsOnlyResetLogger()
		var buf ThreadSafeBuffer
		logging.Init(slog.LevelInfo, &buf)

		_ = app.runServerMode(ctx, mcpSrv, busProvider, "localhost:0", "localhost:0", 1*time.Second, nil, middleware.NewCachingMiddleware(toolManager), nil, serviceRegistry, nil)

		logs := buf.String()
		assert.Contains(t, logs, "No UI directory found")
	})
}
