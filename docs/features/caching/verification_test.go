package features_test

import (
	"os"
	"testing"

	"github.com/mcpany/core/pkg/config"
	pb "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"
)

func TestCachingConfig(t *testing.T) {
	// Read the config.yaml file
	content, err := os.ReadFile("config.yaml")
	require.NoError(t, err)

	// Parse YAML to JSON (since protobuf usually works better with JSON unmarshaling or distinct YAML libs)
	// Here we use sigs.k8s.io/yaml to convert YAML to JSON, then protojson to unmarshal.
	jsonContent, err := yaml.YAMLToJSON(content)
	require.NoError(t, err)

	cfg := &pb.McpAnyServerConfig{}
	err = protojson.Unmarshal(jsonContent, cfg)
	require.NoError(t, err)

	require.Len(t, cfg.UpstreamServices, 1)
	service := cfg.UpstreamServices[0]

	require.Equal(t, "cached-service-example", service.GetName())
	require.NotNil(t, service.Cache)
	require.True(t, service.Cache.GetIsEnabled())
	require.Equal(t, int64(10), service.Cache.GetTtl().GetSeconds())

	err = config.ValidateOrError(service)
	require.NoError(t, err)
}
