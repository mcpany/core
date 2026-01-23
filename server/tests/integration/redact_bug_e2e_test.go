// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactBugE2E(t *testing.T) {
	// 1. Start a mock upstream server that returns 400 with a JSON body
	// containing the sequence triggering the bug.
	// Body: {"msg": 10 / 2, // "password": "secret"}
	// (Note: this is invalid JSON, but that's fine, the redactor runs on bytes)

	buggyPayload := `{"msg": 10 / 2, // "password": "secret"}`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(buggyPayload))
	}))
	defer mockServer.Close()

	// 2. Start MCP Any server
	// We set MCPANY_TEST_DEBUG=true to enable debug mode in the server process,
	// which is required to see the error body in the response (otherwise it's hidden).
	t.Setenv("MCPANY_TEST_DEBUG", "true")
	serverInfo := StartMCPANYServer(t, "RedactBugTest")
	defer serverInfo.CleanupFunc()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := serverInfo.Initialize(ctx)
	require.NoError(t, err)

	// 3. Register the mock service
	serviceID := "buggy-service"
	baseURL := mockServer.URL
	operationID := "trigger_bug"
	endpointPath := "/trigger" // path doesn't matter for the mock
	httpMethod := "POST"

	RegisterHTTPService(t, serverInfo.RegistrationClient, serviceID, baseURL, operationID, endpointPath, httpMethod, nil)

	// 4. Call the tool
	// We expect the tool to fail (return isError: true), but the call itself via JSON-RPC should succeed.

	// Explicitly use json.RawMessage to avoid base64 encoding if the SDK type is []byte alias
	args := json.RawMessage([]byte("{}"))

	result, err := serverInfo.CallTool(ctx, &mcp.CallToolParams{
		Name:      serviceID + "." + operationID,
		Arguments: args,
	})

	require.NoError(t, err) // JSON-RPC call succeeded
	require.True(t, result.IsError, "Tool execution should have failed")

	require.NotEmpty(t, result.Content)

	// Handle content type assertion safely
	var errMsg string
	if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
		errMsg = textContent.Text
	} else {
		// If unmarshalling failed to produce specific types (e.g. resulted in map[string]interface{}),
		// we might need to inspect it differently.
		// But let's assume SDK unmarshals correctly or we inspect via Sprintf for now if type assertion fails.
		errMsg = fmt.Sprintf("%v", result.Content[0])
	}

	t.Logf("Got error message: %s", errMsg)

	// If fixed, the error message should contain the original payload with "secret".
	// If buggy, it would contain "[REDACTED]".

	// Check for correct behavior
	assert.Contains(t, errMsg, `"password": "secret"`, "Error message should contain unredacted secret because it is in a comment")
	assert.NotContains(t, errMsg, `"password": "[REDACTED]"`, "Error message should NOT contain redacted secret")
}
