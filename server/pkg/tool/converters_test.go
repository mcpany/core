package tool

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/upstream/grpc/protobufparser"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestConvertToolDefinitionToProto(t *testing.T) {
	t.Parallel()
	t.Run("nil tool definition", func(t *testing.T) {
		pbTool, err := ConvertToolDefinitionToProto(nil, nil, nil)
		assert.Error(t, err)
		assert.Nil(t, pbTool)
		assert.Contains(t, err.Error(), "cannot convert nil tool definition to proto")
	})

	t.Run("valid tool definition", func(t *testing.T) {
		inputSchema, _ := structpb.NewStruct(map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param1": map[string]interface{}{
					"type": "string",
				},
			},
		})
		outputSchema, _ := structpb.NewStruct(map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"result": map[string]interface{}{
					"type": "string",
				},
			},
		})

		toolDef := configv1.ToolDefinition_builder{
			Name:        proto.String("test-tool"),
			Description: proto.String("A test tool"),
			Title:       proto.String("Test Tool"),
			ServiceId:   proto.String("test-service"),
		}.Build()

		pbTool, err := ConvertToolDefinitionToProto(toolDef, inputSchema, outputSchema)
		assert.NoError(t, err)
		assert.NotNil(t, pbTool)
		assert.Equal(t, "test-tool", pbTool.GetName())
		assert.Equal(t, "A test tool", pbTool.GetDescription())
		assert.Equal(t, "Test Tool", pbTool.GetDisplayName())
		assert.Equal(t, "test-service", pbTool.GetServiceId())
		assert.Equal(t, inputSchema, pbTool.GetAnnotations().GetInputSchema())
		assert.Equal(t, outputSchema, pbTool.GetAnnotations().GetOutputSchema())
	})
}

func TestConvertJSONSchemaToStruct(t *testing.T) {
	t.Parallel()
	t.Run("nil schema", func(t *testing.T) {
		s, err := convertJSONSchemaToStruct(nil)
		assert.NoError(t, err)
		assert.Nil(t, s)
	})

	t.Run("invalid schema type", func(t *testing.T) {
		_, err := convertJSONSchemaToStruct("not a map")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "schema is not a valid JSON object")
	})

	t.Run("valid schema", func(t *testing.T) {
		schema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param1": map[string]interface{}{
					"type": "string",
				},
			},
		}
		s, err := convertJSONSchemaToStruct(schema)
		assert.NoError(t, err)
		assert.NotNil(t, s)
	})
}

func TestGetJSONSchemaForScalarType(t *testing.T) {
	t.Parallel()
	t.Run("unsupported type", func(t *testing.T) {
		_, err := GetJSONSchemaForScalarType("unsupported", "description")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported scalar type: unsupported")
	})

	t.Run("supported types", func(t *testing.T) {
		testCases := []struct {
			scalarType string
			jsonType   string
		}{
			{"TYPE_DOUBLE", "number"},
			{"TYPE_FLOAT", "number"},
			{"TYPE_INT32", "integer"},
			{"TYPE_INT64", "integer"},
			{"TYPE_UINT32", "integer"},
			{"TYPE_UINT64", "integer"},
			{"TYPE_SINT32", "integer"},
			{"TYPE_SINT64", "integer"},
			{"TYPE_FIXED32", "integer"},
			{"TYPE_FIXED64", "integer"},
			{"TYPE_SFIXED32", "integer"},
			{"TYPE_SFIXED64", "integer"},
			{"TYPE_BOOL", "boolean"},
			{"TYPE_STRING", "string"},
			{"TYPE_BYTES", "string"},
		}

		for _, tc := range testCases {
			t.Run(tc.scalarType, func(t *testing.T) {
				schema, err := GetJSONSchemaForScalarType(tc.scalarType, "description")
				assert.NoError(t, err)
				assert.Equal(t, tc.jsonType, schema.Type)
				assert.Equal(t, "description", schema.Description)
			})
		}
	})
}
// ... (skip others if needed or replace block)
// I will just replace the TestGetJSONSchemaForScalarType block first


