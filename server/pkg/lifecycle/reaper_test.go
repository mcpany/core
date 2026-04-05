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
	sessionID := "session-001"
	ttl := 5 * time.Second

	// Test missing intent
	err := reaper.RegisterSubagent("missing-intent", sessionID)
	if err == nil {
		t.Fatal("Expected error for missing intent")
	}

	reaper.RegisterBranch(intentID, ttl)

	// Test valid register
	err = reaper.RegisterSubagent(intentID, sessionID)
	if err != nil {
		t.Fatalf("Failed to register subagent: %v", err)
	}

	// Test inactive lease
	reaper.PruneIntent(intentID)
	err = reaper.RegisterSubagent(intentID, sessionID)
	if err == nil {
		t.Fatal("Expected error for registering to inactive lease")
	}
}

func TestSubagentReaper_RecordHeartbeat(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-002"
	ttl := 2 * time.Second

	// Test missing intent
	err := reaper.RecordHeartbeat("missing-intent", "valid_sig", 5*time.Second)
	if err == nil {
		t.Fatal("Expected error for missing intent")
	}

	reaper.RegisterBranch(intentID, ttl)

	err = reaper.RecordHeartbeat(intentID, "valid_sig", 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to record heartbeat: %v", err)
	}

	// Test inactive lease
	reaper.PruneIntent(intentID)
	err = reaper.RecordHeartbeat(intentID, "valid_sig", 5*time.Second)
	if err == nil {
		t.Fatal("Expected error for heartbeat on inactive lease")
	}
}

func TestSubagentReaper_PruneIntent(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	intentID := "branch-003"
	ttl := 5 * time.Second

	// Test missing intent
	err := reaper.PruneIntent("missing-intent")
	if err == nil {
		t.Fatal("Expected error for missing intent")
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

	// Test missing intent
	_, err := reaper.GetLeaseStatus("missing-intent")
	if err == nil {
		t.Fatal("Expected error for missing intent")
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

func TestSubagentReaper_Start_ContextCancel(t *testing.T) {
	reaper := lifecycle.NewSubagentReaper()
	ctx, cancel := context.WithCancel(context.Background())

	// Start daemon
	reaper.Start(ctx, 1*time.Hour)

	// Cancel context to stop daemon
	cancel()
	time.Sleep(10 * time.Millisecond)
}
