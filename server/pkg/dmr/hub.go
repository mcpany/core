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

// NodeState represents the health and attestation state of a mesh node.
type NodeState struct {
	ID             string
	LastHeartbeat  time.Time
	IsAttested     bool
	ActiveMissions []string
}

// Hub manages the active nodes in the mesh and triggers state migration on failure.
//
// Summary: The authoritative coordinator for mesh resilience and state migration.
type Hub interface {
	RegisterNode(id string, isAttested bool) error
	Heartbeat(id string) error
	CheckHealth(ctx context.Context) []string
	MigrationChannel() <-chan string
}

// hubImpl implements the Hub interface.
type hubImpl struct {
	mu           sync.RWMutex
	nodes        map[string]*NodeState
	timeout      time.Duration
	migrationCh  chan string
}

// NewHub initializes a new Dynamic Mesh Resilience Hub.
//
// Summary: Creates a new DMR Hub with a specified heartbeat timeout.
//
// Parameters:
//   - timeout (time.Duration): The duration after which a node is considered failed if no heartbeat is received.
//
// Returns:
//   - Hub: The initialized DMR Hub interface.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewHub(timeout time.Duration) Hub {
	return &hubImpl{
		nodes:       make(map[string]*NodeState),
		timeout:     timeout,
		migrationCh: make(chan string, 100),
	}
}

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
func (h *hubImpl) MigrationChannel() <-chan string {
	return h.migrationCh
}
