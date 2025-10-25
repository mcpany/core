package examples

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

	"github.com/mcpxy/core/pkg/consts"
	apiv1 "github.com/mcpxy/core/proto/api/v1"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/mcpxy/core/tests/framework"
	"github.com/mcpxy/core/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"github.com/mcpxy/core/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestGRPCExample(t *testing.T) {
	root, err := integration.GetProjectRoot()
	require.NoError(t, err)

	testCase := &framework.E2ETestCase{
		Name:                "gRPC Example",
		UpstreamServiceType: "grpc",
		RegistrationMethods: []framework.RegistrationMethod{
			framework.GRPCRegistration,
		},
		RegisterUpstream: func(
			t *testing.T,
			registrationClient apiv1.RegistrationServiceClient,
			upstreamEndpoint string,
		) {
			integration.RegisterGRPCService(
				t,
				registrationClient,
				"greeter-service",
				upstreamEndpoint,
				nil,
			)
		},
		BuildUpstream: func(t *testing.T) *integration.ManagedProcess {
			// 1. Generate Protobuf Files
			if os.Getenv("SKIP_PROTO_GENERATION") != "true" {
				generateCmd := exec.Command("./generate.sh")
				generateCmd.Dir = root + "/examples/upstream/grpc/greeter_server"
				err = generateCmd.Run()
				require.NoError(t, err, "Failed to generate protobuf files")
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
			err = tidyCmd.Run()
			require.NoError(t, err, "Failed to tidy go module for gRPC server")

			// Find a free port for the upstream server
			port := integration.FindFreePort(t)

			// 2. Build and run the Upstream gRPC Server
			serverPath := filepath.Join(t.TempDir(), "grpc_greeter_server")
			buildCmd2 := exec.Command("go", "build", "-o", serverPath)
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
			config := configv1.McpxServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{
					configv1.UpstreamServiceConfig_builder{
						Name: proto.String("greeter-service"),
						GrpcService: configv1.GrpcUpstreamService_builder{
							Address: proto.String(
								strings.TrimPrefix(upstreamEndpoint, "http://"),
							),
							UseReflection: proto.Bool(true),
						}.Build(),
					}.Build(),
				},
			}.Build()

			jsonBytes, err := protojson.Marshal(config)
			require.NoError(t, err)
			return string(jsonBytes)
		},
		InvokeAIClient: func(t *testing.T, mcpxyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(
				&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"},
				nil,
			)
			cs, err := testMCPClient.Connect(
				ctx,
				&mcp.StreamableClientTransport{Endpoint: mcpxyEndpoint},
				nil,
			)
			require.NoError(t, err, "Failed to connect to MCPXY server")
			defer cs.Close()

			serviceKey, err := util.GenerateServiceKey("greeter-service")
			require.NoError(t, err)
			toolName := fmt.Sprintf("%s%sSayHello", serviceKey, consts.ToolNameServiceSeparator)

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
