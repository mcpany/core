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

func TestPromptsConfig(t *testing.T) {
	content, err := os.ReadFile("config.yaml")
	require.NoError(t, err)

	jsonContent, err := yaml.YAMLToJSON(content)
	require.NoError(t, err)

	cfg := &pb.McpAnyServerConfig{}
	err = protojson.Unmarshal(jsonContent, cfg)
	require.NoError(t, err)

	require.Len(t, cfg.UpstreamServices, 1)
	service := cfg.UpstreamServices[0]

	httpSvc := service.ServiceConfig.(*pb.UpstreamServiceConfig_HttpService).HttpService
	require.Len(t, httpSvc.Prompts, 1)
	prompt := httpSvc.Prompts[0]

	require.Equal(t, "hello-world", prompt.GetName())
	require.Equal(t, "A simple hello world prompt", prompt.GetDescription())
	require.Len(t, prompt.Messages, 1)
	require.Equal(t, pb.PromptMessage_USER, prompt.Messages[0].GetRole())
	require.Equal(t, "Hello world", prompt.Messages[0].GetText().GetText())

	err = config.ValidateOrError(context.Background(), service)
	require.NoError(t, err)
}
