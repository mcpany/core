// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package github_test

import (
	"context"
	"encoding/json"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_GitHub(t *testing.T) {
	// if os.Getenv("GITHUB_TOKEN") == "" {
	// 	t.Skip("GITHUB_TOKEN is not set")
	// }

	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for GitHub Server...")
	t.Parallel()

	// --- 1. Start Mock GitHub API Server ---
	userResponse := `{
		"login": "octocat",
		"id": 1,
		"type": "User",
		"name": "The Octocat"
	}`
	reposResponse := `[
		{
			"id": 1296269,
			"name": "Hello-World",
			"full_name": "octocat/Hello-World",
			"owner": {
				"login": "octocat"
			}
		}
	]`
	mockHandler := integration.DefaultMockHandler(t, map[string]string{
		"/users/octocat": userResponse,
		"/users/octocat/repos": reposResponse,
	})
	mockServer := integration.StartMockServer(t, mockHandler)
	defer mockServer.Close()

	// --- 2. Start MCPANY Server ---
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EGitHubServerTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// --- 3. Register Service Dynamically ---
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("github"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(mockServer.URL),
			Tools: []*configv1.ToolDefinition{
				configv1.ToolDefinition_builder{Name: proto.String("get_user"), CallId: proto.String("get_user")}.Build(),
				configv1.ToolDefinition_builder{Name: proto.String("list_repos"), CallId: proto.String("list_repos")}.Build(),
			},
			Calls: map[string]*configv1.HttpCallDefinition{
				"get_user": configv1.HttpCallDefinition_builder{
					Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
					EndpointPath: proto.String("/users/{{username}}"),
					Parameters: []*configv1.HttpParameterMapping{
						configv1.HttpParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("username"), Type: configv1.ParameterType(configv1.ParameterType_value["STRING"]).Enum()}.Build()}.Build(),
					},
				}.Build(),
				"list_repos": configv1.HttpCallDefinition_builder{
					Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
					EndpointPath: proto.String("/users/{{username}}/repos"),
					Parameters: []*configv1.HttpParameterMapping{
						configv1.HttpParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("username"), Type: configv1.ParameterType(configv1.ParameterType_value["STRING"]).Enum()}.Build()}.Build(),
					},
				}.Build(),
			},
		}.Build(),
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, mcpAnyTestServerInfo.RegistrationClient, req)

	// --- 4. Call Tool via MCPANY ---
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
	require.NoError(t, err)
	// require.Len(t, listToolsResult.Tools, 2, "Expected exactly two tools to be registered")

	foundUser := false
	foundRepos := false
	for _, tool := range listToolsResult.Tools {
		if tool.Name == "github.get_user" {
			foundUser = true
		}
		if tool.Name == "github.list_repos" {
			foundRepos = true
		}
	}
	require.True(t, foundUser, "Expected github.get_user tool")
	require.True(t, foundRepos, "Expected github.list_repos tool")

	// --- 5. Test Cases ---
	t.Run("get_user", func(t *testing.T) {
		args := json.RawMessage(`{"username": "octocat"}`)
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "github.get_user", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, res)

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected text content")

		var response map[string]interface{}
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err, "Failed to unmarshal JSON response")

		require.Equal(t, "octocat", response["login"], "The login should match the input")
	})

	t.Run("list_repos", func(t *testing.T) {
		args := json.RawMessage(`{"username": "octocat"}`)
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "github.list_repos", Arguments: args})
		require.NoError(t, err)
		require.NotNil(t, res)

		require.Len(t, res.Content, 1, "Expected exactly one content item")
		textContent, ok := res.Content[0].(*mcp.TextContent)
		require.True(t, ok, "Expected text content")

		var response []interface{}
		err = json.Unmarshal([]byte(textContent.Text), &response)
		require.NoError(t, err, "Failed to unmarshal JSON response")

		require.NotEmpty(t, response, "The response should not be empty")
	})

	t.Log("INFO: E2E Test Scenario for GitHub Server Completed Successfully!")
}
