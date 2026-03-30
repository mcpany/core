package dmr_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/dmr"
)

// mockHub implements dmr.DMRHub for testing
type mockHub struct {
	nodes map[dmr.NodeID]*dmr.MeshNode
}

func newMockHub() *mockHub {
	return &mockHub{
		nodes: make(map[dmr.NodeID]*dmr.MeshNode),
	}
}

func (h *mockHub) RegisterNode(ctx context.Context, node *dmr.MeshNode) error {
	if node == nil {
		return errors.New("node cannot be nil")
	}
	h.nodes[node.ID] = node
	return nil
}

func (h *mockHub) Heartbeat(ctx context.Context, nodeID dmr.NodeID, drift time.Duration, attestation string) error {
	node, ok := h.nodes[nodeID]
	if !ok {
		return dmr.ErrNodeNotFound
	}

	if drift > 10*time.Millisecond {
		node.Status = dmr.StatusCompromised
		return dmr.ErrClockDriftExceeded
	}

	node.LastHeartbeat = time.Now()
	node.ClockDrift = drift
	node.Status = dmr.StatusHealthy
	return nil
}

func (h *mockHub) GetNodeStatus(ctx context.Context, nodeID dmr.NodeID) (dmr.NodeStatus, error) {
	node, ok := h.nodes[nodeID]
	if !ok {
		return "", dmr.ErrNodeNotFound
	}
	return node.Status, nil
}

func (h *mockHub) InitiateMigration(ctx context.Context, req dmr.MigrationRequest) error {
	return nil
}

func (h *mockHub) Rebalance(ctx context.Context) error {
	return nil
}

func TestHeartbeat_ClockDrift(t *testing.T) {
	hub := newMockHub()
	nodeID := dmr.NodeID("node-1")
	node := &dmr.MeshNode{
		ID:     nodeID,
		Status: dmr.StatusHealthy,
	}

	err := hub.RegisterNode(context.Background(), node)
	if err != nil {
		t.Fatalf("failed to register node: %v", err)
	}

	// Test normal heartbeat
	err = hub.Heartbeat(context.Background(), nodeID, 5*time.Millisecond, "valid-attestation")
	if err != nil {
		t.Errorf("expected no error for valid heartbeat, got %v", err)
	}

	// Test shadow-attestation drift
	err = hub.Heartbeat(context.Background(), nodeID, 50*time.Millisecond, "valid-attestation")
	if !errors.Is(err, dmr.ErrClockDriftExceeded) {
		t.Errorf("expected ErrClockDriftExceeded, got %v", err)
	}

	status, _ := hub.GetNodeStatus(context.Background(), nodeID)
	if status != dmr.StatusCompromised {
		t.Errorf("expected node to be marked compromised after drift exceeded, got %v", status)
	}
}
