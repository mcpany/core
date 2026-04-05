package interop_test

import (
	"context"
	"testing"

	"github.com/mcpany/core/src/interop"
)

func TestOpenClawAdapter_StreamTask(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()
	ctx := context.Background()

	t.Run("SupportedIntent", func(t *testing.T) {
		task := &interop.Task{
			ID:     "task-1",
			Intent: "adaptive_reasoning",
		}

		stream, err := adapter.StreamTask(ctx, task)
		if err != nil {
			t.Fatalf("StreamTask failed: %v", err)
		}

		var results []*interop.TaskResult
		for res := range stream {
			results = append(results, res)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		if results[0].Status != "streaming" {
			t.Errorf("Expected first chunk status 'streaming', got '%s'", results[0].Status)
		}

		if results[1].Status != "streaming" {
			t.Errorf("Expected second chunk status 'streaming', got '%s'", results[1].Status)
		}

		if results[2].Status != "success" {
			t.Errorf("Expected third chunk status 'success', got '%s'", results[2].Status)
		}
	})

	t.Run("UnsupportedIntent", func(t *testing.T) {
		task := &interop.Task{
			ID:     "task-2",
			Intent: "unsupported",
		}

		_, err := adapter.StreamTask(ctx, task)
		if err == nil {
			t.Fatal("Expected error for unsupported intent, got nil")
		}
	})
}
