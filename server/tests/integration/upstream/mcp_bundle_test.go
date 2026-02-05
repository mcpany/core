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
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// simpleFsServerJS is a minimal Node.js MCP server that implements filesystem tools.
// It uses the official SDK via standard input/output.
// To avoid npm install, we will write a vanilla JS server that speaks JSON-RPC over stdio.
// Actually, using the SDK is much easier but requires node_modules.
// For a "bundle", it usually includes node_modules or is a single binary.
// Since we want to test "mcpb", we should probably create a bundle that *works*.
//
// A vanilla JS implementation of MCP is non-trivial to write in a string.
// However, looking at the code, we infer "node" type uses image "node:18-alpine".
// We can use a simple script that just reads stdin and replies.
//
// Let's implement a very basic JSON-RPC handler in JS.
const simpleFsServerJS = `
const readline = require('readline');
const fs = require('fs');
const path = require('path');

const rl = readline.createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

rl.on('line', (line) => {
  if (!line.trim()) return;
  console.error("Received line: " + line);
  try {
    const request = JSON.parse(line);
    handleRequest(request);
  } catch (e) {
    console.error("Failed to parse JSON", e);
  }
});

function sendUrl(msg) {
  process.stdout.write(JSON.stringify(msg) + '\n');
}

console.error("Server starting...");

function handleRequest(req) {
  const method = req.method || req.Method;
  const id = (req.id !== undefined) ? req.id : req.ID;
  const params = req.params || req.Params;

  if (method === 'initialize') {
    sendUrl({
      jsonrpc: '2.0',
      id: id,
      result: {
        protocolVersion: '2024-11-05',
        capabilities: {
          tools: {}
        },
        serverInfo: {
          name: 'simple-fs',
          version: '1.0.0'
        }
      }
    });
  } else if (method === 'notifications/initialized') {
    // Ack
  } else if (method === 'tools/list' || method === 'mcp.listTools') {
    sendUrl({
      jsonrpc: '2.0',
      id: id,
      result: {
        tools: [
          {
            name: "list_directory",
            description: "List files in a directory",
            inputSchema: {
              type: "object",
              properties: {
                path: { type: "string" }
              },
              required: ["path"]
            }
          },
          {
            name: "read_file",
            description: "Read a file",
            inputSchema: {
              type: "object",
              properties: {
                path: { type: "string" }
              },
              required: ["path"]
            }
          }
        ]
      }
    });
  } else if (method === 'tools/call' || method === 'mcp.callTool') {
    const name = params.name || params.Name;
    const args = params.args || params.Arguments || params.arguments;

    try {
      if (name === 'list_directory') {
        const files = fs.readdirSync(args.path);
        sendUrl({
          jsonrpc: '2.0',
          id: id,
          result: {
            content: [{
              type: "text",
              text: JSON.stringify(files)
            }]
          }
        });
      } else if (name === 'read_file') {
        const content = fs.readFileSync(args.path, 'utf8');
        sendUrl({
          jsonrpc: '2.0',
          id: id,
          result: {
            content: [{
              type: "text",
              text: content
            }]
          }
        });
      } else {
         sendUrl({
          jsonrpc: '2.0',
          id: id,
          error: { code: -32601, message: "Method not found: " + method }
        });
      }
    } catch (e) {
      sendUrl({
        jsonrpc: '2.0',
        id: id,
        result: {
            output: {
                text: "Error: " + e.message
            },
            isError: true
        }
      });
    }
  } else {
    // Ping or other methods
    if (id !== undefined) {
        sendUrl({
            jsonrpc: '2.0',
            id: id,
            result: {}
        });
    }
  }
}
`

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

	// Add server.js
	f, err = w.Create("server.js")
	require.NoError(t, err)
	_, err = io.WriteString(f, simpleFsServerJS)
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

	// Check if Docker is available and accessible
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skipf("Skipping Docker tests: docker info failed: %v", err)
	}

	tempDir := t.TempDir()
	bundlePath := createE2EBundle(t, tempDir)

	toolManager := tool.NewManager(nil)
	promptManager := prompt.NewManager()
	resourceManager := resource.NewManager()
	upstreamService := mcp.NewUpstream(nil)
	if impl, ok := upstreamService.(*mcp.Upstream); ok {
		// Use a test-specific temp directory for bundles to ensure isolation
		// and avoid conflicts with global state or other tests.
		// We use a subdirectory "bundles" inside t.TempDir() to keep it clean.
		// If MCP_BUNDLE_DIR is set (e.g. in CI), use that as base to avoid overlay mount issues with /tmp.
		if envDir := os.Getenv("MCP_BUNDLE_DIR"); envDir != "" {
			impl.BundleBaseDir = filepath.Join(envDir, "test-"+t.Name())
		} else {
			impl.BundleBaseDir = filepath.Join(t.TempDir(), "bundles")
		}

		if err := os.MkdirAll(impl.BundleBaseDir, 0755); err != nil {
			t.Fatalf("Failed to create test bundle dir: %v", err)
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
	assert.Contains(t, listResultStr, "server.js")
	assert.Contains(t, listResultStr, "hello.txt")
}
