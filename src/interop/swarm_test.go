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

	t.Run("OpenClaw_UnsupportedCapability", func(t *testing.T) {
		taskOCUnsupported := &interop.Task{
			ID:        "task-oc-unsup",
			Framework: "OpenClaw",
			Intent:    "unsupported_intent",
			Payload:   map[string]string{"data": "test"},
		}

		_, err := hub.RouteTask(ctx, taskOCUnsupported)
		if err == nil {
			t.Error("Expected error for unsupported OpenClaw capability, got nil")
		} else if err.Error() != "OpenClaw does not support capability: unsupported_intent" {
			t.Errorf("Unexpected error message: %v", err)
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

	t.Run("CrewAI_DefaultRole", func(t *testing.T) {
		taskCAIDefaultRole := &interop.Task{
			ID:        "task-cai-003",
			Framework: "CrewAI",
			Intent:    "task_delegation",
		}

		res, err := hub.RouteTask(ctx, taskCAIDefaultRole)
		if err != nil {
			t.Fatalf("Failed to execute CrewAI task: %v", err)
		}

		if res.Telemetry["delegated_role"] != "generalist" {
			t.Errorf("Expected CrewAI delegated_role to be 'generalist', got '%s'", res.Telemetry["delegated_role"])
		}
	})

	t.Run("CrewAI_UnsupportedCapability", func(t *testing.T) {
		taskCAIUnsupported := &interop.Task{
			ID:        "task-cai-unsup",
			Framework: "CrewAI",
			Intent:    "unsupported_intent",
		}

		_, err := hub.RouteTask(ctx, taskCAIUnsupported)
		if err == nil {
			t.Error("Expected error for unsupported CrewAI capability, got nil")
		} else if err.Error() != "CrewAI does not support capability: unsupported_intent" {
			t.Errorf("Unexpected error message: %v", err)
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

	t.Run("AutoGen_UnsupportedCapability", func(t *testing.T) {
		taskAGUnsupported := &interop.Task{
			ID:        "task-ag-unsup",
			Framework: "AutoGen",
			Intent:    "unsupported_intent",
		}

		_, err := hub.RouteTask(ctx, taskAGUnsupported)
		if err == nil {
			t.Error("Expected error for unsupported AutoGen capability, got nil")
		} else if err.Error() != "AutoGen does not support capability: unsupported_intent" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	// 5. Streaming Task Execution
	t.Run("OpenClaw_StreamingTask", func(t *testing.T) {
		taskStream := &interop.Task{
			ID:        "task-oc-stream",
			Framework: "OpenClaw",
			Intent:    "adaptive_reasoning",
		}

		stream, err := hub.StreamRouteTask(ctx, taskStream)
		if err != nil {
			t.Fatalf("Failed to start streaming task: %v", err)
		}

		var results []*interop.TaskResult
		for res := range stream {
			results = append(results, res)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 streaming chunks, got %d", len(results))
		}

		if results[0].Status != "streaming" {
			t.Errorf("Expected first chunk status 'streaming', got '%s'", results[0].Status)
		}

		if results[2].Status != "success" {
			t.Errorf("Expected final chunk status 'success', got '%s'", results[2].Status)
		}
	})

	// 6. Error Case: Unsupported Framework
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

	t.Run("Unsupported_Framework_Stream", func(t *testing.T) {
		taskInvalid := &interop.Task{
			ID:        "task-inv-005",
			Framework: "UnknownFramework",
			Intent:    "do_magic",
		}
		_, err := hub.StreamRouteTask(ctx, taskInvalid)
		if err == nil {
			t.Error("Expected error for streaming task to unknown framework, got nil")
		}
	})

	// 7. UMMB Sync Memory Shard Test
	t.Run("UMMB_SyncMemoryShard", func(t *testing.T) {
		validShard := &interop.MemoryShard{
			ShardID:           "shard-100",
			Intent:            "cross-framework-sync",
			TextContent:       "Mission context update.",
			MultimodalPayload: []byte("<svg>mocked</svg>"),
			Signature:         "valid_hardware_signature",
		}

		invalidShard := &interop.MemoryShard{
			ShardID:           "shard-101",
			Intent:            "unverified-sync",
			TextContent:       "Malicious context.",
			MultimodalPayload: []byte("<svg>malicious</svg>"),
			Signature:         "", // Missing signature should fail
		}

		adapters := []string{"OpenClaw", "CrewAI", "AutoGen"}
		for _, name := range adapters {
			adapter := getAdapterByName(hub, name)

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

	// 8. HandleTask Streaming
	t.Run("AutoGen_HandleTask_Streaming", func(t *testing.T) {
		taskStream := &interop.Task{
			ID:        "task-ag-stream-handletask",
			Framework: "AutoGen",
			Intent:    "subagent_exec",
			Payload:   map[string]string{"stream": "true"},
		}

		resStream, err := hub.RouteTask(ctx, taskStream)
		if err != nil {
			t.Fatalf("Failed to execute AutoGen HandleTask streaming: %v", err)
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

	t.Run("CrewAI_HandleTask_Streaming", func(t *testing.T) {
		taskStream := &interop.Task{
			ID:        "task-cai-stream-handletask",
			Framework: "CrewAI",
			Intent:    "task_delegation",
			Payload:   map[string]string{"stream": "true"},
		}

		resStream, err := hub.RouteTask(ctx, taskStream)
		if err != nil {
			t.Fatalf("Failed to execute CrewAI HandleTask streaming: %v", err)
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

	// 9. StreamRouteTask Tests
	t.Run("AutoGen_StreamRouteTask", func(t *testing.T) {
		taskStream := &interop.Task{
			ID:        "task-ag-streamroute",
			Framework: "AutoGen",
			Intent:    "subagent_exec",
		}

		stream, err := hub.StreamRouteTask(ctx, taskStream)
		if err != nil {
			t.Fatalf("Failed to start streaming task: %v", err)
		}

		var results []*interop.TaskResult
		for res := range stream {
			results = append(results, res)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 streaming chunks, got %d", len(results))
		}

		if results[0].Status != "streaming" {
			t.Errorf("Expected first chunk status 'streaming', got '%s'", results[0].Status)
		}

		if results[1].Status != "success" {
			t.Errorf("Expected final chunk status 'success', got '%s'", results[1].Status)
		}
	})

	t.Run("CrewAI_StreamRouteTask", func(t *testing.T) {
		taskStream := &interop.Task{
			ID:        "task-cai-streamroute",
			Framework: "CrewAI",
			Intent:    "task_delegation",
		}

		stream, err := hub.StreamRouteTask(ctx, taskStream)
		if err != nil {
			t.Fatalf("Failed to start streaming task: %v", err)
		}

		var results []*interop.TaskResult
		for res := range stream {
			results = append(results, res)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 streaming chunks, got %d", len(results))
		}

		if results[0].Status != "streaming" {
			t.Errorf("Expected first chunk status 'streaming', got '%s'", results[0].Status)
		}

		if results[1].Status != "success" {
			t.Errorf("Expected final chunk status 'success', got '%s'", results[1].Status)
		}
	})

	t.Run("OpenClaw_StreamRouteTask_Unsupported", func(t *testing.T) {
		taskStream := &interop.Task{
			ID:        "task-oc-streamroute-unsup",
			Framework: "OpenClaw",
			Intent:    "unsupported_intent",
		}

		_, err := hub.StreamRouteTask(ctx, taskStream)
		if err == nil {
			t.Error("Expected error for unsupported OpenClaw capability, got nil")
		} else if err.Error() != "OpenClaw does not support capability: unsupported_intent" {
			t.Errorf("Unexpected error message: %v", err)
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

func TestStreamTaskCancellations(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewOpenClawAdapter())
	hub.RegisterAdapter(interop.NewCrewAIAdapter())
	hub.RegisterAdapter(interop.NewAutoGenAdapter())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel context to hit early exits

	taskOC := &interop.Task{Framework: "OpenClaw", Intent: "adaptive_reasoning"}
	streamOC, _ := hub.StreamRouteTask(ctx, taskOC)
	for _ = range streamOC {} // Drain stream

	taskCAI := &interop.Task{Framework: "CrewAI", Intent: "task_delegation"}
	streamCAI, _ := hub.StreamRouteTask(ctx, taskCAI)
	for _ = range streamCAI {}

	taskAG := &interop.Task{Framework: "AutoGen", Intent: "subagent_exec"}
	streamAG, _ := hub.StreamRouteTask(ctx, taskAG)
	for _ = range streamAG {}
}

func TestStreamTaskPartialCancellations(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewOpenClawAdapter())
	hub.RegisterAdapter(interop.NewCrewAIAdapter())
	hub.RegisterAdapter(interop.NewAutoGenAdapter())

    // For openclaw which has 3 chunks
	ctx, cancel := context.WithCancel(context.Background())
	taskOC := &interop.Task{Framework: "OpenClaw", Intent: "adaptive_reasoning"}
	streamOC, _ := hub.StreamRouteTask(ctx, taskOC)
	<-streamOC // read first chunk
	cancel() // cancel before second chunk
	for _ = range streamOC {} // drain

	ctx2, cancel2 := context.WithCancel(context.Background())
	taskOC2 := &interop.Task{Framework: "OpenClaw", Intent: "adaptive_reasoning"}
	streamOC2, _ := hub.StreamRouteTask(ctx2, taskOC2)
	<-streamOC2 // read first chunk
	<-streamOC2 // read second chunk
	cancel2() // cancel before third chunk
	for _ = range streamOC2 {} // drain

    // For autogen which has 2 chunks
    ctx3, cancel3 := context.WithCancel(context.Background())
	taskAG := &interop.Task{Framework: "AutoGen", Intent: "subagent_exec"}
	streamAG, _ := hub.StreamRouteTask(ctx3, taskAG)
	<-streamAG // read first chunk
	cancel3() // cancel before second chunk
	for _ = range streamAG {} // drain

    // For crewai which has 2 chunks
    ctx4, cancel4 := context.WithCancel(context.Background())
	taskCAI := &interop.Task{Framework: "CrewAI", Intent: "task_delegation", Payload: map[string]string{}}
	streamCAI, _ := hub.StreamRouteTask(ctx4, taskCAI)
	<-streamCAI // read first chunk
	cancel4() // cancel before second chunk
	for _ = range streamCAI {} // drain
}

func TestStreamTaskPartialCancellations2(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewCrewAIAdapter())
	hub.RegisterAdapter(interop.NewAutoGenAdapter())

    // For autogen, read 0 chunks and cancel
    ctx3, cancel3 := context.WithCancel(context.Background())
    cancel3() // cancel immediately
	taskAG := &interop.Task{Framework: "AutoGen", Intent: "subagent_exec"}
	streamAG, _ := hub.StreamRouteTask(ctx3, taskAG)
	for _ = range streamAG {} // drain

    // For crewai, read 0 chunks and cancel
    ctx4, cancel4 := context.WithCancel(context.Background())
    cancel4() // cancel immediately
	taskCAI := &interop.Task{Framework: "CrewAI", Intent: "task_delegation"}
	streamCAI, _ := hub.StreamRouteTask(ctx4, taskCAI)
	for _ = range streamCAI {} // drain
}

func TestStreamTaskPartialCancellations3(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewOpenClawAdapter())
	hub.RegisterAdapter(interop.NewCrewAIAdapter())
	hub.RegisterAdapter(interop.NewAutoGenAdapter())

    ctx1, cancel1 := context.WithCancel(context.Background())
	taskOC1 := &interop.Task{Framework: "OpenClaw", Intent: "adaptive_reasoning"}
	streamOC1, _ := hub.StreamRouteTask(ctx1, taskOC1)
	cancel1() // cancel before first chunk
	for _ = range streamOC1 {} // drain
}

func TestStreamTaskPartialCancellations4(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewCrewAIAdapter())
	hub.RegisterAdapter(interop.NewAutoGenAdapter())

    ctx1, cancel1 := context.WithCancel(context.Background())
	taskCAI1 := &interop.Task{Framework: "CrewAI", Intent: "task_delegation"}
	streamCAI1, _ := hub.StreamRouteTask(ctx1, taskCAI1)
    <-streamCAI1
	cancel1() // cancel before last chunk
	for _ = range streamCAI1 {} // drain

    ctx2, cancel2 := context.WithCancel(context.Background())
	taskAG1 := &interop.Task{Framework: "AutoGen", Intent: "subagent_exec"}
	streamAG1, _ := hub.StreamRouteTask(ctx2, taskAG1)
    <-streamAG1
	cancel2() // cancel before last chunk
	for _ = range streamAG1 {} // drain
}

func TestCrewAIDefaultRoleStream(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewCrewAIAdapter())

    ctx1 := context.Background()
	taskCAI1 := &interop.Task{Framework: "CrewAI", Intent: "task_delegation"}
	streamCAI1, _ := hub.StreamRouteTask(ctx1, taskCAI1)
    for _ = range streamCAI1 {}
}

func TestStreamTaskCancellationsMore(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewOpenClawAdapter())
	hub.RegisterAdapter(interop.NewCrewAIAdapter())
	hub.RegisterAdapter(interop.NewAutoGenAdapter())

    // Try to trigger the "send final chunk" select with context done
    ctx1, cancel1 := context.WithCancel(context.Background())
	taskOC1 := &interop.Task{Framework: "OpenClaw", Intent: "adaptive_reasoning"}
	streamOC1, _ := hub.StreamRouteTask(ctx1, taskOC1)
    <-streamOC1 // read chunk 1
    <-streamOC1 // read chunk 2
	cancel1() // cancel before final chunk
	for _ = range streamOC1 {}

    ctx2, cancel2 := context.WithCancel(context.Background())
	taskCAI1 := &interop.Task{Framework: "CrewAI", Intent: "task_delegation"}
	streamCAI1, _ := hub.StreamRouteTask(ctx2, taskCAI1)
    <-streamCAI1
	cancel2()
	for _ = range streamCAI1 {}

    ctx3, cancel3 := context.WithCancel(context.Background())
	taskAG1 := &interop.Task{Framework: "AutoGen", Intent: "subagent_exec"}
	streamAG1, _ := hub.StreamRouteTask(ctx3, taskAG1)
    <-streamAG1
	cancel3()
	for _ = range streamAG1 {}
}

func TestStreamTaskCancellationsEvenMore(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewOpenClawAdapter())
	hub.RegisterAdapter(interop.NewCrewAIAdapter())
	hub.RegisterAdapter(interop.NewAutoGenAdapter())

    ctx1, cancel1 := context.WithCancel(context.Background())
	taskOC1 := &interop.Task{Framework: "OpenClaw", Intent: "adaptive_reasoning"}
	streamOC1, _ := hub.StreamRouteTask(ctx1, taskOC1)
    <-streamOC1 // read chunk 1
	cancel1() // cancel before chunk 2
	for _ = range streamOC1 {}
}

func TestStreamTaskCancellationsEvenMore2(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewOpenClawAdapter())

    ctx2, cancel2 := context.WithCancel(context.Background())
	taskOC2 := &interop.Task{Framework: "OpenClaw", Intent: "adaptive_reasoning"}
	streamOC2, _ := hub.StreamRouteTask(ctx2, taskOC2)
    <-streamOC2 // read chunk 1
    <-streamOC2 // read chunk 2
	cancel2() // cancel before chunk 3
	for _ = range streamOC2 {}
}

func TestStreamTaskCancellationsEvenMore3(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewOpenClawAdapter())

    ctx3, cancel3 := context.WithCancel(context.Background())
    cancel3() // cancel before chunk 1
	taskOC3 := &interop.Task{Framework: "OpenClaw", Intent: "adaptive_reasoning"}
	streamOC3, _ := hub.StreamRouteTask(ctx3, taskOC3)
	for _ = range streamOC3 {}
}

func TestStreamTaskCancellationsEvenMore4(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewCrewAIAdapter())
	hub.RegisterAdapter(interop.NewAutoGenAdapter())

    // CrewAI Default Role final chunk
    ctx1, cancel1 := context.WithCancel(context.Background())
	taskCAI1 := &interop.Task{Framework: "CrewAI", Intent: "task_delegation"}
	streamCAI1, _ := hub.StreamRouteTask(ctx1, taskCAI1)
    <-streamCAI1 // Read chunk 1
    cancel1()    // Cancel before chunk 2
	for _ = range streamCAI1 {} // Drain

    // AutoGen final chunk
    ctx2, cancel2 := context.WithCancel(context.Background())
	taskAG1 := &interop.Task{Framework: "AutoGen", Intent: "subagent_exec"}
	streamAG1, _ := hub.StreamRouteTask(ctx2, taskAG1)
    <-streamAG1 // Read chunk 1
    cancel2()    // Cancel before chunk 2
	for _ = range streamAG1 {} // Drain
}

func TestHandleTaskStreaming(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewOpenClawAdapter())
	hub.RegisterAdapter(interop.NewCrewAIAdapter())
	hub.RegisterAdapter(interop.NewAutoGenAdapter())

	ctx := context.Background()

	t.Run("OpenClaw_HandleTaskStreaming", func(t *testing.T) {
		task := &interop.Task{
			ID:        "oc-stream",
			Framework: "OpenClaw",
			Intent:    "adaptive_reasoning",
			Payload:   map[string]string{"stream": "true"},
		}
		res, err := hub.RouteTask(ctx, task)
		if err != nil {
			t.Fatalf("OpenClaw RouteTask failed: %v", err)
		}
		if res.Stream == nil {
			t.Fatalf("Expected stream to be initialized for OpenClaw")
		}
		var chunks []string
		for chunk := range res.Stream {
			chunks = append(chunks, chunk)
		}
		if len(chunks) != 2 {
			t.Errorf("Expected 2 stream chunks for OpenClaw, got %d", len(chunks))
		}
	})

	t.Run("CrewAI_HandleTaskStreaming", func(t *testing.T) {
		task := &interop.Task{
			ID:        "cai-stream",
			Framework: "CrewAI",
			Intent:    "task_delegation",
			Payload:   map[string]string{"stream": "true"},
		}
		res, err := hub.RouteTask(ctx, task)
		if err != nil {
			t.Fatalf("CrewAI RouteTask failed: %v", err)
		}
		if res.Stream == nil {
			t.Fatalf("Expected stream to be initialized for CrewAI")
		}
		var chunks []string
		for chunk := range res.Stream {
			chunks = append(chunks, chunk)
		}
		if len(chunks) != 2 {
			t.Errorf("Expected 2 stream chunks for CrewAI, got %d", len(chunks))
		}
	})

	t.Run("AutoGen_HandleTaskStreaming", func(t *testing.T) {
		task := &interop.Task{
			ID:        "ag-stream",
			Framework: "AutoGen",
			Intent:    "multi_agent_chat",
			Payload:   map[string]string{"stream": "true"},
		}
		res, err := hub.RouteTask(ctx, task)
		if err != nil {
			t.Fatalf("AutoGen RouteTask failed: %v", err)
		}
		if res.Stream == nil {
			t.Fatalf("Expected stream to be initialized for AutoGen")
		}
		var chunks []string
		for chunk := range res.Stream {
			chunks = append(chunks, chunk)
		}
		if len(chunks) != 2 {
			t.Errorf("Expected 2 stream chunks for AutoGen, got %d", len(chunks))
		}
	})
}

func TestStreamTaskUnsupportedIntents(t *testing.T) {
	hub := interop.NewAdapterHub()
	hub.RegisterAdapter(interop.NewOpenClawAdapter())
	hub.RegisterAdapter(interop.NewCrewAIAdapter())
	hub.RegisterAdapter(interop.NewAutoGenAdapter())

	ctx := context.Background()

	t.Run("OpenClaw_StreamTask_Unsupported", func(t *testing.T) {
		task := &interop.Task{Framework: "OpenClaw", Intent: "unsupported"}
		_, err := hub.StreamRouteTask(ctx, task)
		if err == nil {
			t.Error("Expected error for unsupported intent")
		} else if err.Error() != "OpenClaw does not support capability: unsupported" {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("CrewAI_StreamTask_Unsupported", func(t *testing.T) {
		task := &interop.Task{Framework: "CrewAI", Intent: "unsupported"}
		_, err := hub.StreamRouteTask(ctx, task)
		if err == nil {
			t.Error("Expected error for unsupported intent")
		} else if err.Error() != "CrewAI does not support capability: unsupported" {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("AutoGen_StreamTask_Unsupported", func(t *testing.T) {
		task := &interop.Task{Framework: "AutoGen", Intent: "unsupported"}
		_, err := hub.StreamRouteTask(ctx, task)
		if err == nil {
			t.Error("Expected error for unsupported intent")
		} else if err.Error() != "AutoGen does not support capability: unsupported" {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}
