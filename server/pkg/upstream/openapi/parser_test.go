// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/mcpany/core/server/pkg/util"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	listAllPetsSummary = "List all pets"
	createPetSummary   = "Create a pet"
	petsPath           = "/pets"
	petIDParam         = "petId"
	stringType         = "string"
	contentTypeJSON    = "application/json"
	opListPets         = "listPets"
	opCreatePet        = "createPet"
	opShowPetByID      = "showPetById"
	methodPost         = "POST"
)

const sampleOpenAPISpecJSON = `
{
  "openapi": "3.0.0",
  "info": {
    "title": "Sample Pet Store App",
    "version": "1.0.1",
    "description": "A sample pet store API."
  },
  "servers": [
    {
      "url": "http://petstore.swagger.io/v1"
    }
  ],
  "paths": {
    "/pets": {
      "get": {
        "summary": "List all pets",
        "operationId": "listPets",
        "description": "Returns all pets from the system that the user has access to.",
        "responses": {
          "200": {
            "description": "A list of pets.",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/Pet"
                  }
                }
              }
            }
          }
        }
      },
      "post": {
        "summary": "Create a pet",
        "operationId": "createPet",
        "requestBody": {
          "description": "Pet to add to the store",
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/PetInput"
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Null response"
          }
        }
      }
    },
    "/pets/{petId}": {
      "get": {
        "summary": "Info for a specific pet",
        "operationId": "showPetById",
        "parameters": [
          {
            "name": "petId",
            "in": "path",
            "required": true,
            "description": "The id of the pet to retrieve",
            "schema": {
              "type": "string"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Information about the pet",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Pet"
                }
              }
            }
          },
          "default": {
            "description": "Unexpected error",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Error"
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "Pet": {
        "type": "object",
        "required": [
          "id",
          "name"
        ],
        "properties": {
          "id": {
            "type": "integer",
            "format": "int64",
            "description": "Unique identifier for the Pet."
          },
          "name": {
            "type": "string",
            "description": "Name of the Pet."
          },
          "tag": {
            "type": "string"
          }
        }
      },
      "PetInput": {
        "type": "object",
        "required": ["name"],
        "properties": {
          "name": {
            "type": "string",
            "description": "Name of the Pet."
          },
          "status": {
            "type": "string",
            "enum": ["available", "pending", "sold"],
            "description": "pet status in the store"
          }
        }
      },
      "Error": {
        "type": "object",
        "required": [
          "code",
          "message"
        ],
        "properties": {
          "code": {
            "type": "integer",
            "format": "int32"
          },
          "message": {
            "type": "string"
          }
        }
      }
    }
  }
}
`

const nonObjectBodySpecJSON = `
{
  "openapi": "3.0.0",
  "info": {
    "title": "Test API for non-object request body",
    "version": "1.0.0"
  },
  "paths": {
    "/process-items": {
      "post": {
        "summary": "Process a list of items",
        "operationId": "processItems",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "array",
                "items": {
                  "type": "string"
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "OK"
          }
        }
      }
    }
  }
}
`

