/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package http

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculator(t *testing.T) {
	c := NewCalculator()
	defer c.Close()

	client := &http.Client{}

	type testCase struct {
		name           string
		method         string
		path           string
		body           string
		contentType    string
		headers        map[string]string
		expectedStatus int
		expectedBody   string
		isJson         bool
	}

	runTest := func(t *testing.T, tc testCase) {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, c.URL()+tc.path, strings.NewReader(tc.body))
			require.NoError(t, err)

			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedBody != "" {
				respBody, _ := io.ReadAll(resp.Body)
				if tc.isJson {
					assert.JSONEq(t, tc.expectedBody, string(respBody))
				} else {
					assert.Equal(t, tc.expectedBody, string(respBody))
				}
			}
		})
	}

	addTests := []testCase{
		{"success", http.MethodPost, "/add", `{"a": 1, "b": 2}`, "application/json", nil, http.StatusOK, `{"result": 3}`, true},
		{"wrong_method", http.MethodGet, "/add", "", "", nil, http.StatusMethodNotAllowed, "", true},
		{"bad_json", http.MethodPost, "/add", `{"a": 1, "b": "two"}`, "application/json", nil, http.StatusBadRequest, "", true},
	}
	for _, tc := range addTests {
		runTest(t, tc)
	}

	subtractTests := []testCase{
		{"success", http.MethodGet, "/subtract?a=5&b=3", "", "", nil, http.StatusOK, `{"result": 2}`, true},
		{"bad_params", http.MethodGet, "/subtract?a=five&b=3", "", "", nil, http.StatusBadRequest, "", true},
	}
	for _, tc := range subtractTests {
		runTest(t, tc)
	}

	multiplyTests := []testCase{
		{"success", http.MethodPut, "/multiply", `{"a": 4, "b": 5}`, "application/json", nil, http.StatusOK, `{"result": 20}`, true},
	}
	for _, tc := range multiplyTests {
		runTest(t, tc)
	}

	divideTests := []testCase{
		{"success", http.MethodDelete, "/divide", `{"a": 10, "b": 2}`, "application/json", nil, http.StatusOK, `{"result": 5}`, true},
		{"by_zero", http.MethodDelete, "/divide", `{"a": 10, "b": 0}`, "application/json", nil, http.StatusBadRequest, "", true},
	}
	for _, tc := range divideTests {
		runTest(t, tc)
	}

	powerTests := []testCase{
		{"success", http.MethodPatch, "/power", `{"a": 2, "b": 3}`, "application/json", nil, http.StatusOK, `{"result": 8}`, true},
	}
	for _, tc := range powerTests {
		runTest(t, tc)
	}

	pathAddTests := []testCase{
		{"success", http.MethodGet, "/path/add/10/5", "", "", nil, http.StatusOK, `{"result": 15}`, true},
		{"bad_path", http.MethodGet, "/path/add/ten/5", "", "", nil, http.StatusBadRequest, "", true},
	}
	for _, tc := range pathAddTests {
		runTest(t, tc)
	}

	formAddTests := []testCase{
		{"success", http.MethodPost, "/form/add", "a=10&b=5", "application/x-www-form-urlencoded", nil, http.StatusOK, `{"result": 15}`, true},
		{"bad_form", http.MethodPost, "/form/add", "a=ten&b=5", "application/x-www-form-urlencoded", nil, http.StatusBadRequest, "", true},
	}
	for _, tc := range formAddTests {
		runTest(t, tc)
	}

	templateHelloTests := []testCase{
		{"success", http.MethodGet, "/template/hello?name=Jules", "", "", nil, http.StatusOK, `{"result": "Hello, Jules!"}`, true},
	}
	for _, tc := range templateHelloTests {
		runTest(t, tc)
	}

	templateAddTests := []testCase{
		{"success", http.MethodPost, "/template/add", "add 10 and 5", "text/plain", nil, http.StatusOK, `{"result": 15}`, true},
		{"failure", http.MethodPost, "/template/add", "add ten and 5", "text/plain", nil, http.StatusBadRequest, "", true},
	}
	for _, tc := range templateAddTests {
		runTest(t, tc)
	}

	apiKeyTests := []testCase{
		{"success", http.MethodGet, "/apikey", "", "", map[string]string{"X-API-Key": "secret"}, http.StatusOK, `{"result": "ok"}`, true},
		{"failure", http.MethodGet, "/apikey", "", "", map[string]string{"X-API-Key": "wrong-secret"}, http.StatusUnauthorized, "", true},
	}
	for _, tc := range apiKeyTests {
		runTest(t, tc)
	}

	bearerTokenTests := []testCase{
		{"success", http.MethodGet, "/bearertoken", "", "", map[string]string{"Authorization": "Bearer secret"}, http.StatusOK, `{"result": "ok"}`, true},
		{"wrong_token", http.MethodGet, "/bearertoken", "", "", map[string]string{"Authorization": "Bearer wrong-secret"}, http.StatusUnauthorized, "", true},
		{"no_bearer_prefix", http.MethodGet, "/bearertoken", "", "", map[string]string{"Authorization": "secret"}, http.StatusUnauthorized, "", true},
	}
	for _, tc := range bearerTokenTests {
		runTest(t, tc)
	}

	t.Run("echo", func(t *testing.T) {
		template := `{"hello": "{{name}}"}`
		rendered := strings.Replace(template, "{{name}}", "world", -1)
		runTest(t, testCase{
			name:           "template_rendering",
			method:         http.MethodPost,
			path:           "/echo",
			body:           rendered,
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			expectedBody:   rendered,
			isJson:         false,
		})
	})
}
