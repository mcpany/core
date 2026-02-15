//go:build e2e

package features_test

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/config"
	adminv1 "github.com/mcpany/core/proto/admin/v1"
	pb "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"
)

// TestCachingConfig is a unit test for the configuration.
func TestCachingConfig(t *testing.T) {
	// Read the config.yaml file
	content, err := os.ReadFile("config.yaml")
	require.NoError(t, err)

	jsonContent, err := yaml.YAMLToJSON(content)
	require.NoError(t, err)

	cfg := &pb.McpAnyServerConfig{}
	err = protojson.Unmarshal(jsonContent, cfg)
	require.NoError(t, err)

	require.Len(t, cfg.GetUpstreamServices(), 1)
	service := cfg.GetUpstreamServices()[0]

	require.Equal(t, "cached-weather-service", service.GetName())
	require.NotNil(t, service.Cache)
	require.True(t, service.Cache.GetIsEnabled())
	// The config in README says "1h" for service, "5m" for call.
	// We verify basic structure here.
	err = config.ValidateOrError(context.Background(), service)
	require.NoError(t, err)
}

func TestCachingE2E(t *testing.T) {
	metricsPort := 0 // Use 0 to let the OS assign a random available port
	os.Setenv("MCPANY_METRICS_LISTEN_ADDRESS", fmt.Sprintf("127.0.0.1:%d", metricsPort))
	defer os.Unsetenv("MCPANY_METRICS_LISTEN_ADDRESS")

	testCase := &framework.E2ETestCase{
		Name:                "Caching_Verification_Weather",
		UpstreamServiceType: "openapi", // Not used for FileRegistration?
		BuildUpstream:       framework.BuildWebsocketWeatherServer,
		RegistrationMethods: []framework.RegistrationMethod{framework.FileRegistration},
		GenerateUpstreamConfig: func(upstreamAddr string) string {
			// upstreamAddr is 127.0.0.1:PORT
			return fmt.Sprintf(`
upstream_services:
  - name: "cached-weather-service"
    http_service:
      address: "http://%s"
      tools:
        - name: "weather"
          description: "Get weather for a location"
          call_id: "get_weather"
          input_schema:
            type: "object"
            properties:
              location:
                type: "string"
            required: ["location"]
      calls:
        get_weather:
          method: HTTP_METHOD_POST
          endpoint_path: "/weather"
          input_schema:
            type: object
            properties:
              location:
                type: string
            required:
              - location
    cache:
      is_enabled: true
      ttl: "2s" # Short TTL for expiration test
global_settings:
  log_level: LOG_LEVEL_DEBUG
  mcp_listen_address: "127.0.0.1:0" # Random port
`, upstreamAddr)
		},
		InvokeAIClientWithServerInfo: func(t *testing.T, serverInfo *integration.MCPANYTestServerInfo) {
			mcpanyEndpoint := serverInfo.HTTPEndpoint
			// Connect directly to MCP server using mcp-go SDK
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
			// Transport
			transport := &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}
			session, err := client.Connect(ctx, transport, nil)
			require.NoError(t, err)
			defer func() { _ = session.Close() }()

			// Helper to get metrics
			getMetric := func(name string, labelKey, labelValue string) float64 {
				port := metricsPort
				if serverInfo != nil && serverInfo.MetricsEndpoint != "" {
					// Extract port from "host:port"
					parts := strings.Split(serverInfo.MetricsEndpoint, ":")
					if len(parts) == 2 {
						fmt.Sscanf(parts[1], "%d", &port)
					}
				}
				if port == 0 {
					// Fallback for safety, though it should be set
					port = 19091
				}

				resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/metrics", port))
				require.NoError(t, err)
				defer resp.Body.Close()

				scanner := bufio.NewScanner(resp.Body)
				for scanner.Scan() {
					line := scanner.Text()
					if strings.HasPrefix(line, name) {
						// Simple parsing: name{labelKey="labelValue"} value
						if strings.Contains(line, fmt.Sprintf(`%s="%s"`, labelKey, labelValue)) ||
							(labelKey == "" && !strings.Contains(line, "{")) {
							parts := strings.Fields(line)
							if len(parts) >= 2 {
								var val float64
								fmt.Sscanf(parts[len(parts)-1], "%f", &val)
								return val
							}
						}
					}
				}
				return 0
			}

			params := "cached-weather-service.weather"
			_ = params // unused for now

			// Initial state
			hitsBefore := getMetric("mcpany_cache_hits", "tool", "weather")
			missesBefore := getMetric("mcpany_cache_misses", "tool", "weather")

			// 1. Call Tool: weather (london) - EXPECT MISS
			reqLondon := &mcp.CallToolParams{
				Name: "cached-weather-service.weather",
				Arguments: map[string]interface{}{
					"location": "london",
				},
			}
			_, err = session.CallTool(ctx, reqLondon)
			require.NoError(t, err)

			// Allow metric flush?
			time.Sleep(100 * time.Millisecond)

			hits1 := getMetric("mcpany_cache_hits", "tool", "weather")
			misses1 := getMetric("mcpany_cache_misses", "tool", "weather")
			require.Equal(t, missesBefore+1, misses1, "Should be a cache miss")
			require.Equal(t, hitsBefore, hits1, "Should not be a cache hit")

			// 2. Call again (london) - EXPECT HIT
			_, err = session.CallTool(ctx, reqLondon)
			require.NoError(t, err)

			time.Sleep(100 * time.Millisecond)
			hits2 := getMetric("mcpany_cache_hits", "tool", "weather")
			misses2 := getMetric("mcpany_cache_misses", "tool", "weather")
			require.Equal(t, misses1, misses2, "Should NOT be a new miss")
			require.Equal(t, hits1+1, hits2, "Should be a cache hit")

			// 3. Call with different param (tokyo) - EXPECT MISS
			reqTokyo := &mcp.CallToolParams{
				Name: "cached-weather-service.weather",
				Arguments: map[string]interface{}{
					"location": "tokyo",
				},
			}
			_, err = session.CallTool(ctx, reqTokyo)
			require.NoError(t, err)

			time.Sleep(100 * time.Millisecond)
			hits3 := getMetric("mcpany_cache_hits", "tool", "weather")
			misses3 := getMetric("mcpany_cache_misses", "tool", "weather")
			require.Equal(t, misses2+1, misses3, "Should be a cache miss for different param")
			require.Equal(t, hits2, hits3, "Should not be a cache hit for different param")

			// 4. Test Expiration (TTL = 2s)
			time.Sleep(3 * time.Second)
			_, err = session.CallTool(ctx, reqLondon)
			require.NoError(t, err)

			time.Sleep(100 * time.Millisecond)
			hits4 := getMetric("mcpany_cache_hits", "tool", "weather")
			misses4 := getMetric("mcpany_cache_misses", "tool", "weather")
			require.Equal(t, misses3+1, misses4, "Should be a miss after expiration")
			require.Equal(t, hits3, hits4, "Should not be a hit after expiration")

			// 5. Test ClearCache
			t.Logf("Connecting to gRPC endpoint: %s", serverInfo.GrpcRegistrationEndpoint)

			conn, err := grpc.Dial(serverInfo.GrpcRegistrationEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
			require.NoError(t, err)
			defer conn.Close()

			adminClient := adminv1.NewAdminServiceClient(conn)
			_, err = adminClient.ClearCache(ctx, &adminv1.ClearCacheRequest{})
			require.NoError(t, err, "Failed to call ClearCache")

			// Call London again - EXPECT MISS (cache was cleared (re-miss), previous entry deleted)
			// Wait, the previous entry for 'london' expired in step 4, then we refreshed it (miss)?
			// Yes, step 4 was a Miss, so 'london' was re-cached.
			// Now we hit ClearCache. 'london' should be gone.
			// Next call should be Miss.

			_, err = session.CallTool(ctx, reqLondon)
			require.NoError(t, err)

			time.Sleep(100 * time.Millisecond)
			hits5 := getMetric("mcpany_cache_hits", "tool", "weather")
			misses5 := getMetric("mcpany_cache_misses", "tool", "weather")

			require.Equal(t, misses4+1, misses5, "Should be a miss after clear cache")
			require.Equal(t, hits4, hits5, "Should not be a hit after clear cache")
		},
	}

	// Set metrics env var
	os.Setenv("MCPANY_METRICS_LISTEN_ADDRESS", fmt.Sprintf("127.0.0.1:%d", metricsPort))
	defer os.Unsetenv("MCPANY_METRICS_LISTEN_ADDRESS")

	framework.RunE2ETest(t, testCase)
}
