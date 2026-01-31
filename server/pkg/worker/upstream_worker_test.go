// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUpstreamWorker_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 1. Setup Bus
	b, err := bus.NewProvider(nil)
	require.NoError(t, err)

	// 2. Setup Mock Tool Manager
	mockTM := tool.NewMockManagerInterface(ctrl)

	// 3. Create Worker
	w := worker.NewUpstreamWorker(b, mockTM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w.Start(ctx)

	// 4. Subscribe to Result Topic to verify output
	resultBus, err := bus.GetBus[*bus.ToolExecutionResult](b, bus.ToolExecutionResultTopic)
	require.NoError(t, err)

	resultCh := make(chan *bus.ToolExecutionResult, 1)
	unsubscribe := resultBus.Subscribe(ctx, "test-correlation-id", func(res *bus.ToolExecutionResult) {
		select {
		case resultCh <- res:
		default:
		}
	})
	defer unsubscribe()

	// 5. Expectation
	toolName := "test-tool"
	toolInputs := []byte(`{"arg":"value"}`)
	executionResult := map[string]string{"output": "success"}

	mockTM.EXPECT().ExecuteTool(gomock.Any(), gomock.AssignableToTypeOf(&tool.ExecutionRequest{})).DoAndReturn(
		func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
			assert.Equal(t, toolName, req.ToolName)
			assert.Equal(t, json.RawMessage(toolInputs), req.ToolInputs)
			return executionResult, nil
		},
	).Times(1)

	// 6. Publish Request
	requestBus, err := bus.GetBus[*bus.ToolExecutionRequest](b, bus.ToolExecutionRequestTopic)
	require.NoError(t, err)

	req := &bus.ToolExecutionRequest{
		ToolName:   toolName,
		ToolInputs: toolInputs,
	}
	req.SetCorrelationID("test-correlation-id")
	req.Context = ctx

	err = requestBus.Publish(ctx, "request", req)
	require.NoError(t, err)

	// 7. Verify Result
	select {
	case res := <-resultCh:
		assert.Equal(t, "test-correlation-id", res.CorrelationID())
		assert.NoError(t, res.Error)

		var actualResult map[string]string
		err := json.Unmarshal(res.Result, &actualResult)
		require.NoError(t, err)
		assert.Equal(t, "success", actualResult["output"])
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for result")
	}

	cancel() // Signal worker to stop
	w.Stop() // Wait for worker to stop
}

func TestUpstreamWorker_ExecutionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	b, err := bus.NewProvider(nil)
	require.NoError(t, err)

	mockTM := tool.NewMockManagerInterface(ctrl)
	w := worker.NewUpstreamWorker(b, mockTM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	resultBus, err := bus.GetBus[*bus.ToolExecutionResult](b, bus.ToolExecutionResultTopic)
	require.NoError(t, err)

	resultCh := make(chan *bus.ToolExecutionResult, 1)
	unsubscribe := resultBus.Subscribe(ctx, "error-correlation-id", func(res *bus.ToolExecutionResult) {
		select {
		case resultCh <- res:
		default:
		}
	})
	defer unsubscribe()

	expectedErr := errors.New("execution failed")
	mockTM.EXPECT().ExecuteTool(gomock.Any(), gomock.Any()).Return(nil, expectedErr).Times(1)

	requestBus, err := bus.GetBus[*bus.ToolExecutionRequest](b, bus.ToolExecutionRequestTopic)
	require.NoError(t, err)

	req := &bus.ToolExecutionRequest{
		ToolName:   "fail-tool",
		ToolInputs: []byte("{}"),
	}
	req.SetCorrelationID("error-correlation-id")
	req.Context = ctx

	err = requestBus.Publish(ctx, "request", req)
	require.NoError(t, err)

	select {
	case res := <-resultCh:
		assert.Equal(t, "error-correlation-id", res.CorrelationID())
		assert.Error(t, res.Error)
		assert.Equal(t, expectedErr.Error(), res.Error.Error())
		assert.Empty(t, res.Result)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for result")
	}

	cancel()
	w.Stop()
}

func TestUpstreamWorker_Lifecycle(t *testing.T) {
	b, err := bus.NewProvider(nil)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTMGo := tool.NewMockManagerInterface(ctrl)

	w := worker.NewUpstreamWorker(b, mockTMGo)

	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)

	// Ensure it started (async)
	time.Sleep(10 * time.Millisecond)

	cancel() // Cancel context to stop subscriber
	w.Stop() // Wait for wg
}
