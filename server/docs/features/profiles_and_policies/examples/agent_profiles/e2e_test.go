// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	return t.base.RoundTrip(req)
}

func TestAgentProfilesE2E(t *testing.T) {
	// 1. Write config file
	configContent := `
global_settings:
  profiles:
    - "planning"
    - "executor"
  profile_definitions:
    - name: "planning"
      service_config:
        "planning-tools":
          enabled: true
        "shared-tools":
          enabled: true
    - name: "executor"
      service_config:
        "executor-tools":
          enabled: true
        "shared-tools":
          enabled: true

upstream_services:
  - id: "planning-tools"
    name: "planning-tools"
    command_line_service:
      command: "echo"
      local: true
      calls:
        "default":
          args: ["{{args}}"]
      tools:
        - name: "search_web"
          description: "Search the web"
          call_id: "default"
          input_schema: { type: object }
      resources:
        - uri: "planning://guidelines"
          name: "planning-guidelines"
          static:
            text_content: "Always plan before acting."

  - id: "executor-tools"
    name: "executor-tools"
    command_line_service:
      command: "echo"
      local: true
      calls:
        "default":
          args: ["{{args}}"]
      tools:
        - name: "run_code"
          description: "Run code"
          call_id: "default"
          input_schema: { type: object }
      prompts:
        - name: "fix_bug"
          description: "Fix a bug"

  - id: "shared-tools"
    name: "shared-tools"
    command_line_service:
      command: "echo"
      local: true
      calls:
        "default":
          args: ["{{args}}"]
      tools:
        - name: "ping"
          description: "Ping service"
          call_id: "default"
          input_schema: { type: object }
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// 2. Setup Server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fs := afero.NewOsFs()
	store := config.NewFileStore(fs, []string{configPath})
	cfg, err := config.LoadServices(context.Background(), store, "server")
	require.NoError(t, err)

	busProvider, err := bus.NewProvider(nil)
	require.NoError(t, err)

	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	authManager := auth.NewManager()
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, nil)
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, promptManager, resourceManager, authManager)

	mcpServer, err := mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, true)
	require.NoError(t, err)

	// Set Profiles
	if cfg.GetGlobalSettings() != nil {
		toolManager.SetProfiles(cfg.GetGlobalSettings().GetProfiles(), cfg.GetGlobalSettings().GetProfileDefinitions())
	}

	// Register Services
	for _, serviceConfig := range cfg.GetUpstreamServices() {
		upstream, err := upstreamFactory.NewUpstream(serviceConfig)
		require.NoError(t, err)
		_, _, _, err = upstream.Register(ctx, serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		// Register already adds ServiceInfo, no need to add again manually
	}

	// Debug: Print all registered tools
	allTools := toolManager.ListTools()
	t.Logf("DEBUG: All registered tools: %d", len(allTools))
	for _, tool := range allTools {
		t.Logf("DEBUG: Tool: %s (Service: %s)", tool.Tool().GetName(), tool.Tool().GetServiceId())
	}

	// 3. Setup HTTP Server with Profile Injection
	sdkHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return mcpServer.Server()
	}, nil)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		profileID := r.Header.Get("X-Test-Profile-ID")
		if profileID != "" {
			reqCtx := auth.ContextWithProfileID(r.Context(), profileID)
			r = r.WithContext(reqCtx)
		}
		sdkHandler.ServeHTTP(w, r)
	}))
	defer testServer.Close()

	// Helper to check visibility via Client
	checkVisibility := func(t *testing.T, profileID string, expectedTools, expectedResources, expectedPrompts []string) {
		clientTransport := &mcp.StreamableClientTransport{
			Endpoint: testServer.URL,
			HTTPClient: &http.Client{
				Transport: &headerTransport{
					base:    http.DefaultTransport,
					headers: map[string]string{"X-Test-Profile-ID": profileID},
				},
				Timeout: 5 * time.Second,
			},
		}

		client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
		session, err := client.Connect(ctx, clientTransport, nil)
		require.NoError(t, err)
		defer session.Close()

		// 1. List Tools
		toolsResult, err := session.ListTools(ctx, &mcp.ListToolsParams{})
		require.NoError(t, err)
		var visibleTools []string
		for _, tool := range toolsResult.Tools {
			visibleTools = append(visibleTools, tool.Name)
		}
		assert.ElementsMatch(t, expectedTools, visibleTools, "Tools mismatch for profile %s", profileID)

		// 2. List Resources
		resResult, err := session.ListResources(ctx, &mcp.ListResourcesParams{})
		require.NoError(t, err)
		var visibleResources []string
		for _, r := range resResult.Resources {
			visibleResources = append(visibleResources, r.Name)
		}
		assert.ElementsMatch(t, expectedResources, visibleResources, "Resources mismatch for profile %s", profileID)

		// 3. List Prompts
		promptsResult, err := session.ListPrompts(ctx, &mcp.ListPromptsParams{})
		require.NoError(t, err)
		var visiblePrompts []string
		for _, p := range promptsResult.Prompts {
			visiblePrompts = append(visiblePrompts, p.Name)
		}
		assert.ElementsMatch(t, expectedPrompts, visiblePrompts, "Prompts mismatch for profile %s", profileID)
	}

	t.Run("Planning_Profile", func(t *testing.T) {
		checkVisibility(t, "planning",
			[]string{"planning-tools.search_web", "shared-tools.ping"},
			[]string{"planning-guidelines"},
			[]string{},
		)
	})

	t.Run("Executor_Profile", func(t *testing.T) {
		checkVisibility(t, "executor",
			[]string{"executor-tools.run_code", "shared-tools.ping"},
			[]string{},
			[]string{"executor-tools.fix_bug"},
		)
	})
}
