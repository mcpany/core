package vector

import (
	"testing"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertVectorsToColumns_Types(t *testing.T) {
	schema := &entity.Schema{
		Fields: []*entity.Field{
			{Name: "id", DataType: entity.FieldTypeInt64, PrimaryKey: true},
			{Name: "embedding", DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": "2"}},
			{Name: "meta_int", DataType: entity.FieldTypeInt64},
			{Name: "meta_str", DataType: entity.FieldTypeVarChar},
			{Name: "meta_float", DataType: entity.FieldTypeFloat},
			{Name: "meta_double", DataType: entity.FieldTypeDouble},
			{Name: "meta_bool", DataType: entity.FieldTypeBool},
		},
	}

	vectors := []map[string]interface{}{
		{
			"id":     int64(1),
			"values": []interface{}{0.1, 0.2},
			"metadata": map[string]interface{}{
				"meta_int":    int64(10),
				"meta_str":    "test",
				"meta_float":  float32(1.1),
				"meta_double": 2.2,
				"meta_bool":   true,
			},
		},
		{
			"id":     int64(2),
			"values": []interface{}{0.3, 0.4},
			"metadata": map[string]interface{}{
				"meta_int":    20,   // int -> int64 conversion
				"meta_str":    123,  // int -> string conversion
				"meta_float":  1.5,  // float64 -> float32 conversion
				"meta_double": float32(2.5), // float32 -> float64 conversion
				"meta_bool":   false,
			},
		},
	}

	columns, err := convertVectorsToColumns(vectors, schema)
	require.NoError(t, err)
	require.Len(t, columns, 7)

	colMap := make(map[string]entity.Column)
	for _, col := range columns {
		colMap[col.Name()] = col
	}

	// Verify ID column
	idCol := colMap["id"].(*entity.ColumnInt64)
	assert.Equal(t, int64(1), idCol.Data()[0])
	assert.Equal(t, int64(2), idCol.Data()[1])

	// Verify Embedding column
	vecCol := colMap["embedding"].(*entity.ColumnFloatVector)
	assert.Equal(t, []float32{0.1, 0.2}, vecCol.Data()[0])
	assert.Equal(t, []float32{0.3, 0.4}, vecCol.Data()[1])

	// Verify Metadata columns
	assert.Equal(t, int64(10), colMap["meta_int"].(*entity.ColumnInt64).Data()[0])
	assert.Equal(t, int64(20), colMap["meta_int"].(*entity.ColumnInt64).Data()[1])

	assert.Equal(t, "test", colMap["meta_str"].(*entity.ColumnVarChar).Data()[0])
	assert.Equal(t, "123", colMap["meta_str"].(*entity.ColumnVarChar).Data()[1])

	assert.Equal(t, float32(1.1), colMap["meta_float"].(*entity.ColumnFloat).Data()[0])
	assert.Equal(t, float32(1.5), colMap["meta_float"].(*entity.ColumnFloat).Data()[1])

	assert.InDelta(t, 2.2, colMap["meta_double"].(*entity.ColumnDouble).Data()[0], 0.0001)
	assert.InDelta(t, 2.5, colMap["meta_double"].(*entity.ColumnDouble).Data()[1], 0.0001)

	assert.True(t, colMap["meta_bool"].(*entity.ColumnBool).Data()[0])
	assert.False(t, colMap["meta_bool"].(*entity.ColumnBool).Data()[1])
}

func TestConvertVectorsToColumns_IDConversion(t *testing.T) {
	// Scenario 1: ID is string, PK is Int64
	schemaInt := &entity.Schema{Fields: []*entity.Field{{Name: "id", DataType: entity.FieldTypeInt64, PrimaryKey: true}}}
	cols, err := convertVectorsToColumns([]map[string]interface{}{{"id": "123"}}, schemaInt)
	require.NoError(t, err)
	assert.Equal(t, int64(123), cols[0].(*entity.ColumnInt64).Data()[0])

	cols, err = convertVectorsToColumns([]map[string]interface{}{{"id": 123}}, schemaInt)
	require.NoError(t, err)
	assert.Equal(t, int64(123), cols[0].(*entity.ColumnInt64).Data()[0])

	cols, err = convertVectorsToColumns([]map[string]interface{}{{"id": 123.0}}, schemaInt)
	require.NoError(t, err)
	assert.Equal(t, int64(123), cols[0].(*entity.ColumnInt64).Data()[0])

	// Scenario 2: ID is int, PK is VarChar
	schemaStr := &entity.Schema{Fields: []*entity.Field{{Name: "id", DataType: entity.FieldTypeVarChar, PrimaryKey: true}}}
	cols, err = convertVectorsToColumns([]map[string]interface{}{{"id": 123}}, schemaStr)
	require.NoError(t, err)
	assert.Equal(t, "123", cols[0].(*entity.ColumnVarChar).Data()[0])
}

func TestInitializeColumns_Errors(t *testing.T) {
	// Invalid dimension format
	schema := &entity.Schema{
		Fields: []*entity.Field{
			{Name: "vec", DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": "invalid"}},
		},
	}
	_, err := initializeColumns(schema, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse dimension")
}

func TestFillColumnData_MissingVector(t *testing.T) {
	schema := &entity.Schema{
		Fields: []*entity.Field{
			{Name: "vec", DataType: entity.FieldTypeFloatVector, TypeParams: map[string]string{"dim": "2"}},
		},
	}
	// Missing "values" key
	cols, err := convertVectorsToColumns([]map[string]interface{}{{}}, schema)
	require.NoError(t, err)
	// Should be zero-initialized
	assert.Equal(t, []float32{0, 0}, cols[0].(*entity.ColumnFloatVector).Data()[0])
}

func TestFillColumnData_MissingMetadata(t *testing.T) {
	schema := &entity.Schema{
		Fields: []*entity.Field{
			{Name: "meta", DataType: entity.FieldTypeInt64},
		},
	}
	// Missing "metadata" key
	cols, err := convertVectorsToColumns([]map[string]interface{}{{}}, schema)
	require.NoError(t, err)
	assert.Equal(t, int64(0), cols[0].(*entity.ColumnInt64).Data()[0])
}
