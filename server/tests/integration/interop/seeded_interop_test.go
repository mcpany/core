package interop_test

import (
	"context"
	"testing"
	"github.com/mcpany/core/src/interop"
)

// TestInteropSeededIntegration verifies the interop hub using the actual implementations
// simulating an API-to-API integration with database state preparation.
func TestInteropSeededIntegration(t *testing.T) {
	hub := interop.NewAdapterHub()

	// Simulate "Seeding the Database" by configuring the state of our agents
	hub.RegisterAdapter(interop.NewOpenClawAdapter())
	hub.RegisterAdapter(interop.NewCrewAIAdapter())

	// Simulate an external API call to trigger a multi-agent scenario
	ctx := context.Background()

	t.Run("API_Initiates_CrewAI_Delegation", func(t *testing.T) {
		task1 := &interop.Task{
			ID:        "seeded-api-1",
			Framework: "CrewAI",
			Intent:    "task_delegation",
			Payload:   map[string]string{"role": "data_analyst"},
		}

		res1, err := hub.RouteTask(ctx, task1)
		if err != nil {
			t.Fatalf("API to CrewAI task failed: %v", err)
		}

		if res1.Status != "success" || res1.Telemetry["delegated_role"] != "data_analyst" {
			t.Errorf("Expected success and delegated role, got %v", res1)
		}
	})

	t.Run("API_Triggers_OpenClaw_Reasoning", func(t *testing.T) {
		task2 := &interop.Task{
			ID:        "seeded-api-2",
			Framework: "OpenClaw",
			Intent:    "adaptive_reasoning",
			Payload:   map[string]string{"foo": "bar"},
		}

		res2, err := hub.RouteTask(ctx, task2)
		if err != nil {
			t.Fatalf("API to OpenClaw task failed: %v", err)
		}

		if res2.Status != "success" || res2.Telemetry["entropy_score"] != "low" {
			t.Errorf("Expected success, got %v", res2)
		}
	})

}
