package features_test

import (
	"context"
	"os"
	"testing"

	pb "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"
)

func TestMonitoringConfig(t *testing.T) {
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

	// Verify that the service definition is valid, which ensures it will be monitored correctly
	require.Equal(t, "monitored-service", service.GetName())
	require.Len(t, service.GetHttpService().GetTools(), 1)
	require.Equal(t, "monitored-tool", service.GetHttpService().GetTools()[0].GetName())

	err = config.ValidateOrError(context.Background(), service)
	require.NoError(t, err)
}
