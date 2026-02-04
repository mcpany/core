// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"context"
	"fmt"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// MilvusClient implements VectorClient for Milvus.
//
// Summary: Milvus client implementation.
type MilvusClient struct {
	config *configv1.MilvusVectorDB
	client client.Client
}

// NewMilvusClient creates a new Milvus client.
//
// Summary: Creates a new Milvus client.
//
// Parameters:
//   - config: *configv1.MilvusVectorDB. The configuration settings.
//
// Returns:
//   - *MilvusClient: A new Milvus client instance.
//   - error: An error if the operation fails.
func NewMilvusClient(config *configv1.MilvusVectorDB) (*MilvusClient, error) {
	if config.GetAddress() == "" {
		return nil, fmt.Errorf("address is required for Milvus")
	}
	if config.GetCollectionName() == "" {
		return nil, fmt.Errorf("collection_name is required for Milvus")
	}

	// Set a timeout for connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := client.NewClient(ctx, client.Config{
		Address:  config.GetAddress(),
		Username: config.GetUsername(),
		Password: config.GetPassword(),
		APIKey:   config.GetApiKey(),
		DBName:   config.GetDatabaseName(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create milvus client: %w", err)
	}

	// Check if collection exists
	exists, err := c.HasCollection(ctx, config.GetCollectionName())
	if err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("failed to check collection existence: %w", err)
	}
	if !exists {
		_ = c.Close()
		return nil, fmt.Errorf("collection %s does not exist", config.GetCollectionName())
	}

	return &MilvusClient{
		config: config,
		client: c,
	}, nil
}

// Query searches for similar vectors.
//
// Summary: Searches for similar vectors.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - vector: []float32. The query vector.
//   - topK: int64. The number of results to return.
//   - filter: map[string]interface{}. A metadata filter.
//   - namespace: string. The namespace to query.
//
// Returns:
//   - map[string]interface{}: The search results.
//   - error: An error if the operation fails.
func (c *MilvusClient) Query(ctx context.Context, vector []float32, topK int64, filter map[string]interface{}, namespace string) (map[string]interface{}, error) {
	// Milvus uses partitions as namespaces usually, or just metadata fields.
	// Assuming namespace maps to partition names if provided.
	var partitions []string
	if namespace != "" {
		partitions = append(partitions, namespace)
	}

	// Load collection is required before search
	if err := c.client.LoadCollection(ctx, c.config.GetCollectionName(), false); err != nil {
		return nil, fmt.Errorf("failed to load collection: %w", err)
	}

	// Construct expression from filter
	expr := ""
	if filter != nil {
		// Basic filter conversion.
		// Milvus filter syntax: "field > 0 && field < 100"
		// This is complex to convert generically from a map.
		// For now, we support simple equality checks.
		var parts []string
		for k, v := range filter {
			if s, ok := v.(string); ok {
				parts = append(parts, fmt.Sprintf("%s == \"%s\"", k, s))
			} else {
				parts = append(parts, fmt.Sprintf("%s == %v", k, v))
			}
		}
		expr = strings.Join(parts, " && ")
	}

	// Need to discover vector field name first or assume one.
	// For simplicity, we assume the user configured collection appropriately.
	// We'll search across all float vector fields? No, that's ambiguous.
	// We really need to know the vector field name.
	// We can DescribeCollection to find it.
	coll, err := c.client.DescribeCollection(ctx, c.config.GetCollectionName())
	if err != nil {
		return nil, fmt.Errorf("failed to describe collection: %w", err)
	}

	var vectorField string
	var outputFields []string
	for _, field := range coll.Schema.Fields {
		if field.DataType == entity.FieldTypeFloatVector {
			vectorField = field.Name
		} else {
			outputFields = append(outputFields, field.Name)
		}
	}
	if vectorField == "" {
		return nil, fmt.Errorf("no float vector field found in collection")
	}

	sp, _ := entity.NewIndexFlatSearchParam() // Simple search param, might need tuning

	vectors := []entity.Vector{
		entity.FloatVector(vector),
	}

	result, err := c.client.Search(
		ctx,
		c.config.GetCollectionName(),
		partitions,
		expr,
		outputFields,
		vectors,
		vectorField,
		entity.L2, // Default metric type, might need config
		int(topK),
		sp,
	)
	if err != nil {
		return nil, err
	}

	matches := make([]map[string]interface{}, 0)
	if len(result) > 0 {
		// Result is []SearchResult (one per query vector)
		// We only have 1 query vector
		res := result[0]
		for i := 0; i < res.ResultCount; i++ {
			id, err := res.IDs.Get(i)
			if err != nil {
				continue
			}
			match := map[string]interface{}{
				"id":    id,
				"score": res.Scores[i],
			}
			// Add metadata/fields
			metadata := make(map[string]interface{})
			for _, field := range outputFields {
				col := res.Fields.GetColumn(field)
				if col != nil {
					// This is tricky without knowing type, but sdk helps
					if v, err := col.Get(i); err == nil {
						metadata[field] = v
					}
				}
			}
			match["metadata"] = metadata
			matches = append(matches, match)
		}
	}

	return map[string]interface{}{
		"matches": matches,
	}, nil
}

