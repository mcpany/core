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

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	config_v1 "github.com/mcpany/core/proto/config/v1"
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

	// Global Key
	globalKey := "global-secret"
	config.GlobalSettings().SetAPIKey(globalKey)
	defer config.GlobalSettings().SetAPIKey("")

	authManager := auth.NewManager()
	authManager.SetAPIKey(globalKey)

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
			Authentication: &config_v1.AuthenticationConfig{
				AuthMethod: &config_v1.AuthenticationConfig_ApiKey{
					ApiKey: &config_v1.APIKeyAuth{
						KeyValue:  ptr("user-secret"),
						In:        ptr(config_v1.APIKeyAuth_HEADER),
						ParamName: ptr("X-API-Key"),
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

	errChan := make(chan error, 1)
	go func() {
		errChan <- app.runServerMode(ctx, mcpSrv, busProvider, bindAddress, grpcPort, 5*time.Second, users, nil, nil, nil, nil, nil, serviceRegistry)
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
		req.Header.Set("X-API-Key", globalKey)
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
		req.Header.Set("Authorization", "Bearer "+globalKey)
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

	// Case 1: No API Key Configured
	t.Run("No Key - Localhost Allowed", func(t *testing.T) {
		middleware := app.createAuthMiddleware("") // No key
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
		middleware := app.createAuthMiddleware("") // No key
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req, _ := http.NewRequest("GET", "/", nil)
		req.RemoteAddr = "[::1]:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("No Key - Private IP Allowed", func(t *testing.T) {
		middleware := app.createAuthMiddleware("") // No key
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req, _ := http.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.5:12345"
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("No Key - Public IP Denied", func(t *testing.T) {
		middleware := app.createAuthMiddleware("") // No key
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
	t.Run("With Key - Localhost Needs Key", func(t *testing.T) {
		middleware := app.createAuthMiddleware("secret")
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
		middleware := app.createAuthMiddleware("secret")
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
		middleware := app.createAuthMiddleware("secret")
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
}
