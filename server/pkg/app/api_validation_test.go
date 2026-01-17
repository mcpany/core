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
}
