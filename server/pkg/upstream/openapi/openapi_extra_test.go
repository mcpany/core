package openapi

import (
	"context"
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestOpenAPIUpstream_Shutdown(t *testing.T) {
	u := NewOpenAPIUpstream()
	ou := u.(*OpenAPIUpstream)

	// Pre-populate a client with empty service ID, which matches default u.serviceID
	client := ou.getHTTPClient("")
	assert.NotNil(t, client)

	// Shutdown
	err := u.Shutdown(context.Background())
	assert.NoError(t, err)

	ou.mu.Lock()
	_, exists := ou.httpClients[""]
	ou.mu.Unlock()

	assert.False(t, exists, "Client should be removed after shutdown")
}

func TestHttpClientImpl_Do(t *testing.T) {
	// Create a real client but mock the transport to avoid network calls
	client := &http.Client{
		Transport: &mockTransport{},
	}
	wrapper := &httpClientImpl{client: client}

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp, err := wrapper.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

type mockTransport struct{}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
	}, nil
}

func TestConvertMcpOperationsToTools_NonObjectResponseBody(t *testing.T) {
	spec := `
{
  "openapi": "3.0.0",
  "info": { "title": "Test", "version": "1.0" },
  "paths": {
    "/array-response": {
      "get": {
        "operationId": "getArray",
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": { "type": "string" }
                }
              }
            }
          }
        }
      }
    }
  }
}`
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	assert.NoError(t, err)

	ops := extractMcpOperationsFromOpenAPI(doc)
	tools := convertMcpOperationsToTools(ops, doc, "test-service")

	assert.Len(t, tools, 1)
	tool := tools[0]
	outputSchema := tool.GetAnnotations().GetOutputSchema()
	props := outputSchema.GetFields()["properties"].GetStructValue().GetFields()

	// Should be wrapped in "response_body"
	assert.Contains(t, props, "response_body")
	respBody := props["response_body"].GetStructValue().GetFields()
	assert.Equal(t, "array", respBody["type"].GetStringValue())
}

func TestConvertSchemaToStructPB_ArrayTypes(t *testing.T) {
	// Test array without items
	schema := &openapi3.Schema{
		Type: &openapi3.Types{"array"},
	}
	doc := &openapi3.T{}

	val, err := convertSchemaToStructPB("testArray", &openapi3.SchemaRef{Value: schema}, "", doc)
	assert.NoError(t, err)
	fields := val.GetStructValue().GetFields()
	assert.Equal(t, "array", fields["type"].GetStringValue())
	assert.Nil(t, fields["items"], "Items should be nil if not specified")

	// Test array with items
	schemaWithItems := &openapi3.Schema{
		Type: &openapi3.Types{"array"},
		Items: &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type: &openapi3.Types{"string"},
			},
		},
	}
	val, err = convertSchemaToStructPB("testArrayItems", &openapi3.SchemaRef{Value: schemaWithItems}, "", doc)
	assert.NoError(t, err)
	fields = val.GetStructValue().GetFields()
	items := fields["items"].GetStructValue().GetFields()
	assert.Equal(t, "string", items["type"].GetStringValue())
}

func TestResolveSchemaRef_EdgeCases(t *testing.T) {
	// Nil schema ref
	val, err := resolveSchemaRef(nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, val)

	// Ref without doc
	ref := &openapi3.SchemaRef{Ref: "#/components/schemas/Test"}
	_, err = resolveSchemaRef(ref, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil doc")
}

func TestMergeSchemaProperties_Recursion(t *testing.T) {
	// Construct a schema with nested AllOf
	doc := &openapi3.T{
		Components: &openapi3.Components{
			Schemas: make(openapi3.Schemas),
		},
	}

	baseSchema := &openapi3.Schema{
		Properties: map[string]*openapi3.SchemaRef{
			"baseProp": {
				Value: &openapi3.Schema{Type: &openapi3.Types{"string"}},
			},
		},
	}
	doc.Components.Schemas["Base"] = &openapi3.SchemaRef{Value: baseSchema}

	middleSchema := &openapi3.Schema{
		AllOf: openapi3.SchemaRefs{
			{Ref: "#/components/schemas/Base"},
		},
		Properties: map[string]*openapi3.SchemaRef{
			"middleProp": {
				Value: &openapi3.Schema{Type: &openapi3.Types{"integer"}},
			},
		},
	}
	doc.Components.Schemas["Middle"] = &openapi3.SchemaRef{Value: middleSchema}

	finalSchema := &openapi3.Schema{
		AllOf: openapi3.SchemaRefs{
			{Ref: "#/components/schemas/Middle"},
		},
		Properties: map[string]*openapi3.SchemaRef{
			"finalProp": {
				Value: &openapi3.Schema{Type: &openapi3.Types{"boolean"}},
			},
		},
	}

	merged, err := mergeSchemaProperties(finalSchema, doc)
	assert.NoError(t, err)
	assert.Contains(t, merged, "baseProp")
	assert.Contains(t, merged, "middleProp")
	assert.Contains(t, merged, "finalProp")
}

func TestConvertSchemaToStructPB_NoTypeButAllOf(t *testing.T) {
	// Case where Type is nil but AllOf is present -> should be treated as object
	schema := &openapi3.Schema{
		AllOf: openapi3.SchemaRefs{
			{
				Value: &openapi3.Schema{
					Properties: map[string]*openapi3.SchemaRef{
						"prop": {Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
					},
				},
			},
		},
	}
	doc := &openapi3.T{}
	val, err := convertSchemaToStructPB("testAllOf", &openapi3.SchemaRef{Value: schema}, "", doc)
	assert.NoError(t, err)
	fields := val.GetStructValue().GetFields()
	assert.Equal(t, "object", fields["type"].GetStringValue())
	assert.Contains(t, fields["properties"].GetStructValue().GetFields(), "prop")
}

func TestConvertSchemaToStructPB_UnhandledType(t *testing.T) {
	schema := &openapi3.Schema{
		Type: &openapi3.Types{"unknown"},
	}
	doc := &openapi3.T{}
	val, err := convertSchemaToStructPB("testUnknown", &openapi3.SchemaRef{Value: schema}, "", doc)
	assert.NoError(t, err)
	fields := val.GetStructValue().GetFields()
	// Should default to string
	assert.Equal(t, "string", fields["type"].GetStringValue())
}

func TestConvertSchemaToStructPB_WithEnumAndDefault(t *testing.T) {
	schema := &openapi3.Schema{
		Type: &openapi3.Types{"string"},
		Enum: []interface{}{"A", "B"},
		Default: "A",
	}
	doc := &openapi3.T{}
	val, err := convertSchemaToStructPB("testEnum", &openapi3.SchemaRef{Value: schema}, "", doc)
	assert.NoError(t, err)
	fields := val.GetStructValue().GetFields()

	assert.NotNil(t, fields["enum"])
	enums := fields["enum"].GetListValue().GetValues()
	assert.Len(t, enums, 2)
	assert.Equal(t, "A", enums[0].GetStringValue())

	assert.NotNil(t, fields["default"])
	assert.Equal(t, "A", fields["default"].GetStringValue())
}
