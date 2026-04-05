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
		ID:     "task-1",
		Intent: "task_delegation",
		Payload: map[string]string{
			"stream": "true",
			"role":   "data_scientist",
		},
	}

	res, err := adapter.HandleTask(ctx, task)
	if err != nil {
		t.Fatalf("HandleTask failed: %v", err)
	}

	if res.Stream == nil {
		t.Fatal("Expected stream channel to be non-nil")
	}

	var chunks []string
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

	t.Run("SupportedIntent", func(t *testing.T) {
		task := &interop.Task{
			ID:     "task-2",
			Intent: "task_delegation",
			Payload: map[string]string{
				"role": "data_analyst",
			},
		}

		stream, err := adapter.StreamTask(ctx, task)
		if err != nil {
			t.Fatalf("StreamTask failed: %v", err)
		}

		var results []*interop.TaskResult
		for res := range stream {
			results = append(results, res)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}

		if results[0].Status != "streaming" {
			t.Errorf("Expected first chunk status 'streaming', got '%s'", results[0].Status)
		}

		if results[1].Status != "success" {
			t.Errorf("Expected second chunk status 'success', got '%s'", results[1].Status)
		}
	})

	t.Run("UnsupportedIntent", func(t *testing.T) {
		task := &interop.Task{
			ID:     "task-3",
			Intent: "unsupported",
		}

		_, err := adapter.StreamTask(ctx, task)
		if err == nil {
			t.Fatal("Expected error for unsupported intent, got nil")
		}
	})
}
