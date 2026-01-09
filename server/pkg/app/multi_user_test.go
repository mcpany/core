// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spf13/viper"
)

func TestMultiUserToolFiltering(t *testing.T) {
	// Setup mocks and server
	// Setup mocks and server
	// No bus config needed as it defaults to InMemory

	// Define profiles
	// profileDev := "dev-profile"
	// profileProd := "prod-profile"

	// Define users
	// userDev := &configv1.User{
	// 	Id:         proto.String("user-dev"),
	// 	ApiKey:     proto.String("key-dev"),
	// 	ProfileIds: []string{profileDev},
	// }
	// userProd := &configv1.User{
	// 	Id:         proto.String("user-prod"),
	// 	ApiKey:     proto.String("key-prod"),
	// 	ProfileIds: []string{profileProd, profileDev}, // Prod user has access to both
	// }

	// Create Config with users
	fs := afero.NewMemMapFs()
		configContent := `
global_settings:
  profile_definitions:
    - name: "dev-profile"
      service_config:
        dev-service: {enabled: true}
        shared-service: {enabled: true}
    - name: "prod-profile"
      service_config:
        prod-service: {enabled: true}
        shared-service: {enabled: true}
    - name: "secure-profile"
      service_config:
        secure-service: {enabled: true}
    - name: "rbac-profile"
      required_roles: ["admin"]
      service_config:
        rbac-service: {enabled: true}
      selector:
        tags: ["rbac"]

upstream_services:
  - name: "dev-service"
    id: "dev-service"
    http_service:
      address: "http://localhost:8081"
      tools:
        - name: "dev-tool"
          call_id: "dev-call"
      calls:
        dev-call:
          id: "dev-call"
          endpoint_path: "/dev"
          method: "HTTP_METHOD_POST"
  - name: "prod-service"
    id: "prod-service"
    http_service:
      address: "http://localhost:8082"
      tools:
        - name: "prod-tool"
          call_id: "prod-call"
      calls:
        prod-call:
          id: "prod-call"
          endpoint_path: "/prod"
          method: "HTTP_METHOD_POST"
  - name: "secure-service"
    id: "secure-service"
    http_service:
      address: "http://localhost:8084"
      tools:
        - name: "secure-tool"
          call_id: "secure-call"
      calls:
        secure-call:
          id: "secure-call"
          endpoint_path: "/secure"
          method: "HTTP_METHOD_POST"
  - name: "shared-service"
    id: "shared-service"
    http_service:
      address: "http://localhost:8083"
      tools:
        - name: "shared-tool"
          call_id: "shared-call"
      calls:
        shared-call:
          id: "shared-call"
          endpoint_path: "/shared"
          method: "HTTP_METHOD_POST"
  - name: "rbac-service"
    id: "rbac-service"
    http_service:
      address: "http://localhost:8085"
      tools:
        - name: "rbac-tool"
          call_id: "rbac-call"
      calls:
        rbac-call:
          id: "rbac-call"
          endpoint_path: "/rbac"
          method: "HTTP_METHOD_POST"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set enabled profiles for the server instance so it loads all services
	viper.Set("profiles", []string{"dev", "prod", "secure", "rbac", "default"})
	defer viper.Set("profiles", nil)

	app := NewApplication()

	// Start server on random ports
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	httpPort := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	l2, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	grpcPort := l2.Addr().(*net.TCPAddr).Port
	_ = l2.Close()

	configContentWithUsers := configContent + `
users:
  - id: "user-dev"
    authentication:
      api_key:
        param_name: "X-User-Key"
        verification_value: "key-dev"
        in: "HEADER"
    profile_ids: ["dev-profile"]
  - id: "user-prod"
    authentication:
      api_key:
        param_name: "X-User-Key"
        verification_value: "key-prod"
        in: "HEADER"
    profile_ids: ["prod-profile", "dev-profile"]
  - id: "user-secure"
    authentication:
      api_key:
        param_name: "X-User-Key"
        verification_value: "key-secure-user"
        in: "HEADER"
    profile_ids: ["secure-profile"]
  - id: "user-admin"
    authentication:
      api_key:
        param_name: "X-User-Key"
        verification_value: "key-admin"
        in: "HEADER"
    profile_ids: ["rbac-profile"]
    roles: ["admin"]
  - id: "user-guest"
    authentication:
      api_key:
        param_name: "X-User-Key"
        verification_value: "key-guest"
        in: "HEADER"
    profile_ids: ["rbac-profile"]
    roles: ["guest"]
`
	err = afero.WriteFile(fs, "/config.yaml", []byte(configContentWithUsers), 0o644)
	require.NoError(t, err)

	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Run(ctx, fs, false, fmt.Sprintf("%d", httpPort), fmt.Sprintf("%d", grpcPort), []string{"/config.yaml"}, "", 5*time.Second)
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

	// Helper to calls tools/list via SSE
	listTools := func(uid, profileID, headerName, apiKey string) ([]string, error) {
		url := fmt.Sprintf("%s/mcp/u/%s/profile/%s", baseURL, uid, profileID)

		reqBody := `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "tools/list",
			"params": {}
		}`

		req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(reqBody)))
		if err != nil {
			return nil, err
		}

		req.Header.Set(headerName, apiKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("status code: %d", resp.StatusCode)
		}

		// Parse response
		// Response is JSON-RPC: {"jsonrpc": "2.0", "result": { "tools": [...] }, "id": 1}
		var rpcResp struct {
			Result struct {
				Tools []struct {
					Name string `json:"name"`
				} `json:"tools"`
			} `json:"result"`
			Error *struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
			return nil, err
		}

		if rpcResp.Error != nil {
			return nil, fmt.Errorf("rpc error: %s", rpcResp.Error.Message)
		}

		var toolNames []string
		for _, t := range rpcResp.Result.Tools {
			toolNames = append(toolNames, t.Name)
		}
		return toolNames, nil
	}

	// Test Case 1: Dev User (Profile: dev-profile)
	// Should see: dev-tool, shared-tool (if shared logic allows)
	// Shared tool has NO profiles. If logic is "allow if no profiles", then yes.
	// If logic is "restrict to explicit profiles", then no.
	// In server.go I implemented:
	// "if len(info.Config.GetProfiles()) == 0 { hasAccess = true }"
	// So shared-tool should be visible.

	t.Run("UserDev sees dev tools and shared tools", func(t *testing.T) {
		tools, err := listTools("user-dev", "dev-profile", "X-User-Key", "key-dev")
		require.NoError(t, err)
		assert.Contains(t, tools, "dev-tool")
		assert.Contains(t, tools, "shared-tool")
		assert.NotContains(t, tools, "prod-tool")
	})

	// Test Case 2: Prod User (Profile: prod-profile)
	// Should see: prod-tool, shared-tool. Not dev-tool.
	t.Run("UserProd (prod-profile) sees prod tools and shared tools", func(t *testing.T) {
		tools, err := listTools("user-prod", "prod-profile", "X-User-Key", "key-prod")
		require.NoError(t, err)
		assert.Contains(t, tools, "prod-tool")
		assert.Contains(t, tools, "shared-tool")
		assert.NotContains(t, tools, "dev-tool")
	})

	// Test Case 3: Prod User accessing Dev Profile (allowed)
	// Should see: dev-tool, shared-tool.
	t.Run("UserProd accessing dev-profile sees dev tools", func(t *testing.T) {
		tools, err := listTools("user-prod", "dev-profile", "X-User-Key", "key-prod")
		require.NoError(t, err)
		assert.Contains(t, tools, "dev-tool")
		assert.Contains(t, tools, "shared-tool")
		assert.NotContains(t, tools, "prod-tool")
	})


	// Test Case 4: Invalid Auth
	t.Run("Invalid API Key fails", func(t *testing.T) {
		_, err := listTools("user-dev", "dev-profile", "X-User-Key", "wrong-key")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status code: 401")
	})

	// Test Case 5: Invalid Profile Access
	// Dev User accessing Prod Profile (not allowed)
	t.Run("UserDev accessing prod-profile fails", func(t *testing.T) {
		// Wait, if profile is not in user's list, does it return 404 or 403?
		// Logic: "for _, pid := range user.ProfileIds { if pid == profileID { hasAccess = true ... } }"
		// If check fails -> http.Error(w, "Forbidden", http.StatusForbidden)
		// I missed what happens if profileID is not found in loop.
		// I should verify server.go logic handles this denial.
		// Assuming it returns 403.
		_, err := listTools("user-dev", "prod-profile", "X-User-Key", "key-dev")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status code: 403") // Forbidden
	})

	// Test Case 6: RBAC Access
	// User Admin (role: admin) accessing rbac-profile (requires: admin) -> Allowed
	t.Run("UserAdmin accessing rbac-profile allowed", func(t *testing.T) {
		tools, err := listTools("user-admin", "rbac-profile", "X-User-Key", "key-admin")
		require.NoError(t, err)
		assert.Contains(t, tools, "rbac-tool")
	})

	// Test Case 7: RBAC Denial
	// User Guest (role: guest) accessing rbac-profile (requires: admin) -> Denied
	t.Run("UserGuest accessing rbac-profile denied", func(t *testing.T) {
		_, err := listTools("user-guest", "rbac-profile", "X-User-Key", "key-guest")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "status code: 403")
	})
}
