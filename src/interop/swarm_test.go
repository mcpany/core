package interop_test

import (
	"context"
	"strings"
	"testing"

	"github.com/mcpany/core/src/interop"
)

// TestMultiAgentSwarmSimulation simulates a complete end-to-end swarm test
// utilizing multiple agent frameworks managed by the Universal Adapter Hub.
func TestMultiAgentSwarmSimulation(t *testing.T) {
	hub := interop.NewAdapterHub()

	// 1. Register Adapters
	hub.RegisterAdapter(interop.NewOpenClawAdapter())
	hub.RegisterAdapter(interop.NewCrewAIAdapter())
	hub.RegisterAdapter(interop.NewAutoGenAdapter())

	// Context for tasks
	ctx := context.Background()

	// 2. OpenClaw Task: Adaptive Reasoning Simulation
	t.Run("OpenClaw_AdaptiveReasoning", func(t *testing.T) {
		task1 := &interop.Task{
			ID:        "task-oc-001",
			Framework: "OpenClaw",
			Intent:    "adaptive_reasoning",
			Payload:   map[string]string{"context": "high_entropy_data"},
		}

		res1, err := hub.RouteTask(ctx, task1)
		if err != nil {
			t.Fatalf("Failed to execute OpenClaw task: %v", err)
		}

		if res1.Status != "success" {
			t.Errorf("Expected OpenClaw status 'success', got '%s'", res1.Status)
		}

		if res1.Telemetry["entropy_score"] != "low" {
			t.Errorf("Expected OpenClaw low entropy score, got '%s'", res1.Telemetry["entropy_score"])
		}
	})

	// 3. CrewAI Task: Role Delegation
	t.Run("CrewAI_RoleDelegation", func(t *testing.T) {
		task2 := &interop.Task{
			ID:        "task-cai-002",
			Framework: "CrewAI",
			Intent:    "task_delegation",
			Payload:   map[string]string{"role": "data_analyst"},
		}

		res2, err := hub.RouteTask(ctx, task2)
		if err != nil {
			t.Fatalf("Failed to execute CrewAI task: %v", err)
		}

		if !strings.Contains(res2.Output, "data_analyst") {
			t.Errorf("Expected CrewAI output to contain 'data_analyst', got '%s'", res2.Output)
		}

		if res2.Telemetry["auth_status"] != "verified" {
			t.Errorf("Expected CrewAI auth_status to be verified, got '%s'", res2.Telemetry["auth_status"])
		}
	})

	// 4. AutoGen Task: Subagent Checkpoint
	t.Run("AutoGen_SubagentCheckpoint", func(t *testing.T) {
		task3 := &interop.Task{
			ID:        "task-ag-003",
			Framework: "AutoGen",
			Intent:    "subagent_exec",
			Payload:   map[string]string{"action": "compile_report"},
		}

		res3, err := hub.RouteTask(ctx, task3)
		if err != nil {
			t.Fatalf("Failed to execute AutoGen task: %v", err)
		}

		if res3.Telemetry["mailbox_integrity"] != "verified" {
			t.Errorf("Expected AutoGen mailbox integrity 'verified', got '%s'", res3.Telemetry["mailbox_integrity"])
		}

		if res3.Telemetry["history_length"] == "0" {
			t.Errorf("Expected AutoGen chat history to increment, history_length is 0")
		}
	})

	// 5. Error Case: Unsupported Framework
	t.Run("Unsupported_Framework", func(t *testing.T) {
		taskInvalid := &interop.Task{
			ID:        "task-inv-004",
			Framework: "UnknownFramework",
			Intent:    "do_magic",
		}
		_, err := hub.RouteTask(ctx, taskInvalid)
		if err == nil {
			t.Error("Expected error for routing task to unknown framework, got nil")
		}
	})
}
