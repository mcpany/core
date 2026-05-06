package interop_test

import (
	"context"
	"testing"
	"github.com/mcpany/core/src/interop"
)

func TestCrewAIAdapter_New(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()
	if adapter.Name() != "CrewAI" {
		t.Errorf("Expected name 'CrewAI', got '%s'", adapter.Name())
	}
	if !adapter.SupportsCapability("task_delegation") {
		t.Error("Expected support for task_delegation")
	}
	if !adapter.SupportsCapability("role_discovery") {
		t.Error("Expected support for role_discovery")
	}
	if adapter.SupportsCapability("unknown_intent") {
		t.Error("Expected unknown_intent to be unsupported")
	}
}

func TestCrewAIAdapter_HandleTask(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()
	ctx := context.Background()

	// Unsupported Intent
	task := &interop.Task{Intent: "unknown_intent"}
	_, err := adapter.HandleTask(ctx, task)
	if err == nil {
		t.Error("Expected error for unsupported intent, got nil")
	}

	// Supported Intent (Success) with default role
	task = &interop.Task{Intent: "task_delegation", Payload: map[string]string{}}
	res, err := adapter.HandleTask(ctx, task)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", res.Status)
	}
	if res.Telemetry["delegated_role"] != "generalist" {
		t.Errorf("Expected delegated_role 'generalist', got '%s'", res.Telemetry["delegated_role"])
	}
	if _, ok := adapter.RoleRegistry["generalist"]; !ok {
		t.Error("Expected role 'generalist' in RoleRegistry")
	}

	// Supported Intent (Success) with specific role
	task = &interop.Task{Intent: "task_delegation", Payload: map[string]string{"role": "researcher"}}
	res, err = adapter.HandleTask(ctx, task)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if res.Telemetry["delegated_role"] != "researcher" {
		t.Errorf("Expected delegated_role 'researcher', got '%s'", res.Telemetry["delegated_role"])
	}
	if _, ok := adapter.RoleRegistry["researcher"]; !ok {
		t.Error("Expected role 'researcher' in RoleRegistry")
	}

	// Streaming HandleTask
	task = &interop.Task{Intent: "task_delegation", Payload: map[string]string{"stream": "true"}}
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

func TestCrewAIAdapter_SyncMemoryShard(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()
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
}

func TestCrewAIAdapter_StreamTask(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()
	ctx := context.Background()

	// Unsupported Intent
	task := &interop.Task{Intent: "unknown_intent"}
	_, err := adapter.StreamTask(ctx, task)
	if err == nil {
		t.Error("Expected error for unsupported intent, got nil")
	}

	// Supported Intent with default role
	task = &interop.Task{ID: "task-1", Intent: "task_delegation", Payload: map[string]string{}}
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
	if chunks[0].Telemetry["delegated_role"] != "generalist" {
		t.Errorf("Expected delegated_role 'generalist', got '%s'", chunks[0].Telemetry["delegated_role"])
	}

	// Supported Intent with specific role
	task = &interop.Task{ID: "task-2", Intent: "task_delegation", Payload: map[string]string{"role": "writer"}}
	stream, err = adapter.StreamTask(ctx, task)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	chunks = nil
	for chunk := range stream {
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(chunks))
	}
	if chunks[0].Telemetry["delegated_role"] != "writer" {
		t.Errorf("Expected delegated_role 'writer', got '%s'", chunks[0].Telemetry["delegated_role"])
	}
}

func TestCrewAIAdapter_StreamTaskCancel(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()
	ctx, cancel := context.WithCancel(context.Background())

	task := &interop.Task{ID: "task-3", Intent: "task_delegation"}
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
