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
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheck(t *testing.T) {
	t.Parallel()

	t.Run("should return no error when the server is healthy", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		addr := strings.TrimPrefix(server.URL, "http://")
		var out bytes.Buffer
		err := HealthCheck(&out, addr, 1*time.Second)
		assert.NoError(t, err, "HealthCheck should not return an error for a healthy server")
		assert.Contains(t, out.String(), "Health check successful")
	})

	t.Run("should return an error when the server is unhealthy", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		addr := strings.TrimPrefix(server.URL, "http://")
		var out bytes.Buffer
		err := HealthCheck(&out, addr, 1*time.Second)
		assert.Error(t, err, "HealthCheck should return an error for an unhealthy server")
	})

	t.Run("should return an error when the server is unreachable", func(t *testing.T) {
		t.Parallel()
		var out bytes.Buffer
		// Use a port that is unlikely to be in use
		err := HealthCheck(&out, "localhost:9999", 1*time.Second)
		assert.Error(t, err, "HealthCheck should return an error for an unreachable server")
	})

	t.Run("should return an error when the context times out", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		addr := strings.TrimPrefix(server.URL, "http://")
		var out bytes.Buffer

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := HealthCheckWithContext(ctx, &out, addr)
		require.Error(t, err)
		assert.Contains(
			t,
			err.Error(),
			"context deadline exceeded",
			"Expected timeout error due to context cancellation",
		)
	})
	t.Run("should write success message to the provided writer", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		addr := strings.TrimPrefix(server.URL, "http://")
		var out bytes.Buffer
		err := HealthCheck(&out, addr, 1*time.Second)
		require.NoError(t, err)

		expectedOutput := "Health check successful: server is running and healthy."
		assert.Contains(t, out.String(), expectedOutput, fmt.Sprintf("Expected output to contain %q, but got %q", expectedOutput, out.String()))
	})
}
