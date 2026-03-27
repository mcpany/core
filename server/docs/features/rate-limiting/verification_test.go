// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package features_test

import (
	"context"
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/config"
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

	require.Len(t, cfg.GetUpstreamServices(), 1)
	service := cfg.GetUpstreamServices()[0]

	require.Equal(t, "rate-limited-service", service.GetName())
	require.NotNil(t, service.GetRateLimit())
	require.True(t, service.GetRateLimit().GetIsEnabled())
	require.Equal(t, 50.0, service.GetRateLimit().GetRequestsPerSecond())
	require.Equal(t, int64(100), service.GetRateLimit().GetBurst())

	err = config.ValidateOrError(context.Background(), service)
	require.NoError(t, err)
}
