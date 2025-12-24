// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// StartStdioServer starts the MCP server in Stdio mode and returns the client.
func StartStdioServer(t *testing.T, configFile string) (*MCPClient, func()) {
	t.Helper()

	root := ProjectRoot(t)
	serverBin := filepath.Join(root, "../build/bin/server")

	// Use a unique temp DB for each test to avoid conflicts and stale headers
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Create command
	cmd := exec.Command(serverBin, "run", "--stdio", "--config-path", configFile, "--db-path", dbPath, "--metrics-listen-address", "127.0.0.1:0") //nolint:gosec // Test helper
	// We need to set pipe before starting
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	cmd.Stderr = os.Stderr // Pipe stderr to test output for debugging

	err = cmd.Start()
	require.NoError(t, err)

	client := &MCPClient{
		stdin:  stdin,
		stdout: bufio.NewScanner(stdout),
		events: make(chan map[string]interface{}, 100),
	}

	// Start reading loop
	go client.readLoop()

	// Initialize
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = client.Initialize(ctx)
	require.NoError(t, err)

	return client, func() {
		cancel() // Just in case
		_ = stdin.Close()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}
}

// Update MCPClient for Stdio
type MCPClient struct {
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	events chan map[string]interface{}
	nextID int64
}

func (c *MCPClient) readLoop() {
	for c.stdout.Scan() {
		line := c.stdout.Bytes()
		var msg map[string]interface{}
		if err := json.Unmarshal(line, &msg); err == nil {
			c.events <- msg
		}
	}
}

func (c *MCPClient) Initialize(ctx context.Context) error {
	initParams := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"roots": map[string]interface{}{
				"listChanged": true,
			},
		},
		"clientInfo": map[string]interface{}{
			"name":    "integration-test",
			"version": "1.0.0",
		},
	}

	// Stdio Initialize response is immediate
	_, err := c.Call(ctx, "initialize", initParams)
	if err != nil {
		return err
	}

	return c.Notify(ctx, "notifications/initialized", map[string]interface{}{})
}

