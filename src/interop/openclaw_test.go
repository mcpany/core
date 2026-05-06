package interop_test

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/src/interop"
)

func TestOpenClawHandleTaskStreaming(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()

	task := &interop.Task{
		ID:        "task-oc-stream-handletask",
		Framework: "OpenClaw",
		Intent:    "adaptive_reasoning",
		Payload: map[string]string{
			"stream": "true",
		},
	}

	res, err := adapter.HandleTask(context.Background(), task)
	if err != nil {
		t.Fatalf("Failed to execute OpenClaw HandleTask streaming: %v", err)
	}

	if res.Stream == nil {
		t.Fatal("Expected non-nil stream in response")
	}

	var chunks []string
	for chunk := range res.Stream {
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 2 {
		t.Errorf("Expected 2 chunks, got %d", len(chunks))
	}
	if chunks[0] != "chunk 1" || chunks[1] != "chunk 2" {
		t.Errorf("Unexpected chunks: %v", chunks)
	}
}

func TestOpenClawStreamTask(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()

	task := &interop.Task{
		ID:        "task-oc-streamtask",
		Framework: "OpenClaw",
		Intent:    "adaptive_reasoning",
	}

	stream, err := adapter.StreamTask(context.Background(), task)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	var chunks []*interop.TaskResult
	for chunk := range stream {
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 3 {
		t.Errorf("Expected 3 chunks, got %d", len(chunks))
	}
}

func TestOpenClawStreamTaskDirectContextCancel1(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()

	task := &interop.Task{
		ID:        "task-oc-streamtask",
		Framework: "OpenClaw",
		Intent:    "adaptive_reasoning",
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

func TestOpenClawStreamTaskDirectContextCancel2(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()

	task := &interop.Task{
		ID:        "task-oc-streamtask",
		Framework: "OpenClaw",
		Intent:    "adaptive_reasoning",
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

func TestOpenClawStreamTaskDirectContextCancel3(t *testing.T) {
	adapter := interop.NewOpenClawAdapter()

	task := &interop.Task{
		ID:        "task-oc-streamtask",
		Framework: "OpenClaw",
		Intent:    "adaptive_reasoning",
	}

	ctx, cancel := context.WithCancel(context.Background())

	stream, err := adapter.StreamTask(ctx, task)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	<-stream // Receive first chunk
    <-stream // Receive second chunk
	cancel() // Cancel before third chunk

	time.Sleep(10 * time.Millisecond)

	var chunks []*interop.TaskResult
	for chunk := range stream {
		chunks = append(chunks, chunk)
	}

	if len(chunks) != 0 {
		t.Errorf("Expected 0 remaining chunks due to cancel, got %d", len(chunks))
	}
}