// Upsert inserts or updates vectors.
//
// Summary: Inserts or updates vectors.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - vectors: []map[string]interface{}. The list of vectors to upsert.
//   - namespace: string. The namespace to upsert into.
//
// Returns:
//   - map[string]interface{}: The operation result.
//   - error: An error if the operation fails.
func (c *MilvusClient) Upsert(ctx context.Context, vectors []map[string]interface{}, namespace string) (map[string]interface{}, error) {
	// Milvus Upsert (v2.3+)
	if len(vectors) == 0 {
		return nil, nil
	}

	coll, err := c.client.DescribeCollection(ctx, c.config.GetCollectionName())
	if err != nil {
		return nil, fmt.Errorf("failed to describe collection: %w", err)
	}

	columnList, err := convertVectorsToColumns(vectors, coll.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to convert vectors to columns: %w", err)
	}

	partitionName := ""
	if namespace != "" {
		partitionName = namespace
	}

	_, err = c.client.Upsert(ctx, c.config.GetCollectionName(), partitionName, columnList...)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"upserted_count": int64(len(vectors)),
	}, nil
}

// convertVectorsToColumns converts a list of row-based vectors to Milvus column-based format.
// This function is exported for testing purposes.
func convertVectorsToColumns(vectors []map[string]interface{}, schema *entity.Schema) ([]entity.Column, error) {
	columns, err := initializeColumns(schema, len(vectors))
	if err != nil {
		return nil, err
	}

	for i, v := range vectors {
		fillColumnData(columns, i, v, schema)
	}

	columnList := make([]entity.Column, 0, len(columns))
	for _, col := range columns {
		columnList = append(columnList, col)
	}
	return columnList, nil
}

func initializeColumns(schema *entity.Schema, rowCount int) (map[string]entity.Column, error) {
	columns := make(map[string]entity.Column)
	for _, field := range schema.Fields {
		switch field.DataType {
		case entity.FieldTypeInt64:
			data := make([]int64, rowCount)
			columns[field.Name] = entity.NewColumnInt64(field.Name, data)
		case entity.FieldTypeVarChar:
			data := make([]string, rowCount)
			columns[field.Name] = entity.NewColumnVarChar(field.Name, data)
		case entity.FieldTypeFloatVector:
			dimStr := field.TypeParams["dim"]
			var dim int
			if _, err := fmt.Sscanf(dimStr, "%d", &dim); err != nil {
				return nil, fmt.Errorf("failed to parse dimension from field %s: %w", field.Name, err)
			}
			data := make([][]float32, rowCount)
			for i := range data {
				data[i] = make([]float32, dim)
			}
			columns[field.Name] = entity.NewColumnFloatVector(field.Name, dim, data)
		case entity.FieldTypeBool:
			data := make([]bool, rowCount)
			columns[field.Name] = entity.NewColumnBool(field.Name, data)
		case entity.FieldTypeFloat:
			data := make([]float32, rowCount)
			columns[field.Name] = entity.NewColumnFloat(field.Name, data)
		case entity.FieldTypeDouble:
			data := make([]float64, rowCount)
			columns[field.Name] = entity.NewColumnDouble(field.Name, data)
		default:
			// log warning?
		}
	}
	return columns, nil
}

func fillColumnData(columns map[string]entity.Column, i int, v map[string]interface{}, schema *entity.Schema) {
	// ID
	if id, ok := v["id"]; ok {
		// Try to find PK field
		pkField := ""
		for _, f := range schema.Fields {
			if f.PrimaryKey {
				pkField = f.Name
				break
			}
		}
		if pkField != "" {
			if col, ok := columns[pkField]; ok {
				fillIDColumn(col, i, id)
			}
		}
	}

	// Values (Vector)
	if values, ok := v["values"].([]interface{}); ok {
		// Find vector field
		vecField := ""
		for _, f := range schema.Fields {
			if f.DataType == entity.FieldTypeFloatVector {
				vecField = f.Name
				break
			}
		}
		if vecField != "" {
			if col, ok := columns[vecField]; ok {
				vecData := col.(*entity.ColumnFloatVector).Data()
				for j, val := range values {
					if f, ok := val.(float64); ok {
						vecData[i][j] = float32(f)
					}
				}
			}
		}
	}

	// Metadata
	if metadata, ok := v["metadata"].(map[string]interface{}); ok {
		for key, val := range metadata {
			if col, ok := columns[key]; ok {
				fillMetadataColumn(col, i, val)
			}
		}
	}
}

