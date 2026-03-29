package interop_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/src/interop"
)

func TestHubAndAdapters(t *testing.T) {
	ctx := context.Background()
	hub := interop.NewAdapterHub()

	oc := interop.NewOpenClawAdapter()
	cai := interop.NewCrewAIAdapter()
	ag := interop.NewAutoGenAdapter()

	hub.RegisterAdapter(oc)
	hub.RegisterAdapter(cai)
	hub.RegisterAdapter(ag)

	// 1. Test Name Mapping
	if oc.Name() != "OpenClaw" {
		t.Errorf("Expected OpenClaw, got %s", oc.Name())
	}
	if cai.Name() != "CrewAI" {
		t.Errorf("Expected CrewAI, got %s", cai.Name())
	}
	if ag.Name() != "AutoGen" {
		t.Errorf("Expected AutoGen, got %s", ag.Name())
	}

	// 2. Test Capability Routing
	t.Run("Routing", func(t *testing.T) {
		task1 := &interop.Task{
			ID:        "task-1",
			Framework: "OpenClaw",
			Intent:    "adaptive_reasoning",
			Payload:   map[string]string{"input": "test-oc"},
		}

		res1, err := hub.RouteTask(ctx, task1)
		if err != nil {
			t.Fatalf("Failed to route to OpenClaw: %v", err)
		}
		if res1.Status != "success" {
			t.Errorf("Expected success, got %s", res1.Status)
		}

		task2 := &interop.Task{
			ID:        "task-2",
			Framework: "CrewAI",
			Intent:    "role_delegation",
			Payload:   map[string]string{"role": "researcher"},
		}

		res2, err := hub.RouteTask(ctx, task2)
		if err != nil {
			t.Fatalf("Failed to route to CrewAI: %v", err)
		}
		if res2.Telemetry["delegated_role"] != "researcher" {
			t.Errorf("Expected researcher role, got %s", res2.Telemetry["delegated_role"])
		}
	})

	// 3. Test Unsupported Capability
	t.Run("Unsupported", func(t *testing.T) {
		task := &interop.Task{
			ID:        "task-unsupported",
			Framework: "AutoGen",
			Intent:    "non_existent_capability",
		}
		_, err := hub.RouteTask(ctx, task)
		if err == nil {
			t.Error("Expected error for unsupported capability, got nil")
		}
	})

	// 4. Test Multi-Agent Conversation (AutoGen)
	t.Run("AutoGen_Conversation", func(t *testing.T) {
		task := &interop.Task{
			ID:        "task-ag-conv",
			Framework: "AutoGen",
			Intent:    "multi_agent_convo",
			Payload:   map[string]string{"agents": "assistant,coder"},
		}
		res, err := hub.RouteTask(ctx, task)
		if err != nil {
			t.Fatalf("AutoGen task failed: %v", err)
		}
		if res.Telemetry["convo_depth"] != "3" {
			t.Errorf("Expected convo depth 3, got %s", res.Telemetry["convo_depth"])
		}
	})

	// 5. Test Memory Shard Sync
	t.Run("MemorySync", func(t *testing.T) {
		validShard := &interop.MemoryShard{
			ShardID:   "shard-123",
			Intent:    "global_context",
			Signature: "valid-sig",
		}
		invalidShard := &interop.MemoryShard{
			ShardID: "shard-456",
		}

		adapters := map[string]interop.AgentFramework{
			"OpenClaw": oc,
			"CrewAI":   cai,
			"AutoGen":  ag,
		}

		for name, adapter := range adapters {
			err := adapter.SyncMemoryShard(ctx, validShard)
			if err != nil {
				t.Errorf("Expected successful sync for %s, got error: %v", name, err)
			}

			err = adapter.SyncMemoryShard(ctx, invalidShard)
			if err == nil {
				t.Errorf("Expected sync to fail for %s due to missing signature", name)
			}
		}
	})

	// 7. Streaming Task Test
	t.Run("OpenClaw_Streaming", func(t *testing.T) {
		taskStream := &interop.Task{
			ID:        "task-oc-stream",
			Framework: "OpenClaw",
			Intent:    "adaptive_reasoning",
			Payload:   map[string]string{"context": "streaming_data", "stream": "true"},
		}

		resStream, err := hub.RouteTask(ctx, taskStream)
		if err != nil {
			t.Fatalf("Failed to execute OpenClaw streaming task: %v", err)
		}

		if resStream.Stream == nil {
			t.Fatal("Expected Stream channel to be populated")
		}

		chunks := []string{}
		for chunk := range resStream.Stream {
			chunks = append(chunks, chunk)
		}

		if len(chunks) != 2 {
			t.Errorf("Expected 2 chunks, got %d", len(chunks))
		}
	})
}

// Helper to access registered adapters for testing interface direct calls
func getAdapterByName(hub *interop.AdapterHub, name string) interop.AgentFramework {
	// Let's create a temporary hub-like registry or instantiate directly.
	switch name {
	case "OpenClaw":
		return interop.NewOpenClawAdapter()
	case "CrewAI":
		return interop.NewCrewAIAdapter()
	case "AutoGen":
		return interop.NewAutoGenAdapter()
	}
	return nil
}