func (c *MCPClient) Call(ctx context.Context, method string, params interface{}) (interface{}, error) {
	c.nextID++
	id := c.nextID
	reqBody, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      id,
	})

	_, err := c.stdin.Write(append(reqBody, '\n'))
	if err != nil {
		return nil, err
	}

	// Wait for response
	for {
		select {
		case msg := <-c.events:
			if msgID, ok := msg["id"]; ok {
				// Handle float64/int logic if necessary, but here we can just fmt.Sprint check
				if fmt.Sprint(msgID) == fmt.Sprint(id) {
					if errObj, ok := msg["error"]; ok {
						return nil, fmt.Errorf("RPC Error: %v", errObj)
					}
					return msg["result"], nil
				}
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (c *MCPClient) Notify(_ context.Context, method string, params interface{}) error {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	})
	_, err := c.stdin.Write(append(reqBody, '\n'))
	return err
}

func (c *MCPClient) ListTools(ctx context.Context) (*mcp.ListToolsResult, error) {
	res, err := c.Call(ctx, "tools/list", map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(res)
	var result mcp.ListToolsResult
	err = json.Unmarshal(b, &result)
	return &result, err
}

func (c *MCPClient) CallTool(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
	res, err := c.Call(ctx, "tools/call", params)
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(res)
	var result mcp.CallToolResult
	err = json.Unmarshal(b, &result)
	return &result, err
}

func (c *MCPClient) ListPrompts(ctx context.Context) (*mcp.ListPromptsResult, error) {
	res, err := c.Call(ctx, "prompts/list", map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(res)
	var result mcp.ListPromptsResult
	err = json.Unmarshal(b, &result)
	return &result, err
}

func (c *MCPClient) ListResources(ctx context.Context) (*mcp.ListResourcesResult, error) {
	res, err := c.Call(ctx, "resources/list", map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(res)
	var result mcp.ListResourcesResult
	err = json.Unmarshal(b, &result)
	return &result, err
}

func TestAutoDiscoverAndExportPolicy(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer mockServer.Close()

	call1 := "call1"
	call2 := "call2"
	call3 := "hidden_call"

	config := &configv1.UpstreamServiceConfig{
		Name:             proto.String("auto-discover-test"),
		AutoDiscoverTool: proto.Bool(true),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String(mockServer.URL),
				Calls: map[string]*configv1.HttpCallDefinition{
					call1: {
						Id:           proto.String(call1),
						Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
						EndpointPath: proto.String("/call1"),
					},
					call2: {
						Id:           proto.String(call2),
						Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
						EndpointPath: proto.String("/call2"),
					},
					call3: {
						Id:           proto.String(call3),
						Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
						EndpointPath: proto.String("/call3"),
					},
				},
			},
		},
		ToolExportPolicy: &configv1.ExportPolicy{
			DefaultAction: configv1.ExportPolicy_UNEXPORT.Enum(),
			Rules: []*configv1.ExportRule{
				{
					NameRegex: proto.String("^call.*"),
					Action:    configv1.ExportPolicy_EXPORT.Enum(),
				},
			},
		},
	}

	configFile := CreateTempConfigFile(t, config)
	client, cleanup := StartStdioServer(t, configFile)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	assert.Eventually(t, func() bool {
		toolsResult, err := client.ListTools(ctx)
		if err != nil {
			return false
		}

		toolNames := make(map[string]bool)
		for _, tool := range toolsResult.Tools {
			toolNames[tool.Name] = true
		}

		if !toolNames["auto-discover-test.call1"] {
			return false
		}
		if !toolNames["auto-discover-test.call2"] {
			return false
		}
		if toolNames["auto-discover-test.hidden_call"] {
			return false
		}
		return true
	}, 5*time.Second, 100*time.Millisecond, "Expected tools to be discovered and hidden tools to be excluded")
}

func TestCallPolicyExecution(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/allowed":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result": "allowed"}`))
		case "/denied":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"result": "denied"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("call-policy-test"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String(mockServer.URL),
				Calls: map[string]*configv1.HttpCallDefinition{
					"allowed_call": {
						Id:           proto.String("allowed_call"),
						Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
						EndpointPath: proto.String("/allowed"),
					},
					"denied_call": {
						Id:           proto.String("denied_call"),
						Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
						EndpointPath: proto.String("/denied"),
					},
				},
				Tools: []*configv1.ToolDefinition{
					{
						Name:   proto.String("allowed_tool"),
						CallId: proto.String("allowed_call"),
					},
					{
						Name:   proto.String("denied_tool"),
						CallId: proto.String("denied_call"),
					},
				},
			},
		},
		CallPolicies: []*configv1.CallPolicy{
			{
				DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
				Rules: []*configv1.CallPolicyRule{
					{
						Action:    configv1.CallPolicy_DENY.Enum(),
						NameRegex: proto.String("denied.*"),
					},
				},
			},
		},
		ToolExportPolicy: &configv1.ExportPolicy{
			DefaultAction: configv1.ExportPolicy_EXPORT.Enum(),
		},
	}

	configFile := CreateTempConfigFile(t, config)
	client, cleanup := StartStdioServer(t, configFile)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	assert.Eventually(t, func() bool {
		tools, err := client.ListTools(ctx)
		if err != nil {
			return false
		}
		for _, tool := range tools.Tools {
			if tool.Name == "call-policy-test.allowed_tool" {
				return true
			}
		}
		return false
	}, 10*time.Second, 100*time.Millisecond, "Tool call-policy-test.allowed_tool did not appear in list")

	_, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name:      "call-policy-test.allowed_tool",
		Arguments: map[string]interface{}{},
	})
	assert.NoError(t, err)

	_, err = client.CallTool(ctx, &mcp.CallToolParams{
		Name:      "call-policy-test.denied_tool",
		Arguments: map[string]interface{}{},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tool")
}

func TestExportPolicyForPromptsAndResources(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("export-policy-misc"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String(mockServer.URL),
				Prompts: []*configv1.PromptDefinition{
					{Name: proto.String("public_prompt")},
					{Name: proto.String("private_prompt")},
				},
				Resources: []*configv1.ResourceDefinition{
					{Name: proto.String("public_resource"), Uri: proto.String("http://resource/public")},
					{Name: proto.String("private_resource"), Uri: proto.String("http://resource/private")},
				},
			},
		},
		PromptExportPolicy: &configv1.ExportPolicy{
			DefaultAction: configv1.ExportPolicy_UNEXPORT.Enum(),
			Rules: []*configv1.ExportRule{
				{NameRegex: proto.String("^public.*"), Action: configv1.ExportPolicy_EXPORT.Enum()},
			},
		},
		ResourceExportPolicy: &configv1.ExportPolicy{
			DefaultAction: configv1.ExportPolicy_UNEXPORT.Enum(),
			Rules: []*configv1.ExportRule{
				{NameRegex: proto.String("^public.*"), Action: configv1.ExportPolicy_EXPORT.Enum()},
			},
		},
	}

	configFile := CreateTempConfigFile(t, config)
	client, cleanup := StartStdioServer(t, configFile)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify Prompts
	promptsResult, err := client.ListPrompts(ctx)
	require.NoError(t, err)
	promptNames := make(map[string]bool)
	for _, p := range promptsResult.Prompts {
		promptNames[p.Name] = true
	}
	assert.Contains(t, promptNames, "export-policy-misc.public_prompt")
	assert.NotContains(t, promptNames, "export-policy-misc.private_prompt")

	// Verify Resources
	resourcesResult, err := client.ListResources(ctx)
	require.NoError(t, err)
	resNames := make(map[string]bool)
	for _, r := range resourcesResult.Resources {
		resNames[r.Name] = true
	}
	assert.Contains(t, resNames, "public_resource")
	assert.NotContains(t, resNames, "private_resource")
}
