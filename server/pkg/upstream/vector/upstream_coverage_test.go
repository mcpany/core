// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"context"
	"errors"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
)

func TestVectorTools_HandlerErrors(t *testing.T) {
	upstream := &Upstream{}
	mockClient := new(MockVectorClient)
	tools := upstream.getTools(mockClient)
	ctx := context.Background()

	t.Run("query_vectors_validation", func(t *testing.T) {
		var queryTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "query_vectors" {
				queryTool = &tool
				break
			}
		}
		assert.NotNil(t, queryTool)

		// Test missing vector
		_, err := queryTool.Handler(ctx, map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vector is required")

		// Test invalid vector type
		_, err = queryTool.Handler(ctx, map[string]interface{}{"vector": "not-an-array"})
		assert.Error(t, err)

		// Test invalid vector elements
		_, err = queryTool.Handler(ctx, map[string]interface{}{"vector": []interface{}{"not-a-number"}})
		assert.Error(t, err)

		// Test valid vector, invalid top_k type (should default)
		mockClient.On("Query", ctx, []float32{1.0}, int64(10), map[string]interface{}(nil), "").Return(map[string]interface{}{}, nil).Once()
		_, err = queryTool.Handler(ctx, map[string]interface{}{
			"vector": []interface{}{1.0},
			"top_k":  "invalid",
		})
		assert.NoError(t, err)
	})

	t.Run("upsert_vectors_validation", func(t *testing.T) {
		var upsertTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "upsert_vectors" {
				upsertTool = &tool
				break
			}
		}
		assert.NotNil(t, upsertTool)

		// Test missing vectors
		_, err := upsertTool.Handler(ctx, map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vectors is required")

		// Test invalid vectors type
		_, err = upsertTool.Handler(ctx, map[string]interface{}{"vectors": "not-an-array"})
		assert.Error(t, err)

		// Test invalid vector object
		_, err = upsertTool.Handler(ctx, map[string]interface{}{"vectors": []interface{}{"not-a-map"}})
		assert.Error(t, err)
	})

	t.Run("delete_vectors_validation", func(t *testing.T) {
		var deleteTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "delete_vectors" {
				deleteTool = &tool
				break
			}
		}
		assert.NotNil(t, deleteTool)

		// Test with valid IDs (ensure handler works)
		mockClient.On("Delete", ctx, []string{"1"}, "", map[string]interface{}(nil)).Return(map[string]interface{}{}, nil).Once()
		_, err := deleteTool.Handler(ctx, map[string]interface{}{
			"ids": []interface{}{"1"},
		})
		assert.NoError(t, err)

		// Test valid IDs but conversion logic (already covered by main test, but reinforcing)
		mockClient.On("Delete", ctx, []string{"123"}, "", map[string]interface{}(nil)).Return(map[string]interface{}{}, nil).Once()
		_, err = deleteTool.Handler(ctx, map[string]interface{}{
			"ids": []interface{}{123},
		})
		assert.NoError(t, err)
	})
}

func TestVectorCallable_Error(t *testing.T) {
	// Test that callable propagates handler errors
	handler := func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
		return nil, errors.New("handler error")
	}
	c := &vectorCallable{handler: handler}
	// Call wrapper expects ExecutionRequest
	// But Call signature is (ctx, *ExecutionRequest)
	// My mock handler doesn't use args so it's fine passing nil arguments in request
	// But we need to pass a valid struct pointer to avoid panic if the code dereferences it before calling handler?
	// The code in upstream.go: `return c.handler(ctx, req.Arguments)` - it dereferences req.

	req := &tool.ExecutionRequest{Arguments: map[string]interface{}{}}
	_, err := c.Call(context.Background(), req)
	assert.Error(t, err)
	assert.Equal(t, "handler error", err.Error())
}

func TestUpstream_Shutdown(t *testing.T) {
	u := NewUpstream()
	err := u.Shutdown(context.Background())
	assert.NoError(t, err)
}
