// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
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

		// Test Error Cases
		// Missing vector
		_, err = queryTool.Handler(ctx, map[string]interface{}{})
		assert.Error(t, err)
		assert.Equal(t, "vector is required and must be an array", err.Error())

		// Invalid vector type
		_, err = queryTool.Handler(ctx, map[string]interface{}{"vector": "invalid"})
		assert.Error(t, err)

		// Invalid vector elements
		_, err = queryTool.Handler(ctx, map[string]interface{}{"vector": []interface{}{"invalid"}})
		assert.Error(t, err)
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

		// Test Error Cases
		// Missing vectors
		_, err = upsertTool.Handler(ctx, map[string]interface{}{})
		assert.Error(t, err)
		assert.Equal(t, "vectors is required", err.Error())

		// Invalid vectors type
		_, err = upsertTool.Handler(ctx, map[string]interface{}{"vectors": "invalid"})
		assert.Error(t, err)

		// Invalid vector elements
		_, err = upsertTool.Handler(ctx, map[string]interface{}{"vectors": []interface{}{"invalid"}})
		assert.Error(t, err)
	})

	t.Run("delete_vectors", func(t *testing.T) {
		var deleteTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "delete_vectors" {
				deleteTool = &tool
				break
			}
		}
		assert.NotNil(t, deleteTool)

		expectedResult := map[string]interface{}{"success": true}
		mockClient.On("Delete", ctx, []string{"id1", "id2"}, "ns1", map[string]interface{}(nil)).Return(expectedResult, nil)

		args := map[string]interface{}{
			"ids":       []interface{}{"id1", "id2"},
			"namespace": "ns1",
		}
		result, err := deleteTool.Handler(ctx, args)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)

		// Test with filter
		mockClient.On("Delete", ctx, []string(nil), "ns1", map[string]interface{}{"key": "val"}).Return(expectedResult, nil)
		argsWithFilter := map[string]interface{}{
			"namespace": "ns1",
			"filter":    map[string]interface{}{"key": "val"},
		}
		result2, err2 := deleteTool.Handler(ctx, argsWithFilter)
		assert.NoError(t, err2)
		assert.Equal(t, expectedResult, result2)
	})

	t.Run("describe_index_stats", func(t *testing.T) {
		var describeTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "describe_index_stats" {
				describeTool = &tool
				break
			}
		}
		assert.NotNil(t, describeTool)

		expectedResult := map[string]interface{}{
			"dimension":        1536,
			"totalVectorCount": 100,
		}
		mockClient.On("DescribeIndexStats", ctx, map[string]interface{}(nil)).Return(expectedResult, nil)

		args := map[string]interface{}{}
		result, err := describeTool.Handler(ctx, args)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})
}

// MockToolManager is a simple mock for tool.ManagerInterface
type MockToolManager struct {
	tool.ManagerInterface
	mock.Mock
}

func (m *MockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	m.Called(serviceID, info)
}

func (m *MockToolManager) GetToolCount(serviceID string) int {
	return 0
}

func (m *MockToolManager) AddTool(t tool.Tool) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockToolManager) ToolMatchesProfile(tool tool.Tool, profileID string) bool {
	return true
}

func TestRegister(t *testing.T) {
	// Test failure when config is nil
	u := NewUpstream()
	name := "test-vector"
	cfg := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(name),
	}.Build()
	// Missing VectorService config (assuming GetVectorService returns nil if not set or we set it to nil)
	_, _, _, err := u.Register(context.Background(), cfg, &MockToolManager{}, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "vector service config is nil")

	// Test failure with unsupported type (using default factory)
	cfgWithService := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(name),
		VectorService: configv1.VectorUpstreamService_builder{}.Build(),
	}.Build()
	_, _, _, err = u.Register(context.Background(), cfgWithService, &MockToolManager{}, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported vector database type")

	// Test success using injected mock client factory
	uRefactored := &Upstream{}
	mockClient := new(MockVectorClient)

	uRefactored.clientFactory = func(config *configv1.VectorUpstreamService) (Client, error) {
		return mockClient, nil
	}

	// Setup mock tool manager
	mockToolManager := new(MockToolManager)
	mockToolManager.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
	mockToolManager.On("AddTool", mock.Anything).Return(nil)

	// Configure valid service config
	validName := "valid-vector-service"
	validCfg := configv1.UpstreamServiceConfig_builder{
		Name: proto.String(validName),
		VectorService: configv1.VectorUpstreamService_builder{
			Pinecone: configv1.PineconeVectorDB_builder{
				ApiKey:    proto.String("key"),
				IndexName: proto.String("index"),
			}.Build(),
		}.Build(),
	}.Build()

	serviceID, tools, _, err := uRefactored.Register(context.Background(), validCfg, mockToolManager, nil, nil, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, serviceID)
	assert.Len(t, tools, 4) // 4 tools defined
	mockToolManager.AssertExpectations(t)

	// Test client creation failure
	uRefactored.clientFactory = func(config *configv1.VectorUpstreamService) (Client, error) {
		return nil, errors.New("creation failed")
	}
	_, _, _, err = uRefactored.Register(context.Background(), validCfg, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creation failed")
}

func TestVectorCallable(t *testing.T) {
	handler := func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
		return args, nil
	}
	c := &vectorCallable{handler: handler}
	req := &tool.ExecutionRequest{
		Arguments: map[string]interface{}{"key": "value"},
	}
	res, err := c.Call(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, req.Arguments, res)
}

func (m *MockToolManager) GetAllowedServiceIDs(profileID string) (map[string]bool, bool) {
	args := m.Called(profileID)
	return args.Get(0).(map[string]bool), args.Bool(1)
}
