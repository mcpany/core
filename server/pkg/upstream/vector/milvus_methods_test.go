// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"context"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestMilvusClient_Query(t *testing.T) {
	mock := &mockMilvusClient{}
	config := configv1.MilvusVectorDB_builder{
		Address:        proto.String("127.0.0.1:19530"),
		CollectionName: proto.String("test_coll"),
	}.Build()
	c := &MilvusClient{
		config: config,
		client: mock,
	}

	mock.loadCollectionFunc = func(ctx context.Context, name string, async bool, opts ...client.LoadCollectionOption) error {
		return nil
	}
	mock.describeCollectionFunc = func(ctx context.Context, name string) (*entity.Collection, error) {
		return &entity.Collection{
			Schema: &entity.Schema{
				Fields: []*entity.Field{
					{Name: "id", DataType: entity.FieldTypeInt64},
					{Name: "embedding", DataType: entity.FieldTypeFloatVector},
					{Name: "meta", DataType: entity.FieldTypeVarChar},
				},
			},
		}, nil
	}
	mock.searchFunc = func(ctx context.Context, collectionName string, partitions []string, expr string, outputFields []string, vectors []entity.Vector, vectorField string, metricType entity.MetricType, topK int, sp entity.SearchParam, opts ...client.SearchQueryOptionFunc) ([]client.SearchResult, error) {
		assert.Equal(t, "test_coll", collectionName)
		assert.Equal(t, "embedding", vectorField)
		assert.Equal(t, 2, topK)

		// Create result
		ids := entity.NewColumnInt64("id", []int64{1, 2})

		res := client.SearchResult{
			ResultCount: 2,
			IDs:         ids,
			Scores:      []float32{0.9, 0.8},
		}

		return []client.SearchResult{res}, nil
	}

	ctx := context.Background()
	vector := []float32{0.1, 0.2}

	res, err := c.Query(ctx, vector, 2, nil, "")
	require.NoError(t, err)

	matches, ok := res["matches"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, matches, 2)
	assert.Equal(t, int64(1), matches[0]["id"])
	assert.Equal(t, float32(0.9), matches[0]["score"])
}

