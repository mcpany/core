// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resilience

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestHTTPResilience(t *testing.T) {
	t.Parallel()
	t.Run("retry", func(t *testing.T) {
		t.Parallel()
		var requestCount int32
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if atomic.AddInt32(&requestCount, 1) < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, _ = fmt.Fprintln(w, `{"message": "ok"}`)
		}))
		defer ts.Close()

		testCase := &framework.E2ETestCase{
			Name:                "HTTP Retry",
			UpstreamServiceType: "http",
			BuildUpstream: func(_ *testing.T) *integration.ManagedProcess {
				return &integration.ManagedProcess{}
			},
			GenerateUpstreamConfig: func(_ string) string {
				return fmt.Sprintf(`
upstreamServices:
  - name: "retry-service"
    httpService:
      address: "%s"
      calls:
        echo:
          method: "HTTP_METHOD_GET"
          endpointPath: "/"
      tools:
        - callId: "echo"
          name: "echo"
          description: "test tool"
    resilience:
      retryPolicy:
        numberOfRetries: 3
        baseBackoff: "10ms"
`, ts.URL)
			},
			InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
				ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
				defer cancel()

				testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
				cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
				require.NoError(t, err)
				defer func() { _ = cs.Close() }()

				serviceID, _ := util.SanitizeServiceName("retry-service")
				sanitizedToolName, _ := util.SanitizeToolName("echo")
				toolName := serviceID + "." + sanitizedToolName

				res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})
				require.NoError(t, err)
				require.NotNil(t, res)
				switch content := res.Content[0].(type) {
				case *mcp.TextContent:
					require.JSONEq(t, `{"message": "ok"}`, content.Text)
				default:
					t.Fatalf("Unexpected content type: %T", content)
				}
				require.Equal(t, int32(3), requestCount)
			},
		}
		framework.RunE2ETest(t, testCase)
	})

	t.Run("circuit_breaker", func(t *testing.T) {
		t.Parallel()
		var requestCount int32
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			atomic.AddInt32(&requestCount, 1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		testCase := &framework.E2ETestCase{
			Name:                "HTTP Circuit Breaker",
			UpstreamServiceType: "http",
			BuildUpstream: func(_ *testing.T) *integration.ManagedProcess {
				return &integration.ManagedProcess{}
			},
			GenerateUpstreamConfig: func(_ string) string {
				return fmt.Sprintf(`
upstreamServices:
  - name: "cb-service"
    httpService:
      address: "%s"
      calls:
        echo:
          method: "HTTP_METHOD_GET"
          endpointPath: "/"
      tools:
        - callId: "echo"
          name: "echo"
          description: "test tool"
    resilience:
      circuitBreaker:
        consecutiveFailures: 2
        openDuration: "100ms"
`, ts.URL)
			},
			InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
				ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
				defer cancel()

				testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
				cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
				require.NoError(t, err)
				defer func() { _ = cs.Close() }()

				serviceID, _ := util.SanitizeServiceName("cb-service")
				sanitizedToolName, _ := util.SanitizeToolName("echo")
				toolName := serviceID + "." + sanitizedToolName

				// First 2 requests should fail and open the circuit
				_, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})
				require.Error(t, err)
				_, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})
				require.Error(t, err)
				require.Equal(t, int32(2), atomic.LoadInt32(&requestCount))

				// Third request should be blocked by the open circuit
				_, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})
				require.Error(t, err)
				require.Contains(t, err.Error(), "circuit breaker is open")
				require.Equal(t, int32(2), atomic.LoadInt32(&requestCount))

				// Wait for the open duration to elapse
				time.Sleep(150 * time.Millisecond)

				// Fourth request should be allowed (half-open)
				atomic.StoreInt32(&requestCount, 0)
				ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					atomic.AddInt32(&requestCount, 1)
					_, _ = fmt.Fprintln(w, `{"message": "ok"}`)
				})
				res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(`{}`)})
				require.NoError(t, err)
				require.NotNil(t, res)
				switch content := res.Content[0].(type) {
				case *mcp.TextContent:
					require.JSONEq(t, `{"message": "ok"}`, content.Text)
				default:
					t.Fatalf("Unexpected content type: %T", content)
				}
				require.Equal(t, int32(1), atomic.LoadInt32(&requestCount))
			},
		}
		framework.RunE2ETest(t, testCase)
	})
}
