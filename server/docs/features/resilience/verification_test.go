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

func TestResilienceConfig(t *testing.T) {
	content, err := os.ReadFile("config.yaml")
	require.NoError(t, err)

	jsonContent, err := yaml.YAMLToJSON(content)
	require.NoError(t, err)

	cfg := &pb.McpAnyServerConfig{}
	err = protojson.Unmarshal(jsonContent, cfg)
	require.NoError(t, err)

	require.Len(t, cfg.GetUpstreamServices(), 1)
	service := cfg.GetUpstreamServices()[0]

	require.Equal(t, "resilient-service", service.GetName())
	require.NotNil(t, service.GetResilience())
	require.NotNil(t, service.GetResilience().GetCircuitBreaker())
	require.Equal(t, 0.5, service.GetResilience().GetCircuitBreaker().GetFailureRateThreshold())
	require.Equal(t, int64(5), service.GetResilience().GetCircuitBreaker().GetOpenDuration().GetSeconds())

	err = config.ValidateOrError(context.Background(), service)
	require.NoError(t, err)
}
