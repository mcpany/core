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

func TestSubagentReaper_RecordHeartbeat(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-002"
	ttl := 2 * time.Second

	reaper.RegisterBranch(intentID, ttl)

	err := reaper.RecordHeartbeat(intentID, "valid_sig", 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to record heartbeat: %v", err)
	}
}

func TestSubagentReaper_RegisterSubagent(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-register"
	ttl := 5 * time.Second

	// Case 1: Lease not found
	err := reaper.RegisterSubagent(intentID, "session-1")
	if err == nil {
		t.Fatal("Expected error for non-existent lease, got nil")
	}

	// Create lease
	reaper.RegisterBranch(intentID, ttl)

	// Case 2: Success
	err = reaper.RegisterSubagent(intentID, "session-1")
	if err != nil {
		t.Fatalf("Failed to register subagent: %v", err)
	}

	// Case 3: Lease is not Active
	reaper.PruneIntent(intentID)
	err = reaper.RegisterSubagent(intentID, "session-2")
	if err == nil {
		t.Fatal("Expected error for pruned lease, got nil")
	}
}

func TestSubagentReaper_RecordHeartbeatEdgeCases(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-heartbeat"

	// Case 1: Lease not found
	err := reaper.RecordHeartbeat(intentID, "sig", 1*time.Second)
	if err == nil {
		t.Fatal("Expected error for non-existent lease, got nil")
	}

	// Create lease
	reaper.RegisterBranch(intentID, 5*time.Second)

	// Case 2: Lease is not Active
	reaper.PruneIntent(intentID)
	err = reaper.RecordHeartbeat(intentID, "sig", 1*time.Second)
	if err == nil {
		t.Fatal("Expected error for pruned lease, got nil")
	}
}

func TestSubagentReaper_PruneIntentEdgeCases(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-prune-not-found"
	err := reaper.PruneIntent(intentID)
	if err == nil {
		t.Fatal("Expected error for non-existent lease, got nil")
	}
}

func TestSubagentReaper_GetLeaseStatusEdgeCases(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-status-not-found"
	_, err := reaper.GetLeaseStatus(intentID)
	if err == nil {
		t.Fatal("Expected error for non-existent lease, got nil")
	}
}

func TestSubagentReaper_PruneIntent(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-003"
	ttl := 5 * time.Second

	reaper.RegisterBranch(intentID, ttl)

	err := reaper.PruneIntent(intentID)
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
