/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mcp

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func createTestBundle(t *testing.T, dir string) string {
	bundlePath := filepath.Join(dir, "test.mcpb")
	file, err := os.Create(bundlePath)
	require.NoError(t, err)
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

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
	upstream := NewUpstream()
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
		Name: proto.String("test-service-bundle"),
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
	// Bundle unzipping should have happened in /tmp/mcp-bundles/<serviceID>
	// We check mount
	require.Len(t, bd.Mounts, 1)
	assert.Equal(t, "/app/bundle", bd.Mounts[0].Target)
	assert.Contains(t, bd.Mounts[0].Source, "mcp-bundles")
}