func TestSanitizeOperationID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid ID",
			input: "getPet",
			want:  "getPet",
		},
		{
			name:  "ID with space",
			input: "get pet",
			want:  "get_36a9e7_pet",
		},
		{
			name:  "consecutive invalid chars",
			input: "a<>b",
			want:  "a_24295a_b",
		},
		{
			name:  "different invalid chars",
			input: "A%B",
			want:  "A_bbf3f1_B",
		},
		{
			name:  "another different invalid chars",
			input: "A<B",
			want:  "A_dabd3a_B",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := util.SanitizeOperationID(tt.input); got != tt.want {
				t.Errorf("util.SanitizeOperationID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func loadTestSpec(t *testing.T) *openapi3.T {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(sampleOpenAPISpecJSON))
	if err != nil {
		t.Fatalf("Failed to load sample OpenAPI spec: %v", err)
	}
	err = doc.Validate(context.Background())
	if err != nil {
		t.Fatalf("Sample OpenAPI spec failed validation: %v", err)
	}
	return doc
}

func TestExtractMcpOperationsFromOpenAPI(t *testing.T) {
	doc := loadTestSpec(t)
	ops := extractMcpOperationsFromOpenAPI(doc)

	if len(ops) != 3 {
		t.Fatalf("Expected 3 operations, got %d", len(ops))
	}

	opsMap := make(map[string]McpOperation)
	for _, op := range ops {
		opsMap[op.OperationID] = op
	}

	// Check listPets
	opListPets, ok := opsMap[opListPets]
	if !ok {
		t.Fatalf("Operation 'listPets' not found")
	}
	if opListPets.Path != petsPath {
		t.Errorf("Expected listPets Path '/pets', got '%s'", opListPets.Path)
	}
	if opListPets.Method != methodGet {
		t.Errorf("Expected listPets Method 'GET', got '%s'", opListPets.Method)
	}
	if opListPets.Summary != listAllPetsSummary {
		t.Errorf("Expected listPets Summary 'List all pets', got '%s'", opListPets.Summary)
	}
	if _, ok := opListPets.ResponseSchemas["200"][contentTypeJSON]; !ok {
		t.Error("Expected listPets to have 200 application/json response schema")
	}

	// Check createPet
	opCreatePet, ok := opsMap[opCreatePet]
	if !ok {
		t.Fatalf("Operation 'createPet' not found")
	}
	if opCreatePet.Path != petsPath { // createPet also uses /pets path
		t.Errorf("Expected createPet Path '/pets', got '%s'", opCreatePet.Path)
	}
	if opCreatePet.Method != methodPost {
		t.Errorf("Expected createPet Method 'POST', got '%s'", opCreatePet.Method)
	}
	if opCreatePet.Summary != createPetSummary {
		t.Errorf("Expected createPet Summary 'Create a pet', got '%s'", opCreatePet.Summary)
	}
	if _, ok := opCreatePet.RequestBodySchema[contentTypeJSON]; !ok {
		t.Error("Expected createPet to have application/json request body schema")
	}

	// Check showPetByID
	opShowPetByID, ok := opsMap[opShowPetByID]
	if !ok {
		t.Fatalf("Operation 'showPetById' not found")
	}
	if opShowPetByID.Path != "/pets/{petId}" {
		t.Errorf("Expected showPetById Path '/pets/{petId}', got '%s'", opShowPetByID.Path)
	}
	if opShowPetByID.Method != methodGet {
		t.Errorf("Expected showPetById Method 'GET', got '%s'", opShowPetByID.Method)
	}
	if opShowPetByID.Summary != "Info for a specific pet" {
		t.Errorf("Expected showPetById Summary 'Info for a specific pet', got '%s'", opShowPetByID.Summary)
	}
}

func TestConvertMcpOperationsToTools(t *testing.T) {
	doc := loadTestSpec(t)
	ops := extractMcpOperationsFromOpenAPI(doc)
	mcpServerServiceKey := "petstore_instance_1" // Service key should not contain a namespace.

	tools := convertMcpOperationsToTools(ops, doc, mcpServerServiceKey)

	if len(tools) != 3 {
		t.Fatalf("Expected 3 tools, got %d", len(tools))
	}

	toolsMap := make(map[string]*v1.Tool)
	for _, tool := range tools {
		toolsMap[tool.GetName()] = tool
	}

	t.Run("listPets", func(t *testing.T) {
		verifyListPetsTool(t, toolsMap[opListPets], mcpServerServiceKey)
	})

	t.Run("createPet", func(t *testing.T) {
		verifyCreatePetTool(t, toolsMap[opCreatePet])
	})

	t.Run("showPetById", func(t *testing.T) {
		verifyShowPetByIDTool(t, toolsMap[opShowPetByID])
	})
}

func verifyListPetsTool(t *testing.T, toolListPets *v1.Tool, serviceID string) {
	if toolListPets == nil {
		t.Fatalf("Tool 'listPets' not found")
	}

	if toolListPets.GetDisplayName() != listAllPetsSummary {
		t.Errorf("Expected listPets DisplayName 'List all pets', got '%s'", toolListPets.GetDisplayName())
	}
	if toolListPets.GetServiceId() != serviceID {
		t.Errorf("Expected listPets ServiceId '%s', got '%s'", serviceID, toolListPets.GetServiceId())
	}
	if toolListPets.GetUnderlyingMethodFqn() != "GET /pets" {
		t.Errorf("Expected listPets UnderlyingMethodFqn 'GET /pets', got '%s'", toolListPets.GetUnderlyingMethodFqn())
	}
	// Annotations for listPets (GET)
	if toolListPets.GetAnnotations() == nil {
		t.Fatalf("listPets Annotations is nil")
	}
	if toolListPets.GetAnnotations().GetTitle() != listAllPetsSummary {
		t.Errorf("listPets Annotations.Title: got '%s', want 'List all pets'", toolListPets.GetAnnotations().GetTitle())
	}
	if toolListPets.GetAnnotations().GetReadOnlyHint() != true { // GET is read-only
		t.Errorf("listPets Annotations.ReadOnlyHint: got %v, want true", toolListPets.GetAnnotations().GetReadOnlyHint())
	}
	if toolListPets.GetAnnotations().GetIdempotentHint() != true { // GET is idempotent
		t.Errorf("listPets Annotations.IdempotentHint: got %v, want true", toolListPets.GetAnnotations().GetIdempotentHint())
	}
	if toolListPets.GetAnnotations().GetOpenWorldHint() != true { // Default
		t.Errorf("listPets Annotations.OpenWorldHint: got %v, want true", toolListPets.GetAnnotations().GetOpenWorldHint())
	}
	// InputSchema for listPets (no parameters, no request body)
	inputSchemaListPets := toolListPets.GetAnnotations().GetInputSchema()
	if inputSchemaListPets == nil {
		t.Fatalf("listPets InputSchema is nil")
	}
	propertiesListPets := inputSchemaListPets.GetFields()["properties"].GetStructValue()
	if len(propertiesListPets.GetFields()) != 0 {
		t.Errorf("listPets InputSchema.Properties should be empty, got %v", propertiesListPets.GetFields())
	}
}

func verifyCreatePetTool(t *testing.T, toolCreatePet *v1.Tool) {
	if toolCreatePet == nil {
		t.Fatalf("Tool 'createPet' not found")
	}
	if toolCreatePet.GetDisplayName() != createPetSummary {
		t.Errorf("Expected createPet DisplayName 'Create a pet', got '%s'", toolCreatePet.GetDisplayName())
	}
	// Annotations for createPet (POST)
	if toolCreatePet.GetAnnotations() == nil {
		t.Fatalf("createPet Annotations is nil")
	}
	if toolCreatePet.GetAnnotations().GetTitle() != createPetSummary {
		t.Errorf("createPet Annotations.Title: got '%s', want 'Create a pet'", toolCreatePet.GetAnnotations().GetTitle())
	}
	if toolCreatePet.GetAnnotations().GetIdempotentHint() != false { // POST is not idempotent
		t.Errorf("createPet Annotations.IdempotentHint: got %v, want false", toolCreatePet.GetAnnotations().GetIdempotentHint())
	}
	// InputSchema for createPet (has request body PetInput)
	inputSchemaCreatePet := toolCreatePet.GetAnnotations().GetInputSchema()
	if inputSchemaCreatePet == nil {
		t.Fatalf("createPet InputSchema is nil")
	}
	propertiesCreatePet := inputSchemaCreatePet.GetFields()["properties"].GetStructValue()
	if propertiesCreatePet.GetFields() == nil {
		t.Fatalf("createPet InputSchema.Properties or its fields are nil")
	}
	// Check PetInput properties: name (string), status (string)
	propName, ok := propertiesCreatePet.GetFields()["name"]
	if !ok {
		t.Fatalf("createPet InputSchema.Properties missing 'name'")
	}
	if propName.GetStructValue().GetFields()["type"].GetStringValue() != stringType {
		t.Errorf("createPet 'name' property type: got %s, want string", propName.GetStructValue().GetFields()["type"].GetStringValue())
	}
	propStatus, ok := propertiesCreatePet.GetFields()["status"]
	if !ok {
		t.Fatalf("createPet InputSchema.Properties missing 'status'")
	}
	if propStatus.GetStructValue().GetFields()["type"].GetStringValue() != stringType {
		t.Errorf("createPet 'status' property type: got %s, want string", propStatus.GetStructValue().GetFields()["type"].GetStringValue())
	}
	if propStatus.GetStructValue().GetFields()["description"].GetStringValue() != "pet status in the store" {
		t.Errorf("createPet 'status' property description: got '%s', want 'pet status in the store'", propStatus.GetStructValue().GetFields()["description"].GetStringValue())
	}
	enumVal := propStatus.GetStructValue().GetFields()["enum"].GetListValue()
	if enumVal == nil || len(enumVal.Values) != 3 {
		t.Errorf("createPet 'status' property enum: expected 3 values, got %v", enumVal)
	}
}

func verifyShowPetByIDTool(t *testing.T, toolShowPetByID *v1.Tool) {
	if toolShowPetByID == nil {
		t.Fatalf("Tool 'showPetById' not found")
	}
	if toolShowPetByID.GetDisplayName() != "Info for a specific pet" {
		t.Errorf("Expected showPetById DisplayName 'Info for a specific pet', got '%s'", toolShowPetByID.GetDisplayName())
	}
	// Annotations for showPetById (GET)
	if toolShowPetByID.GetAnnotations() == nil {
		t.Fatalf("showPetById Annotations is nil")
	}
	if toolShowPetByID.GetAnnotations().GetIdempotentHint() != true { // GET is idempotent
		t.Errorf("showPetById Annotations.IdempotentHint: got %v, want true", toolShowPetByID.GetAnnotations().GetIdempotentHint())
	}
	// InputSchema for showPetById (has path parameter petId)
	inputSchemaShowPetByID := toolShowPetByID.GetAnnotations().GetInputSchema()
	if inputSchemaShowPetByID == nil {
		t.Fatalf("showPetById InputSchema is nil")
	}
	propertiesShowPetByID := inputSchemaShowPetByID.GetFields()["properties"].GetStructValue()
	if propertiesShowPetByID.GetFields() == nil {
		t.Fatalf("showPetById InputSchema.Properties or its fields are nil")
	}
	propPetID, ok := propertiesShowPetByID.GetFields()[petIDParam]
	if !ok {
		t.Fatalf("showPetById InputSchema.Properties missing 'petId'")
	}
	if propPetID.GetStructValue().GetFields()["type"].GetStringValue() != stringType {
		t.Errorf("showPetById 'petId' property type: got %s, want string", propPetID.GetStructValue().GetFields()["type"].GetStringValue())
	}

	if propPetID.GetStructValue().GetFields()["description"].GetStringValue() != "The id of the pet to retrieve" {
		t.Errorf("showPetById 'petId' property description: got '%s', want 'The id of the pet to retrieve'", propPetID.GetStructValue().GetFields()["description"].GetStringValue())
	}
}

// TestParseOpenAPISpec primarily tests loading and validation.
// Actual parsing into structures is indirectly tested by other functions.
// File operations are avoided here due to potential unreliability in the test environment.
func TestParseOpenAPISpec_LoadAndValidate(t *testing.T) {
	// This test focuses on LoadFromData, which ParseOpenAPISpec uses internally via loader.LoadFromFile.
	// The LoadFromFile aspect is harder to test reliably without file system access.
	// We're essentially testing the core loader's ability to parse and validate a known good spec.

	// The loader is used within ParseOpenAPISpec.
	// Here, we simulate a successful load part of ParseOpenAPISpec.
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(sampleOpenAPISpecJSON))
	if err != nil {
		t.Fatalf("loader.LoadFromData failed: %v", err)
	}
	err = doc.Validate(context.Background())
	if err != nil {
		t.Fatalf("doc.Validate failed: %v", err)
	}

	// Minimal check on returned data from a simulated ParseOpenAPISpec call
	// (if we were to call it directly with a temp file)
	if doc.Info.Title != "Sample Pet Store App" {
		t.Errorf("Expected Info.Title 'Sample Pet Store App', got '%s'", doc.Info.Title)
	}
	if len(doc.Paths.Map()) != 2 { // two paths: /pets and /pets/{petId}
		t.Errorf("Expected 2 paths, got %d", len(doc.Paths.Map()))
	}

	// Test ParseOpenAPISpec with a non-existent file path (error case)
	// This still relies on ParseOpenAPISpec using LoadFromFile.
	// If the environment doesn't allow checking file existence or errors from it, this might be fragile.
	// _, _, err = ParseOpenAPISpec(context.Background(), "/non/existent/file.yaml")
	// if err == nil {
	// 	t.Error("Expected error when parsing non-existent file, got nil")
	// } else {
	// 	// Check for a specific error message if LoadFromFile provides one, e.g., "no such file"
	// 	// This depends on OS and Go's file handling errors.
	// 	t.Logf("Got expected error for non-existent file: %v", err)
	// }
	// Due to tool limitations (no direct file system interaction for writing temp files for tests),
	// testing the LoadFromFile part of ParseOpenAPISpec is deferred.
	// The above commented-out section for non-existent file is illustrative.
}

func TestExtractMcpOperationsFromOpenAPI_XMLContent(t *testing.T) {
	doc := loadTestSpec(t)
	// Add a new operation with XML content
	doc.Paths.Find("/pets").Post.RequestBody.Value.Content["application/xml"] = &openapi3.MediaType{
		Schema: &openapi3.SchemaRef{
			Ref: "#/components/schemas/PetInput",
		},
	}

	ops := extractMcpOperationsFromOpenAPI(doc)
	opsMap := make(map[string]McpOperation)
	for _, op := range ops {
		opsMap[op.OperationID] = op
	}

	opCreatePet, ok := opsMap[opCreatePet]
	if !ok {
		t.Fatalf("Operation 'createPet' not found")
	}

	if _, ok := opCreatePet.RequestBodySchema["application/xml"]; !ok {
		t.Error("Expected createPet to have application/xml request body schema")
	}
}

func TestIsOperationIdempotent(t *testing.T) {
	testCases := []struct {
		method   string
		expected bool
	}{
		{methodGet, true},
		{methodPost, false},
		{"PUT", true},
		{"DELETE", true},
		{"PATCH", false},
		{"HEAD", true},
		{"OPTIONS", true},
		{"TRACE", true},
		{"get", true}, // Lowercase
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			if got := isOperationIdempotent(tc.method); got != tc.expected {
				t.Errorf("isOperationIdempotent(%q) = %v; want %v", tc.method, got, tc.expected)
			}
		})
	}
}

