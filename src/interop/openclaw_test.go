package interop_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/src/interop"
)

func TestOpenClawAdapter_StreamTask(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()
	ctx := context.Background()

	task := &interop.Task{
		ID:        "task-oc-1",
		Framework: "OpenClaw",
		Intent:    "adaptive_reasoning",
	}

	stream, err := adapter.StreamTask(ctx, task)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	results := []*interop.TaskResult{}
	for res := range stream {
		results = append(results, res)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	if results[0].Status != "streaming" {
		t.Errorf("Expected first result to have status 'streaming', got: %s", results[0].Status)
	}

	if results[1].Status != "streaming" {
		t.Errorf("Expected second result to have status 'streaming', got: %s", results[1].Status)
	}

	if results[2].Status != "success" {
		t.Errorf("Expected third result to have status 'success', got: %s", results[2].Status)
	}
}

func TestOpenClawAdapter_StreamTask_Unsupported(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()
	ctx := context.Background()

	task := &interop.Task{
		ID:        "task-oc-unsup",
		Framework: "OpenClaw",
		Intent:    "invalid_intent",
	}

	_, err := adapter.StreamTask(ctx, task)
	if err == nil {
		t.Fatal("Expected error for unsupported capability, got nil")
	}
}
