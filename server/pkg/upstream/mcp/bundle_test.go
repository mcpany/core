// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcp

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func createTestBundle(t *testing.T, dir string) string {
	bundlePath := filepath.Join(dir, "test.mcpb")
	file, err := os.Create(bundlePath) //nolint:gosec // Test file
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	w := zip.NewWriter(file)
	defer func() { _ = w.Close() }()

	manifest := `
{
  "manifest_version": "0.1",
  "name": "test-bundle",
  "version": "1.0",
  "server": {
    "type": "node",
    "mcp_config": {
      "command": "node",
      "args": ["index.js"]
    }
  }
}
`
	f, err := w.Create("manifest.json")
	require.NoError(t, err)
	_, err = io.WriteString(f, manifest)
	require.NoError(t, err)

	return bundlePath
}

func TestUpstream_Register_Bundle(t *testing.T) {
	tempDir := t.TempDir()
	bundlePath := createTestBundle(t, tempDir)

	toolManager := tool.NewManager(nil)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	upstream := NewUpstream(nil)
	// Override BundleBaseDir to use tempDir for isolation
	if impl, ok := upstream.(*Upstream); ok {
		impl.BundleBaseDir = filepath.Join(t.TempDir(), "mcp-bundles")
		err := os.MkdirAll(impl.BundleBaseDir, 0755)
		require.NoError(t, err)
	}

	ctx := context.Background()

	mockCS := &mockClientSession{
		listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
			return &mcp.ListToolsResult{Tools: []*mcp.Tool{{Name: "test-tool-bundle"}}}, nil
		},
	}

	// Mock Connect
	originalConnect := connectForTesting
	var capturedTransport mcp.Transport
	connectForTesting = func(_ *mcp.Client, _ context.Context, transport mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		capturedTransport = transport
		return mockCS, nil
	}
	defer func() { connectForTesting = originalConnect }()

	config := configv1.UpstreamServiceConfig_builder{
		Name:             proto.String("test-service-bundle"),
		AutoDiscoverTool: proto.Bool(true),
		McpService: configv1.McpUpstreamService_builder{
			BundleConnection: configv1.McpBundleConnection_builder{
				BundlePath: proto.String(bundlePath),
			}.Build(),
		}.Build(),
	}.Build()

	serviceID, discoveredTools, _, err := upstream.Register(ctx, config, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	expectedKey, _ := util.SanitizeServiceName("test-service-bundle")
	assert.Equal(t, expectedKey, serviceID)
	require.Len(t, discoveredTools, 1)
	assert.Equal(t, "test-tool-bundle", discoveredTools[0].GetName())

	// Verify transport
	require.IsType(t, &BundleDockerTransport{}, capturedTransport)
	bd := capturedTransport.(*BundleDockerTransport)
	assert.Equal(t, "node:18-alpine", bd.Image) // Inferred
	assert.Equal(t, "node", bd.Command)

	// Verify CopyFiles instead of Mounts
	require.Len(t, bd.Files, 1)
	assert.Equal(t, "/app/bundle", bd.Files[0].Dest)
	// We can't verify exact source path easily without exposing Upstream struct details or complex logic,
	// but we can check if it's set.
	assert.NotEmpty(t, bd.Files[0].Source)

	// Mounts should be empty now
	assert.Empty(t, bd.Mounts)
}

func TestUpstream_Register_Bundle_Failures(t *testing.T) {
	upstream := NewUpstream(nil)
	tm := tool.NewManager(nil)
	pm := prompt.NewManager()
	rm := resource.NewManager()
	ctx := context.Background()

	t.Run("missing bundle path", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test"),
			McpService: configv1.McpUpstreamService_builder{
				BundleConnection: configv1.McpBundleConnection_builder{}.Build(),
			}.Build(),
		}.Build()
		_, _, _, err := upstream.Register(ctx, config, tm, pm, rm, false)
		assert.ErrorContains(t, err, "bundle_path is required")
	})

	t.Run("bundle not found", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test"),
			McpService: configv1.McpUpstreamService_builder{
				BundleConnection: configv1.McpBundleConnection_builder{
					BundlePath: proto.String("/non/existent/path.mcpb"),
				}.Build(),
			}.Build(),
		}.Build()
		_, _, _, err := upstream.Register(ctx, config, tm, pm, rm, false)
		assert.ErrorContains(t, err, "failed to unzip bundle")
	})

	t.Run("invalid zip file", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "invalid.mcpb")
		_ = os.WriteFile(f, []byte("not a zip"), 0600)
		config := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("test"),
			McpService: configv1.McpUpstreamService_builder{
				BundleConnection: configv1.McpBundleConnection_builder{
					BundlePath: proto.String(f),
				}.Build(),
			}.Build(),
		}.Build()
		_, _, _, err := upstream.Register(ctx, config, tm, pm, rm, false)
		assert.ErrorContains(t, err, "failed to unzip bundle")
	})
}

func TestUpstream_Register_Bundle_Python(t *testing.T) {
	tempDir := t.TempDir()
	bundlePath := filepath.Join(tempDir, "python.mcpb")
	file, err := os.Create(bundlePath) //nolint:gosec
	require.NoError(t, err)
	w := zip.NewWriter(file)
	manifest := `{
		"manifest_version": "0.1",
		"name": "python-bundle",
		"version": "1.0",
		"server": {
			"type": "python",
			"entry_point": "main.py",
			"mcp_config": {}
		}
	}`
	f, _ := w.Create("manifest.json")
	_, _ = io.WriteString(f, manifest)
	_ = w.Close()
	_ = file.Close()

	upstream := NewUpstream(nil)
	if impl, ok := upstream.(*Upstream); ok {
		impl.BundleBaseDir = filepath.Join(t.TempDir(), "mcp-bundles")
		_ = os.MkdirAll(impl.BundleBaseDir, 0755)
	}

	// Mock Connect
	originalConnect := connectForTesting
	var capturedTransport mcp.Transport
	connectForTesting = func(_ *mcp.Client, _ context.Context, transport mcp.Transport, _ []mcp.Root) (ClientSession, error) {
		capturedTransport = transport
		return &mockClientSession{
			listToolsFunc: func(_ context.Context, _ *mcp.ListToolsParams) (*mcp.ListToolsResult, error) {
				return &mcp.ListToolsResult{}, nil
			},
		}, nil
	}
	defer func() { connectForTesting = originalConnect }()

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("python-service"),
		McpService: configv1.McpUpstreamService_builder{
			BundleConnection: configv1.McpBundleConnection_builder{
				BundlePath: proto.String(bundlePath),
			}.Build(),
		}.Build(),
	}.Build()

	_, _, _, err = upstream.Register(context.Background(), config, tool.NewManager(nil), prompt.NewManager(), resource.NewManager(), false)
	require.NoError(t, err)

	require.IsType(t, &BundleDockerTransport{}, capturedTransport)
	bd := capturedTransport.(*BundleDockerTransport)
	assert.Equal(t, "python:3.11-slim", bd.Image)
	assert.Equal(t, "python", bd.Command)
	assert.Contains(t, bd.Args, "/app/bundle/main.py")
}