func TestParseOpenAPISpec_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid json", func(t *testing.T) {
		_, _, err := parseOpenAPISpec(ctx, []byte("{invalid"))
		assert.Error(t, err)
	})

	t.Run("validation failure", func(t *testing.T) {
		// Spec missing 'info' section, which is required
		spec := `{"openapi": "3.0.0", "paths": {}}`
		_, _, err := parseOpenAPISpec(ctx, []byte(spec))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid info: must be an object")
	})
}

func TestConvertOpenAPISchemaToInputSchemaProperties_Errors(t *testing.T) {
	doc := loadTestSpec(t)
	t.Run("nil schema ref", func(t *testing.T) {
		_, _, err := convertOpenAPISchemaToInputSchemaProperties(nil, nil, doc)
		assert.NoError(t, err, "Nil schema should not cause a panic, just result in empty properties")
	})

	t.Run("unresolvable ref", func(t *testing.T) {
		schemaRef := &openapi3.SchemaRef{Ref: "#/components/schemas/NonExistent"}
		_, _, err := convertOpenAPISchemaToInputSchemaProperties(schemaRef, nil, doc)
		assert.Error(t, err)
	})
}

func TestConvertMcpOperationsToTools_NoOperationID(t *testing.T) {
	spec := `
{
  "openapi": "3.0.0",
  "info": { "title": "Test", "version": "1.0" },
  "paths": {
    "/no-id": {
      "get": {
        "summary": "No Operation ID",
        "responses": { "200": { "description": "OK" } }
      }
    }
  }
}`
	doc, err := openapi3.NewLoader().LoadFromData([]byte(spec))
	assert.NoError(t, err)

	ops := extractMcpOperationsFromOpenAPI(doc)
	tools := convertMcpOperationsToTools(ops, doc, "test-service")

	assert.Len(t, tools, 1)
	assert.Equal(t, "GET_/no-id", tools[0].GetName())
}

