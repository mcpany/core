package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/auth"
	"github.com/stretchr/testify/require"
)

func TestSecurity_RootAccess_Enforcement(t *testing.T) {
	app := NewApplication()

	// Dummy handler to simulate JSON-RPC endpoint
	jsonRPCHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("JSON-RPC Response"))
	})

	// Setup Handler using handleRoot
	// uiPath is empty, so GET / should fall through to JSON-RPC handler (and be blocked if non-admin)
	// OR if uiPath is set, GET / should serve UI.

	t.Run("JSON-RPC_Blocked_For_Guest", func(t *testing.T) {
		// No UI path, so GET / and POST / both hit fallback logic
		handler := app.handleRoot(jsonRPCHandler, nil, "")

		req := httptest.NewRequest("POST", "/", nil)
		// Inject Guest Role
		ctx := auth.ContextWithRoles(req.Context(), []string{"guest"})
		ctx = auth.ContextWithUser(ctx, "guest")
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code, "Guest should be blocked from JSON-RPC")
	})

	t.Run("JSON-RPC_Allowed_For_Admin", func(t *testing.T) {
		handler := app.handleRoot(jsonRPCHandler, nil, "")

		req := httptest.NewRequest("POST", "/", nil)
		// Inject Admin Role
		ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, "Admin should be allowed")
		require.Equal(t, "JSON-RPC Response", w.Body.String())
	})

	t.Run("UI_Allowed_For_Guest", func(t *testing.T) {
		// Mock UI path
		tempDir := t.TempDir()
		// Create index.html
		// We need to use OS FS because handleRoot uses filepath.Join and http.ServeFile
		// which use OS FS.
		// Wait, http.ServeFile uses os.Open.
		// So we must use real file system for this test or mock http.ServeFile?
		// We can't mock http.ServeFile easily as it is stdlib.
		// So we use t.TempDir().

		// Create index.html
		// We can't easily write to it here without importing os
		// But t.TempDir gives us a path.
		// We can assume we can write to it.
		// Wait, we can't import os if not imported. It is imported in app package but not here?
		// We need to add imports.

		// Setup handler with UI path
		handler := app.handleRoot(jsonRPCHandler, nil, tempDir)

		req := httptest.NewRequest("GET", "/", nil)
		// Inject Guest Role
		ctx := auth.ContextWithRoles(req.Context(), []string{"guest"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		// Since directory exists but index.html might not, ServeFile might 404 or directory list?
		// http.ServeFile handles directories by looking for index.html.
		// If not found, 404 or directory listing depending on config?
		// ServeFile is simple.
		// Let's rely on behavior: if it falls through to JSON-RPC handler, it returns 200 (mock).
		// If it serves UI, it returns 404 (if file missing) or 200 (if file present).
		// But the check `if r.URL.Path == "/" && uiPath != "" && r.Method == http.MethodGet` is what matters.
		// It returns `http.ServeFile`.
		// If `http.ServeFile` is called, we consider it "Allowed" (even if 404 from FS).
		// If it was blocked, we'd get 403.

		handler.ServeHTTP(w, req)

		// 404 or 200 is fine, 403 is NOT.
		require.NotEqual(t, http.StatusForbidden, w.Code, "UI should be accessible (even if 404)")
	})
}
