package interop_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/src/interop"
)

func TestCrewAIAdapter_HandleTask_Streaming(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()
	ctx := context.Background()

	task := &interop.Task{
		ID:        "task-cai-1",
		Framework: "CrewAI",
		Intent:    "task_delegation",
		Payload:   map[string]string{"stream": "true"},
	}

	res, err := adapter.HandleTask(ctx, task)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if res.Stream == nil {
		t.Fatal("Expected stream to be initialized")
	}

	chunks := []string{}
	for chunk := range res.Stream {
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(chunks))
	}
}

func TestCrewAIAdapter_StreamTask(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()
	ctx := context.Background()

	task := &interop.Task{
		ID:        "task-cai-2",
		Framework: "CrewAI",
		Intent:    "task_delegation",
	}

	stream, err := adapter.StreamTask(ctx, task)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	results := []*interop.TaskResult{}
	for res := range stream {
		results = append(results, res)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if results[0].Status != "streaming" {
		t.Errorf("Expected first result to have status 'streaming', got: %s", results[0].Status)
	}

	if results[1].Status != "success" {
		t.Errorf("Expected second result to have status 'success', got: %s", results[1].Status)
	}
}

func TestCrewAIAdapter_StreamTask_Unsupported(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()
	ctx := context.Background()

	task := &interop.Task{
		ID:        "task-cai-unsup",
		Framework: "CrewAI",
		Intent:    "invalid_intent",
	}

	_, err := adapter.StreamTask(ctx, task)
	if err == nil {
		t.Fatal("Expected error for unsupported capability, got nil")
	}
}
