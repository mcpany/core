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
	if _, err := os.Stat(serverBin); err != nil {
		t.Skipf("MCPANY binary not found at %s. Run 'make build'. Skipping test.", serverBin)
	}

	// Use a unique temp DB for each test to avoid conflicts and stale headers
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Create command
	cmd := exec.Command(serverBin, "run", "--stdio", "--config-path", configFile, "--db-path", dbPath, "--metrics-listen-address", LoopbackIP+":0") //nolint:gosec // Test helper
	cmd.Env = append(os.Environ(),
		"MCPANY_DANGEROUS_ALLOW_LOCAL_IPS=true",
		"MCPANY_ENABLE_FILE_CONFIG=true",
		"MCPANY_ALLOW_LOOPBACK_RESOURCES=true",
	)
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

	config := configv1.UpstreamServiceConfig_builder{
		Name:             proto.String("auto-discover-test"),
		AutoDiscoverTool: proto.Bool(true),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(mockServer.URL),
			Calls: map[string]*configv1.HttpCallDefinition{
				call1: configv1.HttpCallDefinition_builder{
					Id:           proto.String(call1),
					Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
					EndpointPath: proto.String("/call1"),
				}.Build(),
				call2: configv1.HttpCallDefinition_builder{
					Id:           proto.String(call2),
					Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
					EndpointPath: proto.String("/call2"),
				}.Build(),
				call3: configv1.HttpCallDefinition_builder{
					Id:           proto.String(call3),
					Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
					EndpointPath: proto.String("/call3"),
				}.Build(),
			},
		}.Build(),
		ToolExportPolicy: configv1.ExportPolicy_builder{
			DefaultAction: configv1.ExportPolicy_UNEXPORT.Enum(),
			Rules: []*configv1.ExportRule{
				configv1.ExportRule_builder{
					NameRegex: proto.String("^call.*"),
					Action:    configv1.ExportPolicy_EXPORT.Enum(),
				}.Build(),
			},
		}.Build(),
	}.Build()

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

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("call-policy-test"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(mockServer.URL),
			Calls: map[string]*configv1.HttpCallDefinition{
				"allowed_call": configv1.HttpCallDefinition_builder{
					Id:           proto.String("allowed_call"),
					Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
					EndpointPath: proto.String("/allowed"),
				}.Build(),
				"denied_call": configv1.HttpCallDefinition_builder{
					Id:           proto.String("denied_call"),
					Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
					EndpointPath: proto.String("/denied"),
				}.Build(),
			},
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("allowed_tool"),
					CallId: proto.String("allowed_call"),
				}.Build(),
				configv1.ToolDefinition_builder{
					Name:   proto.String("denied_tool"),
					CallId: proto.String("denied_call"),
				}.Build(),
			},
		}.Build(),
		CallPolicies: []*configv1.CallPolicy{
			configv1.CallPolicy_builder{
				DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
				Rules: []*configv1.CallPolicyRule{
					configv1.CallPolicyRule_builder{
						Action:    configv1.CallPolicy_DENY.Enum(),
						NameRegex: proto.String("denied.*"),
					}.Build(),
				},
			}.Build(),
		},
		ToolExportPolicy: configv1.ExportPolicy_builder{
			DefaultAction: configv1.ExportPolicy_EXPORT.Enum(),
		}.Build(),
	}.Build()

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
	require.NoError(t, err)

	result, err := client.CallTool(ctx, &mcp.CallToolParams{
		Name:      "call-policy-test.denied_tool",
		Arguments: map[string]interface{}{},
	})
	require.NoError(t, err)
	require.True(t, result.IsError)
    // Check if content contains the error message
    // content is usually a list of text/image
    contentBytes, _ := json.Marshal(result.Content)
	assert.Contains(t, string(contentBytes), "unknown tool")
}

func TestExportPolicyForPromptsAndResources(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("export-policy-misc"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(mockServer.URL),
			Prompts: []*configv1.PromptDefinition{
				configv1.PromptDefinition_builder{Name: proto.String("public_prompt")}.Build(),
				configv1.PromptDefinition_builder{Name: proto.String("private_prompt")}.Build(),
			},
			Resources: []*configv1.ResourceDefinition{
				configv1.ResourceDefinition_builder{Name: proto.String("public_resource"), Uri: proto.String("http://resource/public")}.Build(),
				configv1.ResourceDefinition_builder{Name: proto.String("private_resource"), Uri: proto.String("http://resource/private")}.Build(),
			},
		}.Build(),
		PromptExportPolicy: configv1.ExportPolicy_builder{
			DefaultAction: configv1.ExportPolicy_UNEXPORT.Enum(),
			Rules: []*configv1.ExportRule{
				configv1.ExportRule_builder{NameRegex: proto.String("^public.*"), Action: configv1.ExportPolicy_EXPORT.Enum()}.Build(),
			},
		}.Build(),
		ResourceExportPolicy: configv1.ExportPolicy_builder{
			DefaultAction: configv1.ExportPolicy_UNEXPORT.Enum(),
			Rules: []*configv1.ExportRule{
				configv1.ExportRule_builder{NameRegex: proto.String("^public.*"), Action: configv1.ExportPolicy_EXPORT.Enum()}.Build(),
			},
		}.Build(),
	}.Build()

	configFile := CreateTempConfigFile(t, config)
	client, cleanup := StartStdioServer(t, configFile)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify Prompts and Resources with Eventually to allow for async registration
	assert.Eventually(t, func() bool {
		// Verify Prompts
		promptsResult, err := client.ListPrompts(ctx)
		if err != nil {
			return false
		}
		promptNames := make(map[string]bool)
		for _, p := range promptsResult.Prompts {
			promptNames[p.Name] = true
		}
		if !promptNames["export-policy-misc.public_prompt"] {
			return false
		}
		if promptNames["export-policy-misc.private_prompt"] {
			return false
		}

		// Verify Resources
		resourcesResult, err := client.ListResources(ctx)
		if err != nil {
			return false
		}
		resNames := make(map[string]bool)
		for _, r := range resourcesResult.Resources {
			resNames[r.Name] = true
		}
		if !resNames["public_resource"] {
			return false
		}
		if resNames["private_resource"] {
			return false
		}

		return true
	}, 10*time.Second, 100*time.Millisecond, "Expected prompts and resources to be exported/unexported correctly")
}
