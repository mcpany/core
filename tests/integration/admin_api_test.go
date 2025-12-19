// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"testing"
	"time"

	pb "github.com/mcpany/core/proto/admin/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

func TestAdminAPI(t *testing.T) {
	// Start an in-process MCP server
	serverInfo := StartInProcessMCPANYServer(t, "AdminAPITest")
	defer serverInfo.CleanupFunc()

	// Register a dummy HTTP service so we have something to list
	RegisterHTTPService(t, serverInfo.RegistrationClient, "dummy-service", "http://example.com", "dummyOp", "/dummy", "GET", nil)

	// Connect to Admin API
	// Admin API is exposed on the same gRPC port as Registration Service
	conn, err := grpc.NewClient(serverInfo.GrpcRegistrationEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	adminClient := pb.NewAdminServiceClient(conn)

	var serviceID string

	t.Run("ListServices", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := adminClient.ListServices(ctx, &pb.ListServicesRequest{})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Services)

		found := false
		for _, s := range resp.Services {
			if s.GetName() == "dummy-service" {
				found = true
				serviceID = s.GetId()
				break
			}
		}
		assert.True(t, found, "dummy-service should be listed")
	})

	t.Run("GetService", func(t *testing.T) {
		require.NotEmpty(t, serviceID, "Service ID should have been found in ListServices")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := adminClient.GetService(ctx, &pb.GetServiceRequest{Id: proto.String(serviceID)})
		require.NoError(t, err)
		assert.Equal(t, "dummy-service", resp.Config.GetName())
	})

	t.Run("ListTools", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		resp, err := adminClient.ListTools(ctx, &pb.ListToolsRequest{})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Tools)

		found := false
		for _, tool := range resp.Tools {
			if tool.GetName() == "dummyOp" {
				found = true
				break
			}
		}
		assert.True(t, found, "dummyOp tool should be listed")
	})

	t.Run("GetTool", func(t *testing.T) {
		require.NotEmpty(t, serviceID)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Full Tool Name is ServiceID.SanitizedToolName
		// dummyOp is already sanitary
		fullToolName := serviceID + ".dummyOp"

		resp, err := adminClient.GetTool(ctx, &pb.GetToolRequest{Name: proto.String(fullToolName)})
		require.NoError(t, err)
		assert.Equal(t, "dummyOp", resp.Tool.GetName())
	})
}
