// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

func TestAuthenticationConfig(t *testing.T) {
	content, err := os.ReadFile("config.yaml")
	require.NoError(t, err)

	jsonContent, err := yaml.YAMLToJSON(content)
	require.NoError(t, err)

	cfg := &pb.McpAnyServerConfig{}
	err = protojson.Unmarshal(jsonContent, cfg)
	require.NoError(t, err)

	require.Len(t, cfg.GetUpstreamServices(), 2)

	// Verify Incoming Auth
	incomingSvc := cfg.GetUpstreamServices()[0]
	require.Equal(t, "incoming-auth-service", incomingSvc.GetName())
	require.NotNil(t, incomingSvc.GetAuthentication())
	require.NotNil(t, incomingSvc.GetAuthentication().GetApiKey())
	require.Equal(t, "X-Auth", incomingSvc.GetAuthentication().GetApiKey().GetParamName())
	require.Equal(t, pb.APIKeyAuth_HEADER, incomingSvc.GetAuthentication().GetApiKey().GetIn())
	require.Equal(t, "s3cret", incomingSvc.GetAuthentication().GetApiKey().GetVerificationValue())

	// Verify Outgoing Auth
	outgoingSvc := cfg.GetUpstreamServices()[1]
	require.Equal(t, "outgoing-auth-service", outgoingSvc.GetName())
	require.NotNil(t, outgoingSvc.GetUpstreamAuth())
	require.NotNil(t, outgoingSvc.GetUpstreamAuth().GetBasicAuth())
	require.Equal(t, "admin", outgoingSvc.GetUpstreamAuth().GetBasicAuth().GetUsername())
	require.Equal(t, "password123", outgoingSvc.GetUpstreamAuth().GetBasicAuth().GetPassword().GetPlainText())

	err = config.ValidateOrError(context.Background(), incomingSvc)
	require.NoError(t, err)

	err = config.ValidateOrError(context.Background(), outgoingSvc)
	require.NoError(t, err)
}