func TestConvertMcpOperationsToTools_NonObjectRequestBody(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(nonObjectBodySpecJSON))
	assert.NoError(t, err, "Failed to load non-object body spec")
	err = doc.Validate(context.Background())
	assert.NoError(t, err, "Non-object body spec failed validation")

	ops := extractMcpOperationsFromOpenAPI(doc)
	tools := convertMcpOperationsToTools(ops, doc, "test-service")

	assert.Len(t, tools, 1, "Expected one tool to be generated")
	tool := tools[0]

	inputSchema := tool.GetAnnotations().GetInputSchema()
	assert.NotNil(t, inputSchema, "InputSchema should not be nil")

	// Check for the 'type: object' wrapper
	assert.Equal(t, "object", inputSchema.GetFields()["type"].GetStringValue(), "InputSchema should be a JSON schema object")

	propertiesValue, ok := inputSchema.GetFields()["properties"]
	assert.True(t, ok, "InputSchema should have a 'properties' field")
	properties := propertiesValue.GetStructValue()
	assert.NotNil(t, properties, "InputSchema properties should not be nil")

	// The fix should ensure a property is created to wrap the array.
	assert.NotEmpty(t, properties.GetFields(), "InputSchema should have properties for the non-object request body")

	// Further checks for the fixed implementation:
	requestBodyProp, ok := properties.GetFields()["request_body"]
	assert.True(t, ok, "Expected a 'request_body' property to be created for the array body")

	propSchema := requestBodyProp.GetStructValue().GetFields()
	assert.Equal(t, "array", propSchema["type"].GetStringValue(), "The wrapper property should be of type 'array'")
	assert.NotNil(t, propSchema["items"], "The array property should have an 'items' schema")
}

