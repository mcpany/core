// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultiUserHandler_EdgeCases(t *testing.T) {
	// Setup app with a user
	fs := afero.NewMemMapFs()
	configContent := `
users:
  - id: "user1"
    profile_ids: ["profile1"]
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := NewApplication()

	// Start server on random ports
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	httpPort := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Run(ctx, fs, false, fmt.Sprintf("%d", httpPort), "", []string{"/config.yaml"}, "", 100*time.Millisecond, 5*time.Second)
	}()

	// Wait for server to start
	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)
	require.Eventually(t, func() bool {
		resp, err := http.Get(baseURL + "/healthz")
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond)

	client := &http.Client{Timeout: 2 * time.Second}

	// Helper to send request
	sendReq := func(method, path string, body []byte) (int, string, error) {
		req, err := http.NewRequest(method, baseURL+path, bytes.NewReader(body))
		if err != nil {
			return 0, "", err
		}
		resp, err := client.Do(req)
		if err != nil {
			return 0, "", err
		}
		defer resp.Body.Close()
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return resp.StatusCode, buf.String(), nil
	}

	// 1. Invalid Path Format
	t.Run("Invalid Path Format", func(t *testing.T) {
		code, _, err := sendReq("GET", "/mcp/u/user1/invalid", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, code)
	})

	// 2. User Not Found
	t.Run("User Not Found", func(t *testing.T) {
		code, _, err := sendReq("GET", "/mcp/u/unknown/profile/p1", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, code)
	})

	// 3. Stateless JSON-RPC Invalid JSON
	t.Run("Stateless JSON-RPC Invalid JSON", func(t *testing.T) {
		// Needs to match the path structure to reach the delegate handler
		// But verify first that we can reach it.
		// /mcp/u/user1/profile/profile1 is the path.
		// However, user1 only has profile1.
		// And profile1 is not defined in upstream services?
		// If profile definition is missing, hasAccess logic depends on user.ProfileIds.
		// Code: "for _, pid := range user.GetProfileIds() { if pid == profileID { hasAccess = true ... } }"
		// So access check passes if it's in user's list.

		// Wait, RBAC check: "if def, ok := a.ProfileManager.GetProfileDefinition(profileID); ok && len(def.RequiredRoles) > 0"
		// If profile definition is missing, it skips RBAC check. So it proceeds.

		code, _, err := sendReq("POST", "/mcp/u/user1/profile/profile1", []byte("invalid-json"))
		require.NoError(t, err)
		// It goes to delegate handler.
		// "if r.Method == http.MethodPost && (r.URL.Path == "/" || r.URL.Path == "")"
		// Wait, http.StripPrefix strips "/mcp/u/user1/profile/profile1".
		// So path seen by delegate is "" or "/".
		assert.Equal(t, http.StatusBadRequest, code)
	})

	// 4. Stateless JSON-RPC Unsupported Method
	t.Run("Stateless JSON-RPC Unsupported Method", func(t *testing.T) {
		body := `{"jsonrpc":"2.0","method":"unknown/method","id":1}`
		code, _, err := sendReq("POST", "/mcp/u/user1/profile/profile1", []byte(body))
		require.NoError(t, err)
		assert.Equal(t, http.StatusMethodNotAllowed, code)
	})

	// 5. Request Body Too Large (Simulated)
	// Hard to simulate 5MB upload in unit test easily without allocation,
	// but we can try small limit if we could configure it.
	// But limit is hardcoded `5<<20`.
	// Skipping this for now.

	// 6. User Auth configured but failed (Missing header)
	// We need to update config to require auth for a user.
	// Since we can't update config dynamically easily in this test structure without ReloadConfig,
	// We can try to use ReloadConfig.
}

func TestMultiUserHandler_UserAuth(t *testing.T) {
	fs := afero.NewMemMapFs()
	configContent := `
