package interop_test

import (
	"context"
	"testing"
	"github.com/mcpany/core/src/interop"
)

func TestOpenClawAdapter_StreamTask(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()
	task := &interop.Task{
		ID:        "test-oc-streamtask",
		Framework: "OpenClaw",
		Intent:    "adaptive_reasoning",
	}

	stream, err := adapter.StreamTask(context.Background(), task)
	if err != nil {
		t.Fatalf("StreamTask failed: %v", err)
	}

	results := []*interop.TaskResult{}
	for res := range stream {
		results = append(results, res)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 streaming results, got %d", len(results))
	}

	if results[0].Status != "streaming" {
		t.Errorf("Expected first result status 'streaming', got %s", results[0].Status)
	}

	if results[1].Status != "streaming" {
		t.Errorf("Expected second result status 'streaming', got %s", results[1].Status)
	}

	if results[2].Status != "success" {
		t.Errorf("Expected third result status 'success', got %s", results[2].Status)
	}
}

func TestOpenClawAdapter_StreamTask_Unsupported(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()
	task := &interop.Task{
		ID:        "test-oc-unsupported",
		Framework: "OpenClaw",
		Intent:    "unsupported_intent",
	}

	_, err := adapter.StreamTask(context.Background(), task)
	if err == nil {
		t.Fatal("Expected error for unsupported intent, got nil")
	}
}