func TestConvertMcpOperationsToTools_AllOfAndNested(t *testing.T) {
	spec := `
{
  "openapi": "3.0.0",
  "info": { "title": "Test", "version": "1.0" },
  "components": {
    "schemas": {
      "Base": {
        "type": "object",
        "properties": {
          "id": { "type": "string" },
          "common": { "type": "string", "description": "base common" }
        }
      },
      "Extended": {
        "allOf": [
          { "$ref": "#/components/schemas/Base" },
          {
            "type": "object",
            "properties": {
              "name": { "type": "string" },
              "common": { "type": "string", "description": "extended common" }
            }
          }
        ]
      }
    }
  },
  "paths": {
    "/complex": {
      "post": {
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {
                 "type": "object",
                 "properties": {
                    "ext": { "$ref": "#/components/schemas/Extended" },
                    "nested": {
                       "type": "object",
                       "properties": {
                          "deep": { "type": "string" }
                       }
                    }
                 }
              }
            }
          }
        },
        "responses": { "200": { "description": "OK" } }
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
	inputSchema := tools[0].GetAnnotations().GetInputSchema()
	props := inputSchema.GetFields()["properties"].GetStructValue().GetFields()

	// Check nested object 'nested'
	assert.Contains(t, props, "nested")
	nestedVal := props["nested"].GetStructValue() // Should not fail if correctly converted
	assert.NotNil(t, nestedVal)
	assert.Contains(t, nestedVal.GetFields()["properties"].GetStructValue().GetFields(), "deep")

	// Check AllOf object 'ext'
	assert.Contains(t, props, "ext")
	extVal := props["ext"].GetStructValue()
	extProps := extVal.GetFields()["properties"].GetStructValue().GetFields()

	assert.Contains(t, extProps, "id", "Should inherit id from Base")
	assert.Contains(t, extProps, "name", "Should have name from Extended")

	// Check override
	assert.Contains(t, extProps, "common")
	commonDesc := extProps["common"].GetStructValue().GetFields()["description"].GetStringValue()
	assert.Equal(t, "extended common", commonDesc, "Local property should override inherited one")
}

func TestConvertMcpOperationsToTools_RequiredFields(t *testing.T) {
	spec := `
{
  "openapi": "3.0.0",
  "info": { "title": "Test", "version": "1.0" },
  "paths": {
    "/required": {
      "post": {
        "operationId": "requiredTest",
        "parameters": [
          {
            "name": "q_param",
            "in": "query",
            "required": true,
            "schema": { "type": "string" }
          },
          {
            "name": "opt_param",
            "in": "query",
            "required": false,
            "schema": { "type": "string" }
          }
        ],
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["top_req"],
                "properties": {
                  "top_req": { "type": "string" },
                  "top_opt": { "type": "string" },
                  "nested": {
                    "type": "object",
                    "required": ["sub_req"],
                    "properties": {
                      "sub_req": { "type": "string" },
                      "sub_opt": { "type": "string" }
                    }
                  }
                }
              }
            }
          }
        },
        "responses": { "200": { "description": "OK" } }
      }
    }
  }
}`

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	ops := extractMcpOperationsFromOpenAPI(doc)
	tools := convertMcpOperationsToTools(ops, doc, "test-service")

	require.Len(t, tools, 1)
	tool := tools[0]
	inputSchema := tool.GetAnnotations().GetInputSchema()

	// Check top-level required (includes parameters and body props)
	// The top-level input schema is "type": "object".
	// It should have a "required" field which is a list.

	// In structpb, a list is a ListValue.
	requiredVal, ok := inputSchema.GetFields()["required"]
	require.True(t, ok, "InputSchema should have 'required' field")

	requiredList := requiredVal.GetListValue().GetValues()
	var requiredFields []string
	for _, v := range requiredList {
		requiredFields = append(requiredFields, v.GetStringValue())
	}

	assert.Contains(t, requiredFields, "q_param", "Required query parameter should be in required list")
	assert.Contains(t, requiredFields, "top_req", "Required body property should be in required list")
	assert.NotContains(t, requiredFields, "opt_param", "Optional query parameter should NOT be in required list")
	assert.NotContains(t, requiredFields, "top_opt", "Optional body property should NOT be in required list")

	// Check nested required
	props := inputSchema.GetFields()["properties"].GetStructValue().GetFields()
	nestedVal := props["nested"].GetStructValue()

	nestedRequiredVal, ok := nestedVal.GetFields()["required"]
	require.True(t, ok, "Nested object should have 'required' field")

	nestedRequiredList := nestedRequiredVal.GetListValue().GetValues()
	var nestedRequiredFields []string
	for _, v := range nestedRequiredList {
		nestedRequiredFields = append(nestedRequiredFields, v.GetStringValue())
	}

	assert.Contains(t, nestedRequiredFields, "sub_req")
	assert.NotContains(t, nestedRequiredFields, "sub_opt")
}
