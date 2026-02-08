// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

// recursiveSpecJSON defines a schema where 'Category' has a property 'subcategories'
// which is an array of 'Category'.
const recursiveSpecJSON = `
{
  "openapi": "3.0.0",
  "info": {
    "title": "Recursive Spec",
    "version": "1.0.0"
  },
  "components": {
    "schemas": {
      "Category": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string"
          },
          "subcategories": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/Category"
            }
          }
        }
      }
    }
  },
  "paths": {
    "/categories": {
      "post": {
        "operationId": "createCategory",
        "requestBody": {
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/Category"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "OK",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Category"
                }
              }
            }
          }
        }
      }
    }
  }
}
`

func TestConvertMcpOperationsToTools_RecursiveSchema(t *testing.T) {
	// Load the spec
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(recursiveSpecJSON))
	assert.NoError(t, err)

	err = doc.Validate(context.Background())
	assert.NoError(t, err)

	// Extract operations
	ops := extractMcpOperationsFromOpenAPI(doc)

	// Convert to tools - this is where we expect infinite recursion if not handled
	tools := convertMcpOperationsToTools(ops, doc, "test-service")

	assert.NotEmpty(t, tools)
	tool := tools[0]

	// Verify InputSchema
	inputSchema := tool.GetAnnotations().GetInputSchema()
	assert.NotNil(t, inputSchema)

	props := inputSchema.GetFields()["properties"].GetStructValue().GetFields()
	assert.Contains(t, props, "subcategories")

	// Check depth handling (manually inspecting structure)
	// Level 0: Category
	// Level 1: subcategories (array)
	// Level 2: items (Category)
	// Level 3: subcategories (array)
	// ...

	// We just want to ensure it didn't crash and produced *something*
	// The depth limit we plan to implement is 10, so we can check if it goes reasonably deep but stops.
}
