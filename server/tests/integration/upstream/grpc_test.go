// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestUpstreamService_GRPC(t *testing.T) {
	testCase := &framework.E2ETestCase{
		Name:                "gRPC Weather Server",
		UpstreamServiceType: "grpc",
		BuildUpstream:       framework.BuildGRPCWeatherServer,
		RegisterUpstream:    framework.RegisterGRPCWeatherService,
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer func() { _ = cs.Close() }()

			listToolsResult, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
			require.NoError(t, err)
			for _, tool := range listToolsResult.Tools {
				t.Logf("Discovered tool from MCPANY: %s", tool.Name)
			}

			const weatherServiceID = "e2e_grpc_weather"
			serviceID, _ := util.SanitizeServiceName(weatherServiceID)
			sanitizedToolName, _ := util.SanitizeToolName("GetWeather")
			toolName := serviceID + "." + sanitizedToolName
			addArgs := `{"location": "london"}`
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(addArgs)})
			require.NoError(t, err, "Error calling GetWeather tool")
			require.NotNil(t, res, "Nil response from GetWeather tool")
			switch content := res.Content[0].(type) {
			case *mcp.TextContent:
				require.JSONEq(t, `{"weather": "Cloudy, 15°C"}`, content.Text, "The weather is incorrect")
			default:
				t.Fatalf("Unexpected content type: %T", content)
			}
		},
	}

	framework.RunE2ETest(t, testCase)
}

func TestUpstreamService_GRPCExample(t *testing.T) {
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	testCase := &framework.E2ETestCase{
		Name:                "gRPC Greeter Example",
		UpstreamServiceType: "grpc",
		BuildUpstream: func(t *testing.T) *integration.ManagedProcess {
			// 1. Generate Protobuf Files
			if os.Getenv("SKIP_PROTO_GENERATION") != "true" {
				generateCmd := exec.Command("./generate.sh")
				generateCmd.Dir = root + "/examples/upstream_service_demo/grpc/greeter_server"
				if err := generateCmd.Run(); err != nil {
					require.NoError(t, err, "Failed to generate protobuf files")
				}
			}

			// Tidy dependencies for the upstream server
			serverDir := filepath.Join(
				root,
				"examples",
				"upstream",
				"grpc",
				"greeter_server",
				"server",
			)
			tidyCmd := exec.Command("go", "mod", "tidy")
			tidyCmd.Dir = serverDir
			if err := tidyCmd.Run(); err != nil {
				require.NoError(t, err, "Failed to tidy go module for gRPC server")
			}

			// Find a free port for the upstream server
			port := integration.FindFreePort(t)

			// 2. Build and run the Upstream gRPC Server
			serverPath := filepath.Join(t.TempDir(), "grpc_greeter_server")
			buildCmd2 := exec.Command("go", "build", "-o", serverPath) //nolint:gosec // Test utility
			buildCmd2.Dir = filepath.Join(
				root,
				"examples",
				"upstream",
				"grpc",
				"greeter_server",
				"server",
			)
			buildOutput, err := buildCmd2.CombinedOutput()
			require.NoError(
				t,
				err,
				"Failed to build gRPC server binary. Output:\n%s",
				string(buildOutput),
			)

			upstreamServerProcess := integration.NewManagedProcess(
				t,
				"upstream-grpc-server",
				serverPath,
				nil,
				[]string{"GRPC_PORT=" + strconv.Itoa(port), "SKIP_PROTO_GENERATION=true"},
			)
			upstreamServerProcess.Port = port
			return upstreamServerProcess
		},
		GenerateUpstreamConfig: func(upstreamEndpoint string) string {
			return `{"upstream_services": [{"name": "greeter-service", "grpc_service": {"address": "` + strings.TrimPrefix(upstreamEndpoint, "http://") + `", "use_reflection": true}}]}`
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(
				&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"},
				nil,
			)
			cs, err := testMCPClient.Connect(
				ctx,
				&mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint},
				nil,
			)
			require.NoError(t, err, "Failed to connect to MCPANY server")
			defer func() { _ = cs.Close() }()

			serviceID, err := util.SanitizeServiceName("greeter-service")
			require.NoError(t, err)
			sanitizedToolName, err := util.SanitizeToolName("SayHello")
			require.NoError(t, err)
			toolName := fmt.Sprintf("%s.%s", serviceID, sanitizedToolName)

			// Wait for the tool to be available
			require.Eventually(t, func() bool {
				result, err := cs.ListTools(ctx, &mcp.ListToolsParams{})
				if err != nil {
					t.Logf("Failed to list tools: %v", err)
					return false
				}
				for _, tool := range result.Tools {
					if tool.Name == toolName {
						return true
					}
				}
				t.Logf("Tool %s not yet available", toolName)
				return false
			}, integration.TestWaitTimeLong, 1*time.Second, "Tool %s did not become available in time", toolName)

			params := json.RawMessage(`{"name": "World"}`)

			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: params})
			require.NoError(t, err, "Error calling tool '%s'", toolName)
			require.NotNil(t, res, "Nil response from tool '%s'", toolName)
			require.Len(
				t,
				res.Content,
				1,
				"Expected exactly one content item from tool '%s'",
				toolName,
			)

			textContent, ok := res.Content[0].(*mcp.TextContent)
			require.True(t, ok, "Expected content to be of type TextContent")

			var jsonResponse map[string]interface{}
			err = json.Unmarshal([]byte(textContent.Text), &jsonResponse)
			require.NoError(t, err, "Failed to unmarshal JSON response from tool")

			require.Equal(t, "Hello World", jsonResponse["message"])
		},
	}

	framework.RunE2ETest(t, testCase)
}
