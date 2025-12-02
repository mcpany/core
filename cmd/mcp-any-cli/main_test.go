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

package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI(t *testing.T) {
	t.Run("generate http service", func(t *testing.T) {
		input := "http\ntest-http-service\nhttp://localhost:8080\napiKey\nX-API-Key\nmy-api-key\ny\nget-user\nGet a user\nGET\n/users/{id}\ny\nid\nuserId\nn\nn\n"
		expectedOutput := `upstreamServices:
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

		oldStdin := os.Stdin
		defer func() { os.Stdin = oldStdin }()

		r, w, err := os.Pipe()
		require.NoError(t, err)

		_, err = w.WriteString(input)
		require.NoError(t, err)
		w.Close()

		os.Stdin = r

		oldStdout := os.Stdout
		defer func() { os.Stdout = oldStdout }()

		r, w, err = os.Pipe()
		require.NoError(t, err)

		os.Stdout = w

		err = run()
		require.NoError(t, err)

		w.Close()

		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		require.NoError(t, err)

		output := strings.TrimSpace(buf.String())
		assert.Contains(t, output, "Generated configuration:")
		assert.YAMLEq(t, expectedOutput, strings.Split(output, "Generated configuration:")[1])
	})
}
