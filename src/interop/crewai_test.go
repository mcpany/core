package interop_test

import (
	"context"
	"testing"
	"github.com/mcpany/core/src/interop"
)

func TestCrewAIAdapter_HandleTask_Streaming(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()
	task := &interop.Task{
		ID:        "test-cai-stream",
		Framework: "CrewAI",
		Intent:    "task_delegation",
		Payload:   map[string]string{"stream": "true"},
	}

	res, err := adapter.HandleTask(context.Background(), task)
	if err != nil {
		t.Fatalf("HandleTask failed: %v", err)
	}

	if res.Stream == nil {
		t.Fatal("Expected Stream channel to be populated")
	}

	chunks := 0
	for range res.Stream {
		chunks++
	}

	if chunks != 2 {
		t.Errorf("Expected 2 chunks, got %d", chunks)
	}
}

func TestCrewAIAdapter_StreamTask(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()
	task := &interop.Task{
		ID:        "test-cai-streamtask",
		Framework: "CrewAI",
		Intent:    "task_delegation",
		Payload:   map[string]string{"role": "analyst"},
	}

	stream, err := adapter.StreamTask(context.Background(), task)
	if err != nil {
		t.Fatalf("StreamTask failed: %v", err)
	}

	results := []*interop.TaskResult{}
	for res := range stream {
		results = append(results, res)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 streaming results, got %d", len(results))
	}

	if results[0].Status != "streaming" {
		t.Errorf("Expected first result status 'streaming', got %s", results[0].Status)
	}

	if results[1].Status != "success" {
		t.Errorf("Expected second result status 'success', got %s", results[1].Status)
	}

	if results[0].Telemetry["delegated_role"] != "analyst" {
		t.Errorf("Expected delegated_role 'analyst', got %s", results[0].Telemetry["delegated_role"])
	}
}

func TestCrewAIAdapter_StreamTask_Unsupported(t *testing.T) {
	adapter := interop.NewCrewAIAdapter()
	task := &interop.Task{
		ID:        "test-cai-unsupported",
		Framework: "CrewAI",
		Intent:    "unsupported_intent",
	}

	_, err := adapter.StreamTask(context.Background(), task)
	if err == nil {
		t.Fatal("Expected error for unsupported intent, got nil")
	}
}
