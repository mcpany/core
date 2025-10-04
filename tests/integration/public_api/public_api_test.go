/*
 * Copyright 2025 Author(s) of MCPXY
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

package public_api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_PublicHttpPost(t *testing.T) {
	t.Log("INFO: Starting E2E Test Scenario for Public HTTP POST API...")
	mcpxTestServerInfo := integration.StartMCPXYServer(t, "E2EPublicHttpPostTest")
	defer mcpxTestServerInfo.CleanupFunc()

	const serviceID = "public_http_post"
	const baseURL = "https://httpbin.org"
	const operationID = "postAnything"
	const endpointPath = "/post"
	const httpMethod = "POST"

	t.Logf("INFO: Registering '%s' with MCPXY at endpoint %s%s...", serviceID, baseURL, endpointPath)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterHTTPService(t, registrationGRPCClient, serviceID, baseURL, operationID, endpointPath, httpMethod, nil)
	t.Logf("INFO: '%s' registered.", serviceID)

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err, "Failed to connect to MCPXY server")
	defer cs.Close()

	toolName := fmt.Sprintf("%s/-/%s", serviceID, operationID)
	params := json.RawMessage(`{"message": "hello world"}`)

	// Adding a retry loop to handle intermittent failures from httpbin.org
	var res *mcp.CallToolResult
	require.Eventually(t, func() bool {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
		if err != nil {
			t.Logf("retrying after error calling tool '%s': %v", toolName, err)
			// If we get an error that is not a 5xx error, we should fail fast.
			if !strings.Contains(err.Error(), "502") && !strings.Contains(err.Error(), "503") && !strings.Contains(err.Error(), "504") {
				t.Errorf("unrecoverable error calling httpbin post tool: %v", err)
			}
			return false // retry on error
		}
		return true
	}, 1*time.Minute, 5*time.Second, "tool call to httpbin.org failed after multiple retries")

	require.NotNil(t, res, "Nil response from tool '%s'", toolName)
	require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected content to be of type TextContent")

	var jsonResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &jsonResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response from httpbin")

	jsonData, ok := jsonResponse["json"].(map[string]interface{})
	require.True(t, ok, "Response should contain a 'json' object")

	message, ok := jsonData["message"].(string)
	require.True(t, ok, "The 'json' object should contain a 'message' string")
	require.Equal(t, "hello world", message, "The message should be 'hello world'")

	t.Log("INFO: E2E Test Scenario for Public HTTP POST API Completed Successfully!")
}

func TestUpstreamService_PublicWebsocket(t *testing.T) {
	t.Log("INFO: Starting E2E Test Scenario for Public Websocket API...")
	mcpxTestServerInfo := integration.StartMCPXYServer(t, "E2EPublicWebsocketTest")
	defer mcpxTestServerInfo.CleanupFunc()

	const serviceID = "public_websocket_echo"
	const baseURL = "wss://ws.postman-echo.com/raw"
	const operationID = "echo"

	t.Logf("INFO: Registering '%s' with MCPXY at endpoint %s...", serviceID, baseURL)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterWebsocketService(t, registrationGRPCClient, serviceID, baseURL, operationID, nil)
	t.Logf("INFO: '%s' registered.", serviceID)

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err, "Failed to connect to MCPXY server")
	defer cs.Close()

	toolName := fmt.Sprintf("%s/-/%s", serviceID, operationID)
	params := json.RawMessage(`{"message": "hello from websocket"}`)
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
	require.NoError(t, err, "Error calling tool '%s'", toolName)
	require.NotNil(t, res, "Nil response from tool '%s'", toolName)
	require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected content to be of type TextContent")
	require.Contains(t, textContent.Text, "hello from websocket", "Response should contain the message we sent")

	t.Log("INFO: E2E Test Scenario for Public Websocket API Completed Successfully!")
}

func TestUpstreamService_JsonPlaceholderPost(t *testing.T) {
	t.Log("INFO: Starting E2E Test Scenario for JSONPlaceholder POST API...")
	mcpxTestServerInfo := integration.StartMCPXYServer(t, "E2EJsonPlaceholderPostTest")
	defer mcpxTestServerInfo.CleanupFunc()

	const serviceID = "jsonplaceholder"
	const baseURL = "https://jsonplaceholder.typicode.com"
	const operationID = "createPost"
	const endpointPath = "/posts"
	const httpMethod = "POST"

	t.Logf("INFO: Registering '%s' with MCPXY at endpoint %s%s...", serviceID, baseURL, endpointPath)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterHTTPService(t, registrationGRPCClient, serviceID, baseURL, operationID, endpointPath, httpMethod, nil)
	t.Logf("INFO: '%s' registered.", serviceID)

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err, "Failed to connect to MCPXY server")
	defer cs.Close()

	toolName := fmt.Sprintf("%s/-/%s", serviceID, operationID)
	params := json.RawMessage(`{"title": "foo","body": "bar","userId": 1}`)
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
	require.NoError(t, err, "Error calling tool '%s'", toolName)
	require.NotNil(t, res, "Nil response from tool '%s'", toolName)
	require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)

	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected content to be of type TextContent")

	var jsonResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &jsonResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Contains(t, jsonResponse, "id", "Response should contain an 'id' field")
	require.Equal(t, "foo", jsonResponse["title"])
	require.Equal(t, "bar", jsonResponse["body"])
	require.Equal(t, float64(1), jsonResponse["userId"])

	t.Log("INFO: E2E Test Scenario for JSONPlaceholder POST API Completed Successfully!")
}

func TestUpstreamService_LanyardWebsocket(t *testing.T) {
	t.Log("INFO: Starting E2E Test Scenario for Lanyard Websocket API...")
	mcpxTestServerInfo := integration.StartMCPXYServer(t, "E2ELanyardWebsocketTest")
	defer mcpxTestServerInfo.CleanupFunc()

	const serviceID = "lanyard_websocket"
	const baseURL = "wss://api.lanyard.rest/socket"
	const operationID = "subscribe"

	t.Logf("INFO: Registering '%s' with MCPXY at endpoint %s...", serviceID, baseURL)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterWebsocketService(t, registrationGRPCClient, serviceID, baseURL, operationID, nil)
	t.Logf("INFO: '%s' registered.", serviceID)

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
	defer cancel()

	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err, "Failed to connect to MCPXY server")
	defer cs.Close()

	toolName := fmt.Sprintf("%s/-/%s", serviceID, operationID)
	// Subscribe to a known Discord user ID for Lanyard
	params := json.RawMessage(`{"op": 2, "d": {"subscribe_to_id": "138332767946997760"}}`)

	// We only check for the initial response, not the stream of presence updates
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
	require.NoError(t, err, "Error calling tool '%s'", toolName)
	require.NotNil(t, res, "Nil response from tool '%s'", toolName)
	require.Len(t, res.Content, 1, "Expected exactly one content item from tool '%s'", toolName)

	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected content to be of type TextContent")

	var jsonResponse map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &jsonResponse)
	require.NoError(t, err, "Failed to unmarshal JSON response")

	require.Equal(t, float64(1), jsonResponse["op"], "Expected OP code 1 for Hello message")

	t.Log("INFO: E2E Test Scenario for Lanyard Websocket API Completed Successfully!")
}
