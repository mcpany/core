package dmr

import (
	"context"
	"errors"
	"time"
)

// ShardID represents a unique identifier for a piece of entangled state.
type ShardID string

// NodeID represents a unique identifier for a physical node in the mesh.
type NodeID string

// Shard represents a piece of entangled state managed by the mesh.
type Shard struct {
	ID        ShardID
	State     []byte
	Signature string // Hardware-attested signature of the state
	Owner     NodeID // Current node responsible for this shard
}

// MeshNode represents a physical node participating in the Dynamic Mesh.
type MeshNode struct {
	ID             NodeID
	Status         NodeStatus
	LastHeartbeat  time.Time
	ClockDrift     time.Duration // To neutralize "Shadow-Attestation"
	AssignedShards []ShardID
}

// NodeStatus defines the current health and attestation state of a node.
type NodeStatus string

const (
	StatusHealthy    NodeStatus = "HEALTHY"
	StatusDegraded   NodeStatus = "DEGRADED"
	StatusCompromised NodeStatus = "COMPROMISED"
	StatusOffline    NodeStatus = "OFFLINE"
)

// MigrationRequest defines the parameters for moving a shard from one node to another.
type MigrationRequest struct {
	ShardID     ShardID
	SourceNode  NodeID
	TargetNode  NodeID
	Reason      string
	Attestation string // Proof authorizing the migration
}

// DMRHub is the central nervous system for Dynamic Mesh Resilience.
// It monitors node health, orchestrates state migrations, and ensures
// HACA (Hardware-Attested Cost Attribution) compliance during re-sharding.
type DMRHub interface {
	// RegisterNode adds a new physical node to the mesh.
	RegisterNode(ctx context.Context, node *MeshNode) error

	// Heartbeat processes a liveness and attestation ping from a node.
	// Returns an error if the node's clock drift exceeds the acceptable threshold
	// (neutralizing "Shadow-Attestation" exploits).
	Heartbeat(ctx context.Context, nodeID NodeID, drift time.Duration, attestation string) error

	// GetNodeStatus retrieves the current status of a specific node.
	GetNodeStatus(ctx context.Context, nodeID NodeID) (NodeStatus, error)

	// InitiateMigration triggers the atomic transfer of a shard.
	InitiateMigration(ctx context.Context, req MigrationRequest) error

	// Rebalance dynamically calculates and executes a new shard distribution
	// across healthy nodes.
	Rebalance(ctx context.Context) error
}

var (
	ErrNodeNotFound       = errors.New("node not found in mesh")
	ErrNodeCompromised    = errors.New("node attestation failed or compromised")
	ErrClockDriftExceeded = errors.New("node clock drift exceeds security threshold (potential shadow-attestation)")
	ErrMigrationFailed    = errors.New("shard migration failed")
)
