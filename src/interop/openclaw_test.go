package interop_test

import (
	"context"
	"testing"
	"github.com/mcpany/core/src/interop"
)

func TestOpenClawAdapter_Name(t *testing.T) {
	a := interop.NewOpenClawAdapter()
	if a.Name() != "OpenClaw" {
		t.Errorf("expected OpenClaw, got %s", a.Name())
	}
}

func TestOpenClawAdapter_HandleTask(t *testing.T) {
	a := interop.NewOpenClawAdapter()
	ctx := context.Background()

	t.Run("UnsupportedCapability", func(t *testing.T) {
		_, err := a.HandleTask(ctx, &interop.Task{Intent: "unsupported"})
		if err == nil {
			t.Fatal("expected error for unsupported capability")
		}
	})

	t.Run("SupportedCapability", func(t *testing.T) {
		res, err := a.HandleTask(ctx, &interop.Task{
			Intent: "adaptive_reasoning",
			ID: "abc",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.TaskID != "abc" {
			t.Errorf("expected task ID abc, got %s", res.TaskID)
		}
		if res.Status != "success" {
			t.Errorf("expected success, got %s", res.Status)
		}
	})

	t.Run("Streaming", func(t *testing.T) {
		res, err := a.HandleTask(ctx, &interop.Task{
			Intent: "adaptive_reasoning",
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

func TestOpenClawAdapter_SyncMemoryShard(t *testing.T) {
	a := interop.NewOpenClawAdapter()
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

func TestOpenClawAdapter_StreamTask(t *testing.T) {
	a := interop.NewOpenClawAdapter()
	ctx := context.Background()

	t.Run("UnsupportedCapability", func(t *testing.T) {
		_, err := a.StreamTask(ctx, &interop.Task{Intent: "unsupported"})
		if err == nil {
			t.Fatal("expected error for unsupported capability")
		}
	})

	t.Run("Success", func(t *testing.T) {
		stream, err := a.StreamTask(ctx, &interop.Task{
			Intent: "adaptive_reasoning",
			ID: "def",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		count := 0
		for res := range stream {
			if res.TaskID != "def" {
				t.Errorf("expected task ID def, got %s", res.TaskID)
			}
			count++
		}
		if count != 3 {
			t.Errorf("expected 3 chunks, got %d", count)
		}
	})
}
