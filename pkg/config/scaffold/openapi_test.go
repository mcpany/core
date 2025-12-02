// Copyright 2024 Author(s) of MCP Any
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

package scaffold

import (
	"os"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAPIScaffolder_Scaffold(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		spec := &openapi3.T{
			Info: &openapi3.Info{
				Title: "Test API",
			},
			Servers: openapi3.Servers{
				{
					URL: "http://localhost:8080",
				},
			},
		}

		scaffolder := NewOpenAPIScaffolder(spec)
		config, err := scaffolder.Scaffold()
		require.NoError(t, err)

		require.NotNil(t, config)
		require.Len(t, config.GetUpstreamServices(), 1)

		service := config.GetUpstreamServices()[0]
		assert.Equal(t, "Test API", service.GetName())
		assert.NotNil(t, service.GetOpenapiService())
		assert.Equal(t, "http://localhost:8080", service.GetOpenapiService().GetAddress())
	})

	t.Run("nil info", func(t *testing.T) {
		spec := &openapi3.T{
			Servers: openapi3.Servers{
				{
					URL: "http://localhost:8080",
				},
			},
		}

		scaffolder := NewOpenAPIScaffolder(spec)
		config, err := scaffolder.Scaffold()
		require.NoError(t, err)

		require.NotNil(t, config)
		require.Len(t, config.GetUpstreamServices(), 1)

		service := config.GetUpstreamServices()[0]
		assert.Equal(t, "", service.GetName())
	})
}

func TestScaffoldFile(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// Create a temporary OpenAPI spec file
		openapiFile, err := os.CreateTemp("", "openapi-*.yaml")
		require.NoError(t, err)
		defer os.Remove(openapiFile.Name())

		_, err = openapiFile.WriteString(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths: {}
`)
		require.NoError(t, err)
		openapiFile.Close()

		// Create a temporary output file
		outputFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(outputFile.Name())
		outputFile.Close()

		err = ScaffoldFile(openapiFile.Name(), outputFile.Name())
		require.NoError(t, err)

		// Read the output file and check its contents
		data, err := os.ReadFile(outputFile.Name())
		require.NoError(t, err)
		assert.Contains(t, string(data), "name: Test API")
		assert.Contains(t, string(data), "address: http://localhost:8080")
	})

	t.Run("invalid openapi file path", func(t *testing.T) {
		err := ScaffoldFile("/invalid/path", "config.yaml")
		require.Error(t, err)
	})

	t.Run("valid openapi spec with paths", func(t *testing.T) {
		// Create a temporary OpenAPI spec file
		openapiFile, err := os.CreateTemp("", "openapi-*.yaml")
		require.NoError(t, err)
		defer os.Remove(openapiFile.Name())

		_, err = openapiFile.WriteString(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /users:
    get:
      summary: Get Users
      responses:
        '200':
          description: OK
`)
		require.NoError(t, err)
		openapiFile.Close()

		// Create a temporary output file
		outputFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(outputFile.Name())
		outputFile.Close()

		err = ScaffoldFile(openapiFile.Name(), outputFile.Name())
		require.NoError(t, err)
	})

	t.Run("malformed openapi spec", func(t *testing.T) {
		// Create a temporary OpenAPI spec file
		openapiFile, err := os.CreateTemp("", "openapi-*.yaml")
		require.NoError(t, err)
		defer os.Remove(openapiFile.Name())

		_, err = openapiFile.WriteString(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: http://localhost:8080
paths:
  /users:
    get:
      summary: Get Users
      responses:
        '200':
          description: OK
  malformed
`)
		require.NoError(t, err)
		openapiFile.Close()

		// Create a temporary output file
		outputFile, err := os.CreateTemp("", "config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(outputFile.Name())
		outputFile.Close()

		err = ScaffoldFile(openapiFile.Name(), outputFile.Name())
		require.Error(t, err)
	})
}