func fillIDColumn(col entity.Column, i int, id interface{}) {
	if col.Type() == entity.FieldTypeVarChar {
		col.(*entity.ColumnVarChar).Data()[i] = fmt.Sprint(id)
	} else if col.Type() == entity.FieldTypeInt64 {
		var val int64
		switch v := id.(type) {
		case float64:
			val = int64(v)
		case int:
			val = int64(v)
		case int64:
			val = v
		case string:
			// Try to parse string
			_, _ = fmt.Sscanf(v, "%d", &val)
		}
		col.(*entity.ColumnInt64).Data()[i] = val
	}
}

func fillMetadataColumn(col entity.Column, i int, val interface{}) {
	// Assign value
	// This is tedious in Go without generics reflection helpers
	switch col.Type() {
	case entity.FieldTypeVarChar:
		col.(*entity.ColumnVarChar).Data()[i] = fmt.Sprint(val)
	case entity.FieldTypeInt64:
		var v int64
		switch t := val.(type) {
		case float64:
			v = int64(t)
		case int:
			v = int64(t)
		case int64:
			v = t
		case string:
			_, _ = fmt.Sscanf(t, "%d", &v)
		}
		col.(*entity.ColumnInt64).Data()[i] = v
	case entity.FieldTypeFloat:
		var v float32
		switch t := val.(type) {
		case float64:
			v = float32(t)
		case float32:
			v = t
		}
		col.(*entity.ColumnFloat).Data()[i] = v
	case entity.FieldTypeDouble:
		var v float64
		switch t := val.(type) {
		case float64:
			v = t
		case float32:
			v = float64(t)
		}
		col.(*entity.ColumnDouble).Data()[i] = v
	case entity.FieldTypeBool:
		var v bool
		if b, ok := val.(bool); ok {
			v = b
		}
		col.(*entity.ColumnBool).Data()[i] = v
	}
}

// Delete removes vectors.
//
// Summary: Removes vectors.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - ids: []string. The list of IDs to delete.
//   - namespace: string. The namespace to delete from.
//   - filter: map[string]interface{}. An optional metadata filter.
//
// Returns:
//   - map[string]interface{}: The operation result.
//   - error: An error if the operation fails.
func (c *MilvusClient) Delete(ctx context.Context, ids []string, namespace string, filter map[string]interface{}) (map[string]interface{}, error) {
	// Construct expression
	var expr string
	switch {
	case len(ids) > 0:
		// id in ["1", "2"] or id in [1, 2]
		// Assuming PK is named "id" or "pk"? We need to know PK name.
		coll, err := c.client.DescribeCollection(ctx, c.config.GetCollectionName())
		if err != nil {
			return nil, fmt.Errorf("failed to describe collection: %w", err)
		}
		pkField := ""
		var pkType entity.FieldType
		for _, f := range coll.Schema.Fields {
			if f.PrimaryKey {
				pkField = f.Name
				pkType = f.DataType
				break
			}
		}
		if pkField == "" {
			return nil, fmt.Errorf("PK field not found")
		}

		if pkType == entity.FieldTypeInt64 {
			expr = fmt.Sprintf("%s in [%s]", pkField, strings.Join(ids, ", "))
		} else {
			expr = fmt.Sprintf("%s in [\"%s\"]", pkField, strings.Join(ids, "\", \""))
		}
	case filter != nil:
		var parts []string
		for k, v := range filter {
			if s, ok := v.(string); ok {
				parts = append(parts, fmt.Sprintf("%s == \"%s\"", k, s))
			} else {
				parts = append(parts, fmt.Sprintf("%s == %v", k, v))
			}
		}
		expr = strings.Join(parts, " && ")
	default:
		return nil, fmt.Errorf("must provide ids or filter")
	}

	partitionName := ""
	if namespace != "" {
		partitionName = namespace
	}

	err := c.client.Delete(ctx, c.config.GetCollectionName(), partitionName, expr)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"success": true,
	}, nil
}

// DescribeIndexStats returns statistics about the index.
//
// Summary: Returns statistics about the index.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - _ : map[string]interface{}. Unused.
//
// Returns:
//   - map[string]interface{}: The index statistics.
//   - error: An error if the operation fails.
func (c *MilvusClient) DescribeIndexStats(ctx context.Context, _ map[string]interface{}) (map[string]interface{}, error) {
	coll, err := c.client.DescribeCollection(ctx, c.config.GetCollectionName())
	if err != nil {
		return nil, err
	}

	stats, err := c.client.GetCollectionStatistics(ctx, c.config.GetCollectionName())
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":  coll.Name,
		"stats": stats,
	}, nil
}
