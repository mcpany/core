// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package amt

import (
	"testing"
)

func TestAMTBroker_EstablishTunnel(t *testing.T) {
	broker := NewBroker()

	// Test valid establishment
	tunnelID, err := broker.EstablishTunnel("node-1", "mission-token-12345")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if tunnelID == "" {
		t.Fatal("Expected valid tunnelID, got empty string")
	}

	// Test missing remote node ID
	_, err = broker.EstablishTunnel("", "mission-token-12345")
	if err == nil || err.Error() != "remote node id cannot be empty" {
		t.Fatalf("Expected 'remote node id cannot be empty' error, got %v", err)
	}

	// Test missing mission token
	_, err = broker.EstablishTunnel("node-1", "")
	if err == nil || err.Error() != "invalid mission token" {
		t.Fatalf("Expected 'invalid mission token' error, got %v", err)
	}
}

func TestAMTBroker_InvokeRemote(t *testing.T) {
	broker := NewBroker()
	tunnelID, _ := broker.EstablishTunnel("node-1", "mission-token-12345")

	// Test valid invocation
	result, err := broker.InvokeRemote(tunnelID, "test-tool")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != "mock-result-for-test-tool" {
		t.Fatalf("Expected 'mock-result-for-test-tool', got %s", result)
	}

	// Test invalid tunnel ID
	_, err = broker.InvokeRemote("invalid-tunnel", "test-tool")
	if err == nil || err.Error() != "tunnel not found" {
		t.Fatalf("Expected 'tunnel not found' error, got %v", err)
	}
}

func TestAMTBroker_ResumeTunnel(t *testing.T) {
	broker := NewBroker()
	tunnelID, _ := broker.EstablishTunnel("node-1", "mission-token-12345")

	// Create a ticket based on the implementation detail (for testing only)
	ticket := "ticket-" + tunnelID

	// Test valid resumption
	resumedID, err := broker.ResumeTunnel(ticket)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resumedID != tunnelID {
		t.Fatalf("Expected resumed ID to match original tunnel ID")
	}

	// Test missing ticket
	_, err = broker.ResumeTunnel("")
	if err == nil || err.Error() != "invalid mesh ticket" {
		t.Fatalf("Expected 'invalid mesh ticket' error, got %v", err)
	}

	// Test unknown ticket
	_, err = broker.ResumeTunnel("unknown-ticket")
	if err == nil || err.Error() != "invalid mesh ticket" {
		t.Fatalf("Expected 'invalid mesh ticket' error, got %v", err)
	}
}