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

func TestSubagentReaper_RegisterSubagent(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-005"
	ttl := 5 * time.Second

	// Not found
	err := reaper.RegisterSubagent("not-found", "session-1")
	if err == nil {
		t.Fatal("Expected error when registering subagent for non-existent intent")
	}

	reaper.RegisterBranch(intentID, ttl)

	// Success
	err = reaper.RegisterSubagent(intentID, "session-1")
	if err != nil {
		t.Fatalf("Failed to register subagent: %v", err)
	}

	// Not active
	reaper.PruneIntent(intentID)
	err = reaper.RegisterSubagent(intentID, "session-2")
	if err == nil {
		t.Fatal("Expected error when registering subagent for pruned intent")
	}
}

func TestSubagentReaper_RecordHeartbeat_EdgeCases(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-006"
	ttl := 2 * time.Second

	// Not found
	err := reaper.RecordHeartbeat("not-found", "valid_sig", 5*time.Second)
	if err == nil {
		t.Fatal("Expected error when recording heartbeat for non-existent intent")
	}

	reaper.RegisterBranch(intentID, ttl)
	reaper.PruneIntent(intentID)

	// Not active
	err = reaper.RecordHeartbeat(intentID, "valid_sig", 5*time.Second)
	if err == nil {
		t.Fatal("Expected error when recording heartbeat for pruned intent")
	}
}

func TestSubagentReaper_PruneIntent_NotFound(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	err := reaper.PruneIntent("not-found")
	if err == nil {
		t.Fatal("Expected error when pruning non-existent intent")
	}
}

func TestSubagentReaper_GetLeaseStatus_NotFound(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	_, err := reaper.GetLeaseStatus("not-found")
	if err == nil {
		t.Fatal("Expected error when getting status of non-existent intent")
	}
}

func TestSubagentReaper_Start_ContextDone(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()

	ctx, cancel := context.WithCancel(context.Background())

	// Start the daemon with a long interval so it doesn't tick immediately
	reaper.Start(ctx, 10*time.Second)

	// Immediately cancel context
	cancel()
	time.Sleep(50 * time.Millisecond) // allow goroutine to exit

	// If we got here and it didn't panic, it gracefully handled ctx.Done()
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
