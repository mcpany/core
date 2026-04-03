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
	sessionID := "session-001"
	ttl := 5 * time.Second

	// Test registering a subagent without a lease
	err := reaper.RegisterSubagent(intentID, sessionID)
	if err == nil {
		t.Fatal("Expected error when registering subagent for non-existent lease")
	}

	// Create lease
	reaper.RegisterBranch(intentID, ttl)

	// Test successful registration
	err = reaper.RegisterSubagent(intentID, sessionID)
	if err != nil {
		t.Fatalf("Failed to register subagent: %v", err)
	}

	// Test registering a subagent on pruned lease
	reaper.PruneIntent(intentID)
	err = reaper.RegisterSubagent(intentID, sessionID)
	if err == nil {
		t.Fatal("Expected error when registering subagent on pruned lease")
	}
}

func TestSubagentReaper_RecordHeartbeat(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-002"
	ttl := 2 * time.Second

	// Test heartbeat on non-existent lease
	err := reaper.RecordHeartbeat(intentID, "valid_sig", 5*time.Second)
	if err == nil {
		t.Fatal("Expected error when recording heartbeat for non-existent lease")
	}

	// Create lease
	reaper.RegisterBranch(intentID, ttl)

	// Test successful heartbeat
	err = reaper.RecordHeartbeat(intentID, "valid_sig", 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to record heartbeat: %v", err)
	}

	// Test heartbeat on pruned lease
	reaper.PruneIntent(intentID)
	err = reaper.RecordHeartbeat(intentID, "valid_sig", 5*time.Second)
	if err == nil {
		t.Fatal("Expected error when recording heartbeat on pruned lease")
	}
}

func TestSubagentReaper_PruneIntent(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-003"
	ttl := 5 * time.Second

	// Test prune on non-existent lease
	err := reaper.PruneIntent("non-existent")
	if err == nil {
		t.Fatal("Expected error when pruning non-existent lease")
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

	// Test on non-existent lease
	_, err := reaper.GetLeaseStatus("non-existent")
	if err == nil {
		t.Fatal("Expected error when getting status for non-existent lease")
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
