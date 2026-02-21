// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package upstream

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/mcp"
	"github.com/mcpany/core/server/pkg/util"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// MockDockerClient implements mcp.DockerClient for testing without real Docker
type MockDockerClient struct {
	conn2 net.Conn // Server side connection
}

func (m *MockDockerClient) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	return container.CreateResponse{ID: "mock-container-id"}, nil
}

func (m *MockDockerClient) ContainerAttach(ctx context.Context, containerID string, options container.AttachOptions) (types.HijackedResponse, error) {
	c1, c2 := net.Pipe()
	m.conn2 = c2

	go m.serve()

	return types.HijackedResponse{
		Conn:   c1,
		Reader: bufio.NewReader(c1),
	}, nil
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	return nil
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	if m.conn2 != nil {
		_ = m.conn2.Close()
	}
	return nil
}

func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	return nil
}

func (m *MockDockerClient) Close() error {
	return nil
}

func (m *MockDockerClient) serve() {
	if m.conn2 == nil {
		return
	}
	defer m.conn2.Close()
	scanner := bufio.NewScanner(m.conn2)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var req struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      interface{}     `json:"id"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			continue
		}

		// Simple mock logic matching simpleFsServerJS
		var result interface{}

		if req.Method == "initialize" {
			result = map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
				"serverInfo": map[string]interface{}{
					"name":    "simple-fs",
					"version": "1.0.0",
				},
			}
		} else if req.Method == "tools/list" || req.Method == "mcp.listTools" {
			result = map[string]interface{}{
				"tools": []map[string]interface{}{
					{
						"name": "list_directory",
						"description": "List files in a directory",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"path": map[string]interface{}{"type": "string"},
							},
							"required": []string{"path"},
						},
					},
					{
						"name": "read_file",
						"description": "Read a file",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"path": map[string]interface{}{"type": "string"},
							},
							"required": []string{"path"},
						},
					},
				},
			}
		} else if req.Method == "tools/call" || req.Method == "mcp.callTool" {
			// Try manual extraction
			var rawParams map[string]interface{}
			_ = json.Unmarshal(req.Params, &rawParams)
			name, _ := rawParams["name"].(string)
			// args handling is tricky with json.RawMessage

			if name == "read_file" {
				result = map[string]interface{}{
					"content": []map[string]interface{}{
						{"type": "text", "text": "Hello MCP Bundle!"},
					},
				}
			} else if name == "list_directory" {
				files := []string{"manifest.json", "server.js", "hello.txt"}
				filesJSON, _ := json.Marshal(files)
				result = map[string]interface{}{
					"content": []map[string]interface{}{
						{"type": "text", "text": string(filesJSON)},
					},
				}
			} else {
				// Error
			}
		} else {
			result = map[string]interface{}{}
		}

		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  result,
		}
		b, _ := json.Marshal(resp)
		// Wrap in Docker multiplexing header (Stream 1 = Stdout)
		header := make([]byte, 8)
		header[0] = 1 // Stdout
		payload := append(b, '\n')
		binary.BigEndian.PutUint32(header[4:], uint32(len(payload)))
		m.conn2.Write(header)
		m.conn2.Write(payload)
	}
}

const manifestJSON = `
{
  "manifest_version": "0.1",
  "name": "simple-fs-bundle",
  "version": "1.0",
  "server": {
    "type": "node",
    "entry_point": "server.js",
    "mcp_config": {

    }
  }
}
`

// No serverJS needed for mock client test as we simulate responses

func createE2EBundle(t *testing.T, dir string) string {
	bundlePath := filepath.Join(dir, "e2e_test.mcpb")
	file, err := os.Create(bundlePath) //nolint:gosec // Test file
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	w := zip.NewWriter(file)
	defer func() { _ = w.Close() }()

	// Add manifest.json
	f, err := w.Create("manifest.json")
	require.NoError(t, err)
	_, err = io.WriteString(f, manifestJSON)
	require.NoError(t, err)

	// Add server.js (dummy)
	f, err = w.Create("server.js")
	require.NoError(t, err)
	_, err = io.WriteString(f, "// Dummy")
	require.NoError(t, err)

	return bundlePath
}

func TestE2E_Bundle_Filesystem(t *testing.T) {
	// Use Mock Docker Client
	mcp.SetNewDockerClientForTesting(func(ops ...client.Opt) (mcp.DockerClient, error) {
		return &MockDockerClient{}, nil
	})
	// We should reset it, but unexported var... assuming test isolation or subsequent tests don't rely on real docker.
	// We can't reset easily. But other tests likely use NewClientWithOpts if they don't override.
	// Since we set the factory function, subsequent tests will use it if they use BundleDockerTransport.
	// This might break other tests if they expect real docker.
	// Ideally we set it to something that calls real client at the end.
	defer mcp.SetNewDockerClientForTesting(func(ops ...client.Opt) (mcp.DockerClient, error) {
		return client.NewClientWithOpts(ops...)
	})

	tempDir := t.TempDir()
	bundlePath := createE2EBundle(t, tempDir)

	toolManager := tool.NewManager(nil)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	upstreamService := mcp.NewUpstream(nil)
	if impl, ok := upstreamService.(*mcp.Upstream); ok {
		impl.BundleBaseDir = filepath.Join(t.TempDir(), "bundles")
		if err := os.MkdirAll(impl.BundleBaseDir, 0755); err != nil {
			t.Fatalf("Failed to create test bundle dir: %v", err)
		}
	}
	ctx := context.Background()

	config := configv1.UpstreamServiceConfig_builder{
		Name:             proto.String("fs-bundle-service"),
		AutoDiscoverTool: proto.Bool(true),
		McpService: configv1.McpUpstreamService_builder{
			BundleConnection: configv1.McpBundleConnection_builder{
				BundlePath: proto.String(bundlePath),
			}.Build(),
		}.Build(),
	}.Build()

	serviceID, discoveredTools, _, err := upstreamService.Register(ctx, config, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)
	expectedKey, _ := util.SanitizeServiceName("fs-bundle-service")
	assert.Equal(t, expectedKey, serviceID)

	// Check if tools were discovered
	assert.GreaterOrEqual(t, len(discoveredTools), 2)

	toolNames := make(map[string]bool)
	for _, def := range discoveredTools {
		toolNames[def.GetName()] = true
	}
	assert.Contains(t, toolNames, "list_directory")
	assert.Contains(t, toolNames, "read_file")

	// Call tool
	sanitizedToolName, _ := util.SanitizeToolName("read_file")
	toolID := serviceID + "." + sanitizedToolName

	mcpTool, ok := toolManager.GetTool(toolID)
	require.True(t, ok, "Tool should be registered: %s", toolID)

	callArgs := json.RawMessage(`{"path": "/app/bundle/hello.txt"}`)
	req := &tool.ExecutionRequest{
		ToolName:   toolID,
		ToolInputs: callArgs,
	}
	result, err := mcpTool.Execute(ctx, req)

	require.NoError(t, err)

	resultStr, ok := result.(string)
	require.True(t, ok, "Result should be a string")
	assert.Equal(t, "Hello MCP Bundle!", resultStr)
}
