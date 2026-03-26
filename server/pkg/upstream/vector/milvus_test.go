// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// TestNewMilvusClient tests the creation of a new Milvus client.
// Note: We cannot easily mock the Milvus SDK client interface without generating a mock for it,
// and it's an external library. For unit tests, we mainly check configuration validation.
// E2E tests would require a running Milvus instance.
func TestNewMilvusClient(t *testing.T) {
	tests := []struct {
		name        string
		config      *configv1.MilvusVectorDB
		expectError bool
	}{
		{
			name: "Valid Config",
			config: configv1.MilvusVectorDB_builder{
				Address:        proto.String("127.0.0.1:19530"),
				CollectionName: proto.String("test_collection"),
			}.Build(),
			expectError: true, // Will fail connection in NewMilvusClient
		},
		{
			name: "Missing Address",
			config: configv1.MilvusVectorDB_builder{
				CollectionName: proto.String("test_collection"),
			}.Build(),
			expectError: true,
		},
		{
			name: "Missing Collection",
			config: configv1.MilvusVectorDB_builder{
				Address: proto.String("127.0.0.1:19530"),
			}.Build(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMilvusClient(tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConvertVectorsToColumns(t *testing.T) {
	schema := &entity.Schema{
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
			},
			{
				Name:     "embedding",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": "4",
				},
			},
			{
				Name:     "category",
				DataType: entity.FieldTypeVarChar,
			},
		},
	}

	vectors := []map[string]interface{}{
		{
			"id":     "vec1",
			"values": []interface{}{0.1, 0.2, 0.3, 0.4},
			"metadata": map[string]interface{}{
				"category": "news",
			},
		},
		{
			"id":     "vec2",
			"values": []interface{}{0.5, 0.6, 0.7, 0.8},
			"metadata": map[string]interface{}{
				"category": "sports",
			},
		},
	}

	columns, err := convertVectorsToColumns(vectors, schema)
	assert.NoError(t, err)
	assert.Len(t, columns, 3)

	colMap := make(map[string]entity.Column)
	for _, col := range columns {
		colMap[col.Name()] = col
	}

	assert.Contains(t, colMap, "id")
	assert.Contains(t, colMap, "embedding")
	assert.Contains(t, colMap, "category")

	// Check ID column
	idCol := colMap["id"].(*entity.ColumnVarChar)
	assert.Equal(t, "vec1", idCol.Data()[0])
	assert.Equal(t, "vec2", idCol.Data()[1])

	// Check Embedding column
	embCol := colMap["embedding"].(*entity.ColumnFloatVector)
	assert.Equal(t, []float32{0.1, 0.2, 0.3, 0.4}, embCol.Data()[0])
	assert.Equal(t, []float32{0.5, 0.6, 0.7, 0.8}, embCol.Data()[1])

	// Check Category column
	catCol := colMap["category"].(*entity.ColumnVarChar)
	assert.Equal(t, "news", catCol.Data()[0])
	assert.Equal(t, "sports", catCol.Data()[1])
}

func TestConvertVectorsToColumns_Int64PK_MixedMetadata(t *testing.T) {
	schema := &entity.Schema{
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
			},
			{
				Name:     "embedding",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": "2",
				},
			},
			{
				Name:     "count",
				DataType: entity.FieldTypeInt64,
			},
			{
				Name:     "score",
				DataType: entity.FieldTypeDouble,
			},
			{
				Name:     "active",
				DataType: entity.FieldTypeBool,
			},
		},
	}

	vectors := []map[string]interface{}{
		{
			"id":     101, // Int
			"values": []interface{}{0.1, 0.2},
			"metadata": map[string]interface{}{
				"count":  10,
				"score":  0.95,
				"active": true,
			},
		},
		{
			"id":     "102", // String representing int
			"values": []interface{}{0.3, 0.4},
			"metadata": map[string]interface{}{
				"count":  "20", // String representing int
				"score":  0.88,
				"active": false,
			},
		},
	}

	columns, err := convertVectorsToColumns(vectors, schema)
	assert.NoError(t, err)
	assert.Len(t, columns, 5)

	colMap := make(map[string]entity.Column)
	for _, col := range columns {
		colMap[col.Name()] = col
	}

	// Check ID column (Int64)
	idCol := colMap["id"].(*entity.ColumnInt64)
	assert.Equal(t, int64(101), idCol.Data()[0])
	assert.Equal(t, int64(102), idCol.Data()[1])

	// Check Count (Int64)
	countCol := colMap["count"].(*entity.ColumnInt64)
	assert.Equal(t, int64(10), countCol.Data()[0])
	assert.Equal(t, int64(20), countCol.Data()[1])

	// Check Score (Double)
	scoreCol := colMap["score"].(*entity.ColumnDouble)
	assert.InDelta(t, 0.95, scoreCol.Data()[0], 0.0001)
	assert.InDelta(t, 0.88, scoreCol.Data()[1], 0.0001)

	// Check Active (Bool)
	activeCol := colMap["active"].(*entity.ColumnBool)
	assert.Equal(t, true, activeCol.Data()[0])
	assert.Equal(t, false, activeCol.Data()[1])
}
