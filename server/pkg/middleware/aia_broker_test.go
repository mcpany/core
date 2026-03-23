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

func TestAIABroker_Disabled(t *testing.T) {
	broker := NewAIABroker(AIABrokerConfig{Enabled: false})

	req := &tool.ExecutionRequest{
		ToolName: "test_tool",
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := broker.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestAIABroker_MissingArguments(t *testing.T) {
	broker := NewAIABroker(AIABrokerConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "test_tool",
		Arguments: nil,
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := broker.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "Intent Drift Detected: missing mission-root alignment heartbeat")
	assert.False(t, nextCalled)
}

func TestAIABroker_MissingAlignmentHeaders(t *testing.T) {
	broker := NewAIABroker(AIABrokerConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "test_tool",
		Arguments: map[string]interface{}{
			"param1": "value1",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := broker.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "Intent Drift Detected: missing mission-root alignment heartbeat or intent")
	assert.False(t, nextCalled)
}

func TestAIABroker_InvalidAlignmentHeaders(t *testing.T) {
	broker := NewAIABroker(AIABrokerConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "test_tool",
		Arguments: map[string]interface{}{
			"_x_alignment_heartbeat": 123,
			"_x_mission_root_intent": "intent1",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := broker.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "Intent Drift Detected: invalid alignment heartbeat format")
	assert.False(t, nextCalled)
}

func TestAIABroker_AttestationFailed(t *testing.T) {
	broker := NewAIABroker(AIABrokerConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "test_tool",
		Arguments: map[string]interface{}{
			"_x_alignment_heartbeat": "invalid-prefix-heartbeat",
			"_x_mission_root_intent": "intent1",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := broker.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), fmt.Sprintf("Intent Drift Detected: heartbeat attestation failed for intent '%s'", "intent1"))
	assert.False(t, nextCalled)
}

func TestAIABroker_Success(t *testing.T) {
	broker := NewAIABroker(AIABrokerConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "test_tool",
		Arguments: map[string]interface{}{
			"_x_alignment_heartbeat": "attested-signature-123",
			"_x_mission_root_intent": "root-intent-1",
			"param1":                 "value1",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true

		// Verify args were cleaned
		args := req.Arguments
		_, hasHeartbeat := args["_x_alignment_heartbeat"]
		_, hasIntent := args["_x_mission_root_intent"]
		assert.False(t, hasHeartbeat)
		assert.False(t, hasIntent)
		assert.Equal(t, "value1", args["param1"])

		return "success", nil
	}

	res, err := broker.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestAIABroker_RequiredFor_Skipped(t *testing.T) {
	broker := NewAIABroker(AIABrokerConfig{
		Enabled:     true,
		RequiredFor: []string{"critical_tool"},
	})

	req := &tool.ExecutionRequest{
		ToolName: "safe_tool",
		Arguments: map[string]interface{}{},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := broker.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestAIABroker_RequiredFor_Enforced(t *testing.T) {
	broker := NewAIABroker(AIABrokerConfig{
		Enabled:     true,
		RequiredFor: []string{"critical_tool"},
	})

	req := &tool.ExecutionRequest{
		ToolName: "critical_tool",
		Arguments: map[string]interface{}{},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return nil, nil
	}

	res, err := broker.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "Intent Drift Detected: missing mission-root alignment heartbeat or intent")
	assert.False(t, nextCalled)
}
