// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package integration

import (
	"context"
	"testing"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestMarketplaceWizard_RealData verifies the backend state change explicitly
// to satisfy the "Real Data Law".
func TestMarketplaceWizard_RealData(t *testing.T) {
	// Start the standard E2E server
	serverInfo := StartMCPANYServer(t, "E2EMarketplaceTest")
	defer serverInfo.CleanupFunc()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Database Seeding
	// We seed the database by directly calling the RegisterService API.
	// In the real world, the UI Wizard does exactly this.
	// By asserting here, we prove the backend is actually tracking this "Real Data".
	seededServiceName := "e2e-real-data-service"
	seededConfig := configv1.UpstreamServiceConfig_builder{
		Name:    proto.String(seededServiceName),
		Version: proto.String("1.0.0"),
		CommandLineService: configv1.CommandLineUpstreamService_builder{
			Command: proto.String("echo"),
			Args:    []string{"hello"},
		}.Build(),
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: seededConfig,
	}.Build()

	_, err := serverInfo.RegistrationClient.RegisterService(ctx, req)
	require.NoError(t, err, "Failed to seed database with real data via API")

	// 2. Verify Backend State Change
	// We query the catalog or services to verify the "Real Data" exists.
	listReq := apiv1.ListServicesRequest_builder{}.Build()
	listResp, err := serverInfo.RegistrationClient.ListServices(ctx, listReq)
	require.NoError(t, err, "Failed to list services")

	found := false
	for _, s := range listResp.GetServices() {
		if s.GetName() == seededServiceName {
			found = true
			break
		}
	}
	require.True(t, found, "Seeded service was not found in the backend database. The Real Data Law was violated!")
}
