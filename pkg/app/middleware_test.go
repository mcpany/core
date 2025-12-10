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

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	testCases := []struct {
		name               string
		apiKey             string
		headerKey          string
		path               string
		expectedStatusCode int
	}{
		{
			name:               "No API Key Configured",
			apiKey:             "",
			headerKey:          "",
			path:               "/",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Correct API Key Provided",
			apiKey:             "my-secret-key",
			headerKey:          "my-secret-key",
			path:               "/",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Incorrect API Key Provided",
			apiKey:             "my-secret-key",
			headerKey:          "wrong-key",
			path:               "/",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "No API Key Provided When Required",
			apiKey:             "my-secret-key",
			headerKey:          "",
			path:               "/",
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:               "Metrics Endpoint Exempted",
			apiKey:             "my-secret-key",
			headerKey:          "",
			path:               "/metrics",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Healthz Endpoint Exempted",
			apiKey:             "my-secret-key",
			headerKey:          "",
			path:               "/healthz",
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			viper.Set("api-key", tc.apiKey)
			defer viper.Set("api-key", "")

			req := httptest.NewRequest("GET", tc.path, nil)
			if tc.headerKey != "" {
				req.Header.Set("X-API-Key", tc.headerKey)
			}
			rr := httptest.NewRecorder()

			handler := AuthMiddleware(nextHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatusCode, rr.Code)
		})
	}
}
