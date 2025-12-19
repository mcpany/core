// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestCallPolicyMiddleware(t *testing.T) {
	const successResult = "success"

	actionPtr := func(a configv1.CallPolicy_Action) *configv1.CallPolicy_Action {
		return &a
	}

	t.Run("no policies -> allowed", func(t *testing.T) {
		// No policy
		cpMiddleware := middleware.NewCallPolicyMiddleware("service", nil)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}

		nextCalled := false
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return successResult, nil
		}

		result, err := cpMiddleware.Execute(context.Background(), req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, result)
		assert.True(t, nextCalled)
	})

	t.Run("name regex deny -> blocked", func(t *testing.T) {
		policy := &configv1.CallPolicy{
			DefaultAction: actionPtr(configv1.CallPolicy_ALLOW),
			Rules: []*configv1.CallPolicyRule{
				{
					Action:    actionPtr(configv1.CallPolicy_DENY),
					NameRegex: proto.String(".*test-tool"),
				},
			},
		}

		cpMiddleware := middleware.NewCallPolicyMiddleware("service", policy)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}

		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			return successResult, nil
		}

		_, err := cpMiddleware.Execute(context.Background(), req, next)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execution denied by policy")
	})

	t.Run("argument regex deny -> blocked", func(t *testing.T) {
		policy := &configv1.CallPolicy{
			DefaultAction: actionPtr(configv1.CallPolicy_ALLOW),
			Rules: []*configv1.CallPolicyRule{
				{
					Action:        actionPtr(configv1.CallPolicy_DENY),
					ArgumentRegex: proto.String(".*dangerous.*"),
				},
			},
		}

		cpMiddleware := middleware.NewCallPolicyMiddleware("service", policy)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{"cmd": "dangerous command"}`),
		}

		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			return successResult, nil
		}

		_, err := cpMiddleware.Execute(context.Background(), req, next)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execution denied by policy")
	})

	t.Run("argument regex mismatch -> allowed", func(t *testing.T) {
		policy := &configv1.CallPolicy{
			DefaultAction: actionPtr(configv1.CallPolicy_ALLOW),
			Rules: []*configv1.CallPolicyRule{
				{
					Action:        actionPtr(configv1.CallPolicy_DENY),
					ArgumentRegex: proto.String(".*dangerous.*"),
				},
			},
		}

		cpMiddleware := middleware.NewCallPolicyMiddleware("service", policy)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{"cmd": "safe command"}`),
		}

		nextCalled := false
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return successResult, nil
		}

		result, err := cpMiddleware.Execute(context.Background(), req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, result)
		assert.True(t, nextCalled)
	})

	t.Run("default deny -> blocked", func(t *testing.T) {
		policy := &configv1.CallPolicy{
			DefaultAction: actionPtr(configv1.CallPolicy_DENY),
		}

		cpMiddleware := middleware.NewCallPolicyMiddleware("service", policy)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}

		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			return successResult, nil
		}

		_, err := cpMiddleware.Execute(context.Background(), req, next)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "execution denied by default policy")
	})

	t.Run("default deny but allowed by rule -> allowed", func(t *testing.T) {
		policy := &configv1.CallPolicy{
			DefaultAction: actionPtr(configv1.CallPolicy_DENY),
			Rules: []*configv1.CallPolicyRule{
				{
					Action:    actionPtr(configv1.CallPolicy_ALLOW),
					NameRegex: proto.String(".*test-tool"),
				},
			},
		}

		cpMiddleware := middleware.NewCallPolicyMiddleware("service", policy)

		req := &tool.ExecutionRequest{
			ToolName:   "service.test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}

		nextCalled := false
		next := func(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
			nextCalled = true
			return successResult, nil
		}

		result, err := cpMiddleware.Execute(context.Background(), req, next)
		assert.NoError(t, err)
		assert.Equal(t, successResult, result)
		assert.True(t, nextCalled)
	})
}
