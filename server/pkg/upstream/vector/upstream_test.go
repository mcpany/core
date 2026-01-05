// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"context"
	"errors"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
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

		// Test arg validation
		_, err = queryTool.Handler(ctx, map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vector is required")

		_, err = queryTool.Handler(ctx, map[string]interface{}{"vector": []interface{}{"not-a-number"}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vector elements must be numbers")
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

		// Validation
		_, err = upsertTool.Handler(ctx, map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vectors is required")

		_, err = upsertTool.Handler(ctx, map[string]interface{}{"vectors": []interface{}{"not-object"}})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vectors must be objects")
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
		mockClient.On("Delete", ctx, []string{"id1"}, "ns1", map[string]interface{}(nil)).Return(expectedResult, nil)

		args := map[string]interface{}{
			"ids":       []interface{}{"id1"},
			"namespace": "ns1",
		}
		result, err := deleteTool.Handler(ctx, args)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("describe_index_stats", func(t *testing.T) {
		var statsTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "describe_index_stats" {
				statsTool = &tool
				break
			}
		}
		assert.NotNil(t, statsTool)

		expectedResult := map[string]interface{}{"totalVectorCount": 100}
		mockClient.On("DescribeIndexStats", ctx, map[string]interface{}{"foo": "bar"}).Return(expectedResult, nil)

		args := map[string]interface{}{
			"filter": map[string]interface{}{"foo": "bar"},
		}
		result, err := statsTool.Handler(ctx, args)
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
func (m *MockToolManager) AddTool(t tool.Tool) error {
	args := m.Called(t)
	return args.Error(0)
}

func TestRegister(t *testing.T) {
	// Setup
	mockClient := new(MockVectorClient)
	mockToolManager := new(MockToolManager)

	u := &Upstream{
		clientFactory: func(config *configv1.VectorUpstreamService) (Client, error) {
			return mockClient, nil
		},
	}

	name := "test-vector"
	cfg := &configv1.UpstreamServiceConfig{
		Name: &name,
		ServiceConfig: &configv1.UpstreamServiceConfig_VectorService{
			VectorService: &configv1.VectorUpstreamService{
				VectorDbType: &configv1.VectorUpstreamService_Pinecone{
					Pinecone: &configv1.PineconeVectorDB{
						ApiKey:      proto.String("key"),
						Environment: proto.String("env"),
						ProjectId:   proto.String("proj"),
						IndexName:   proto.String("idx"),
					},
				},
			},
		},
	}

	// Mock expectations
	mockToolManager.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
	mockToolManager.On("AddTool", mock.Anything).Return(nil)

	// Execute
	id, tools, _, err := u.Register(context.Background(), cfg, mockToolManager, nil, nil, false)

	// Verify
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.NotEmpty(t, tools)
	assert.Equal(t, 4, len(tools)) // 4 tools defined

	mockToolManager.AssertExpectations(t)
}

func TestRegister_Errors(t *testing.T) {
	t.Run("MissingConfig", func(t *testing.T) {
		u := NewUpstream()
		name := "test-vector"
		cfg := &configv1.UpstreamServiceConfig{
			Name: &name,
		}
		// Missing VectorService config
		_, _, _, err := u.Register(context.Background(), cfg, &MockToolManager{}, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "vector service config is nil")
	})

	t.Run("ClientError", func(t *testing.T) {
		u := &Upstream{
			clientFactory: func(config *configv1.VectorUpstreamService) (Client, error) {
				return nil, errors.New("client creation failed")
			},
		}
		name := "test-vector"
		cfg := &configv1.UpstreamServiceConfig{
			Name: &name,
			ServiceConfig: &configv1.UpstreamServiceConfig_VectorService{
				VectorService: &configv1.VectorUpstreamService{
					VectorDbType: &configv1.VectorUpstreamService_Pinecone{
						Pinecone: &configv1.PineconeVectorDB{},
					},
				},
			},
		}

		_, _, _, err := u.Register(context.Background(), cfg, &MockToolManager{}, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client creation failed")
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		// Use default factory to test switch default case
		u := NewUpstream()
		name := "test-vector"
		cfg := &configv1.UpstreamServiceConfig{
			Name: &name,
			ServiceConfig: &configv1.UpstreamServiceConfig_VectorService{
				VectorService: &configv1.VectorUpstreamService{
					// No type set
				},
			},
		}

		_, _, _, err := u.Register(context.Background(), cfg, &MockToolManager{}, nil, nil, false)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported vector database type")
	})
}

func TestVectorCallable_Call(t *testing.T) {
	handler := func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": "ok"}, nil
	}
	vc := &vectorCallable{handler: handler}

	req := &tool.ExecutionRequest{
		Arguments: map[string]interface{}{"a": 1},
	}
	res, err := vc.Call(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"result": "ok"}, res)
}

func TestShutdown(t *testing.T) {
	u := NewUpstream()
	assert.NoError(t, u.Shutdown(context.Background()))
}

// Integration test for helper method (optional, but good for coverage)
func TestDefaultClientFactory(t *testing.T) {
	// This will try to create a real client, but validation will fail or it will succeed struct creation
	// Since NewPineconeClient just returns a struct without making network calls, it should succeed
	cfg := &configv1.VectorUpstreamService{
		VectorDbType: &configv1.VectorUpstreamService_Pinecone{
			Pinecone: &configv1.PineconeVectorDB{
				ApiKey:      proto.String("key"),
				Environment: proto.String("env"),
				ProjectId:   proto.String("proj"),
				IndexName:   proto.String("idx"),
			},
		},
	}
	client, err := defaultClientFactory(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestGetTools_ExtraCoverage(t *testing.T) {
	// Add coverage for specific error paths in handlers if any
	// Already covered most in TestVectorTools
}
