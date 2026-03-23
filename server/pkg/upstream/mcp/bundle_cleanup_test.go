// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func createMockBundle(t *testing.T, path string) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create bundle file: %v", err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	// manifest.json
	manifest := `
{
  "manifest_version": "1.0",
  "name": "test-bundle",
  "version": "1.0.0",
  "description": "Test Bundle",
  "server": {
    "type": "python",
    "mcp_config": {
      "command": "python",
      "args": ["main.py"]
    }
  }
}
`
	mf, err := w.Create("manifest.json")
	if err != nil {
		t.Fatalf("failed to create manifest in zip: %v", err)
	}
	_, _ = mf.Write([]byte(manifest))

	// main.py
	mainPy, err := w.Create("main.py")
	if err != nil {
		t.Fatalf("failed to create main.py in zip: %v", err)
	}
	_, _ = mainPy.Write([]byte("print('hello')"))
}

func TestBundleCleanup(t *testing.T) {
	// Create a temp bundle file
	tmpBundleDir, err := os.MkdirTemp("", "bundle-source")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpBundleDir)

	bundlePath := filepath.Join(tmpBundleDir, "bundle.zip")
	createMockBundle(t, bundlePath)

	u := NewUpstream(nil).(*Upstream)
	ctx := context.Background()

	busProvider, _ := bus.NewProvider(nil)
	toolManager := tool.NewManager(busProvider)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()

	bundleConn := configv1.McpBundleConnection_builder{}.Build()
	bundleConn.SetBundlePath(bundlePath)

	mcpSvc := configv1.McpUpstreamService_builder{}.Build()
	mcpSvc.SetBundleConnection(bundleConn)

	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	serviceConfig.SetName("test-bundle-service")
	serviceConfig.SetMcpService(mcpSvc)

	// Set mock connection
	SetConnectForTesting(func(_ *mcp.Client, _ context.Context, _ mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		return &CleanupMockClientSession{}, nil
	})
	defer SetConnectForTesting(nil)

	// Register
	serviceID, _, _, err := u.Register(ctx, serviceConfig, toolManager, promptManager, resourceManager, false)
	assert.NoError(t, err)

	// Verify temp dir exists
	expectedTempDir := filepath.Join(bundleBaseDir, serviceID)
	_, err = os.Stat(expectedTempDir)
	assert.NoError(t, err, "Temp dir should exist after registration")

	// Shutdown
	err = u.Shutdown(ctx)
	assert.NoError(t, err)

	// Verify temp dir is gone
	_, err = os.Stat(expectedTempDir)
	assert.True(t, os.IsNotExist(err), "Temp dir should be removed after shutdown")
}

// CleanupMockClientSession needed for interface satisfaction
type CleanupMockClientSession struct {}

func (m *CleanupMockClientSession) ListTools(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
	return &mcp.ListToolsResult{}, nil
}

func (m *CleanupMockClientSession) ListPrompts(_ context.Context, _ *mcp.ListPromptsParams) (*mcp.ListPromptsResult, error) {
	return &mcp.ListPromptsResult{}, nil
}

func (m *CleanupMockClientSession) ListResources(_ context.Context, _ *mcp.ListResourcesParams) (*mcp.ListResourcesResult, error) {
	return &mcp.ListResourcesResult{}, nil
}

func (m *CleanupMockClientSession) GetPrompt(_ context.Context, _ *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
	return nil, nil
}

func (m *CleanupMockClientSession) ReadResource(_ context.Context, _ *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	return nil, nil
}

func (m *CleanupMockClientSession) CallTool(_ context.Context, _ *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	return nil, nil
}

func (m *CleanupMockClientSession) Close() error { return nil }
