package interop_test

import (
	"context"
	"testing"
	"github.com/mcpany/core/src/interop"
)

func TestOpenClawAdapter_New(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()
	if adapter.Name() != "OpenClaw" {
		t.Errorf("Expected name 'OpenClaw', got '%s'", adapter.Name())
	}
	if !adapter.SupportsCapability("adaptive_reasoning") {
		t.Error("Expected support for adaptive_reasoning")
	}
	if !adapter.SupportsCapability("context_sync") {
		t.Error("Expected support for context_sync")
	}
	if adapter.SupportsCapability("unknown_intent") {
		t.Error("Expected unknown_intent to be unsupported")
	}
	if adapter.CurrentEpoch != 1 {
		t.Errorf("Expected CurrentEpoch to be 1, got %d", adapter.CurrentEpoch)
	}
}

func TestOpenClawAdapter_HandleTask(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()
	ctx := context.Background()

	// Unsupported Intent
	task := &interop.Task{Intent: "unknown_intent"}
	_, err := adapter.HandleTask(ctx, task)
	if err == nil {
		t.Error("Expected error for unsupported intent, got nil")
	}

	// Supported Intent (Success)
	task = &interop.Task{Intent: "adaptive_reasoning", Payload: map[string]string{}}
	res, err := adapter.HandleTask(ctx, task)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", res.Status)
	}
	if adapter.CurrentEpoch != 2 {
		t.Errorf("Expected CurrentEpoch 2, got %d", adapter.CurrentEpoch)
	}

	// Streaming HandleTask
	task = &interop.Task{Intent: "adaptive_reasoning", Payload: map[string]string{"stream": "true"}}
	res, err = adapter.HandleTask(ctx, task)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res.Stream == nil {
		t.Error("Expected stream channel, got nil")
	}
	// Drain stream
	<-res.Stream
	<-res.Stream
}

func TestOpenClawAdapter_SyncMemoryShard(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()
	ctx := context.Background()

	// Invalid signature
	err := adapter.SyncMemoryShard(ctx, &interop.MemoryShard{Signature: ""})
	if err == nil {
		t.Error("Expected error for missing signature, got nil")
	}

	// Valid signature
	err = adapter.SyncMemoryShard(ctx, &interop.MemoryShard{Signature: "valid_sig", ShardID: "test_shard", TextContent: "test", MultimodalPayload: []byte{1,2}})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestOpenClawAdapter_StreamTask(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()
	ctx := context.Background()

	// Unsupported Intent
	task := &interop.Task{Intent: "unknown_intent"}
	_, err := adapter.StreamTask(ctx, task)
	if err == nil {
		t.Error("Expected error for unsupported intent, got nil")
	}

	// Supported Intent
	task = &interop.Task{ID: "task-1", Intent: "adaptive_reasoning"}
	stream, err := adapter.StreamTask(ctx, task)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	var chunks []*interop.TaskResult
	for chunk := range stream {
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 3 {
		t.Errorf("Expected 3 chunks, got %d", len(chunks))
	}
	if chunks[0].Status != "streaming" {
		t.Errorf("Expected first chunk status 'streaming', got '%s'", chunks[0].Status)
	}
	if chunks[1].Status != "streaming" {
		t.Errorf("Expected second chunk status 'streaming', got '%s'", chunks[1].Status)
	}
	if chunks[2].Status != "success" {
		t.Errorf("Expected third chunk status 'success', got '%s'", chunks[2].Status)
	}
}

func TestOpenClawAdapter_StreamTaskCancel(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()

	task := &interop.Task{ID: "task-2", Intent: "adaptive_reasoning"}
	ctx1, cancel1 := context.WithCancel(context.Background())
	stream1, err := adapter.StreamTask(ctx1, task)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	cancel1() // cancel before first chunk
	for _ = range stream1 {}

	ctx2, cancel2 := context.WithCancel(context.Background())
	stream2, _ := adapter.StreamTask(ctx2, task)
	<-stream2 // read first chunk
	cancel2() // cancel before second chunk
	for _ = range stream2 {}

	ctx3, cancel3 := context.WithCancel(context.Background())
	stream3, _ := adapter.StreamTask(ctx3, task)
	<-stream3 // read first chunk
	<-stream3 // read second chunk
	cancel3() // cancel before third chunk
	for _ = range stream3 {}
}
