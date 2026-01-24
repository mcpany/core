// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewUpstream(t *testing.T) {
	u := NewUpstream()
	assert.NotNil(t, u)
}

func TestUpstream_Shutdown(t *testing.T) {
	u := NewUpstream()
	err := u.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestUpstream_Register_AddToolError(t *testing.T) {
	uRefactored := &Upstream{}
	mockClient := new(MockVectorClient)

	uRefactored.clientFactory = func(config *configv1.VectorUpstreamService) (Client, error) {
		return mockClient, nil
	}

	// Setup mock tool manager to fail on AddTool
	mockToolManager := new(MockToolManager)
	mockToolManager.On("AddServiceInfo", mock.Anything, mock.Anything).Return()
	// Fail for the first tool, succeed for others (or fail for all, here we simulate mixed)
	mockToolManager.On("AddTool", mock.Anything).Return(errors.New("failed to add tool"))

	// Configure valid service config
	validName := "valid-vector-service"
	validCfg := &configv1.UpstreamServiceConfig{
		Name: &validName,
		ServiceConfig: &configv1.UpstreamServiceConfig_VectorService{
			VectorService: &configv1.VectorUpstreamService{
				VectorDbType: &configv1.VectorUpstreamService_Milvus{},
			},
		},
	}

	serviceID, tools, _, err := uRefactored.Register(context.Background(), validCfg, mockToolManager, nil, nil, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, serviceID)
	// Tools should be empty or partial because AddTool failed
	// The implementation continues on error, but does NOT append to discoveredTools if AddTool fails?
	// Let's check the code:
	// if err := toolManager.AddTool(callableTool); err != nil {
	// 	log.Error("Failed to add tool", "tool", toolName, "error", err)
	// 	continue
	// }
	// discoveredTools = append(discoveredTools, toolDef)
	// So if AddTool fails, it is NOT appended.
	assert.Len(t, tools, 0)
}

func TestVectorTools_EdgeCases(t *testing.T) {
	upstream := &Upstream{}
	mockClient := new(MockVectorClient)
	tools := upstream.getTools(mockClient)

	ctx := context.Background()

	t.Run("query_vectors_defaults", func(t *testing.T) {
		var queryTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "query_vectors" {
				queryTool = &tool
				break
			}
		}
		assert.NotNil(t, queryTool)

		// Expect default topK=10
		expectedResult := map[string]interface{}{"matches": []map[string]interface{}{}}
		mockClient.On("Query", ctx, []float32{0.1}, int64(10), map[string]interface{}(nil), "").Return(expectedResult, nil)

		// Invalid top_k type (should default to 10)
		args := map[string]interface{}{
			"vector": []interface{}{0.1},
			"top_k":  "invalid",
		}
		result, err := queryTool.Handler(ctx, args)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("upsert_vectors_types", func(t *testing.T) {
		var upsertTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "upsert_vectors" {
				upsertTool = &tool
				break
			}
		}
		assert.NotNil(t, upsertTool)

		// Test vectors must be objects
		_, err := upsertTool.Handler(ctx, map[string]interface{}{
			"vectors": []interface{}{"not-an-object"},
		})
		assert.Error(t, err)
		assert.Equal(t, "vectors must be objects", err.Error())
	})

	t.Run("delete_vectors_edge_cases", func(t *testing.T) {
		var deleteTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "delete_vectors" {
				deleteTool = &tool
				break
			}
		}
		assert.NotNil(t, deleteTool)

		expectedResult := map[string]interface{}{"success": true}

		// Case: ids provided but not a list (should be treated as empty list/ignored if implementation casts safely)
		// The implementation:
		// if idsInterface, ok := args["ids"].([]interface{}); ok { ... }
		// So if not []interface{}, ids remains nil.
		mockClient.On("Delete", ctx, []string(nil), "", map[string]interface{}(nil)).Return(expectedResult, nil)

		args := map[string]interface{}{
			"ids": "not-a-list",
		}
		result, err := deleteTool.Handler(ctx, args)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)

		// Case: filter provided but not a map (should be ignored)
		// The implementation:
		// if f, ok := args["filter"].(map[string]interface{}); ok { filter = f }
		argsInvalidFilter := map[string]interface{}{
			"filter": "not-a-map",
		}
		result2, err2 := deleteTool.Handler(ctx, argsInvalidFilter)
		assert.NoError(t, err2)
		assert.Equal(t, expectedResult, result2)
	})

	t.Run("describe_index_stats_edge_cases", func(t *testing.T) {
		var describeTool *vectorToolDef
		for _, tool := range tools {
			if tool.Name == "describe_index_stats" {
				describeTool = &tool
				break
			}
		}
		assert.NotNil(t, describeTool)

		expectedResult := map[string]interface{}{"totalVectorCount": 0}
		// Expect nil filter if invalid type provided
		mockClient.On("DescribeIndexStats", ctx, map[string]interface{}(nil)).Return(expectedResult, nil)

		args := map[string]interface{}{
			"filter": "invalid-filter",
		}
		result, err := describeTool.Handler(ctx, args)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})
}
