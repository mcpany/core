package features_test

import (
	"context"
	"os"
	"testing"

	"github.com/mcpany/core/pkg/config"
	pb "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"
)

func TestRateLimitConfig(t *testing.T) {
	content, err := os.ReadFile("config.yaml")
	require.NoError(t, err)

	jsonContent, err := yaml.YAMLToJSON(content)
	require.NoError(t, err)

	cfg := &pb.McpAnyServerConfig{}
	err = protojson.Unmarshal(jsonContent, cfg)
	require.NoError(t, err)

	require.Len(t, cfg.UpstreamServices, 1)
	service := cfg.UpstreamServices[0]

	require.Equal(t, "rate-limited-service", service.GetName())
	require.NotNil(t, service.RateLimit)
	require.True(t, service.RateLimit.GetIsEnabled())
	require.Equal(t, 50.0, service.RateLimit.GetRequestsPerSecond())
	require.Equal(t, int64(100), service.RateLimit.GetBurst())

	err = config.ValidateOrError(context.Background(), service)
	require.NoError(t, err)
}
