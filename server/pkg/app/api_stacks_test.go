// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleStackConfig_Get(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleStackConfig(store)

	// Setup: Create a stack (collection)
	stackID := "test-stack"
	collection := configv1.Collection_builder{
		Name: proto.String(stackID),
		Services: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: proto.String("svc1"),
				HttpService: configv1.HttpUpstreamService_builder{
					Address: proto.String("http://example.com"),
				}.Build(),
			}.Build(),
		},
	}.Build()
	require.NoError(t, store.SaveServiceCollection(context.Background(), collection))

	// Test GET
	req := httptest.NewRequest(http.MethodGet, "/stacks/test-stack/config", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")

	body := w.Body.String()
	assert.Contains(t, body, "name: test-stack")
	assert.Contains(t, body, "svc1")
	assert.Contains(t, body, "http_service")
}

func TestHandleStackConfig_Post(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleStackConfig(store)

	stackID := "new-stack"
	yamlConfig := `
name: new-stack
services:
  - name: svc2
    http_service:
      address: http://example.org
`

	// Test POST
	req := httptest.NewRequest(http.MethodPost, "/stacks/new-stack/config", strings.NewReader(yamlConfig))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify update
	collection, err := store.GetServiceCollection(context.Background(), stackID)
	require.NoError(t, err)
	require.NotNil(t, collection)
	assert.Equal(t, stackID, collection.GetName())
	require.Len(t, collection.GetServices(), 1)
	assert.Equal(t, "svc2", collection.GetServices()[0].GetName())
	assert.Equal(t, "http://example.org", collection.GetServices()[0].GetHttpService().GetAddress())
}

func TestHandleStackConfig_Post_Invalid(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleStackConfig(store)

	t.Run("Invalid YAML", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/stacks/test-stack/config", strings.NewReader("invalid: : yaml"))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid YAML")
	})

	t.Run("ID Mismatch", func(t *testing.T) {
		yamlConfig := `
name: other-stack
services: []
`
		req := httptest.NewRequest(http.MethodPost, "/stacks/test-stack/config", strings.NewReader(yamlConfig))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Stack name in config must match URL path")
	})

	t.Run("Invalid Service Config", func(t *testing.T) {
		// Missing address for HTTP service
		yamlConfig := `
name: test-stack
services:
  - name: svc1
    http_service: {}
`
		req := httptest.NewRequest(http.MethodPost, "/stacks/test-stack/config", strings.NewReader(yamlConfig))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid service in stack")
	})
}

func TestIsUnsafeAllowed(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleStackConfig(store)

	// Unsafe config (command line service)
	unsafeYaml := `
name: unsafe-stack
services:
  - name: cmd-svc
    command_line_service:
      command: echo
`

	t.Run("Reject by Default", func(t *testing.T) {
		os.Unsetenv("MCPANY_ALLOW_UNSAFE_CONFIG")
		req := httptest.NewRequest(http.MethodPost, "/stacks/unsafe-stack/config", strings.NewReader(unsafeYaml))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "Unsafe service configuration not allowed")
	})

	t.Run("Allow via Env Var", func(t *testing.T) {
		os.Setenv("MCPANY_ALLOW_UNSAFE_CONFIG", "true")
		defer os.Unsetenv("MCPANY_ALLOW_UNSAFE_CONFIG")

		req := httptest.NewRequest(http.MethodPost, "/stacks/unsafe-stack/config", strings.NewReader(unsafeYaml))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Allow via Admin Role", func(t *testing.T) {
		os.Unsetenv("MCPANY_ALLOW_UNSAFE_CONFIG")

		req := httptest.NewRequest(http.MethodPost, "/stacks/unsafe-stack/config", strings.NewReader(unsafeYaml))
		// Inject admin role
		ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
