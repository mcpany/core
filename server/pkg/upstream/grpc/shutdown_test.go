// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstream_Shutdown(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewUpstream(poolManager)

	// We need to set serviceID on the upstream, but it's private.
	// However, Register sets it.

	var promptManager prompt.ManagerInterface
	var resourceManager resource.ManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(addr)
	grpcService.SetUseReflection(true) // Enable reflection to make Register work smoothly with mock server

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-shutdown-service")
	serviceConfig.SetGrpcService(grpcService)

	tm := NewMockToolManager()

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
	require.NoError(t, err)
	assert.NotEmpty(t, serviceID)

	err = upstream.Shutdown(context.Background())
	assert.NoError(t, err)
}
