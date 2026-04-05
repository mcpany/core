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
	intentID := "branch-register-001"
	ttl := 5 * time.Second

	// Error when registering to non-existent lease
	err := reaper.RegisterSubagent(intentID, "session-001")
	if err == nil {
		t.Fatal("Expected error when registering to non-existent lease, got nil")
	}

	reaper.RegisterBranch(intentID, ttl)

	// Success case
	err = reaper.RegisterSubagent(intentID, "session-001")
	if err != nil {
		t.Fatalf("Failed to register subagent: %v", err)
	}

	// Error when registering to a pruned lease
	err = reaper.PruneIntent(intentID)
	if err != nil {
		t.Fatalf("Failed to prune intent: %v", err)
	}

	err = reaper.RegisterSubagent(intentID, "session-002")
	if err == nil {
		t.Fatal("Expected error when registering to pruned lease, got nil")
	}
}

func TestSubagentReaper_RecordHeartbeat(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-002"
	ttl := 2 * time.Second

	// Error on missing lease
	err := reaper.RecordHeartbeat(intentID, "valid_sig", 5*time.Second)
	if err == nil {
		t.Fatal("Expected error when recording heartbeat for non-existent lease, got nil")
	}

	reaper.RegisterBranch(intentID, ttl)

	// Success case
	err = reaper.RecordHeartbeat(intentID, "valid_sig", 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to record heartbeat: %v", err)
	}

	// Error on pruned lease
	err = reaper.PruneIntent(intentID)
	if err != nil {
		t.Fatalf("Failed to prune intent: %v", err)
	}

	err = reaper.RecordHeartbeat(intentID, "valid_sig", 5*time.Second)
	if err == nil {
		t.Fatal("Expected error when recording heartbeat for pruned lease, got nil")
	}
}

func TestSubagentReaper_PruneIntent(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-003"
	ttl := 5 * time.Second

	// Error on missing lease
	err := reaper.PruneIntent(intentID)
	if err == nil {
		t.Fatal("Expected error when pruning non-existent intent, got nil")
	}

	reaper.RegisterBranch(intentID, ttl)

	// Success case
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
	intentID := "branch-status-001"

	// Error on missing lease
	_, err := reaper.GetLeaseStatus(intentID)
	if err == nil {
		t.Fatal("Expected error when getting status of non-existent lease, got nil")
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
