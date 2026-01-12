// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/upstream/vector/mocks"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestMilvusClient_Upsert(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		config := &configv1.MilvusVectorDB{
			Address:        proto.String("localhost:19530"),
			CollectionName: proto.String("test_collection"),
		}
		mc := &MilvusClient{
			config: config,
			client: mockClient,
		}

		// Mock DescribeCollection
		schema := &entity.Schema{
			Fields: []*entity.Field{
				{Name: "id", DataType: entity.FieldTypeInt64, PrimaryKey: true},
				{Name: "vector", DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": "2"}},
			},
		}
		mockClient.On("DescribeCollection", ctx, "test_collection").Return(&entity.Collection{Schema: schema}, nil)

		// Mock Upsert
		// We expect 2 columns: id and vector.
		// Return nil for entity.Column return type.
		mockClient.On("Upsert", ctx, "test_collection", "", mock.Anything, mock.Anything).Return(nil, nil)

		vectors := []map[string]interface{}{
			{
				"id":     int64(1),
				"values": []interface{}{0.1, 0.2},
			},
		}

		res, err := mc.Upsert(ctx, vectors, "")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), res["upserted_count"])
	})

	t.Run("Empty Vectors", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		config := &configv1.MilvusVectorDB{
			Address:        proto.String("localhost:19530"),
			CollectionName: proto.String("test_collection"),
		}
		mc := &MilvusClient{
			config: config,
			client: mockClient,
		}

		res, err := mc.Upsert(ctx, nil, "")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}

func TestMilvusClient_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("Delete by IDs (Int64)", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		config := &configv1.MilvusVectorDB{
			Address:        proto.String("localhost:19530"),
			CollectionName: proto.String("test_collection"),
		}
		mc := &MilvusClient{
			config: config,
			client: mockClient,
		}

		schema := &entity.Schema{
			Fields: []*entity.Field{
				{Name: "id", DataType: entity.FieldTypeInt64, PrimaryKey: true},
			},
		}
		mockClient.On("DescribeCollection", ctx, "test_collection").Return(&entity.Collection{Schema: schema}, nil)
		mockClient.On("Delete", ctx, "test_collection", "", "id in [1, 2]").Return(nil)

		res, err := mc.Delete(ctx, []string{"1", "2"}, "", nil)
		assert.NoError(t, err)
		assert.Equal(t, true, res["success"])
	})

	t.Run("Delete by IDs (VarChar)", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		config := &configv1.MilvusVectorDB{
			Address:        proto.String("localhost:19530"),
			CollectionName: proto.String("test_collection"),
		}
		mc := &MilvusClient{
			config: config,
			client: mockClient,
		}

		schema := &entity.Schema{
			Fields: []*entity.Field{
				{Name: "id", DataType: entity.FieldTypeVarChar, PrimaryKey: true},
			},
		}
		mockClient.On("DescribeCollection", ctx, "test_collection").Return(&entity.Collection{Schema: schema}, nil)
		mockClient.On("Delete", ctx, "test_collection", "", "id in [\"a\", \"b\"]").Return(nil)

		res, err := mc.Delete(ctx, []string{"a", "b"}, "", nil)
		assert.NoError(t, err)
		assert.Equal(t, true, res["success"])
	})

	t.Run("Delete by Filter", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		config := &configv1.MilvusVectorDB{
			Address:        proto.String("localhost:19530"),
			CollectionName: proto.String("test_collection"),
		}
		mc := &MilvusClient{
			config: config,
			client: mockClient,
		}

		mockClient.On("Delete", ctx, "test_collection", "", "field == \"value\"").Return(nil)

		res, err := mc.Delete(ctx, nil, "", map[string]interface{}{"field": "value"})
		assert.NoError(t, err)
		assert.Equal(t, true, res["success"])
	})
}

func TestMilvusClient_Query(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		config := &configv1.MilvusVectorDB{
			Address:        proto.String("localhost:19530"),
			CollectionName: proto.String("test_collection"),
		}
		mc := &MilvusClient{
			config: config,
			client: mockClient,
		}

		mockClient.On("LoadCollection", ctx, "test_collection", false).Return(nil)

		schema := &entity.Schema{
			Fields: []*entity.Field{
				{Name: "id", DataType: entity.FieldTypeInt64, PrimaryKey: true},
				{Name: "vector", DataType: entity.FieldTypeFloatVector},
				{Name: "field1", DataType: entity.FieldTypeVarChar},
			},
		}
		mockClient.On("DescribeCollection", ctx, "test_collection").Return(&entity.Collection{Schema: schema}, nil)

		// Mock Search result
		searchResults := []client.SearchResult{
			{
				ResultCount: 1,
				IDs:         entity.NewColumnInt64("id", []int64{1}),
				Scores:      []float32{0.9},
				Fields:      client.ResultSet{},
			},
		}

		mockClient.On("Search", ctx, "test_collection", []string(nil), "",
			mock.MatchedBy(func(fields []string) bool {
				hasID := false
				hasField1 := false
				for _, f := range fields {
					if f == "id" {
						hasID = true
					}
					if f == "field1" {
						hasField1 = true
					}
				}
				return hasID && hasField1 && len(fields) == 2
			}),
			mock.Anything, "vector", entity.L2, 10, mock.Anything).
			Return(searchResults, nil)

		res, err := mc.Query(ctx, []float32{0.1, 0.2}, 10, nil, "")
		assert.NoError(t, err)
		matches := res["matches"].([]map[string]interface{})
		assert.Len(t, matches, 1)
		assert.Equal(t, int64(1), matches[0]["id"])
		assert.Equal(t, float32(0.9), matches[0]["score"])
	})
}

func TestMilvusClient_DescribeIndexStats(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		config := &configv1.MilvusVectorDB{
			Address:        proto.String("localhost:19530"),
			CollectionName: proto.String("test_collection"),
		}
		mc := &MilvusClient{
			config: config,
			client: mockClient,
		}

		mockClient.On("DescribeCollection", ctx, "test_collection").Return(&entity.Collection{Name: "test_collection"}, nil)
		stats := map[string]string{"row_count": "100"}
		mockClient.On("GetCollectionStatistics", ctx, "test_collection").Return(stats, nil)

		res, err := mc.DescribeIndexStats(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, "test_collection", res["name"])
		assert.Equal(t, stats, res["stats"])
	})
}