func TestMilvusClient_Query_Errors(t *testing.T) {
	ctx := context.Background()
	config := configv1.MilvusVectorDB_builder{
		Address:        proto.String("localhost:19530"),
		CollectionName: proto.String("test_coll"),
	}.Build()

	t.Run("LoadCollection_Fail", func(t *testing.T) {
		mock := &mockMilvusClient{}
		c := &MilvusClient{config: config, client: mock}
		mock.loadCollectionFunc = func(ctx context.Context, name string, async bool, opts ...client.LoadCollectionOption) error {
			return errors.New("load failed")
		}
		_, err := c.Query(ctx, []float32{0.1}, 1, nil, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "load failed")
	})

	t.Run("DescribeCollection_Fail", func(t *testing.T) {
		mock := &mockMilvusClient{}
		c := &MilvusClient{config: config, client: mock}
		mock.loadCollectionFunc = func(ctx context.Context, name string, async bool, opts ...client.LoadCollectionOption) error {
			return nil
		}
		mock.describeCollectionFunc = func(ctx context.Context, name string) (*entity.Collection, error) {
			return nil, errors.New("describe failed")
		}
		_, err := c.Query(ctx, []float32{0.1}, 1, nil, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "describe failed")
	})

	t.Run("Search_Fail", func(t *testing.T) {
		mock := &mockMilvusClient{}
		c := &MilvusClient{config: config, client: mock}
		mock.loadCollectionFunc = func(ctx context.Context, name string, async bool, opts ...client.LoadCollectionOption) error { return nil }
		mock.describeCollectionFunc = func(ctx context.Context, name string) (*entity.Collection, error) {
			return &entity.Collection{
				Schema: &entity.Schema{
					Fields: []*entity.Field{
						{Name: "embedding", DataType: entity.FieldTypeFloatVector},
					},
				},
			}, nil
		}
		mock.searchFunc = func(ctx context.Context, collectionName string, partitions []string, expr string, outputFields []string, vectors []entity.Vector, vectorField string, metricType entity.MetricType, topK int, sp entity.SearchParam, opts ...client.SearchQueryOptionFunc) ([]client.SearchResult, error) {
			return nil, errors.New("search failed")
		}
		_, err := c.Query(ctx, []float32{0.1}, 1, nil, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "search failed")
	})

	t.Run("No_Vector_Field", func(t *testing.T) {
		mock := &mockMilvusClient{}
		c := &MilvusClient{config: config, client: mock}
		mock.loadCollectionFunc = func(ctx context.Context, name string, async bool, opts ...client.LoadCollectionOption) error { return nil }
		mock.describeCollectionFunc = func(ctx context.Context, name string) (*entity.Collection, error) {
			return &entity.Collection{
				Schema: &entity.Schema{
					Fields: []*entity.Field{
						{Name: "id", DataType: entity.FieldTypeInt64},
					},
				},
			}, nil
		}
		_, err := c.Query(ctx, []float32{0.1}, 1, nil, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no float vector field found")
	})
}

func TestMilvusClient_Upsert(t *testing.T) {
	mock := &mockMilvusClient{}
	config := configv1.MilvusVectorDB_builder{
		Address:        proto.String("127.0.0.1:19530"),
		CollectionName: proto.String("test_coll"),
	}.Build()
	c := &MilvusClient{
		config: config,
		client: mock,
	}

	mock.describeCollectionFunc = func(ctx context.Context, name string) (*entity.Collection, error) {
		return &entity.Collection{
			Schema: &entity.Schema{
				Fields: []*entity.Field{
					{Name: "id", DataType: entity.FieldTypeInt64, PrimaryKey: true},
					{Name: "embedding", DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": "2"}},
				},
			},
		}, nil
	}
	mock.upsertFunc = func(ctx context.Context, collectionName string, partitionName string, columns ...entity.Column) (entity.Column, error) {
		assert.Equal(t, "test_coll", collectionName)
		assert.Len(t, columns, 2) // id and embedding
		return nil, nil // Return is ignored by milvus.go logic which just checks error
	}

	ctx := context.Background()
	vectors := []map[string]interface{}{
		{
			"id": int64(1),
			"values": []interface{}{0.1, 0.2},
		},
	}

	res, err := c.Upsert(ctx, vectors, "")
	require.NoError(t, err)
	assert.Equal(t, int64(1), res["upserted_count"])
}

func TestMilvusClient_Upsert_Errors(t *testing.T) {
	ctx := context.Background()
	config := configv1.MilvusVectorDB_builder{
		Address:        proto.String("localhost:19530"),
		CollectionName: proto.String("test_coll"),
	}.Build()

	t.Run("DescribeCollection_Fail", func(t *testing.T) {
		mock := &mockMilvusClient{}
		c := &MilvusClient{config: config, client: mock}
		mock.describeCollectionFunc = func(ctx context.Context, name string) (*entity.Collection, error) {
			return nil, errors.New("describe failed")
		}
		_, err := c.Upsert(ctx, []map[string]interface{}{{"id": 1}}, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "describe failed")
	})

	t.Run("Conversion_Fail", func(t *testing.T) {
		mock := &mockMilvusClient{}
		c := &MilvusClient{config: config, client: mock}
		mock.describeCollectionFunc = func(ctx context.Context, name string) (*entity.Collection, error) {
			return &entity.Collection{
				Schema: &entity.Schema{
					Fields: []*entity.Field{
						{Name: "embedding", DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": "invalid"}},
					},
				},
			}, nil
		}
		_, err := c.Upsert(ctx, []map[string]interface{}{{"values": []interface{}{0.1}}}, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse dimension")
	})

	t.Run("Upsert_Fail", func(t *testing.T) {
		mock := &mockMilvusClient{}
		c := &MilvusClient{config: config, client: mock}
		mock.describeCollectionFunc = func(ctx context.Context, name string) (*entity.Collection, error) {
			return &entity.Collection{
				Schema: &entity.Schema{
					Fields: []*entity.Field{
						{Name: "id", DataType: entity.FieldTypeInt64, PrimaryKey: true},
					},
				},
			}, nil
		}
		mock.upsertFunc = func(ctx context.Context, collectionName string, partitionName string, columns ...entity.Column) (entity.Column, error) {
			return nil, errors.New("upsert failed")
		}
		_, err := c.Upsert(ctx, []map[string]interface{}{{"id": 1}}, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "upsert failed")
	})
}

func TestMilvusClient_Delete(t *testing.T) {
	mock := &mockMilvusClient{}
	config := configv1.MilvusVectorDB_builder{
		Address:        proto.String("127.0.0.1:19530"),
		CollectionName: proto.String("test_coll"),
	}.Build()
	c := &MilvusClient{
		config: config,
		client: mock,
	}

	mock.describeCollectionFunc = func(ctx context.Context, name string) (*entity.Collection, error) {
		return &entity.Collection{
			Schema: &entity.Schema{
				Fields: []*entity.Field{
					{Name: "id", DataType: entity.FieldTypeInt64, PrimaryKey: true},
				},
			},
		}, nil
	}
	mock.deleteFunc = func(ctx context.Context, collectionName string, partitionName string, expr string) error {
		assert.Equal(t, "test_coll", collectionName)
		assert.Equal(t, "id in [1, 2]", expr)
		return nil
	}

	ctx := context.Background()
	ids := []string{"1", "2"}

	res, err := c.Delete(ctx, ids, "", nil)
	require.NoError(t, err)
	assert.Equal(t, true, res["success"])
}

func TestMilvusClient_Delete_Errors(t *testing.T) {
	ctx := context.Background()
	config := configv1.MilvusVectorDB_builder{
		Address:        proto.String("localhost:19530"),
		CollectionName: proto.String("test_coll"),
	}.Build()

	t.Run("DescribeCollection_Fail", func(t *testing.T) {
		mock := &mockMilvusClient{}
		c := &MilvusClient{config: config, client: mock}
		mock.describeCollectionFunc = func(ctx context.Context, name string) (*entity.Collection, error) {
			return nil, errors.New("describe failed")
		}
		_, err := c.Delete(ctx, []string{"1"}, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "describe failed")
	})

	t.Run("PK_Not_Found", func(t *testing.T) {
		mock := &mockMilvusClient{}
		c := &MilvusClient{config: config, client: mock}
		mock.describeCollectionFunc = func(ctx context.Context, name string) (*entity.Collection, error) {
			return &entity.Collection{
				Schema: &entity.Schema{Fields: []*entity.Field{}},
			}, nil
		}
		_, err := c.Delete(ctx, []string{"1"}, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PK field not found")
	})

	t.Run("Delete_Fail", func(t *testing.T) {
		mock := &mockMilvusClient{}
		c := &MilvusClient{config: config, client: mock}
		mock.describeCollectionFunc = func(ctx context.Context, name string) (*entity.Collection, error) {
			return &entity.Collection{
				Schema: &entity.Schema{
					Fields: []*entity.Field{
						{Name: "id", DataType: entity.FieldTypeInt64, PrimaryKey: true},
					},
				},
			}, nil
		}
		mock.deleteFunc = func(ctx context.Context, collectionName string, partitionName string, expr string) error {
			return errors.New("delete failed")
		}
		_, err := c.Delete(ctx, []string{"1"}, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
	})

	t.Run("Missing_Args", func(t *testing.T) {
		mock := &mockMilvusClient{}
		c := &MilvusClient{config: config, client: mock}
		_, err := c.Delete(ctx, nil, "", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must provide ids or filter")
	})
}

func TestMilvusClient_DescribeIndexStats(t *testing.T) {
	mock := &mockMilvusClient{}
	config := configv1.MilvusVectorDB_builder{
		Address:        proto.String("127.0.0.1:19530"),
		CollectionName: proto.String("test_coll"),
	}.Build()
	c := &MilvusClient{
		config: config,
		client: mock,
	}

	mock.describeCollectionFunc = func(ctx context.Context, name string) (*entity.Collection, error) {
		return &entity.Collection{Name: "test_coll"}, nil
	}
	mock.getCollectionStatisticsFunc = func(ctx context.Context, name string) (map[string]string, error) {
		return map[string]string{"row_count": "100"}, nil
	}

	ctx := context.Background()
	res, err := c.DescribeIndexStats(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, "test_coll", res["name"])
	stats, ok := res["stats"].(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "100", stats["row_count"])
}
