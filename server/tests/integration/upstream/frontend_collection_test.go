// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package upstream

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/command"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrontendReactCollection(t *testing.T) {
	// 1. Load config
	fs := afero.NewOsFs()
	configPath := "../../../../marketplace/upstream_service_collection/frontend_react.yaml"

	// Verify file exists
	_, err := fs.Stat(configPath)
	require.NoError(t, err, "Config file not found at %s", configPath)

	store := config.NewFileStore(fs, []string{configPath})
	cfg, err := store.Load(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// 2. Set up managers
	toolManager := tool.NewManager(nil)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()

	// 3. Register npm service
	cmdUpstream := command.NewUpstream()
	ctx := context.Background()

	foundNpm := false
	foundNode := false

	// Check if npm is installed locally
	hasNpm := false
	if _, err := exec.LookPath("npm"); err == nil {
		hasNpm = true
	} else {
		t.Log("npm not found locally, skipping execution tests")
	}

	for _, svc := range cfg.GetUpstreamServices() {
		if svc.GetName() == "npm-service-wrapper" {
			foundNpm = true
			_, tools, _, err := cmdUpstream.Register(ctx, svc, toolManager, promptManager, resourceManager, false)
			require.NoError(t, err)
			assert.NotEmpty(t, tools)

			// Verify tools are registered
			checkTools := []string{"npm_install", "npm_run", "npm_exec"}
			for _, toolName := range checkTools {
				// Name is sanitized? usually "npm_service_wrapper.npm_install"
				// Register returns sanitizedName which we need.
				// But we can check discoveredTools.
				found := false
				for _, d := range tools {
					if d.GetName() == toolName {
						found = true
						break
					}
				}
				assert.True(t, found, "Tool %s should be discovered", toolName)
			}

			if hasNpm {
				// Execute npm_exec --version
				// Service ID is sanitized name.
				// "npm-service-wrapper" -> "npm_service_wrapper" (likely)
				// Actually Register returns serviceID.
				// Since we called Register inside the loop and didn't capture return per service carefully,
				// we relied on it not erroring.
				// We need to look up the tool in toolManager.

				// Let's rely on iteration over tools in toolManager if needed, or guess ID.
				// util.SanitizeServiceName("npm-service-wrapper") -> "npm_service_wrapper"

				toolID := svc.GetSanitizedName() + ".npm_exec"
				tTool, ok := toolManager.GetTool(toolID)
				require.True(t, ok, "Tool %s should be in manager", toolID)

				// Create inputs: args=["root"] to avoid argument injection check (disallows flags starting with -)
				inputs := json.RawMessage(`{"args": ["root"]}`)
				req := &tool.ExecutionRequest{
					ToolName:   toolID,
					ToolInputs: inputs,
				}

				res, err := tTool.Execute(ctx, req)
				require.NoError(t, err)

				// Result is map for CommandTool
				resMap, ok := res.(map[string]interface{})
				require.True(t, ok, "Result should be map")

				stdout, _ := resMap["stdout"].(string)
				t.Logf("npm root output: %s", stdout)
				assert.NotEmpty(t, stdout)
			}
		}
		if svc.GetName() == "node-service-wrapper" {
			foundNode = true
			_, tools, _, err := cmdUpstream.Register(ctx, svc, toolManager, promptManager, resourceManager, false)
			require.NoError(t, err)
			assert.NotEmpty(t, tools)
		}
	}
	assert.True(t, foundNpm, "npm-service-wrapper not found in config")
	assert.True(t, foundNode, "node-service-wrapper not found in config")
}
