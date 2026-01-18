// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

func TestCheckURLReachability(t *testing.T) {
	t.Run("Reachable 200 OK", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		err := checkURLReachability(context.Background(), server.URL)
		assert.NoError(t, err)
	})

	t.Run("Reachable 404 (Considered Reachable)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		// 404 is considered reachable in current implementation (>= 400 && < 500 except 405/401 is complicated logic in api.go)
		// api.go:
		// if resp.StatusCode >= 400 && resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusUnauthorized {
		//   if resp.StatusCode >= 500 { ... }
		// }
		// So 404 is not 405 or 401. It is >= 400. It is < 500. So it returns nil error.
		err := checkURLReachability(context.Background(), server.URL)
		assert.NoError(t, err)
	})

	t.Run("Reachable 500 (Error)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		err := checkURLReachability(context.Background(), server.URL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "server returned error status")
	})

	t.Run("Unreachable", func(t *testing.T) {
		// Connect to a closed port
		err := checkURLReachability(context.Background(), "http://127.0.0.1:0")
		assert.Error(t, err)
	})
}

func TestHandleServiceValidate(t *testing.T) {
	app := &Application{}

	t.Run("Valid Static Config", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("http://example.com"),
				},
			},
		}
		body, _ := protojson.Marshal(svc)
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/services/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler := app.handleServiceValidate()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp, "valid")
	})

	t.Run("Invalid Static Config", func(t *testing.T) {
		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String(""), // Missing name
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String("http://example.com"),
				},
			},
		}
		body, _ := protojson.Marshal(svc)
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/services/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler := app.handleServiceValidate()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["valid"])
		assert.Contains(t, resp["error"], "service name is empty")
	})

	t.Run("Connectivity Check Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String(server.URL),
				},
			},
		}
		body, _ := protojson.Marshal(svc)
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/services/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler := app.handleServiceValidate()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, true, resp["valid"])
	})

	t.Run("Connectivity Check Failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		svc := &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String(server.URL),
				},
			},
		}
		body, _ := protojson.Marshal(svc)
		req := httptest.NewRequestWithContext(context.Background(), "POST", "/services/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler := app.handleServiceValidate()
		handler.ServeHTTP(w, req)

		// Returns 200 OK but with valid=false
		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["valid"])
		assert.Contains(t, resp["error"], "server returned error status")
	})
}
