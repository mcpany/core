// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/profile"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func ptr[T any](v T) *T {
	return &v
}

func TestRunServerMode_Auth(t *testing.T) {
	// Setup dependencies
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Find free port
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	bindAddress := fmt.Sprintf("localhost:%d", port)
	grpcPort := "" // Disable gRPC for this test

	busProvider, _ := bus.NewProvider(nil)
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager)

	// Create app
	app := NewApplication()

	// Initialize SettingsManager
	app.SettingsManager = NewGlobalSettingsManager("global-secret", nil, nil)

	// Global Key - config.GlobalSettings is used by some components, ensure sync
	config.GlobalSettings().SetAPIKey("global-secret")
	defer config.GlobalSettings().SetAPIKey("")

	authManager := auth.NewManager()
	authManager.SetAPIKey("global-secret")

	app.AuthManager = authManager
	app.ProfileManager = profile.NewManager(nil) // Empty profiles

	serviceRegistry := serviceregistry.New(
		upstreamFactory,
		app.ToolManager,
		app.PromptManager,
		app.ResourceManager,
		authManager,
	)

	mcpSrv, err := mcpserver.NewServer(
		ctx,
		app.ToolManager,
		app.PromptManager,
		app.ResourceManager,
		authManager,
		serviceRegistry,
		busProvider,
		true, // debug
	)
	require.NoError(t, err)

	// Users Setup
	users := []*config_v1.User{
		{
			Id: proto.String("user_with_auth"),
			Authentication: &config_v1.Authentication{
				AuthMethod: &config_v1.Authentication_ApiKey{
					ApiKey: &config_v1.APIKeyAuth{
						VerificationValue: ptr("user-secret"),
						In:                ptr(config_v1.APIKeyAuth_HEADER),
						ParamName:         ptr("X-API-Key"),
					},
				},
			},
			ProfileIds: []string{"profileA"},
		},
		{
			Id: proto.String("user_no_auth"),
			// No auth config, falls back to global
			ProfileIds: []string{"profileB"},
		},
		{
			Id:         proto.String("user_blocked"),
			ProfileIds: []string{}, // No access to any profile
		},
	}

	// Pre-populate users in AuthManager
	authManager.SetUsers(users)

	cachingMiddleware := middleware.NewCachingMiddleware(app.ToolManager)

	errChan := make(chan error, 1)
	go func() {
		// Pass nil for storage, it should be fine if not used by these handlers
		errChan <- app.runServerMode(ctx, mcpSrv, busProvider, bindAddress, grpcPort, 5*time.Second, nil, cachingMiddleware, app.Storage, serviceRegistry, nil)
	}()

	// Wait for server to be ready
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", bindAddress, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
		return false
	}, 2*time.Second, 100*time.Millisecond)

	baseURL := fmt.Sprintf("http://%s", bindAddress)

	// Case 1: Route not found (invalid path)
	t.Run("Invalid Path", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/mcp/u/foo")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// Case 2: User Not Found
	t.Run("User Not Found", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/mcp/u/unknown_user/profile/any")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// Case 3: User With Auth - Missing Key
	t.Run("User Auth - Missing Key", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/mcp/u/user_with_auth/profile/profileA", nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// Case 4: User With Auth - Wrong Key
	t.Run("User Auth - Wrong Key", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/mcp/u/user_with_auth/profile/profileA", nil)
		req.Header.Set("X-API-Key", "wrong-key")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// Case 5: User With Auth - Correct Key
	t.Run("User Auth - Correct Key", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/mcp/u/user_with_auth/profile/profileA", nil)
		req.Header.Set("X-API-Key", "user-secret")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		// Assuming 405 or 200 or 404 depending entirely on what handler is there
		// But definitively NOT 401/403
		assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode)
		assert.NotEqual(t, http.StatusForbidden, resp.StatusCode)
	})

	// Case 6: User No Auth - Fallback to Global - Missing Key
	t.Run("User No Auth - Global Missing", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/mcp/u/user_no_auth/profile/profileB", nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// Case 7: User No Auth - Fallback to Global - Correct Key
	t.Run("User No Auth - Global Correct", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/mcp/u/user_no_auth/profile/profileB", nil)
		req.Header.Set("X-API-Key", "global-secret")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// Case 8: Profile Access Denied
	t.Run("Profile Access Denied", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/mcp/u/user_with_auth/profile/profileB", nil) // user_with_auth only has profileA
		req.Header.Set("X-API-Key", "user-secret")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	// Case 9: Global Auth via Bearer Token
	t.Run("Global Auth Bearer", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/mcp/u/user_no_auth/profile/profileB", nil)
		req.Header.Set("Authorization", "Bearer global-secret")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// Clean up
	cancel()
	<-errChan
}

func TestAuthMiddleware_LocalhostSecurity(t *testing.T) {
	app := NewApplication()
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil) // Empty key

	// Case 1: No API Key Configured
	t.Run("No Key - Localhost Allowed", func(t *testing.T) {
		middleware := app.createAuthMiddleware() // No key
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req, _ := http.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("No Key - IPv6 Localhost Allowed", func(t *testing.T) {
		middleware := app.createAuthMiddleware() // No key
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req, _ := http.NewRequest("GET", "/", nil)
		req.RemoteAddr = "[::1]:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("No Key - Private IP Blocked", func(t *testing.T) {
		middleware := app.createAuthMiddleware() // No key
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req, _ := http.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.5:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("No Key - Public IP Denied", func(t *testing.T) {
		middleware := app.createAuthMiddleware() // No key
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req, _ := http.NewRequest("GET", "/", nil)
		req.RemoteAddr = "8.8.8.8:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	// Case 2: API Key Configured
	app.SettingsManager.Update(nil, "secret") // Set key

	t.Run("With Key - Localhost Needs Key", func(t *testing.T) {
		middleware := app.createAuthMiddleware()
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req, _ := http.NewRequest("GET", "/", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("With Key - External Needs Key", func(t *testing.T) {
		middleware := app.createAuthMiddleware()
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req, _ := http.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.5:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("With Key - External With Correct Key Allowed", func(t *testing.T) {
		middleware := app.createAuthMiddleware()
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req, _ := http.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.5:12345"
		req.Header.Set("X-API-Key", "secret")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("With Key - External With Authorization Bearer Allowed", func(t *testing.T) {
		middleware := app.createAuthMiddleware()
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req, _ := http.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.5:12345"
		req.Header.Set("Authorization", "Bearer secret")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestAuthMiddleware_AuthDisabled(t *testing.T) {
	// Setup dependencies to simulate disabled auth
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Find free port
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	bindAddress := fmt.Sprintf("localhost:%d", port)

	app := NewApplication()
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	busProvider, _ := bus.NewProvider(nil)
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager)
	authManager := auth.NewManager()

	app.AuthManager = authManager

	serviceRegistry := serviceregistry.New(
		upstreamFactory,
		app.ToolManager,
		app.PromptManager,
		app.ResourceManager,
		authManager,
	)
	mcpSrv, err := mcpserver.NewServer(
		ctx,
		app.ToolManager,
		app.PromptManager,
		app.ResourceManager,
		authManager,
		serviceRegistry,
		busProvider,
		true,
	)
	require.NoError(t, err)

	cachingMiddleware := middleware.NewCachingMiddleware(app.ToolManager)

	// Configure settings to Disable Auth
	origMiddlewares := config.GlobalSettings().Middlewares()
	defer config.GlobalSettings().SetMiddlewares(origMiddlewares)

	// Disable Auth Middleware
	config.GlobalSettings().SetMiddlewares([]*config_v1.Middleware{
		{
			Name:     proto.String("auth"),
			Priority: proto.Int32(1),
			Disabled: proto.Bool(true),
		},
	})

	errChan := make(chan error, 1)
	go func() {
		errChan <- app.runServerMode(ctx, mcpSrv, busProvider, bindAddress, "", 5*time.Second, nil, cachingMiddleware, nil, serviceRegistry, nil)
	}()

	// Wait for server to be ready
	require.Eventually(t, func() bool {
		conn, err := net.DialTimeout("tcp", bindAddress, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
		return false
	}, 2*time.Second, 100*time.Millisecond)

	baseURL := fmt.Sprintf("http://%s", bindAddress)

	// Verify Public Access is BLOCKED (StatusForbidden) because of our fix
	// Even though auth is disabled, we enforce private IP check.
	// We simulate public access by mocking RemoteAddr?
	// We can't mock RemoteAddr easily with http.Get/Client unless we use a custom Dialer or explicit Request with test setup.
	// But here we are integration testing with a real server listening on localhost.
	// So requests WILL come from localhost (127.0.0.1).
	// So they SHOULD SUCCEED.

	// Wait, if I want to test that PUBLIC access is BLOCKED, I need to send a request that LOOKS like it comes from public IP.
	// Real network stack won't let me spoof source IP easily on TCP connection (handshake won't complete).
	// But `runServerMode` uses `createAuthMiddleware` which uses `r.RemoteAddr`.

	// If I can't spoof RemoteAddr in integration test, I can at least verify that LOCALHOST access works when auth is disabled.
	// Before my fix, localhost access worked.
	// After my fix, localhost access STILL works (because it's private IP).

	// To verify the "Security" part (blocking public), I need to unit test `createAuthMiddleware("")` logic, which I already did in `TestAuthMiddleware_LocalhostSecurity`.
	// AND I need to verify that `runServerMode` actually WIRES it up.
	// By running this test and asserting Localhost works, I confirm it didn't break functionality.
	// To confirm it BLOCKS public, I would need to unit test `runServerMode` logic or rely on the fact that I proved `createAuthMiddleware("")` blocks public.

	// Ideally, I should verify the log message "Enforcing private-IP-only access" was logged.
	// But capturing logs is hard here.

	// So:
	// 1. Verify Localhost access works (200 OK or 404 Not Found, but NOT 401/403).
	// 2. We trust `TestAuthMiddleware_LocalhostSecurity` for the "Public Block" part.

	req, _ := http.NewRequest("GET", baseURL+"/healthz", nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Should be OK (200) because we are on localhost
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	cancel()
	<-errChan
}
