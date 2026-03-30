// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp "github.com/modelcontextprotocol/go-sdk/pkg/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHITLIntegration(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer mockServer.Close()

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("hitl-integration-test"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(mockServer.URL),
			Calls: map[string]*configv1.HttpCallDefinition{
				"sensitive_call": configv1.HttpCallDefinition_builder{
					Id:           proto.String("sensitive_call"),
					Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
					EndpointPath: proto.String("/sensitive"),
				}.Build(),
			},
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("sensitive_tool"),
					CallId: proto.String("sensitive_call"),
				}.Build(),
			},
		}.Build(),
		Hitl: configv1.HITLConfig_builder{
			Enabled:        proto.Bool(true),
			RequireMfa:     proto.Bool(true),
			TimeoutSeconds: proto.Int32(5),
			SensitiveTools: []string{"hitl-integration-test.sensitive_tool"},
		}.Build(),
	}.Build()

	configFile := CreateTempConfigFile(t, config)

	// Create app configuration
	serverConfig := &configv1.ServerConfig{
		Server: &configv1.ServerSettings{
			RestPort:  8082,
			AdminPort: 9092,
			BusProvider: &configv1.ServerSettings_Bus{
				Config: &configv1.ServerSettings_Bus_InMemory{},
			},
		},
	}

	appCtx, cancelApp := context.WithCancel(context.Background())
	defer cancelApp()

	// Need a way to start the actual server and get the HTTP addresses.
	// Use StartServer helper from e2e_helpers.go
	httpURL, _, _, err := StartServer(appCtx, t, serverConfig, configFile)
	require.NoError(t, err)

	// Since we are not doing direct stdin/stdout we'll use a websocket or rest client
	// For testing the HITL flow, let's trigger it via the prompt API or stdio?
	// Wait, standard StdioServer wrapper is better for calling tools.

	// Start StdioServer to act as the agent caller
	client, cleanup := StartStdioServer(t, configFile)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Give the server a moment to start up and subscribe
	time.Sleep(1 * time.Second)

	// Call the sensitive tool in a goroutine because it will block waiting for HITL approval
	callErrCh := make(chan error, 1)
	go func() {
		_, err := client.CallTool(ctx, &mcp.CallToolParams{
			Name:      "hitl-integration-test.sensitive_tool",
			Arguments: map[string]interface{}{},
		})
		callErrCh <- err
	}()

	// Give the tool call a moment to reach the middleware and publish the hitl request
	time.Sleep(1 * time.Second)

	// Verify the request is pending via the UI endpoint
	resp, err := http.Get(httpURL + "/api/v1/hitl/approvals")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var approvals []map[string]interface{}
	err = json.Unmarshal(body, &approvals)
	require.NoError(t, err)

	// Find our approval request
	var approvalID string
	var found bool
	for _, a := range approvals {
		if a["tool"] == "hitl-integration-test.sensitive_tool" {
			approvalID = a["id"].(string)
			assert.Equal(t, true, a["requireMfa"])
			assert.Equal(t, "pending", a["status"])
			found = true
			break
		}
	}
	require.True(t, found, "Expected to find pending HITL approval for sensitive_tool")

	// Approve the request with MFA code
	approveBody := []byte(`{"action": "approved", "mfaCode": "123456"}`)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/hitl/approvals/%s", httpURL, approvalID), bytes.NewReader(approveBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	approveResp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer approveResp.Body.Close()

	assert.Equal(t, http.StatusOK, approveResp.StatusCode)

	// Wait for the tool call to complete
	select {
	case err := <-callErrCh:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Tool call did not complete after approval")
	}

	// Verify approval list is empty
	resp2, err := http.Get(httpURL + "/api/v1/hitl/approvals")
	require.NoError(t, err)
	defer resp2.Body.Close()

	body2, _ := io.ReadAll(resp2.Body)
	var approvals2 []map[string]interface{}
	_ = json.Unmarshal(body2, &approvals2)

	for _, a := range approvals2 {
		assert.NotEqual(t, "hitl-integration-test.sensitive_tool", a["tool"])
	}
}

func TestHITLIntegration_Deny(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer mockServer.Close()

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("hitl-integration-deny"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(mockServer.URL),
			Calls: map[string]*configv1.HttpCallDefinition{
				"sensitive_call_deny": configv1.HttpCallDefinition_builder{
					Id:           proto.String("sensitive_call_deny"),
					Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
					EndpointPath: proto.String("/sensitive-deny"),
				}.Build(),
			},
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{
					Name:   proto.String("sensitive_tool_deny"),
					CallId: proto.String("sensitive_call_deny"),
				}.Build(),
			},
		}.Build(),
		Hitl: configv1.HITLConfig_builder{
			Enabled:        proto.Bool(true),
			RequireMfa:     proto.Bool(false),
			TimeoutSeconds: proto.Int32(5),
			SensitiveTools: []string{"hitl-integration-deny.sensitive_tool_deny"},
		}.Build(),
	}.Build()

	configFile := CreateTempConfigFile(t, config)

	serverConfig := &configv1.ServerConfig{
		Server: &configv1.ServerSettings{
			RestPort:  8083,
			AdminPort: 9093,
			BusProvider: &configv1.ServerSettings_Bus{
				Config: &configv1.ServerSettings_Bus_InMemory{},
			},
		},
	}

	appCtx, cancelApp := context.WithCancel(context.Background())
	defer cancelApp()

	httpURL, _, _, err := StartServer(appCtx, t, serverConfig, configFile)
	require.NoError(t, err)

	client, cleanup := StartStdioServer(t, configFile)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	time.Sleep(1 * time.Second)

	callErrCh := make(chan error, 1)
	go func() {
		_, err := client.CallTool(ctx, &mcp.CallToolParams{
			Name:      "hitl-integration-deny.sensitive_tool_deny",
			Arguments: map[string]interface{}{},
		})
		callErrCh <- err
	}()

	time.Sleep(1 * time.Second)

	resp, err := http.Get(httpURL + "/api/v1/hitl/approvals")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var approvals []map[string]interface{}
	json.Unmarshal(body, &approvals)

	var approvalID string
	for _, a := range approvals {
		if a["tool"] == "hitl-integration-deny.sensitive_tool_deny" {
			approvalID = a["id"].(string)
			break
		}
	}
	require.NotEmpty(t, approvalID)

	// Deny
	denyBody := []byte(`{"action": "denied"}`)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/hitl/approvals/%s", httpURL, approvalID), bytes.NewReader(denyBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	denyResp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer denyResp.Body.Close()

	// Wait for tool call to error out
	select {
	case err := <-callErrCh:
		require.Error(t, err)
		assert.Contains(t, err.Error(), "human denied request")
	case <-time.After(5 * time.Second):
		t.Fatal("Tool call did not complete after deny")
	}
}
