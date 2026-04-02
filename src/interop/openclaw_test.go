package interop

import (
	"context"
	"testing"
)

func TestOpenClawAdapter_Name(t *testing.T) {
	adapter := NewOpenClawAdapter()
	if name := adapter.Name(); name != "OpenClaw" {
		t.Errorf("expected Name() to be 'OpenClaw', got '%s'", name)
	}
}

func TestOpenClawAdapter_SupportsCapability(t *testing.T) {
	adapter := NewOpenClawAdapter()
	tests := []struct {
		name       string
		capability string
		expected   bool
	}{
		{"supported capability adaptive_reasoning", "adaptive_reasoning", true},
		{"supported capability context_sync", "context_sync", true},
		{"unsupported capability unknown", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := adapter.SupportsCapability(tt.capability); got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestOpenClawAdapter_HandleTask(t *testing.T) {
	tests := []struct {
		name          string
		task          *Task
		expectError   bool
		expectedError string
		checkStream   bool
	}{
		{
			name: "successful task execution",
			task: &Task{
				ID:        "task-1",
				Intent:    "adaptive_reasoning",
				Payload:   map[string]string{},
			},
			expectError: false,
		},
		{
			name: "unsupported capability",
			task: &Task{
				ID:        "task-2",
				Intent:    "unknown_capability",
				Payload:   map[string]string{},
			},
			expectError:   true,
			expectedError: "OpenClaw does not support capability: unknown_capability",
		},
		{
			name: "task with stream payload",
			task: &Task{
				ID:        "task-3",
				Intent:    "context_sync",
				Payload:   map[string]string{"stream": "true"},
			},
			expectError: false,
			checkStream: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewOpenClawAdapter()
			initialEpoch := adapter.CurrentEpoch
			ctx := context.Background()

			result, err := adapter.HandleTask(ctx, tt.task)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if err.Error() != tt.expectedError {
					t.Errorf("expected error message '%s', got '%s'", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatalf("expected result, got nil")
			}

			if result.TaskID != tt.task.ID {
				t.Errorf("expected TaskID '%s', got '%s'", tt.task.ID, result.TaskID)
			}

			if result.Status != "success" {
				t.Errorf("expected Status 'success', got '%s'", result.Status)
			}

			if adapter.CurrentEpoch != initialEpoch+1 {
				t.Errorf("expected epoch to increment to %d, got %d", initialEpoch+1, adapter.CurrentEpoch)
			}

			if tt.checkStream {
				if result.Stream == nil {
					t.Errorf("expected Stream to be initialized, got nil")
				} else {
					// Consume stream
					chunks := []string{}
					for chunk := range result.Stream {
						chunks = append(chunks, chunk)
					}
					if len(chunks) != 2 {
						t.Errorf("expected 2 chunks, got %d", len(chunks))
					}
				}
			} else {
				if result.Stream != nil {
					t.Errorf("expected Stream to be nil, got non-nil")
				}
			}
		})
	}
}

func TestOpenClawAdapter_SyncMemoryShard(t *testing.T) {
	tests := []struct {
		name          string
		shard         *MemoryShard
		expectError   bool
		expectedError string
	}{
		{
			name: "valid shard",
			shard: &MemoryShard{
				ShardID:     "shard-1",
				TextContent: "hello",
				Signature:   "valid-signature",
			},
			expectError: false,
		},
		{
			name: "missing signature",
			shard: &MemoryShard{
				ShardID:     "shard-2",
				TextContent: "hello",
				Signature:   "",
			},
			expectError:   true,
			expectedError: "invalid memory shard: missing signature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewOpenClawAdapter()
			ctx := context.Background()

			err := adapter.SyncMemoryShard(ctx, tt.shard)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if err.Error() != tt.expectedError {
					t.Errorf("expected error message '%s', got '%s'", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestOpenClawAdapter_StreamTask(t *testing.T) {
	t.Run("successful stream execution", func(t *testing.T) {
		adapter := NewOpenClawAdapter()
		initialEpoch := adapter.CurrentEpoch
		ctx := context.Background()

		task := &Task{
			ID:      "stream-task-1",
			Intent:  "adaptive_reasoning",
			Payload: map[string]string{},
		}

		stream, err := adapter.StreamTask(ctx, task)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var results []*TaskResult
		for res := range stream {
			results = append(results, res)
		}

		if len(results) != 3 {
			t.Fatalf("expected 3 chunks, got %d", len(results))
		}

		if adapter.CurrentEpoch != initialEpoch+1 {
			t.Errorf("expected epoch to increment to %d, got %d", initialEpoch+1, adapter.CurrentEpoch)
		}

		if results[0].Status != "streaming" || results[0].Telemetry["chunk_index"] != "0" {
			t.Errorf("unexpected first chunk: %+v", results[0])
		}
		if results[1].Status != "streaming" || results[1].Telemetry["chunk_index"] != "1" {
			t.Errorf("unexpected second chunk: %+v", results[1])
		}
		if results[2].Status != "success" || results[2].Telemetry["chunk_index"] != "2" {
			t.Errorf("unexpected final chunk: %+v", results[2])
		}
	})

	t.Run("unsupported capability", func(t *testing.T) {
		adapter := NewOpenClawAdapter()
		ctx := context.Background()

		task := &Task{
			ID:      "stream-task-2",
			Intent:  "unknown",
			Payload: map[string]string{},
		}

		stream, err := adapter.StreamTask(ctx, task)
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if err.Error() != "OpenClaw does not support capability: unknown" {
			t.Errorf("unexpected error message: %v", err)
		}

		if stream != nil {
			t.Errorf("expected nil stream, got non-nil")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		adapter := NewOpenClawAdapter()
		ctx, cancel := context.WithCancel(context.Background())

		task := &Task{
			ID:      "stream-task-3",
			Intent:  "adaptive_reasoning",
			Payload: map[string]string{},
		}

		stream, err := adapter.StreamTask(ctx, task)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Cancel the context before fully reading
		cancel()

		// Read remaining chunks to avoid goroutine leak
		count := 0
		for range stream {
			count++
		}

		if count == 3 {
			t.Errorf("expected stream to be cancelled early, but received all 3 chunks")
		}
	})
}
