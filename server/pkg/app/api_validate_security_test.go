// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHandleServiceValidate_Unsafe(t *testing.T) {
	app := &Application{}

	// Setup an unsafe filesystem configuration
	osFs := &configv1.OsFs{}
	fsSvc := &configv1.FilesystemUpstreamService{}
	fsSvc.SetRootPaths(map[string]string{"/": "/"})
	fsSvc.SetOs(osFs)

	svc := &configv1.UpstreamServiceConfig{}
	svc.SetName("unsafe-fs-validate")
	svc.SetFilesystemService(fsSvc)

	body, _ := protojson.Marshal(svc)

	t.Run("Forbid Unsafe Validation (Regular User)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()

		app.handleServiceValidate().ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Allow Unsafe Validation (Admin User)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
		// Inject admin role
		ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		app.handleServiceValidate().ServeHTTP(w, req)

		assert.NotEqual(t, http.StatusForbidden, w.Code)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
