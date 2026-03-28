package interop_test

import (
	"context"
	"github.com/mcpany/core/src/interop"
	"testing"
)

// TestInteropE2EFlow simulates an end-to-end execution utilizing the interop mechanism
// demonstrating that the adapter hub can correctly process and simulate a real-world scenario.
func TestInteropE2EFlow(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewOpenClawAdapter())
	hub.RegisterAdapter(interop.NewCrewAIAdapter())

	ctx := context.Background()

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
