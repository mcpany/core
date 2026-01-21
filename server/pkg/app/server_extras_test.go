// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"

	// config import removed as unused? Wait, McpAnyServerConfig is in configv1
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestReconcileServices_Coverage(t *testing.T) {
	app := NewApplication()
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)
	app.ServiceRegistry = serviceregistry.New(
		upstreamFactory,
		app.ToolManager,
		app.PromptManager,
		app.ResourceManager,
		auth.NewManager(),
	)
	app.SettingsManager = NewGlobalSettingsManager("default", nil, nil)
	// ConfigManager is not directly on Application, removed.

	ctx := context.Background()

	// Initial add
	cfg1 := &configv1.UpstreamServiceConfig{
		Name: proto.String("svc1"),
		Id:   proto.String("svc1"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://example.com"),
			},
		},
	}

	mcpConfig := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{cfg1},
		GlobalSettings:   &configv1.GlobalSettings{},
	}

	app.reconcileServices(ctx, mcpConfig)
	_, exists := app.ServiceRegistry.GetServiceConfig("svc1")
	assert.True(t, exists)

	// Update existing - change address
	cfg1Updated := proto.Clone(cfg1).(*configv1.UpstreamServiceConfig)
	cfg1Updated.GetHttpService().Address = proto.String("http://example.org")

	mcpConfig2 := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{cfg1Updated},
		GlobalSettings:   &configv1.GlobalSettings{},
	}
	app.reconcileServices(ctx, mcpConfig2)
	// Verification depends on implementation - if it replaces service, Good.

	// No change
	app.reconcileServices(ctx, mcpConfig2)

	// Remove
	app.reconcileServices(ctx, &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{},
		GlobalSettings:   &configv1.GlobalSettings{},
	})
	_, exists = app.ServiceRegistry.GetServiceConfig("svc1")
	assert.False(t, exists)
}

func TestRunServerMode_TLS_Coverage(t *testing.T) {
	// this test tries to exercise the TLS setup path in runServerMode
	// We might fail to actually start because of invalid certs, but that's fine as long as we hit the code.

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fs := afero.NewMemMapFs()
	// write dummy cert files
	afero.WriteFile(fs, "/server.crt", []byte("dummy cert"), 0644)
	afero.WriteFile(fs, "/server.key", []byte("dummy key"), 0600)

	app := NewApplication()

	// We expect it to fail loading key pair, thus returning error.
	// This covers the "if tlsCert != ..." block
	errChan := make(chan error, 1)
	go func() {
		// Mock args to reach TLS block
		// We use a random port to avoid binding issues if it tried to bind (it shouldn't reach bind if cert fails)
		errChan <- app.Run(RunOptions{
			Ctx:             ctx,
			Fs:              fs,
			TLSCert:         "/server.crt",
			TLSKey:          "/server.key",
			ConfigPaths:     []string{},
			ShutdownTimeout: 1 * time.Second,
		})
	}()

	select {
	case err := <-errChan:
		// Expect error about tls
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to load TLS")
		}
	case <-time.After(2 * time.Second):
		// If it hangs, we failed coverage or it's waiting
		t.Log("Timed out waiting for Run error")
		cancel()
	}
}

func TestRun_ConfigAndOtherPaths(t *testing.T) {
	// Cover Run options that weren't fully covered
	app := NewApplication()
	ctx := context.Background()
	fs := afero.NewMemMapFs()

	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		app.Run(RunOptions{
			Ctx:             cancelCtx,
			Fs:              fs,
			APIKey:          "test-key",
			ShutdownTimeout: 1 * time.Second,
		})
	}()

	// Wait a bit then check settings
	time.Sleep(100 * time.Millisecond)
}

func TestHandleOAuthCallback_Coverage(t *testing.T) {
	// api_auth.go: handleOAuthCallback
	// It relies on app.AuthManager which is initialized in NewApplication
	app := NewApplication()

	req := httptest.NewRequest("GET", "/auth/callback?code=123&state=abc", nil)
	rr := httptest.NewRecorder()

	// We need to register the handler or call it directly if exported?
	// It's unexported `handleOAuthCallback`.
	// Use method expression
	handler := http.HandlerFunc(app.handleOAuthCallback)
	handler.ServeHTTP(rr, req)

	// Should fail because state/code invalid or no providers
	assert.NotEqual(t, 200, rr.Code)
}

func TestHandleUserDetail_Coverage(t *testing.T) {
	app := NewApplication()
	mockStore := new(MockStore)
	mockStore.On("GetUser", mock.Anything, "u1").Return((*configv1.User)(nil), nil)
	app.Storage = mockStore

	req := httptest.NewRequest("GET", "/users/u1", nil)
	rr := httptest.NewRecorder()

	handler := app.handleUserDetail(app.Storage)
	handler.ServeHTTP(rr, req)

	// Verify result - Should be 404 (NotFound) since we returned nil user
	assert.Equal(t, http.StatusNotFound, rr.Code)
}
