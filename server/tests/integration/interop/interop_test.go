// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package interop_test

import (
	"context"
	"testing"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/src/interop"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/stretchr/testify/require"
)

// TestInteropIntegration verifies the interop hub using the actual implementations.
// Data MUST be seeded via the database. NO data mocks in the backend are allowed.
func TestInteropIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Start Server
	serverInfo := integration.StartMCPANYServer(t, "InteropIntegrationTest")
	defer serverInfo.CleanupFunc()

	// 2. Seed Data
	integration.SeedStandardData(t, serverInfo)

	// Verify Data via API to ensure seeding was successful
	resp, err := serverInfo.RegistrationClient.ListServices(ctx, apiv1.ListServicesRequest_builder{}.Build())
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Run Interop Hub Logic
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewOpenClawAdapter())

	task := &interop.Task{
		ID:        "int-1",
		Framework: "OpenClaw",
		Intent:    "adaptive_reasoning",
		Payload:   map[string]string{"foo": "bar"},
	}

	res, err := hub.RouteTask(ctx, task)
	if err != nil {
		t.Fatalf("Integration task failed: %v", err)
	}

	if res.Status != "success" {
		t.Errorf("Expected success, got %s", res.Status)
	}
}
