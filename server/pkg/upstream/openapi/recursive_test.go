// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package openapi

import (
	"context"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestConvertMcpOperationsToTools_RecursiveSchema(t *testing.T) {
	// Define a recursive schema: a Category with a list of sub-Categories.
	recursiveSpec := `
{
  "openapi": "3.0.0",
  "info": {
    "title": "Recursive Schema Test",
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
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(recursiveSpec))
	assert.NoError(t, err)

	// Validate the doc to ensure references are resolved correctly by the loader
	err = doc.Validate(context.Background())
	assert.NoError(t, err)

	ops := extractMcpOperationsFromOpenAPI(doc)

	// Create a channel to signal completion
	done := make(chan bool)

	go func() {
		// This should finish quickly if recursion is handled.
		// If not, it will stack overflow or hang.
		tools := convertMcpOperationsToTools(ops, doc, "test-service")
		assert.NotEmpty(t, tools)
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Test timed out, likely due to infinite recursion")
	}
}
