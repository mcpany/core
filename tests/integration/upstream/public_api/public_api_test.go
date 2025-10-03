/*
 * Copyright 2025 Author(s) of MCPX
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

package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mcpxy/mcpx/pkg/consts"
	"github.com/mcpxy/mcpx/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_PublicHttpPost(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeMedium) // Increased timeout for network resiliency
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Public HTTP POST API...")
	t.Parallel()

	// --- 1. Start MCPX Server ---
	mcpxTestServerInfo := integration.StartMCPXServer(t, "E2EPublicHttpPostTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register Public HTTP POST Service with MCPX ---
	const serviceID = "public_http_post"
	serviceURL := "https://httpbin.org"
	endpointPath := "/post"
	operationID := "postAnything"
	t.Logf("INFO: Registering '%s' with MCPX at endpoint %s%s...", serviceID, serviceURL, endpointPath)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterHTTPService(t, registrationGRPCClient, serviceID, serviceURL, operationID, endpointPath, "POST", nil)
	t.Logf("INFO: '%s' registered.", serviceID)

	// --- 3. Call Tool via MCPX ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	toolName := fmt.Sprintf("%s%s%s", serviceID, consts.ToolNameServiceSeparator, operationID)
	postPayload := `{"key": "value", "number": 123}`

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(postPayload)})
		if err == nil {
			break // Success
		}

		// If the error is a 503 or a timeout, we can retry. Otherwise, fail fast.
		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") {
			t.Logf("Attempt %d/%d: Call to httpbin.org failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		// For any other error, fail the test immediately.
		require.NoError(t, err, "unrecoverable error calling httpbin post tool")
	}

	// If all retries failed due to transient errors, skip the test.
	if err != nil {
		t.Skipf("Skipping test: all %d retries to httpbin.org failed with transient errors. Last error: %v", maxRetries, err)
	}

	require.NotNil(t, res, "Nil response from httpbin post tool after retries")
	require.Len(t, res.Content, 1, "Expected exactly one content block in the response")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content but got %T", res.Content[0])

	var httpbinResponse struct {
		JSON json.RawMessage `json:"json"`
	}
	err = json.Unmarshal([]byte(textContent.Text), &httpbinResponse)
	require.NoError(t, err, "Failed to unmarshal httpbin response")
	assert.JSONEq(t, postPayload, string(httpbinResponse.JSON), "The posted JSON does not match the response")

	t.Log("INFO: E2E Test Scenario for Public HTTP POST API Completed Successfully!")
}

func TestUpstreamService_PublicWebsocket(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Public Websocket API...")
	t.Parallel()

	// --- 1. Start MCPX Server ---
	mcpxTestServerInfo := integration.StartMCPXServer(t, "E2EPublicWebsocketTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register Public Websocket Service with MCPX ---
	const serviceID = "public_websocket_echo"
	// Using a known public echo server
	serviceEndpoint := "wss://ws.postman-echo.com/raw"
	t.Logf("INFO: Registering '%s' with MCPX at endpoint %s...", serviceID, serviceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterWebsocketService(t, registrationGRPCClient, serviceID, serviceEndpoint, "echo", nil)
	t.Logf("INFO: '%s' registered.", serviceID)

	// --- 3. Call Tool via MCPX ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	toolName := fmt.Sprintf("%s%secho", serviceID, consts.ToolNameServiceSeparator)
	echoMessage := `{"message": "hello public websocket"}`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(echoMessage)})
	require.NoError(t, err, "Error calling echo tool")
	require.NotNil(t, res, "Nil response from echo tool")

	require.Len(t, res.Content, 1, "Expected exactly one content block in the response")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content but got %T", res.Content[0])
	require.JSONEq(t, echoMessage, textContent.Text, "The echoed message does not match the original")

	t.Log("INFO: E2E Test Scenario for Public Websocket API Completed Successfully!")
}

func TestUpstreamService_JsonPlaceholderPost(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for JSONPlaceholder POST API...")
	t.Parallel()

	// --- 1. Start MCPX Server ---
	mcpxTestServerInfo := integration.StartMCPXServer(t, "E2EJsonPlaceholderPostTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register JSONPlaceholder Service with MCPX ---
	const serviceID = "jsonplaceholder"
	serviceURL := "https://jsonplaceholder.typicode.com"
	endpointPath := "/posts"
	operationID := "createPost"
	t.Logf("INFO: Registering '%s' with MCPX at endpoint %s%s...", serviceID, serviceURL, endpointPath)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterHTTPService(t, registrationGRPCClient, serviceID, serviceURL, operationID, endpointPath, "POST", nil)
	t.Logf("INFO: '%s' registered.", serviceID)

	// --- 3. Call Tool via MCPX ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	toolName := fmt.Sprintf("%s%s%s", serviceID, consts.ToolNameServiceSeparator, operationID)
	postPayload := `{"title": "foo", "body": "bar", "userId": 1}`
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(postPayload)})
	require.NoError(t, err, "Error calling jsonplaceholder post tool")
	require.NotNil(t, res, "Nil response from jsonplaceholder post tool")

	require.Len(t, res.Content, 1, "Expected exactly one content block in the response")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content but got %T", res.Content[0])

	var apiResponse struct {
		ID     int    `json:"id"`
		Title  string `json:"title"`
		Body   string `json:"body"`
		UserID int    `json:"userId"`
	}
	err = json.Unmarshal([]byte(textContent.Text), &apiResponse)
	require.NoError(t, err, "Failed to unmarshal jsonplaceholder response")

	require.Equal(t, "foo", apiResponse.Title, "Title does not match")
	require.Equal(t, "bar", apiResponse.Body, "Body does not match")
	require.Equal(t, 1, apiResponse.UserID, "UserID does not match")
	require.NotZero(t, apiResponse.ID, "ID should not be zero")

	t.Log("INFO: E2E Test Scenario for JSONPlaceholder POST API Completed Successfully!")
}

func TestUpstreamService_LanyardWebsocket(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Lanyard Websocket API...")
	t.Parallel()

	// --- 1. Start MCPX Server ---
	mcpxTestServerInfo := integration.StartMCPXServer(t, "E2ELanyardWebsocketTest")
	defer mcpxTestServerInfo.CleanupFunc()

	// --- 2. Register Lanyard Websocket Service with MCPX ---
	const serviceID = "lanyard_websocket"
	serviceEndpoint := "wss://api.lanyard.rest/socket"
	operationID := "subscribe"
	t.Logf("INFO: Registering '%s' with MCPX at endpoint %s...", serviceID, serviceEndpoint)
	registrationGRPCClient := mcpxTestServerInfo.RegistrationClient
	integration.RegisterWebsocketService(t, registrationGRPCClient, serviceID, serviceEndpoint, operationID, nil)
	t.Logf("INFO: '%s' registered.", serviceID)

	// --- 3. Call Tool via MCPX ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	toolName := fmt.Sprintf("%s%s%s", serviceID, consts.ToolNameServiceSeparator, operationID)
	// This is a valid Discord user ID to test with, taken from Lanyard's documentation
	discordUserID := "94490510688792576"
	subscribePayload := fmt.Sprintf(`{"op": 2, "d": {"subscribe_to_id": "%s"}}`, discordUserID)

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(subscribePayload)})
	require.NoError(t, err, "Error calling lanyard websocket tool")
	require.NotNil(t, res, "Nil response from lanyard websocket tool")

	// We expect multiple messages from a subscription, but for this test,
	// we'll just check the first one (INIT_STATE)
	require.NotEmpty(t, res.Content, "Expected at least one content block in the response")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content but got %T", res.Content[0])

	var lanyardResponse struct {
		Op int             `json:"op"`
		T  string          `json:"t"`
		D  json.RawMessage `json:"d"`
	}
	err = json.Unmarshal([]byte(textContent.Text), &lanyardResponse)
	require.NoError(t, err, "Failed to unmarshal lanyard response")

	// Opcode 1 is "Hello", which contains the heartbeat interval.
	// This is the first message we should receive.
	require.Equal(t, 1, lanyardResponse.Op, "Expected Opcode 1 (Hello)")

	t.Log("INFO: E2E Test Scenario for Lanyard Websocket API Completed Successfully!")
}
