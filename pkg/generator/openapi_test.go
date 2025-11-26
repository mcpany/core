// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package generator_test

import (
	"testing"

	"github.com/mcpany/core/pkg/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleOpenAPIV2Spec = `
swagger: "2.0"
info:
  title: "Echo Service"
  version: "1.0.0"
host: "localhost:8080"
schemes:
  - "http"
paths:
  /echo:
    post:
      operationId: "echo"
      summary: "Echoes a message"
      parameters:
        - in: "body"
          name: "body"
          required: true
          schema:
            type: "object"
            properties:
              message:
                type: "string"
      responses:
        "200":
          description: "Successful response"
`

const sampleOpenAPIV3Spec = `
openapi: 3.0.0
info:
  title: Petstore API
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /pets:
    get:
      operationId: listPets
      summary: List all pets
      responses:
        '200':
          description: A paged array of pets
`

const malformedSpec = `
this is not a valid openapi spec
`

const specMissingServer = `
openapi: 3.0.0
info:
  title: Petstore API
  version: 1.0.0
paths:
  /pets:
    get:
      operationId: listPets
      summary: List all pets
      responses:
        '200':
          description: A paged array of pets
`

func TestGenerateMCPConfigFromOpenAPI(t *testing.T) {
	t.Run("valid v2 spec", func(t *testing.T) {
		output, err := generator.GenerateMCPConfigFromOpenAPI([]byte(sampleOpenAPIV2Spec), "test.yaml")
		require.NoError(t, err)

		expectedOutput := `upstreamServices:
- httpService:
    address: http://localhost:8080
    calls:
      echo:
        endpointPath: /echo
        method: HTTP_METHOD_POST
  name: Echo Service
`
		assert.Equal(t, expectedOutput, string(output))
	})

	t.Run("valid v3 spec", func(t *testing.T) {
		output, err := generator.GenerateMCPConfigFromOpenAPI([]byte(sampleOpenAPIV3Spec), "test.yaml")
		require.NoError(t, err)

		expectedOutput := `upstreamServices:
- httpService:
    address: http://localhost:8080
    calls:
      listPets:
        endpointPath: /pets
        method: HTTP_METHOD_GET
  name: Petstore API
`
		assert.Equal(t, expectedOutput, string(output))
	})

	t.Run("malformed spec", func(t *testing.T) {
		_, err := generator.GenerateMCPConfigFromOpenAPI([]byte(malformedSpec), "test.yaml")
		require.Error(t, err)
	})

	t.Run("spec missing server", func(t *testing.T) {
		_, err := generator.GenerateMCPConfigFromOpenAPI([]byte(specMissingServer), "test.yaml")
		require.Error(t, err)
	})
}
