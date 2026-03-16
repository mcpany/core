// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGRPCProvider_Name(t *testing.T) {
	provider := &GRPCProvider{}
	assert.Equal(t, "grpc", provider.Name())
}

func TestGRPCProvider_Discover(t *testing.T) {
	provider := &GRPCProvider{Endpoint: "localhost:50051"}

	svcs, err := provider.Discover(context.Background())
	require.NoError(t, err)
	require.Len(t, svcs, 1)

	svc := svcs[0]
	assert.Equal(t, "Auto-discovered gRPC", svc.GetName())
	assert.Equal(t, "localhost:50051", svc.GetGrpcService().GetAddress())
	assert.True(t, svc.GetGrpcService().GetUseReflection())
	assert.Contains(t, svc.GetTags(), "grpc")
	assert.Contains(t, svc.GetTags(), "auto-discovered")
}