users:
  - id: "user_auth"
    profile_ids: ["p1"]
    authentication:
      api_key:
        param_name: "X-Key"
        verification_value: "secret"
        in: "HEADER"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := NewApplication()

	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	httpPort := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	go func() {
		app.Run(ctx, fs, false, fmt.Sprintf("%d", httpPort), "", []string{"/config.yaml"}, "", 100*time.Millisecond, 5*time.Second)
	}()

	baseURL := fmt.Sprintf("http://localhost:%d", httpPort)
	require.Eventually(t, func() bool {
		resp, err := http.Get(baseURL + "/healthz")
		if err != nil { return false }
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond)

	client := &http.Client{Timeout: 2 * time.Second}

	t.Run("Missing Auth", func(t *testing.T) {
		req, _ := http.NewRequest("GET", baseURL+"/mcp/u/user_auth/profile/p1", nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Correct Auth", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/mcp/u/user_auth/profile/p1", bytes.NewReader([]byte(`{}`))) // Valid JSON but empty method
		req.Header.Set("X-Key", "secret")
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		// Should get 405 Method Not Allowed (reached handler) instead of 401
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}

func TestRun_ConflictingPorts(t *testing.T) {
    // This tests if Run returns error when ports are already taken.
    // It's partially covered in server_test.go but adding explicit check here.
    l, err := net.Listen("tcp", "localhost:0")
    require.NoError(t, err)
    defer l.Close()
    port := l.Addr().(*net.TCPAddr).Port

    app := NewApplication()
    fs := afero.NewMemMapFs()

    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    err = app.Run(ctx, fs, false, fmt.Sprintf("%d", port), "", nil, "", 100*time.Millisecond, 5*time.Second)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "failed to start a server")
}

func TestReloadConfig_DynamicUpdates(t *testing.T) {
    // Test that middleware updates on reload (coverage for updateGlobalSettings)
    fs := afero.NewMemMapFs()
    config1 := `
global_settings:
  allowed_ips: ["127.0.0.1"]
`
    err := afero.WriteFile(fs, "/config.yaml", []byte(config1), 0o644)
    require.NoError(t, err)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    app := NewApplication()

    // We need to Mock Run or just init what we need.
    // Init minimal state
    app.SettingsManager = NewGlobalSettingsManager("", nil, nil)
    app.ToolManager = nil // Not needed for this test part?
    // Run initializes app.ipMiddleware.
    // We can manually init it.

    // Actually simpler to run full app.
    l, err := net.Listen("tcp", "localhost:0")
    require.NoError(t, err)
    port := l.Addr().(*net.TCPAddr).Port
    _ = l.Close()

    go func() {
        app.Run(ctx, fs, false, fmt.Sprintf("%d", port), "", []string{"/config.yaml"}, "", 100*time.Millisecond, 5*time.Second)
    }()

    require.NoError(t, app.WaitForStartup(ctx))

    // Check allow list
    assert.True(t, app.ipMiddleware.Allow("127.0.0.1"))
    assert.False(t, app.ipMiddleware.Allow("10.0.0.1"))

    // Update config
    config2 := `
global_settings:
  allowed_ips: ["127.0.0.1", "10.0.0.1"]
`
    err = afero.WriteFile(fs, "/config.yaml", []byte(config2), 0o644)
    require.NoError(t, err)

    err = app.ReloadConfig(ctx, fs, []string{"/config.yaml"})
    require.NoError(t, err)

    // Check allow list updated
    assert.True(t, app.ipMiddleware.Allow("10.0.0.1"))
}

// Mocking required interfaces for middleware test if needed
// But above integration test is better.

func TestMultiUserHandler_RBAC_RoleMismatch(t *testing.T) {
    // Tests that even if user has access to profile ID, if they lack required role, it fails.
    fs := afero.NewMemMapFs()
    configContent := `
global_settings:
  profile_definitions:
    - name: "admin_profile"
      required_roles: ["admin"]

users:
  - id: "user_regular"
    profile_ids: ["admin_profile"]
    roles: ["user"]
`
    err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
    require.NoError(t, err)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    app := NewApplication()

    l, err := net.Listen("tcp", "localhost:0")
    require.NoError(t, err)
    port := l.Addr().(*net.TCPAddr).Port
    _ = l.Close()

    go func() {
        app.Run(ctx, fs, false, fmt.Sprintf("%d", port), "", []string{"/config.yaml"}, "", 100*time.Millisecond, 5*time.Second)
    }()

    baseURL := fmt.Sprintf("http://localhost:%d", port)
    require.Eventually(t, func() bool {
		resp, err := http.Get(baseURL + "/healthz")
		if err != nil { return false }
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 100*time.Millisecond)

    resp, err := http.Get(fmt.Sprintf("http://localhost:%d/mcp/u/user_regular/profile/admin_profile", port))
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
