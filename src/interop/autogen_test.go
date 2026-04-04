package interop_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/src/interop"
)

func TestAutoGenAdapter_HandleTask_Streaming(t *testing.T) {
	adapter := interop.NewAutoGenAdapter()
	ctx := context.Background()

	task := &interop.Task{
		ID:        "task-ag-1",
		Framework: "AutoGen",
		Intent:    "subagent_exec",
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

func TestAutoGenAdapter_StreamTask(t *testing.T) {
	adapter := interop.NewAutoGenAdapter()
	ctx := context.Background()

	task := &interop.Task{
		ID:        "task-ag-2",
		Framework: "AutoGen",
		Intent:    "subagent_exec",
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

func TestAutoGenAdapter_StreamTask_Unsupported(t *testing.T) {
	adapter := interop.NewAutoGenAdapter()
	ctx := context.Background()

	task := &interop.Task{
		ID:        "task-ag-unsup",
		Framework: "AutoGen",
		Intent:    "invalid_intent",
	}

	_, err := adapter.StreamTask(ctx, task)
	if err == nil {
		t.Fatal("Expected error for unsupported capability, got nil")
	}
}
