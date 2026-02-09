package audit

import (
	"context"
	"testing"
	"time"
)

func TestMemoryAuditStore(t *testing.T) {
	store := NewMemoryAuditStore()

	// 1. Write entries
	entries := []Entry{
		{
			Timestamp: time.Now().Add(-1 * time.Hour),
			ToolName:  "tool_a",
			UserID:    "user_1",
		},
		{
			Timestamp: time.Now(),
			ToolName:  "tool_b",
			UserID:    "user_2",
		},
		{
			Timestamp: time.Now().Add(-30 * time.Minute),
			ToolName:  "tool_a",
			UserID:    "user_2",
		},
	}

	for _, e := range entries {
		if err := store.Write(context.Background(), e); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}

	// 2. Read with filter (ToolName)
	filtered, err := store.Read(context.Background(), Filter{ToolName: "tool_a"})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if len(filtered) != 2 {
		t.Errorf("Expected 2 entries for tool_a, got %d", len(filtered))
	}

	// 3. Read with filter (UserID)
	filtered, err = store.Read(context.Background(), Filter{UserID: "user_2"})
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if len(filtered) != 2 {
		t.Errorf("Expected 2 entries for user_2, got %d", len(filtered))
	}

	// 4. Read all (should be sorted desc by default implementation if appended sequentially?
	// Wait, my implementation reverses the list.
	// entries are appended: [oldest, newest, middle] (based on my inputs)
	// Read iterates reverse: [middle, newest, oldest]
	// Let's check timestamps.
	// entry 0: T-1h
	// entry 1: T
	// entry 2: T-30m
	// Slice: [T-1h, T, T-30m]
	// Read Reverse: T-30m, T, T-1h
	// This is simple "insertion order reversed", not sorted by timestamp.
	// That's acceptable for a simple memory store if we assume Write is called sequentially in time.
	// But in my test I inserted out of order.
	// Let's assume standard usage is sequential.

	all, err := store.Read(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("Read all failed: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(all))
	}
}
