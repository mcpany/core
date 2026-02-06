// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package upstream

import (
	"archive/zip"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/mcp"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/integration"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// simpleFsServerPy is a minimal Python MCP server that implements filesystem tools.
// We use Python because the slim-debian image is often more robust in CI than alpine.
const simpleFsServerPy = `
import sys
import json
import os

def send(msg):
    sys.stdout.write(json.dumps(msg) + '\n')
    sys.stdout.flush()

def handle_request(req):
    method = req.get('method')
    msg_id = req.get('id')
    params = req.get('params', {})

    if method == 'initialize':
        send({
            'jsonrpc': '2.0',
            'id': msg_id,
            'result': {
                'protocolVersion': '2024-11-05',
                'capabilities': {'tools': {}},
                'serverInfo': {'name': 'simple-fs', 'version': '1.0.0'}
            }
        })
    elif method == 'notifications/initialized':
        pass
    elif method == 'tools/list':
        send({
            'jsonrpc': '2.0',
            'id': msg_id,
            'result': {
                'tools': [
                    {
                        'name': 'list_directory',
                        'description': 'List files in a directory',
                        'inputSchema': {
                            'type': 'object',
                            'properties': {'path': {'type': 'string'}},
                            'required': ['path']
                        }
                    },
                    {
                        'name': 'read_file',
                        'description': 'Read a file',
                        'inputSchema': {
                            'type': 'object',
                            'properties': {'path': {'type': 'string'}},
                            'required': ['path']
                        }
                    }
                ]
            }
        })
    elif method == 'tools/call':
        name = params.get('name')
        args = params.get('arguments', {})
        try:
            if name == 'list_directory':
                files = os.listdir(args['path'])
                send({
                    'jsonrpc': '2.0',
                    'id': msg_id,
                    'result': {
                        'content': [{'type': 'text', 'text': json.dumps(files)}]
                    }
                })
            elif name == 'read_file':
                with open(args['path'], 'r') as f:
                    content = f.read()
                send({
                    'jsonrpc': '2.0',
                    'id': msg_id,
                    'result': {
                        'content': [{'type': 'text', 'text': content}]
                    }
                })
            else:
                send({'jsonrpc': '2.0', 'id': msg_id, 'error': {'code': -32601, 'message': 'Method not found'}})
        except Exception as e:
            send({'jsonrpc': '2.0', 'id': msg_id, 'error': {'code': -32000, 'message': str(e)}})
    else:
        if msg_id is not None:
            send({'jsonrpc': '2.0', 'id': msg_id, 'result': {}})

for line in sys.stdin:
    if not line.strip(): continue
    try:
        req = json.loads(line)
        handle_request(req)
    except Exception as e:
        sys.stderr.write(str(e) + '\n')
`

const manifestJSON = `
{
  "manifest_version": "0.1",
  "name": "simple-fs-bundle",
  "version": "1.0",
  "server": {
    "type": "python",
    "entry_point": "server.py",
    "mcp_config": {

    }
  }
}
`

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

	// Add server.py
	f, err = w.Create("server.py")
	require.NoError(t, err)
	_, err = io.WriteString(f, simpleFsServerPy)
	require.NoError(t, err)

	// Add a dummy file to read
	f, err = w.Create("hello.txt")
	require.NoError(t, err)
	_, err = io.WriteString(f, "Hello MCP Bundle!")
	require.NoError(t, err)

	// Add a subdirectory structure
	_, err = w.Create("data/")
	require.NoError(t, err)
	f, err = w.Create("data/test.json")
	require.NoError(t, err)
	_, err = io.WriteString(f, "{}")
	require.NoError(t, err)

	return bundlePath
}