func TestConvertMCPToolToProto(t *testing.T) {
	t.Parallel()
	t.Run("nil tool", func(t *testing.T) {
		pbTool, err := ConvertMCPToolToProto(nil)
		assert.Error(t, err)
		assert.Nil(t, pbTool)
		assert.Contains(t, err.Error(), "cannot convert nil mcp tool to proto")
	})

	t.Run("valid tool", func(t *testing.T) {
		mcpTool := &mcp.Tool{
			Name:        "test-tool",
			Description: "A test tool",
			Title:       "Test Tool",
		}

		pbTool, err := ConvertMCPToolToProto(mcpTool)
		assert.NoError(t, err)
		assert.NotNil(t, pbTool)
		assert.Equal(t, "test-tool", pbTool.GetName())
		assert.Equal(t, "A test tool", pbTool.GetDescription())
		assert.Equal(t, "Test Tool", pbTool.GetDisplayName())
	})

	t.Run("full annotations and schemas", func(t *testing.T) {
		destructiveHint := true
		openWorldHint := false
		mcpTool := &mcp.Tool{
			Name:        "test-tool",
			Description: "A test tool",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Annotation Title",
				DestructiveHint: &destructiveHint,
				OpenWorldHint:   &openWorldHint,
			},
			InputSchema:  map[string]interface{}{"type": "string"},
			OutputSchema: map[string]interface{}{"type": "number"},
		}

		pbTool, err := ConvertMCPToolToProto(mcpTool)
		assert.NoError(t, err)
		assert.NotNil(t, pbTool)
		assert.Equal(t, "Annotation Title", pbTool.GetDisplayName())
		assert.True(t, pbTool.GetAnnotations().GetDestructiveHint())
		assert.False(t, pbTool.GetAnnotations().GetOpenWorldHint())
		assert.NotNil(t, pbTool.GetAnnotations().GetInputSchema())
		assert.NotNil(t, pbTool.GetAnnotations().GetOutputSchema())
	})

	t.Run("nil annotations", func(t *testing.T) {
		mcpTool := &mcp.Tool{
			Name:        "test-tool",
			Description: "A test tool",
			Annotations: nil,
		}
		pbTool, err := ConvertMCPToolToProto(mcpTool)
		assert.NoError(t, err)
		assert.NotNil(t, pbTool)
		assert.NotNil(t, pbTool.GetAnnotations())
	})
}

func TestConvertMcpFieldsToInputSchemaProperties(t *testing.T) {
	t.Parallel()
	t.Run("empty fields", func(t *testing.T) {
		properties, err := ConvertMcpFieldsToInputSchemaProperties(nil)
		assert.NoError(t, err)
		assert.NotNil(t, properties)
		assert.Empty(t, properties.Fields)
	})

	t.Run("valid fields", func(t *testing.T) {
		fields := []*protobufparser.McpField{
			{
				Name:        "param1",
				Type:        "TYPE_STRING",
				Description: "A string parameter",
			},
			{
				Name:        "param2",
				Type:        "TYPE_INT32",
				Description: "An integer parameter",
			},
		}

		properties, err := ConvertMcpFieldsToInputSchemaProperties(fields)
		assert.NoError(t, err)
		assert.NotNil(t, properties)
		assert.Len(t, properties.Fields, 2)
	})
}

func TestConvertMCPToolToProto_NilInputSchema(t *testing.T) {
	t.Parallel()
	mcpTool := &mcp.Tool{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: nil,
	}

	pbTool, err := ConvertMCPToolToProto(mcpTool)
	assert.NoError(t, err)
	assert.NotNil(t, pbTool)
	assert.NotNil(t, pbTool.GetAnnotations().GetInputSchema())
	assert.Equal(t, "object", pbTool.GetAnnotations().GetInputSchema().GetFields()["type"].GetStringValue())
}

func TestConvertProtoToMCPTool_NilTool(t *testing.T) {
	t.Parallel()
	mcpTool, err := ConvertProtoToMCPTool(nil)
	assert.Error(t, err)
	assert.Nil(t, mcpTool)
	assert.EqualError(t, err, "cannot convert nil pb tool to mcp tool")
}

func TestConvertProtoToMCPTool_EmptyToolName(t *testing.T) {
	t.Parallel()
	pbTool := configv1.ToolDefinition_builder{
		Name: proto.String(""),
	}.Build()
	pbToolProto, err := ConvertToolDefinitionToProto(pbTool, nil, nil)
	assert.NoError(t, err)

	mcpTool, err := ConvertProtoToMCPTool(pbToolProto)
	assert.Error(t, err)
	assert.Nil(t, mcpTool)
	assert.EqualError(t, err, "tool name cannot be empty")
}

func TestConvertProtoToMCPTool(t *testing.T) {
	t.Parallel()
	t.Run("valid tool", func(t *testing.T) {
		pbTool := configv1.ToolDefinition_builder{
			Name:        proto.String("test-tool"),
			Description: proto.String("A test tool"),
			Title:       proto.String("Test Tool"),
			ServiceId:   proto.String("test-service"),
		}.Build()
		pbToolProto, err := ConvertToolDefinitionToProto(pbTool, nil, nil)
		assert.NoError(t, err)

		mcpTool, err := ConvertProtoToMCPTool(pbToolProto)
		assert.NoError(t, err)
		assert.NotNil(t, mcpTool)
		assert.Equal(t, "test-service.test-tool", mcpTool.Name)
		assert.Equal(t, "A test tool", mcpTool.Description)
		assert.Equal(t, "Test Tool", mcpTool.Title)
	})
}
