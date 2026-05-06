// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mcpany/core/server/pkg/tool"
)

func TestAMTBroker_Execute_Disabled(t *testing.T) {
	broker := NewAMTBroker(AMTBrokerConfig{Enabled: false})

	req := &tool.ExecutionRequest{
		ToolName:  "remote_exec",
		Arguments: map[string]interface{}{},
	}

	nextCalled := false
	next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := broker.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestAMTBroker_Execute_InvalidTicketFormat(t *testing.T) {
	broker := NewAMTBroker(AMTBrokerConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "remote_exec",
		Arguments: map[string]interface{}{
			"meshTicket": 12345, // Not a string
		},
	}

	next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
		return nil, errors.New("should not be called")
	}

	res, err := broker.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "invalid mesh ticket format")
}

func TestAMTBroker_Execute_InvalidTicket(t *testing.T) {
	broker := NewAMTBroker(AMTBrokerConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "remote_exec",
		Arguments: map[string]interface{}{
			"meshTicket": "bad-ticket",
		},
	}

	next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
		return nil, errors.New("should not be called")
	}

	res, err := broker.Execute(context.Background(), req, next)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "invalid mesh ticket: bad-ticket")
}

func TestAMTBroker_Execute_Success(t *testing.T) {
	broker := NewAMTBroker(AMTBrokerConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "remote_exec",
		Arguments: map[string]interface{}{
			"meshTicket": "valid-mission-bound-ticket",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := broker.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestAMTBroker_Execute_NoTicket(t *testing.T) {
	broker := NewAMTBroker(AMTBrokerConfig{Enabled: true})

	req := &tool.ExecutionRequest{
		ToolName: "local_exec",
		Arguments: map[string]interface{}{},
	}

	nextCalled := false
	next := func(ctx context.Context, r *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := broker.Execute(context.Background(), req, next)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}
