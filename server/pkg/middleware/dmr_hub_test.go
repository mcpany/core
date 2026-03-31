// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
)

func TestDMRHub_Disabled(t *testing.T) {
	hub := NewDMRHub(DMRHubConfig{Enabled: false})

	req := &tool.ExecutionRequest{
		ToolName: "stateful_tool",
		Arguments: map[string]interface{}{
			"_x_dmr_node_status": "failed",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := hub.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestDMRHub_NoArguments(t *testing.T) {
	hub := NewDMRHub(DMRHubConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName:  "stateful_tool",
		Arguments: nil,
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := hub.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestDMRHub_HealthyNode(t *testing.T) {
	hub := NewDMRHub(DMRHubConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "stateful_tool",
		Arguments: map[string]interface{}{
			"_x_dmr_node_status": "healthy",
			"param1":             "value1",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		// Verify args were cleaned
		args := req.Arguments
		_, hasStatus := args["_x_dmr_node_status"]
		assert.False(t, hasStatus)
		assert.Equal(t, "value1", args["param1"])
		return "success", nil
	}

	res, err := hub.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestDMRHub_FailedNode_MissingProof(t *testing.T) {
	hub := NewDMRHub(DMRHubConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "stateful_tool",
		Arguments: map[string]interface{}{
			"_x_dmr_node_status": "failed",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := hub.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "DMR Error: migration required, missing ZKSA proof")
	assert.False(t, nextCalled)
}

func TestDMRHub_FailedNode_InvalidProof(t *testing.T) {
	hub := NewDMRHub(DMRHubConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "stateful_tool",
		Arguments: map[string]interface{}{
			"_x_dmr_node_status":      "failed",
			"_x_zksa_migration_proof": "invalid-proof-xyz",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := hub.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "DMR Error: invalid ZKSA migration proof")
	assert.False(t, nextCalled)
}

func TestDMRHub_FailedNode_Success(t *testing.T) {
	hub := NewDMRHub(DMRHubConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "stateful_tool",
		Arguments: map[string]interface{}{
			"_x_dmr_node_status":      "failed",
			"_x_zksa_migration_proof": "zksa-proof-12345",
			"param1":                  "value1",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		// Verify args were cleaned
		args := req.Arguments
		_, hasStatus := args["_x_dmr_node_status"]
		_, hasProof := args["_x_zksa_migration_proof"]
		assert.False(t, hasStatus)
		assert.False(t, hasProof)
		assert.Equal(t, "value1", args["param1"])
		return "success", nil
	}

	res, err := hub.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestDMRHub_InvalidNodeStatusFormat(t *testing.T) {
	hub := NewDMRHub(DMRHubConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "stateful_tool",
		Arguments: map[string]interface{}{
			"_x_dmr_node_status": 123, // Invalid format
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := hub.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "DMR Error: invalid node status format")
	assert.False(t, nextCalled)
}

func TestDMRHub_UnknownNodeStatus(t *testing.T) {
	hub := NewDMRHub(DMRHubConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "stateful_tool",
		Arguments: map[string]interface{}{
			"_x_dmr_node_status": "unknown-status",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := hub.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), fmt.Sprintf("DMR Error: unknown node status '%s'", "unknown-status"))
	assert.False(t, nextCalled)
}

func TestDMRHub_StatefulTools_Skipped(t *testing.T) {
	hub := NewDMRHub(DMRHubConfig{
		Enabled:       true,
		StatefulTools: []string{"critical_db"},
	})

	req := &tool.ExecutionRequest{
		ToolName: "stateless_tool",
		Arguments: map[string]interface{}{
			"_x_dmr_node_status": "failed", // Should be ignored
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := hub.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestDMRHub_StatefulTools_Enforced(t *testing.T) {
	hub := NewDMRHub(DMRHubConfig{
		Enabled:       true,
		StatefulTools: []string{"critical_db"},
	})

	req := &tool.ExecutionRequest{
		ToolName: "critical_db",
		Arguments: map[string]interface{}{
			"_x_dmr_node_status": "failed",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := hub.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "DMR Error: migration required, missing ZKSA proof")
	assert.False(t, nextCalled)
}
