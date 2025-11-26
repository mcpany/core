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

package e2e_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleOpenAPISpec = `
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

func TestBootstrapOpenAPI(t *testing.T) {
	// Build the mcpany binary
	cmd := exec.Command("go", "build", "-o", "mcpany", "../../cmd/server")
	err := cmd.Run()
	require.NoError(t, err)
	defer os.Remove("mcpany")

	t.Run("from local file", func(t *testing.T) {
		// Create a temporary file for the OpenAPI spec
		specFile, err := os.CreateTemp("", "openapi-*.yaml")
		require.NoError(t, err)
		defer os.Remove(specFile.Name())

		_, err = specFile.WriteString(sampleOpenAPISpec)
		require.NoError(t, err)
		specFile.Close()

		// Run the bootstrap command
		cmd = exec.Command("./mcpany", "bootstrap", "openapi", specFile.Name())
		output, err := cmd.CombinedOutput()
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

	t.Run("from remote url", func(t *testing.T) {
		// Create a test server to serve the OpenAPI spec
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, sampleOpenAPISpec)
		}))
		defer server.Close()

		// Run the bootstrap command
		cmd = exec.Command("./mcpany", "bootstrap", "openapi", server.URL)
		output, err := cmd.CombinedOutput()
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
}
