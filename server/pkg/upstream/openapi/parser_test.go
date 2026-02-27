// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
)

var petstoreSpec = []byte(`
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
servers:
  - url: http://petstore.swagger.io/v1
paths:
  /pets:
    get:
      summary: List all pets
      operationId: listPets
      tags:
        - pets
      parameters:
        - name: limit
          in: query
          description: How many items to return at one time (max 100)
          required: false
          schema:
            type: integer
            format: int32
      responses:
        '200':
          description: A paged array of pets
          headers:
            x-next:
              description: A link to the next page of responses
              schema:
                type: string
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pets"
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
    post:
      summary: Create a pet
      operationId: createPet
      tags:
        - pets
      responses:
        '201':
          description: Null response
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
  /pets/{petId}:
    get:
      summary: Info for a specific pet
      operationId: showPetById
      tags:
        - pets
      parameters:
        - name: petId
          in: path
          required: true
          description: The id of the pet to retrieve
          schema:
            type: string
      responses:
        '200':
          description: Expected response to a valid request
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pet"
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
components:
  schemas:
    Pet:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        tag:
          type: string
    Pets:
      type: array
      items:
        $ref: "#/components/schemas/Pet"
    Error:
      type: object
      required:
        - code
        - message
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string
`)

func TestSanitizeOperationID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid ID", "listPets", "listPets"},
		{"ID with space", "list Pets", "list_Pets"},
		{"consecutive invalid chars", "list  Pets", "list_Pets"},
		{"different invalid chars", "list-Pets.V1", "list_Pets_V1"},
		{"another different invalid chars", "list-Pets/V1", "list_Pets_V1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.SanitizeOperationID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractMcpOperationsFromOpenAPI(t *testing.T) {
	_, doc, err := parseOpenAPISpec(context.Background(), petstoreSpec)
	require.NoError(t, err)

	ops := extractMcpOperationsFromOpenAPI(doc)
	assert.Len(t, ops, 3)

	opMap := make(map[string]McpOperation)
	for _, op := range ops {
		opMap[op.OperationID] = op
	}

	assert.Contains(t, opMap, "listPets")
	assert.Contains(t, opMap, "createPet")
	assert.Contains(t, opMap, "showPetById")

	listPets := opMap["listPets"]
	assert.Equal(t, "GET", listPets.Method)
	assert.Equal(t, "/pets", listPets.Path)
}

func TestConvertMcpOperationsToTools(t *testing.T) {
	_, doc, err := parseOpenAPISpec(context.Background(), petstoreSpec)
	require.NoError(t, err)

	ops := extractMcpOperationsFromOpenAPI(doc)
	tools := convertMcpOperationsToTools(ops, doc, "test-service")

	assert.Len(t, tools, 3)

	toolMap := make(map[string]*pb.Tool)
	for _, tool := range tools {
		toolMap[tool.GetName()] = tool
	}

	t.Run("listPets", func(t *testing.T) {
		assert.Contains(t, toolMap, "listPets")
		// Check tool details if needed, e.g. input schema having 'limit'
	})

	t.Run("createPet", func(t *testing.T) {
		assert.Contains(t, toolMap, "createPet")
	})

	t.Run("showPetById", func(t *testing.T) {
		assert.Contains(t, toolMap, "showPetById")
	})
}

func TestParseOpenAPISpec_LoadAndValidate(t *testing.T) {
	parsed, doc, err := parseOpenAPISpec(context.Background(), petstoreSpec)
	require.NoError(t, err)
	assert.NotNil(t, parsed)
	assert.NotNil(t, doc)
	assert.Equal(t, "Swagger Petstore", parsed.Info.Title)
}

func TestExtractMcpOperationsFromOpenAPI_XMLContent(t *testing.T) {
	// Simple spec with XML content type to verify map logic
	spec := []byte(`
openapi: "3.0.0"
info:
  title: XML Test
  version: 1.0.0
paths:
  /xml:
    post:
      operationId: postXML
      requestBody:
        content:
          application/xml:
            schema:
              type: object
      responses:
        '200':
          description: ok
`)
	_, doc, err := parseOpenAPISpec(context.Background(), spec)
	require.NoError(t, err)

	ops := extractMcpOperationsFromOpenAPI(doc)
	require.Len(t, ops, 1)
	op := ops[0]
	assert.Equal(t, "postXML", op.OperationID)
	assert.Contains(t, op.RequestBodySchema, "application/xml")
}

func TestIsOperationIdempotent(t *testing.T) {
	tests := []struct {
		method   string
		expected bool
	}{
		{"GET", true},
		{"POST", false},
		{"PUT", true},
		{"DELETE", true},
		{"PATCH", false},
		{"HEAD", true},
		{"OPTIONS", true},
		{"TRACE", true},
		{"get", true}, // Case insensitivity check
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			assert.Equal(t, tt.expected, isOperationIdempotent(tt.method))
		})
	}
}

