package interop_test

import (
	"context"
	"testing"
	"github.com/mcpany/core/src/interop"
)

func TestAutoGenAdapter_New(t *testing.T) {
	adapter := interop.NewAutoGenAdapter()
	if adapter.Name() != "AutoGen" {
		t.Errorf("Expected name 'AutoGen', got '%s'", adapter.Name())
	}
	if !adapter.SupportsCapability("multi_agent_chat") {
		t.Error("Expected support for multi_agent_chat")
	}
	if !adapter.SupportsCapability("subagent_exec") {
		t.Error("Expected support for subagent_exec")
	}
	if adapter.SupportsCapability("unknown_intent") {
		t.Error("Expected unknown_intent to be unsupported")
	}
}

func TestAutoGenAdapter_HandleTask(t *testing.T) {
	adapter := interop.NewAutoGenAdapter()
	ctx := context.Background()

	// Unsupported Intent
	task := &interop.Task{Intent: "unknown_intent"}
	_, err := adapter.HandleTask(ctx, task)
	if err == nil {
		t.Error("Expected error for unsupported intent, got nil")
	}

	// Supported Intent (Success)
	task = &interop.Task{Intent: "subagent_exec", Payload: map[string]string{}}
	res, err := adapter.HandleTask(ctx, task)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", res.Status)
	}
	if len(adapter.ChatHistory) != 1 {
		t.Errorf("Expected 1 item in ChatHistory, got %d", len(adapter.ChatHistory))
	}

	// Streaming HandleTask
	task = &interop.Task{Intent: "subagent_exec", Payload: map[string]string{"stream": "true"}}
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

func TestAutoGenAdapter_SyncMemoryShard(t *testing.T) {
	adapter := interop.NewAutoGenAdapter()
	ctx := context.Background()

	// Invalid signature
	err := adapter.SyncMemoryShard(ctx, &interop.MemoryShard{Signature: ""})
	if err == nil {
		t.Error("Expected error for missing signature, got nil")
	}

	// Valid signature
	err = adapter.SyncMemoryShard(ctx, &interop.MemoryShard{Signature: "valid_sig", ShardID: "test_shard"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(adapter.ChatHistory) != 1 {
		t.Errorf("Expected 1 item in ChatHistory after SyncMemoryShard, got %d", len(adapter.ChatHistory))
	}
}

func TestAutoGenAdapter_StreamTask(t *testing.T) {
	adapter := interop.NewAutoGenAdapter()
	ctx := context.Background()

	// Unsupported Intent
	task := &interop.Task{Intent: "unknown_intent"}
	_, err := adapter.StreamTask(ctx, task)
	if err == nil {
		t.Error("Expected error for unsupported intent, got nil")
	}

	// Supported Intent
	task = &interop.Task{ID: "task-1", Intent: "subagent_exec"}
	stream, err := adapter.StreamTask(ctx, task)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	var chunks []*interop.TaskResult
	for chunk := range stream {
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(chunks))
	}

	if chunks[0].Status != "streaming" {
		t.Errorf("Expected first chunk status 'streaming', got '%s'", chunks[0].Status)
	}
	if chunks[1].Status != "success" {
		t.Errorf("Expected second chunk status 'success', got '%s'", chunks[1].Status)
	}

	if len(adapter.ChatHistory) != 2 {
		t.Errorf("Expected 2 items in ChatHistory after StreamTask, got %d", len(adapter.ChatHistory))
	}
}

func TestAutoGenAdapter_StreamTaskCancel(t *testing.T) {
	adapter := interop.NewAutoGenAdapter()
	ctx, cancel := context.WithCancel(context.Background())

	task := &interop.Task{ID: "task-2", Intent: "subagent_exec"}
	stream, err := adapter.StreamTask(ctx, task)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	cancel() // Cancel before reading to hit select ctx.Done()
	var chunks []*interop.TaskResult
	for chunk := range stream {
		chunks = append(chunks, chunk)
	}
    // Might get 0 or 1 chunk depending on timing, just checking it doesn't block
}
