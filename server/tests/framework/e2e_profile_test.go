// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/tests/integration"
	"github.com/stretchr/testify/require"
)

func TestProfileAuthentication(t *testing.T) {
	// 1. Setup - Start Upstream Echo Server
	upstreamPort := integration.FindFreePort(t)
	// We can use the simple echo server or just a dummy listener if we only care about auth reaching the handler.
	// But to get a 200 OK from "tools/list", we need a valid backend or just use a mock.
	// StartMCPANYServerWithConfig loads services from config.
	// We can define a simplified config that points to a "stdio" mock or similar?
	// Or just point to a non-existent upstream? Auth happens BEFORE forwarding.
	// If auth fails, we get 401. If auth succeeds, we might get 502 (if upstream down) or 404.
	// To distinguish "Auth Success" from "Auth Fail", 502/404 is distinct from 401.
	// So we don't strictly need a running upstream to test AUTH logic, but having one is better for "OK" response.

	// Let's spin up a dummy HTTP server that returns 200 OK for anything.
	// This simulates the upstream.
	upstreamHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":{"tools":[]}}`))
	})
	upstreamServer := &http.Server{Handler: upstreamHandler, Addr: fmt.Sprintf(":%d", upstreamPort), ReadHeaderTimeout: 5 * time.Second}
	go func() { _ = upstreamServer.ListenAndServe() }()
	defer func() { _ = upstreamServer.Close() }()

	// 2. Define Configuration
	// 2. Define Configuration
	configJSON := fmt.Sprintf(`{
  "global_settings": {
    "log_level": "LOG_LEVEL_DEBUG",
    "profiles": ["dev", "prod"]
  },
  "users": [
    {
      "id": "alice",
      "authentication": {
        "api_key": {
          "param_name": "X-User-Key",
          "key_value": "alice-secret",
          "in": "HEADER"
        }
      },
      "profile_ids": ["dev", "prod"]
    },
    {
      "id": "bob",
      "authentication": {
        "api_key": {
          "param_name": "X-User-Key",
          "key_value": "bob-secret",
          "in": "HEADER"
        }
      },
      "profile_ids": ["dev"]
    }
  ],
  "upstream_services": [
    {
      "name": "test-service",
      "http_service": {
        "address": "http://localhost:%d"
      },
      "profiles": [
        {
          "id": "dev",
          "name": "dev",
          "authentication": {
            "api_key": {
              "param_name": "X-Profile-Key",
              "key_value": "dev-secret",
              "in": "HEADER"
            }
          }
        },
        {
          "id": "prod",
          "name": "prod",
          "authentication": {
             "api_key": {
               "param_name": "X-Profile-Key",
               "key_value": "prod-secret",
               "in": "HEADER"
             }
          }
        }
      ]
    }
  ]
}`, upstreamPort)

	// 3. Start MCPANY Server
	serverInfo := integration.StartMCPANYServerWithConfig(t, "profile_test", configJSON)
	defer serverInfo.CleanupFunc()

	time.Sleep(5 * time.Second) // Wait for server to be ready and services to be registered

	baseURL := serverInfo.HTTPEndpoint

	client := &http.Client{Timeout: 5 * time.Second}

	// Helper to make request
	checkAuth := func(t *testing.T, user, profile string, headers map[string]string, expectedStatus int, msg string) {
		// Create a valid JSON-RPC request body
		url := fmt.Sprintf("%s/u/%s/profile/%s", baseURL, user, profile)
		body := strings.NewReader(`{"jsonrpc": "2.0", "method": "tools/list", "id": 1}`)
		req, err := http.NewRequest(http.MethodPost, url, body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != expectedStatus {
			t.Errorf("%s: expected status %d, got %d", msg, expectedStatus, resp.StatusCode)
		}
	}

	t.Run("Valid_Profile_Key", func(t *testing.T) {
		checkAuth(t, "alice", "dev", map[string]string{"X-Profile-Key": "dev-secret"}, http.StatusOK, "Valid Profile Key")
	})

	t.Run("Valid_User_Key", func(t *testing.T) {
		// Expect 401 because Profile Auth is configured and takes precedence over User Auth.
		checkAuth(t, "alice", "dev", map[string]string{"X-User-Key": "alice-secret"}, http.StatusUnauthorized, "Valid User Key (Should fail due to Profile Priority)")
	})

	t.Run("Invalid_Key", func(t *testing.T) {
		checkAuth(t, "alice", "dev", map[string]string{"X-Profile-Key": "wrong"}, http.StatusUnauthorized, "Invalid key should fail")
	})

	t.Run("Mixed_Keys_(Profile_Priority)", func(t *testing.T) {
		checkAuth(t, "alice", "dev", map[string]string{"X-Profile-Key": "dev-secret", "X-User-Key": "wrong"}, http.StatusOK, "Profile Key should overwrite User Key")
	})

	t.Run("Mixed_Keys_(Profile_Wrong)", func(t *testing.T) {
		checkAuth(t, "alice", "dev", map[string]string{"X-Profile-Key": "wrong", "X-User-Key": "alice-secret"}, http.StatusUnauthorized, "Wrong Profile Key should fail even if User Key is valid")
	})

	t.Run("Access_Forbidden_Profile", func(t *testing.T) {
		// Bob is not in "prod".
		checkAuth(t, "bob", "prod", map[string]string{"X-Profile-Key": "prod-secret"}, http.StatusForbidden, "Bob should be forbidden from prod")
	})

	t.Run("Prod Profile Auth", func(t *testing.T) {
		// Alice accesses Prod with Prod key
		checkAuth(t, "alice", "prod", map[string]string{"X-Profile-Key": "prod-secret"}, http.StatusOK, "Alice accessing Prod with Prod key")
	})
}
