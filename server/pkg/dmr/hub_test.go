// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package dmr

import (
	"context"
	"testing"
	"time"
)

func TestDMRHub_Registration(t *testing.T) {
	hub := NewHub(time.Second)

	err := hub.RegisterNode("node-1", true)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = hub.RegisterNode("", true)
	if err == nil {
		t.Error("Expected error for empty node id, got nil")
	}
}

func TestDMRHub_Heartbeat(t *testing.T) {
	hub := NewHub(time.Second)
	hub.RegisterNode("node-1", true)

	err := hub.Heartbeat("node-1")
	if err != nil {
		t.Errorf("Expected no error on valid heartbeat, got %v", err)
	}

	err = hub.Heartbeat("node-unknown")
	if err == nil {
		t.Error("Expected error for unknown node heartbeat, got nil")
	}
}

func TestDMRHub_HealthCheck(t *testing.T) {
	hub := NewHub(10 * time.Millisecond)
	hub.RegisterNode("node-1", true)
	hub.RegisterNode("node-2", false) // Attestation failure

	// node-2 should fail immediately because IsAttested is false
	failed := hub.CheckHealth(context.Background())
	if len(failed) != 1 || failed[0] != "node-2" {
		t.Errorf("Expected node-2 to fail attestation, got %v", failed)
	}

	// Wait for node-1 to timeout
	time.Sleep(15 * time.Millisecond)

	failedAfterTimeout := hub.CheckHealth(context.Background())

	// node-1 should now fail too due to timeout. node-2 was already removed.
	if len(failedAfterTimeout) != 1 || failedAfterTimeout[0] != "node-1" {
		t.Errorf("Expected node-1 to fail, got %v", failedAfterTimeout)
	}
}
