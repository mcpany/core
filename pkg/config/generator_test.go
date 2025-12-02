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

package config

import (
	"bufio"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestGenerator(t *testing.T) {
	t.Run("generateHTTPService", func(t *testing.T) {
		input := "test-http-service\nhttp://localhost:8080\napiKey\nX-API-Key\nmy-api-key\ny\nget-user\nGet a user\nGET\n/users/{id}\ny\nid\nuserId\nn\nn\n"
		reader := bufio.NewReader(strings.NewReader(input))
		generator := &Generator{reader: reader}

		configBytes, err := generator.generateHTTPService()
		require.NoError(t, err)

		var actualConfig Config
		err = yaml.Unmarshal(configBytes, &actualConfig)
		require.NoError(t, err)

		expectedYAML := `upstreamServices:
    - name: test-http-service
      httpService:
        address: http://localhost:8080
        calls:
            - operationId: get-user
              description: Get a user
              method: GET
              endpointPath: /users/{id}
              parameterMappings:
                - inputParameterName: id
                  targetParameterName: userId
      upstreamAuthentication:
        apiKey:
            headerName: X-API-Key
            apiKey:
                plainText: my-api-key
`
		var expectedConfig Config
		err = yaml.Unmarshal([]byte(expectedYAML), &expectedConfig)
		require.NoError(t, err)

		assert.Equal(t, expectedConfig, actualConfig)
	})

	t.Run("generateGRPCService", func(t *testing.T) {
		input := "test-grpc-service\nlocalhost:50051\ny\n"
		reader := bufio.NewReader(strings.NewReader(input))
		generator := &Generator{reader: reader}

		configBytes, err := generator.generateGRPCService()
		require.NoError(t, err)

		var actualConfig Config
		err = yaml.Unmarshal(configBytes, &actualConfig)
		require.NoError(t, err)

		expectedYAML := `upstreamServices:
    - name: test-grpc-service
      grpcService:
        address: localhost:50051
        reflection:
            enabled: true
`
		var expectedConfig Config
		err = yaml.Unmarshal([]byte(expectedYAML), &expectedConfig)
		require.NoError(t, err)

		assert.Equal(t, expectedConfig, actualConfig)
	})

	t.Run("generateOpenAPIService", func(t *testing.T) {
		input := "test-openapi-service\n./openapi.json\n"
		reader := bufio.NewReader(strings.NewReader(input))
		generator := &Generator{reader: reader}

		configBytes, err := generator.generateOpenAPIService()
		require.NoError(t, err)

		var actualConfig Config
		err = yaml.Unmarshal(configBytes, &actualConfig)
		require.NoError(t, err)

		expectedYAML := `upstreamServices:
    - name: test-openapi-service
      openapiService:
        spec:
            path: ./openapi.json
`
		var expectedConfig Config
		err = yaml.Unmarshal([]byte(expectedYAML), &expectedConfig)
		require.NoError(t, err)

		assert.Equal(t, expectedConfig, actualConfig)
	})

	t.Run("generateGraphQLService", func(t *testing.T) {
		input := "test-graphql-service\nhttp://localhost:8080/graphql\ny\nuser\n{ id name }\nn\n"
		reader := bufio.NewReader(strings.NewReader(input))
		generator := &Generator{reader: reader}

		configBytes, err := generator.generateGraphQLService()
		require.NoError(t, err)

		var actualConfig Config
		err = yaml.Unmarshal(configBytes, &actualConfig)
		require.NoError(t, err)

		expectedYAML := `upstreamServices:
    - name: test-graphql-service
      graphqlService:
        address: http://localhost:8080/graphql
        calls:
            - name: user
              selectionSet: "{ id name }"
`
		var expectedConfig Config
		err = yaml.Unmarshal([]byte(expectedYAML), &expectedConfig)
		require.NoError(t, err)

		assert.Equal(t, expectedConfig, actualConfig)
	})
}
