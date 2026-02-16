// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestHandleTemplates_Security(t *testing.T) {
	app, _ := setupApiTestApp()
	handler := app.handleTemplates()

	t.Run("Unauthorized Create", func(t *testing.T) {
		// Non-admin user tries to create a template
		tmpl := configv1.ServiceTemplate_builder{
			Id:   proto.String("malicious-tmpl"),
			Name: proto.String("Malicious Template"),
		}.Build()

		opts := protojson.MarshalOptions{UseProtoNames: true}
		body, _ := opts.Marshal(tmpl)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/service-templates", bytes.NewReader(body))

		// Simulate Authenticated User: regular-user
		ctx := auth.ContextWithUser(req.Context(), "regular-user")
		ctx = auth.ContextWithRoles(ctx, []string{"user"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Expect Forbidden
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected 403 Forbidden, got %d", w.Code)
		}
	})

	t.Run("XSS Sanitization", func(t *testing.T) {
		// Admin creates a template with XSS payload
		payload := "<script>alert('xss')</script>"
		tmpl := configv1.ServiceTemplate_builder{
			Id:          proto.String("xss-tmpl"),
			Name:        proto.String("XSS Template " + payload),
			Description: proto.String("Description " + payload),
			Icon:        proto.String("icon " + payload),
		}.Build()

		opts := protojson.MarshalOptions{UseProtoNames: true}
		body, _ := opts.Marshal(tmpl)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/service-templates", bytes.NewReader(body))

		// Simulate Authenticated User: admin
		ctx := auth.ContextWithUser(req.Context(), "admin")
		ctx = auth.ContextWithRoles(ctx, []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Expect Created (if we allow admin to create sanitized content)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Verify content in storage
		reqGet := httptest.NewRequest(http.MethodGet, "/api/v1/service-templates", nil)
		wGet := httptest.NewRecorder()
		handler.ServeHTTP(wGet, reqGet)

		var list []map[string]any
		_ = json.Unmarshal(wGet.Body.Bytes(), &list)

		found := false
		for _, item := range list {
			if item["id"] == "xss-tmpl" {
				found = true
				name := item["name"].(string)
				desc := item["description"].(string)
				icon := item["icon"].(string)

				if name == "XSS Template "+payload || desc == "Description "+payload || icon == "icon "+payload {
					t.Errorf("XSS payload NOT sanitized! Name: %s, Desc: %s, Icon: %s", name, desc, icon)
				} else {
					// Verify it is encoded
					assert.Contains(t, name, "&lt;script&gt;")
					assert.Contains(t, desc, "&lt;script&gt;")
				}
			}
		}
		if !found {
			t.Errorf("Created template not found")
		}
	})
}

func TestHandleTemplateDetail_Security(t *testing.T) {
	app, _ := setupApiTestApp()
	handler := app.handleTemplateDetail()

	t.Run("Unauthorized Delete", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/service-templates/some-id", nil)

		// Simulate Authenticated User: regular-user
		ctx := auth.ContextWithUser(req.Context(), "regular-user")
		ctx = auth.ContextWithRoles(ctx, []string{"user"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Expect Forbidden
		if w.Code != http.StatusForbidden {
			t.Errorf("Expected 403 Forbidden, got %d", w.Code)
		}
	})
}
