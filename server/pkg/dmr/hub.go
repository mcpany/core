// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package dmr implements the Dynamic Mesh Resilience (DMR) Hub.
package dmr

import (
	"context"
	"errors"
	"sync"
	"time"
)

// NodeState represents the public NodeState entity.
//
// Summary: Defines the structured data model representing a state.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type NodeState struct {
	ID             string
	LastHeartbeat  time.Time
	IsAttested     bool
	ActiveMissions []string
}

// Hub represents the public Hub entity.
//
// Summary: Defines the structured data model representing a .
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Hub interface {
	// RegisterNode adds a new node to the mesh or updates its state.
	//
	// Summary: Registers a node in the DMR Hub.
	//
	// Parameters:
	//   - id (string): The unique identifier of the node.
	//   - isAttested (bool): Whether the node has provided valid hardware attestation.
	//
	// Returns:
	//   - error: An error if the id is empty.
	//
	// Errors:
	//   - Returns "node id cannot be empty" if id is empty.
	//
	// Side Effects:
	//   - Modifies the internal nodes map.
	RegisterNode(id string, isAttested bool) error

	// Heartbeat processes a heartbeat signal from a mesh node.
	//
	// Summary: Updates the last heartbeat timestamp for a node.
	//
	// Parameters:
	//   - id (string): The unique identifier of the node.
	//
	// Returns:
	//   - error: An error if the node is not found.
	//
	// Errors:
	//   - Returns "node not found" if the node is not registered.
	//
	// Side Effects:
	//   - Updates the LastHeartbeat time for the node.
	Heartbeat(id string) error

	// CheckHealth scans the registered nodes for timeouts or attestation failures.
	//
	// Summary: Evaluates node health and triggers migration for failed nodes.
	//
	// Parameters:
	//   - ctx (context.Context): The context for the health check.
	//
	// Returns:
	//   - []string: A list of node IDs that have failed and require migration.
	//
	// Errors:
	//   - None.
	//
	// Side Effects:
	//   - Can send failed node IDs to the migration channel.
	CheckHealth(ctx context.Context) []string

	// MigrationChannel returns a read-only channel for listening to migration events.
	//
	// Summary: Provides access to the stream of failed node IDs that require migration.
	//
	// Parameters:
	//   - None.
	//
	// Returns:
	//   - <-chan string: A channel emitting failed node IDs.
	//
	// Errors:
	//   - None.
	//
	// Side Effects:
	//   - None.
	MigrationChannel() <-chan string
}

// hubImpl implements the Hub interface.
type hubImpl struct {
	mu           sync.RWMutex
	nodes        map[string]*NodeState
	timeout      time.Duration
	migrationCh  chan string
}

// NewHub serves as a public interface for interacting with NewHub.
//
// Summary: Constructs and returns an initialized hub ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewHub(timeout time.Duration) Hub {
	return &hubImpl{
		nodes:       make(map[string]*NodeState),
		timeout:     timeout,
		migrationCh: make(chan string, 100),
	}
}

// RegisterNode serves as a public interface for interacting with RegisterNode.
//
// Summary: Register the node appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (h *hubImpl) RegisterNode(id string, isAttested bool) error {
	if id == "" {
		return errors.New("node id cannot be empty")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.nodes[id] = &NodeState{
		ID:            id,
		LastHeartbeat: time.Now(),
		IsAttested:    isAttested,
	}
	return nil
}

// Heartbeat serves as a public interface for interacting with Heartbeat.
//
// Summary: Heartbeat the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (h *hubImpl) Heartbeat(id string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	node, exists := h.nodes[id]
	if !exists {
		return errors.New("node not found")
	}

	node.LastHeartbeat = time.Now()
	return nil
}

// CheckHealth serves as a public interface for interacting with CheckHealth.
//
// Summary: Check the health appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (h *hubImpl) CheckHealth(ctx context.Context) []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	var failedNodes []string
	now := time.Now()

	for id, node := range h.nodes {
		if now.Sub(node.LastHeartbeat) > h.timeout || !node.IsAttested {
			failedNodes = append(failedNodes, id)
			delete(h.nodes, id)
			select {
			case h.migrationCh <- id:
			default:
				// Channel is full, migration might be delayed
			}
		}
	}

	return failedNodes
}

// MigrationChannel serves as a public interface for interacting with MigrationChannel.
//
// Summary: Migration the channel appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (h *hubImpl) MigrationChannel() <-chan string {
	return h.migrationCh
}
