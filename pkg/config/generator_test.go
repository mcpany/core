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
	"errors"
	"strings"
	"testing"
)

func TestGenerator_generateHTTPService(t *testing.T) {
	input := "my-service\nhttp://example.com\nget-user\nGet a user\nHTTP_METHOD_GET\n/users/{id}\n"
	reader := bufio.NewReader(strings.NewReader(input))

	g := &Generator{
		reader: reader,
	}

	output, err := g.generateHTTPService()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedOutput := `upstreamServices:
  - name: "my-service"
    httpService:
      address: "http://example.com"
      calls:
        - operationId: "get-user"
          description: "Get a user"
          method: "HTTP_METHOD_GET"
          endpointPath: "/users/{id}"
`
	if string(output) != expectedOutput {
		t.Errorf("unexpected output:\ngot:\n%s\nwant:\n%s", string(output), expectedOutput)
	}
}

func TestGenerator_Generate_unsupportedServiceType(t *testing.T) {
	input := "invalid-service-type\n"
	reader := bufio.NewReader(strings.NewReader(input))

	g := &Generator{
		reader: reader,
	}

	_, err := g.Generate()
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}

	expectedError := "unsupported service type: invalid-service-type"
	if err.Error() != expectedError {
		t.Errorf("unexpected error:\ngot:\n%s\nwant:\n%s", err.Error(), expectedError)
	}
}

type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

func TestGenerator_Generate_ioError(t *testing.T) {
	reader := bufio.NewReader(&errorReader{})

	g := &Generator{
		reader: reader,
	}

	_, err := g.Generate()
	if err == nil {
		t.Fatal("expected an error, but got nil")
	}

	expectedError := "test error"
	if err.Error() != expectedError {
		t.Errorf("unexpected error:\ngot:\n%s\nwant:\n%s", err.Error(), expectedError)
	}
}
