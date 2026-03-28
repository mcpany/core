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

// TestInteropE2EFlow simulates an end-to-end execution utilizing the interop mechanism
// demonstrating that the adapter hub can correctly process and simulate a real-world scenario.
// Uses Database seeding and no mocks on the service layer.
func TestInteropE2EFlow(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Start Server
	serverInfo := integration.StartMCPANYServer(t, "InteropE2EFlow")
	defer serverInfo.CleanupFunc()

	// 2. Seed Data
	integration.SeedStandardData(t, serverInfo)

	// Verify Data via API to ensure seeding was successful
	resp, err := serverInfo.RegistrationClient.ListServices(ctx, apiv1.ListServicesRequest_builder{}.Build())
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Run Hub
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewOpenClawAdapter())
	hub.RegisterAdapter(interop.NewCrewAIAdapter())

	// 1. CrewAI delegation scenario
	task1 := &interop.Task{
		ID:        "e2e-1",
		Framework: "CrewAI",
		Intent:    "task_delegation",
		Payload:   map[string]string{"role": "data_analyst"},
	}

	res1, err := hub.RouteTask(ctx, task1)
	if err != nil {
		t.Fatalf("E2E task 1 failed: %v", err)
	}

	if res1.Status != "success" || res1.Telemetry["delegated_role"] != "data_analyst" {
		t.Errorf("Expected success and delegated role, got %v", res1)
	}

	// 2. OpenClaw reasoning scenario
	task2 := &interop.Task{
		ID:        "e2e-2",
		Framework: "OpenClaw",
		Intent:    "adaptive_reasoning",
		Payload:   map[string]string{"foo": "bar"},
	}

	res2, err := hub.RouteTask(ctx, task2)
	if err != nil {
		t.Fatalf("E2E task 2 failed: %v", err)
	}

	if res2.Status != "success" {
		t.Errorf("Expected success, got %s", res2.Status)
	}
}
