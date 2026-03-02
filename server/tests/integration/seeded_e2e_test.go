// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package integration

import (
	"context"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/stretchr/testify/require"
)

func TestSeededE2E(t *testing.T) {
	t.Log("INFO: Starting Seeded E2E Test...")
	t.Parallel()

	// Use a longer timeout for this specific test as it has timed out previously
	ctx, cancel := context.WithTimeout(context.Background(), TestWaitTimeLong)
	defer cancel()

	// 1. Start Server
	serverInfo := StartMCPANYServer(t, "SeededE2ETest")
	defer serverInfo.CleanupFunc()

	// 2. Seed Data
	SeedStandardData(t, serverInfo)

	// 3. Verify Data via API
	// List services
	resp, err := serverInfo.RegistrationClient.ListServices(ctx, &apiv1.ListServicesRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)

	foundCore := false
	foundTools := false
	for _, svc := range resp.Services {
		if svc.GetName() == "seed-core" {
			foundCore = true
		}
		if svc.GetName() == "seed-tools" {
			foundTools = true
		}
	}
	require.True(t, foundCore, "seed-core service not found")
	require.True(t, foundTools, "seed-tools service not found")

	t.Log("SUCCESS: Seeded services verified.")
}
