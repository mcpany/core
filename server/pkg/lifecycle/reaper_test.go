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

func TestSubagentReaper_RegisterSubagent(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "intent-001"
	ttl := 5 * time.Second

	// 1. Happy Path: Registering a subagent on an ACTIVE lease
	reaper.RegisterBranch(intentID, ttl)

	err := reaper.RegisterSubagent(intentID, "session-001")
	if err != nil {
		t.Fatalf("Expected successful registration, got err: %v", err)
	}

	// 2. Error Path: Registering on a non-existent lease
	err = reaper.RegisterSubagent("non-existent", "session-002")
	if err == nil {
		t.Fatalf("Expected error for non-existent lease, got nil")
	}
	expectedErrMsg1 := "lease not found for intent: non-existent"
	if err.Error() != expectedErrMsg1 {
		t.Errorf("Expected error '%s', got '%v'", expectedErrMsg1, err)
	}

	// 3. Error Path: Registering on a non-ACTIVE lease
	intentID2 := "intent-002"
	reaper.RegisterBranch(intentID2, ttl)

	// Transition it to PRUNED manually to test the failure
	err = reaper.PruneIntent(intentID2)
	if err != nil {
		t.Fatalf("Failed to prune intent: %v", err)
	}

	err = reaper.RegisterSubagent(intentID2, "session-003")
	if err == nil {
		t.Fatalf("Expected error for non-ACTIVE lease, got nil")
	}
	expectedErrMsg2 := "cannot register subagent: lease is PRUNED"
	if err.Error() != expectedErrMsg2 {
		t.Errorf("Expected error '%s', got '%v'", expectedErrMsg2, err)
	}
}
