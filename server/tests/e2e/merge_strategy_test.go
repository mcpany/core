// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/app"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeStrategyAndFiltering(t *testing.T) {
	t.Setenv("MCPANY_ENABLE_FILE_CONFIG", "true")
	// Create a temporary directory for config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Define a mock upstream config
	// We will use a stdio upstream that echoes tools, but for simplicity in E2E,
	// we can use the "mock" upstream if available, or just use a simple stdio script.
	// Actually, we can use `tests/integration/examples/tools/echo.sh` or similar if it exists.
	// Or write a small script.

	// Create a mock MCP server in Go
	mockServerSrc := filepath.Join(tmpDir, "mock_main.go")
	mockServerBin := filepath.Join(tmpDir, "mock_server")

	mockSrcContent := "package main\n\n" +
		"import (\n" +
		"\t\"encoding/json\"\n" +
		"\t\"fmt\"\n" +
		"\t\"io\"\n" +
		"\t\"os\"\n" +
		")\n\n" +
		"type JSONRPCRequest struct {\n" +
		"\tJSONRPC string          `json:\"jsonrpc\"`\n" +
		"\tID      interface{}     `json:\"id\"`\n" +
		"\tMethod  string          `json:\"method\"`\n" +
		"\tParams  json.RawMessage `json:\"params\"`\n" +
		"}\n\n" +
		"type JSONRPCResponse struct {\n" +
		"\tJSONRPC string      `json:\"jsonrpc\"`\n" +
		"\tID      interface{} `json:\"id\"`\n" +
		"\tResult  interface{} `json:\"result,omitempty\"`\n" +
		"\tError   interface{} `json:\"error,omitempty\"`\n" +
		"}\n\n" +
		"func main() {\n" +
		"\tdecoder := json.NewDecoder(os.Stdin)\n" +
		"\tencoder := json.NewEncoder(os.Stdout)\n\n" +
		"\tfor {\n" +
		"\t\tvar req JSONRPCRequest\n" +
		"\t\tif err := decoder.Decode(&req); err != nil {\n" +
		"\t\t\tif err == io.EOF {\n" +
		"\t\t\t\treturn\n" +
		"\t\t\t}\n" +
		"\t\t\tfmt.Fprintf(os.Stderr, \"Decode error: %v\\n\", err)\n" +
		"\t\t\tcontinue\n" +
		"\t\t}\n\n" +
		"\t\tvar result interface{}\n" +
		"\t\tswitch req.Method {\n" +
		"\t\tcase \"initialize\":\n" +
		"\t\t\tresult = map[string]interface{}{\n" +
		"\t\t\t\t\"protocolVersion\": \"2024-11-05\",\n" +
		"\t\t\t\t\"capabilities\":    map[string]interface{}{},\n" +
		"\t\t\t\t\"serverInfo\": map[string]interface{}{\n" +
		"\t\t\t\t\t\"name\":    \"mock-server\",\n" +
		"\t\t\t\t\t\"version\": \"1.0\",\n" +
		"\t\t\t\t},\n" +
		"\t\t\t}\n" +
		"\t\tcase \"notifications/initialized\":\n" +
		"\t\t\tcontinue // No response needed\n" +
		"\t\tcase \"tools/list\":\n" +
		"\t\t\tresult = map[string]interface{}{\n" +
		"\t\t\t\t\"tools\": []map[string]interface{}{\n" +
		"\t\t\t\t\t{\n" +
		"\t\t\t\t\t\t\"name\": \"tool_a\",\n" +
		"\t\t\t\t\t\t\"description\": \"Original Description A\",\n" +
		"\t\t\t\t\t\t\"inputSchema\": map[string]interface{}{\"type\": \"object\", \"properties\": map[string]interface{}{}},\n" +
		"\t\t\t\t\t},\n" +
		"\t\t\t\t\t{\n" +
		"\t\t\t\t\t\t\"name\": \"tool_b\",\n" +
		"\t\t\t\t\t\t\"description\": \"Original Description B\",\n" +
		"\t\t\t\t\t\t\"inputSchema\": map[string]interface{}{\"type\": \"object\", \"properties\": map[string]interface{}{}},\n" +
		"\t\t\t\t\t},\n" +
		"\t\t\t\t},\n" +
		"\t\t\t}\n" +
		"\t\tdefault:\n" +
		"\t\t\t// Ignore other methods or return nil result\n" +
		"\t\t}\n\n" +
		"\t\tif req.ID != nil {\n" +
		"\t\t\tresp := JSONRPCResponse{\n" +
		"\t\t\t\tJSONRPC: \"2.0\",\n" +
		"\t\t\t\tID:      req.ID,\n" +
		"\t\t\t\tResult:  result,\n" +
		"\t\t\t}\n" +
		"\t\t\tencoder.Encode(resp)\n" +
		"\t\t}\n" +
		"\t}\n" +
		"}\n"
	err := os.WriteFile(mockServerSrc, []byte(mockSrcContent), 0644)
	require.NoError(t, err)

	// Build the mock server
	cmd := exec.Command("go", "build", "-o", mockServerBin, mockServerSrc)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	require.NoError(t, err)

	configContent := fmt.Sprintf(`
global_settings:
  db_driver: "sqlite"
  db_path: "%s"
  profiles: ["custom_profile"]
  profile_definitions:
    - name: "custom_profile"
      selector:
        tags: ["visible"]
      service_config:
        "mock-service":
          enabled: true

users:
  - id: "default"
    profile_ids: ["custom_profile"]

upstream_services:
  - name: "mock-service"
    id: "mock-service"

    mcp_service:
      tool_auto_discovery: true
      stdio_connection:
        command: "%s"
      tools:
        - name: "tool_a"
          description: "Overridden Description A"
          merge_strategy: MERGE_STRATEGY_OVERRIDE
          tags: ["visible"]
        - name: "tool_b"
          tags: ["hidden"] # Should be filtered out by profile "custom_profile" (selecting "visible")
`, filepath.Join(tmpDir, "test.db"), mockServerBin)

	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Run the server
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fs := afero.NewOsFs()
	runner := app.NewApplication()

	// Pick random ports
	jsonrpcPort := getFreePort(t)
	grpcPort := "" // Disable gRPC for this test

	go func() {
		opts := app.RunOptions{
			Ctx:             ctx,
			Fs:              fs,
			Stdio:           false,
			JSONRPCPort:     jsonrpcPort,
			GRPCPort:        grpcPort,
			ConfigPaths:     []string{configPath},
			APIKey:          "",
			ShutdownTimeout: 5 * time.Second,
		}
		if err := runner.Run(opts); err != nil {
			// It might return error on cancel, which is expected
			if ctx.Err() == nil {
				t.Logf("Server failed: %v", err)
			}
		}
	}()

	// Wait for server to be ready
	require.Eventually(t, func() bool {
		return app.HealthCheck(os.Stderr, jsonrpcPort, 1*time.Second) == nil
	}, 10*time.Second, 100*time.Millisecond)

	// Verify Tools via JSON-RPC
	// We expect tool_a to be present (Overridden description)
	// We expect tool_b to be absent (Filtered out)

	callListTools := func() ([]map[string]any, error) {
		// Use stateless endpoint /mcp/u/default/profile/custom_profile
		url := fmt.Sprintf("http://%s/mcp/u/default/profile/custom_profile", jsonrpcPort)
		reqBody := `{"jsonrpc":"2.0","method":"tools/list","id":1,"params":{}}`
		req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		// Stateless endpoint returns JSON, no SSE needed

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		var rpcResp struct {
			Result struct {
				Tools []map[string]any `json:"tools"`
			} `json:"result"`
			Error any `json:"error"`
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		// Reset body for decoder (needs a new reader)
		// Actually, just decode from bytes or print bytes on error
		if err := json.Unmarshal(bodyBytes, &rpcResp); err != nil {
			return nil, fmt.Errorf("json decode error: %v, body: %s", err, string(bodyBytes))
		}
		if rpcResp.Error != nil {
			return nil, fmt.Errorf("rpc error: %v", rpcResp.Error)
		}
		return rpcResp.Result.Tools, nil
	}

	// Verify Tools via JSON-RPC
	// Poll until tools are registered
	require.Eventually(t, func() bool {
		tools, err := callListTools()
		if err != nil {
			t.Logf("ListTools error: %v", err)
			return false
		}

		for _, tool := range tools {
			name, _ := tool["name"].(string)
			if name == "tool_a" || name == "mock-service.tool_a" {
				// Verify override worked
				if desc, ok := tool["description"].(string); ok && desc == "Overridden Description A" {
					return true
				}
			}
		}
		return false
	}, 10*time.Second, 500*time.Millisecond, "tool_a with overridden description should be found")

	// Now verify tool_b is NOT present
	tools, err := callListTools()
	require.NoError(t, err)
	for _, tool := range tools {
		name, _ := tool["name"].(string)
		if name == "tool_b" || name == "mock-service.tool_b" {
			assert.Fail(t, "tool_b should be filtered out")
		}
	}
}

func getFreePort(t *testing.T) string {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()
	return l.Addr().String()
}
