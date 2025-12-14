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

func TestProfilesConfig(t *testing.T) {
	content, err := os.ReadFile("config.yaml")
	require.NoError(t, err)

	jsonContent, err := yaml.YAMLToJSON(content)
	require.NoError(t, err)

	cfg := &pb.McpAnyServerConfig{}
	err = protojson.Unmarshal(jsonContent, cfg)
	require.NoError(t, err)

	require.Len(t, cfg.UpstreamServices, 1)
	service := cfg.UpstreamServices[0]

	require.Equal(t, "profiled-service", service.GetName())
	require.Len(t, service.Profiles, 2)
	require.Equal(t, "dev", service.Profiles[0].GetName())
	require.Equal(t, "staging", service.Profiles[1].GetName())

	// Validate config using internal validator
	err = config.ValidateOrError(service)
	require.NoError(t, err)
}
