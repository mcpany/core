package interop_test

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/src/interop"
)

func TestCrewAIStreamTaskDirect(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()

	task := &interop.Task{
		ID:        "task-cai-streamtask",
		Framework: "CrewAI",
		Intent:    "task_delegation",
	}

	stream, err := adapter.StreamTask(context.Background(), task)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var chunks []*interop.TaskResult
	for chunk := range stream {
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(chunks))
	}
}

func TestCrewAIStreamTaskDirectContextCancel1(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()

	task := &interop.Task{
		ID:        "task-cai-streamtask",
		Framework: "CrewAI",
		Intent:    "task_delegation",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel before doing anything

	stream, err := adapter.StreamTask(ctx, task)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	var chunks []*interop.TaskResult
	for chunk := range stream {
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 0 {
		t.Errorf("Expected 0 chunks due to early cancel, got %d", len(chunks))
	}
}

func TestCrewAIStreamTaskDirectContextCancel2(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()

	task := &interop.Task{
		ID:        "task-cai-streamtask",
		Framework: "CrewAI",
		Intent:    "task_delegation",
	}

	ctx, cancel := context.WithCancel(context.Background())

	stream, err := adapter.StreamTask(ctx, task)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	<-stream // Receive first chunk
	cancel() // Cancel before second chunk

	time.Sleep(10 * time.Millisecond)

	var chunks []*interop.TaskResult
	for chunk := range stream {
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 0 {
		t.Errorf("Expected 0 remaining chunks due to cancel, got %d", len(chunks))
	}
}


func TestCrewAIStreamTaskUnsupported(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()

	task := &interop.Task{
		ID:        "task-cai-streamtask-unsupported",
		Framework: "CrewAI",
		Intent:    "unsupported_intent",
	}

	_, err := adapter.StreamTask(context.Background(), task)
	if err == nil {
		t.Fatal("Expected error for unsupported intent")
	}
}

func TestCrewAIStreamTaskDirectContextCancel3(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()

	task := &interop.Task{
		ID:        "task-cai-streamtask",
		Framework: "CrewAI",
		Intent:    "task_delegation",
	}

	ctx, cancel := context.WithCancel(context.Background())

	stream, err := adapter.StreamTask(ctx, task)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	<-stream // Receive first chunk
	cancel() // Cancel before second chunk

	time.Sleep(10 * time.Millisecond)

    // Just exhaust
	for _ = range stream {
	}
}