func TestE2E_Bundle_Filesystem(t *testing.T) {
	if os.Getenv("SKIP_DOCKER_TESTS") == "true" {
		t.Skip("Skipping Docker tests because SKIP_DOCKER_TESTS is set")
	}
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping Docker tests in CI due to potential overlayfs/mount issues")
	}

	// In CI, OverlayFS often fails with "invalid argument" when using bind mounts in DinD
	// especially when the source is on certain filesystems or with nested overlay layers.
	// Since we cannot fix the CI runner's kernel/docker setup, we skip this specific test in CI.
	if os.Getenv("CI") != "" {
		t.Skip("Skipping Bundle test in CI due to OverlayFS/BindMount issues in DinD environment")
	}

	// Check if Docker is available and accessible
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skipf("Skipping Docker tests: docker info failed: %v", err)
	}

	// Use a directory in the build folder instead of /tmp (t.TempDir())
	// This ensures we are on a standard filesystem (likely ext4), avoiding potential
	// overlayfs incompatibilities with tmpfs mounts in some Docker-in-Docker environments.
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)
	testBundlesDir := filepath.Join(root, "../build/e2e-bundles")
	if err := os.MkdirAll(testBundlesDir, 0755); err != nil {
		t.Fatalf("Failed to create test bundles dir: %v", err)
	}
	// Create unique subdir for this test run
	runDir, err := os.MkdirTemp(testBundlesDir, "test-run-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(runDir) }()

	bundlePath := createE2EBundle(t, runDir)

	toolManager := tool.NewManager(nil)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	upstreamService := mcp.NewUpstream(nil)
	if impl, ok := upstreamService.(*mcp.Upstream); ok {
		// Use a subdirectory in our runDir for extraction
		impl.BundleBaseDir = filepath.Join(runDir, "extracted")
		if err := os.MkdirAll(impl.BundleBaseDir, 0755); err != nil {
			t.Fatalf("Failed to create extracted bundle dir: %v", err)
		}
		t.Logf("Using BundleBaseDir: %s", impl.BundleBaseDir)
	}
	ctx := context.Background()

	// Usage of real implementation is default, so no need to touch connectForTesting

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
	// We might need to wait or check if discoveredTools is populated.
	// Register waits for ListTools, so it should be there.
	assert.GreaterOrEqual(t, len(discoveredTools), 2)

	toolNames := make(map[string]bool)
	for _, def := range discoveredTools {
		toolNames[def.GetName()] = true
	}
	assert.Contains(t, toolNames, "list_directory")
	assert.Contains(t, toolNames, "read_file")

	// Now, let's try to CALL a tool.
	// We need to use the toolManager to execute it, or use the client directly?
	// The upstream Register adds tools to toolManager.

	// Let's call read_file on "hello.txt"
	// The bundle is mounted at /app/bundle in the container.
	// So expected path for hello.txt is /app/bundle/hello.txt

	sanitizedToolName, _ := util.SanitizeToolName("read_file")
	toolID := serviceID + "." + sanitizedToolName // Note: separator might be / or . depending on implementation
	// Upstream uses "." joiner usually?
	// Looking at streamable_http.go: toolID := serviceID + "." + sanitizedToolName

	// Wait, streamable_http.go:502: toolID := serviceID + "." + sanitizedToolName

	mcpTool, ok := toolManager.GetTool(toolID)
	require.True(t, ok, "Tool should be registered: %s", toolID)

	// Prepare Call
	callArgs := json.RawMessage(`{"path": "/app/bundle/hello.txt"}`)
	req := &tool.ExecutionRequest{
		ToolName:   toolID,
		ToolInputs: callArgs,
	}
	result, err := mcpTool.Execute(ctx, req)

	require.NoError(t, err)

	// Result should be the content string since it's not a JSON map
	resultStr, ok := result.(string)
	require.True(t, ok, "Result should be a string")
	assert.Equal(t, "Hello MCP Bundle!", resultStr)

	// Verify list_directory
	sanitizedListTool, _ := util.SanitizeToolName("list_directory")
	listToolID := serviceID + "." + sanitizedListTool
	mcpListTool, ok := toolManager.GetTool(listToolID)
	require.True(t, ok)

	listArgs := json.RawMessage(`{"path": "/app/bundle"}`)
	listReq := &tool.ExecutionRequest{
		ToolName:   listToolID,
		ToolInputs: listArgs,
	}
	listResult, err := mcpListTool.Execute(ctx, listReq)
	require.NoError(t, err)

	// Result is a JSON string of files array (because it's not a map)
	listResultStr, ok := listResult.(string)
	require.True(t, ok, "List result should be a string")
	assert.Contains(t, listResultStr, "manifest.json")
	assert.Contains(t, listResultStr, "server.py")
	assert.Contains(t, listResultStr, "hello.txt")
}