func TestParseOpenAPISpec_Errors(t *testing.T) {
	t.Run("invalid_json", func(t *testing.T) {
		_, _, err := parseOpenAPISpec(context.Background(), []byte("{invalid"))
		assert.Error(t, err)
	})

	t.Run("validation_failure", func(t *testing.T) {
		// Missing version/info
		_, _, err := parseOpenAPISpec(context.Background(), []byte(`{"openapi": "3.0.0"}`))
		assert.Error(t, err)
	})
}

func TestConvertOpenAPISchemaToInputSchemaProperties_Errors(t *testing.T) {
	_, doc, err := parseOpenAPISpec(context.Background(), petstoreSpec)
	require.NoError(t, err)

	t.Run("nil_schema_ref", func(t *testing.T) {
		props, err := convertOpenAPISchemaToInputSchemaProperties(nil, nil, doc)
		assert.NoError(t, err) // Should be valid empty struct
		assert.Empty(t, props.Fields)
	})

	t.Run("unresolvable_ref", func(t *testing.T) {
		badRef := &openapi3.SchemaRef{Ref: "#/components/schemas/NonExistent"}
		_, err := convertOpenAPISchemaToInputSchemaProperties(badRef, nil, doc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not resolve request body schema reference")
	})
}

func TestConvertMcpOperationsToTools_NoOperationID(t *testing.T) {
	spec := []byte(`
openapi: "3.0.0"
info:
  title: NoID
  version: 1.0.0
paths:
  /noid:
    get:
      summary: No Operation ID
      responses:
        '200':
          description: ok
`)
	_, doc, err := parseOpenAPISpec(context.Background(), spec)
	require.NoError(t, err)

	ops := extractMcpOperationsFromOpenAPI(doc)
	tools := convertMcpOperationsToTools(ops, doc, "svc")
	require.Len(t, tools, 1)
	assert.Equal(t, "GET_noid", tools[0].GetName()) // Fallback to Method_Path
}

func TestConvertMcpOperationsToTools_NonObjectRequestBody(t *testing.T) {
	spec := []byte(`
openapi: "3.0.0"
info:
  title: ArrayBody
  version: 1.0.0
paths:
  /array:
    post:
      operationId: postArray
      requestBody:
        content:
          application/json:
            schema:
              type: array
              items:
                type: string
      responses:
        '200':
          description: ok
`)
	_, doc, err := parseOpenAPISpec(context.Background(), spec)
	require.NoError(t, err)

	ops := extractMcpOperationsFromOpenAPI(doc)
	tools := convertMcpOperationsToTools(ops, doc, "svc")
	require.Len(t, tools, 1)

	// Should wrap non-object in "request_body"
	fields := tools[0].GetInputSchema().GetFields()["properties"].GetStructValue().GetFields()
	assert.Contains(t, fields, "request_body")
	assert.Equal(t, "array", fields["request_body"].GetStructValue().GetFields()["type"].GetStringValue())
}

func TestConvertMcpOperationsToTools_AllOfAndNested(t *testing.T) {
	spec := []byte(`
openapi: "3.0.0"
info:
  title: AllOf
  version: 1.0.0
paths:
  /allof:
    post:
      operationId: postAllOf
      requestBody:
        content:
          application/json:
            schema:
              allOf:
                - $ref: "#/components/schemas/Base"
                - type: object
                  properties:
                    extended:
                      type: string
      responses:
        '200':
          description: ok
components:
  schemas:
    Base:
      type: object
      properties:
        base:
          type: string
`)
	_, doc, err := parseOpenAPISpec(context.Background(), spec)
	require.NoError(t, err)

	ops := extractMcpOperationsFromOpenAPI(doc)
	tools := convertMcpOperationsToTools(ops, doc, "svc")
	require.Len(t, tools, 1)

	fields := tools[0].GetInputSchema().GetFields()["properties"].GetStructValue().GetFields()
	assert.Contains(t, fields, "base")
	assert.Contains(t, fields, "extended")
}

func TestConvertSchemaToStructPB_NilSchemaValue(t *testing.T) {
	_, doc, err := parseOpenAPISpec(context.Background(), petstoreSpec)
	require.NoError(t, err)

	// Manually create a SchemaRef with nil Value but no Ref (should imply value is nil)
	sr := &openapi3.SchemaRef{Value: nil}

	val, err := convertSchemaToStructPB("nilField", sr, "", doc)
	assert.Error(t, err)
	assert.Nil(t, val)
}

func TestConvertSchemaToStructPB_UnsupportedType(t *testing.T) {
	_, doc, err := parseOpenAPISpec(context.Background(), petstoreSpec)
	require.NoError(t, err)

	// Unknown type
	sr := &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"unknown"}}}

	val, err := convertSchemaToStructPB("unknownField", sr, "", doc)
	assert.NoError(t, err)
	// Defaults to string
	assert.Equal(t, "string", val.GetStructValue().GetFields()["type"].GetStringValue())
}
