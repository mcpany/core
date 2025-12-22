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

func TestMonitoringConfig(t *testing.T) {
	// Read the config.yaml file
	content, err := os.ReadFile("config.yaml")
	require.NoError(t, err)

	jsonContent, err := yaml.YAMLToJSON(content)
	require.NoError(t, err)

	cfg := &pb.McpAnyServerConfig{}
	err = protojson.Unmarshal(jsonContent, cfg)
	require.NoError(t, err)

	require.Len(t, cfg.UpstreamServices, 1)
	service := cfg.UpstreamServices[0]

	// Verify that the service definition is valid, which ensures it will be monitored correctly
	require.Equal(t, "monitored-service", service.GetName())
	require.Len(t, service.ServiceConfig.(*pb.UpstreamServiceConfig_HttpService).HttpService.Tools, 1)
	require.Equal(t, "monitored-tool", service.ServiceConfig.(*pb.UpstreamServiceConfig_HttpService).HttpService.Tools[0].GetName())

	err = config.ValidateOrError(context.Background(), service)
	require.NoError(t, err)
}
