// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockVectorClient is a mock implementation of VectorClient
type MockVectorClient struct {
	mock.Mock
}

func (m *MockVectorClient) Query(ctx context.Context, vector []float32, topK int64, filter map[string]interface{}, namespace string) (map[string]interface{}, error) {
	args := m.Called(ctx, vector, topK, filter, namespace)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockVectorClient) Upsert(ctx context.Context, vectors []map[string]interface{}, namespace string) (map[string]interface{}, error) {
	args := m.Called(ctx, vectors, namespace)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockVectorClient) Delete(ctx context.Context, ids []string, namespace string, filter map[string]interface{}) (map[string]interface{}, error) {
	args := m.Called(ctx, ids, namespace, filter)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockVectorClient) DescribeIndexStats(ctx context.Context, filter map[string]interface{}) (map[string]interface{}, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func TestVectorTools(t *testing.T) {
	upstream := &Upstream{}
	mockClient := new(MockVectorClient)
	tools := upstream.getTools(mockClient)

	ctx := context.Background()

	t.Run("query_vectors", func(t *testing.T) {
		// Find query_vectors tool
		var queryTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "query_vectors" {
				queryTool = &tool
				break
			}
		}
		assert.NotNil(t, queryTool)

		// Mock response
		expectedResult := map[string]interface{}{
			"matches": []map[string]interface{}{
				{"id": "vec1", "score": 0.9},
			},
		}

		vector := []float32{0.1, 0.2, 0.3}
		mockClient.On("Query", ctx, vector, int64(5), map[string]interface{}(nil), "").Return(expectedResult, nil)

		// Call handler
		args := map[string]interface{}{
			"vector": []interface{}{0.1, 0.2, 0.3},
			"top_k":  5.0,
		}
		result, err := queryTool.Handler(ctx, args)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("upsert_vectors", func(t *testing.T) {
		// Find upsert_vectors tool
		var upsertTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "upsert_vectors" {
				upsertTool = &tool
				break
			}
		}
		assert.NotNil(t, upsertTool)

		// Mock response
		expectedResult := map[string]interface{}{
			"upserted_count": 1,
		}

		vectors := []map[string]interface{}{
			{"id": "vec1", "values": []interface{}{0.1, 0.2, 0.3}},
		}
		mockClient.On("Upsert", ctx, vectors, "ns1").Return(expectedResult, nil)

		// Call handler
		args := map[string]interface{}{
			"vectors": []interface{}{
				map[string]interface{}{"id": "vec1", "values": []interface{}{0.1, 0.2, 0.3}},
			},
			"namespace": "ns1",
		}
		result, err := upsertTool.Handler(ctx, args)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("delete_vectors", func(t *testing.T) {
		// Find delete_vectors tool
		var deleteTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "delete_vectors" {
				deleteTool = &tool
				break
			}
		}
		assert.NotNil(t, deleteTool)

		// Mock response
		expectedResult := map[string]interface{}{
			"success": true,
		}

		ids := []string{"vec1", "vec2"}
		mockClient.On("Delete", ctx, ids, "ns1", map[string]interface{}(nil)).Return(expectedResult, nil)

		// Call handler
		args := map[string]interface{}{
			"ids":       []interface{}{"vec1", "vec2"},
			"namespace": "ns1",
		}
		result, err := deleteTool.Handler(ctx, args)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("describe_index_stats", func(t *testing.T) {
		// Find describe_index_stats tool
		var statsTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "describe_index_stats" {
				statsTool = &tool
				break
			}
		}
		assert.NotNil(t, statsTool)

		// Mock response
		expectedResult := map[string]interface{}{
			"totalVectorCount": int64(100),
		}

		mockClient.On("DescribeIndexStats", ctx, map[string]interface{}(nil)).Return(expectedResult, nil)

		// Call handler
		args := map[string]interface{}{}
		result, err := statsTool.Handler(ctx, args)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("query_vectors_error_invalid_vector", func(t *testing.T) {
		var queryTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "query_vectors" {
				queryTool = &tool
				break
			}
		}
		assert.NotNil(t, queryTool)

		// Invalid vector type (not array)
		args := map[string]interface{}{
			"vector": "invalid",
		}
		_, err := queryTool.Handler(ctx, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vector is required and must be an array")

		// Invalid vector elements
		args = map[string]interface{}{
			"vector": []interface{}{"not-a-number"},
		}
		_, err = queryTool.Handler(ctx, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vector elements must be numbers")
	})

	t.Run("upsert_vectors_error_invalid_input", func(t *testing.T) {
		var upsertTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "upsert_vectors" {
				upsertTool = &tool
				break
			}
		}
		assert.NotNil(t, upsertTool)

		// Missing vectors
		args := map[string]interface{}{}
		_, err := upsertTool.Handler(ctx, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vectors is required")

		// Invalid vectors format
		args = map[string]interface{}{
			"vectors": []interface{}{"not-a-map"},
		}
		_, err = upsertTool.Handler(ctx, args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vectors must be objects")
	})
}

// MockToolManager is a simple mock for tool.ManagerInterface
type MockToolManager struct {
	tool.ManagerInterface
}

func (m *MockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {}
func (m *MockToolManager) AddTool(t tool.Tool) error { return nil }

func TestRegister(t *testing.T) {
	// This test mainly verifies that Register logic runs without error and calls ToolManager
	// However, we can't easily inject the mockClient into Register because it creates it internally.
	// So we can only test the error path or success path if we mock NewPineconeClient or integration test it.
	// Since we can't mock package level functions easily in Go without a variable, we'll skip detailed Register test
	// relying on unit tests of getTools and PineconeClient.

	// But we can verify it fails if config is missing
	u := NewUpstream()
	name := "test-vector"
	cfg := &configv1.UpstreamServiceConfig{
		Name: &name,
	}
	// Missing VectorService config
	_, _, _, err := u.Register(context.Background(), cfg, &MockToolManager{}, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vector service config is nil")
}
