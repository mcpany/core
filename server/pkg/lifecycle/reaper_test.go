package lifecycle_test

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/lifecycle"
)

func TestSubagentReaper_RegisterBranch(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-001"
	ttl := 5 * time.Second

	lease := reaper.RegisterBranch(intentID, ttl)

	if lease == nil {
		t.Fatal("Expected lease to be created")
	}

	status, err := reaper.GetLeaseStatus(intentID)
	if err != nil {
		t.Fatalf("Failed to get lease status: %v", err)
	}

	if status != lifecycle.StatusActive {
		t.Errorf("Expected lease status to be ACTIVE, got %s", status)
	}
}

func TestSubagentReaper_RegisterSubagent(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-001a"
	ttl := 5 * time.Second
	sessionID := "session-123"

	// Test Lease Not Found
	err := reaper.RegisterSubagent("nonexistent", sessionID)
	if err == nil {
		t.Error("Expected error when registering subagent for nonexistent intent, got nil")
	}

	reaper.RegisterBranch(intentID, ttl)

	// Test Happy Path
	err = reaper.RegisterSubagent(intentID, sessionID)
	if err != nil {
		t.Fatalf("Failed to register subagent: %v", err)
	}

	// Test Lease Not Active (by expiring it manually or simulating)
	// We will manually register and then prune it
	reaper.PruneIntent(intentID)
	err = reaper.RegisterSubagent(intentID, sessionID)
	if err == nil {
		t.Error("Expected error when registering subagent for non-active lease, got nil")
	}
}

func TestSubagentReaper_RecordHeartbeat(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-002"
	ttl := 2 * time.Second

	// Test Lease Not Found
	err := reaper.RecordHeartbeat("nonexistent", "sig", 1*time.Second)
	if err == nil {
		t.Error("Expected error for non-existent lease, got nil")
	}

	reaper.RegisterBranch(intentID, ttl)

	err = reaper.RecordHeartbeat(intentID, "valid_sig", 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to record heartbeat: %v", err)
	}

	// Test Lease Not Active
	reaper.PruneIntent(intentID)
	err = reaper.RecordHeartbeat(intentID, "valid_sig", 5*time.Second)
	if err == nil {
		t.Error("Expected error when recording heartbeat for non-active lease, got nil")
	}
}

func TestSubagentReaper_PruneIntent(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-003"
	ttl := 5 * time.Second

	// Test Lease Not Found
	err := reaper.PruneIntent("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent lease, got nil")
	}

	reaper.RegisterBranch(intentID, ttl)

	err = reaper.PruneIntent(intentID)
	if err != nil {
		t.Fatalf("Failed to prune intent: %v", err)
	}

	status, err := reaper.GetLeaseStatus(intentID)
	if err != nil {
		t.Fatalf("Failed to get lease status: %v", err)
	}

	if status != lifecycle.StatusPruned {
		t.Errorf("Expected lease status to be PRUNED, got %s", status)
	}
}

func TestSubagentReaper_GetLeaseStatus(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	_, err := reaper.GetLeaseStatus("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent lease, got nil")
	}
}

func TestSubagentReaper_SweepExpired(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-004"

	// Create a lease that expires immediately
	reaper.RegisterBranch(intentID, 1*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the daemon with a fast interval
	reaper.Start(ctx, 10*time.Millisecond)

	// Wait for the sweep to occur
	time.Sleep(50 * time.Millisecond)
	reaper.Stop()

	status, err := reaper.GetLeaseStatus(intentID)
	if err != nil {
		t.Fatalf("Failed to get lease status: %v", err)
	}

	if status != lifecycle.StatusExpired {
		t.Errorf("Expected lease status to be EXPIRED, got %s", status)
	}
}

func TestSubagentReaper_StartContextCancellation(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	ctx, cancel := context.WithCancel(context.Background())

	// Start it, then immediately cancel the context
	reaper.Start(ctx, 10*time.Millisecond)
	cancel()

	// Give the goroutine time to exit cleanly
	time.Sleep(50 * time.Millisecond)
}

func TestSubagentReaper_StartQuitChannel(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	ctx := context.Background()

	// Start it, then stop it using the quit channel logic
	reaper.Start(ctx, 10*time.Millisecond)
	reaper.Stop()

	// Give the goroutine time to exit cleanly
	time.Sleep(50 * time.Millisecond)
}
