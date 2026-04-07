package interop_test

import (
	"context"
	"testing"
	"github.com/mcpany/core/src/interop"
)

func TestCrewAIAdapter_Name(t *testing.T) {
	a := interop.NewCrewAIAdapter()
	if a.Name() != "CrewAI" {
		t.Errorf("expected CrewAI, got %s", a.Name())
	}
}

func TestCrewAIAdapter_HandleTask(t *testing.T) {
	a := interop.NewCrewAIAdapter()
	ctx := context.Background()

	t.Run("UnsupportedCapability", func(t *testing.T) {
		_, err := a.HandleTask(ctx, &interop.Task{Intent: "unsupported"})
		if err == nil {
			t.Fatal("expected error for unsupported capability")
		}
	})

	t.Run("SupportedCapability", func(t *testing.T) {
		res, err := a.HandleTask(ctx, &interop.Task{
			Intent: "task_delegation",
			ID: "789",
			Payload: map[string]string{"role": "researcher"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.TaskID != "789" {
			t.Errorf("expected task ID 789, got %s", res.TaskID)
		}
		if res.Status != "success" {
			t.Errorf("expected success, got %s", res.Status)
		}
	})

	t.Run("Streaming", func(t *testing.T) {
		res, err := a.HandleTask(ctx, &interop.Task{
			Intent: "task_delegation",
			Payload: map[string]string{"stream": "true"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Stream == nil {
			t.Fatal("expected stream channel to be non-nil")
		}
		for range res.Stream {
			// consume stream
		}
	})
}

func TestCrewAIAdapter_SyncMemoryShard(t *testing.T) {
	a := interop.NewCrewAIAdapter()
	ctx := context.Background()

	t.Run("MissingSignature", func(t *testing.T) {
		err := a.SyncMemoryShard(ctx, &interop.MemoryShard{})
		if err == nil {
			t.Fatal("expected error for missing signature")
		}
	})

	t.Run("ValidShard", func(t *testing.T) {
		err := a.SyncMemoryShard(ctx, &interop.MemoryShard{Signature: "valid"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCrewAIAdapter_StreamTask(t *testing.T) {
	a := interop.NewCrewAIAdapter()
	ctx := context.Background()

	t.Run("UnsupportedCapability", func(t *testing.T) {
		_, err := a.StreamTask(ctx, &interop.Task{Intent: "unsupported"})
		if err == nil {
			t.Fatal("expected error for unsupported capability")
		}
	})

	t.Run("Success", func(t *testing.T) {
		stream, err := a.StreamTask(ctx, &interop.Task{
			Intent: "task_delegation",
			ID: "012",
			Payload: map[string]string{"role": "writer"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		count := 0
		for res := range stream {
			if res.TaskID != "012" {
				t.Errorf("expected task ID 012, got %s", res.TaskID)
			}
			count++
		}
		if count != 2 {
			t.Errorf("expected 2 chunks, got %d", count)
		}
	})
}
