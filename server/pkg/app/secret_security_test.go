// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretLeak(t *testing.T) {
	// Setup SQLite DB
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_secrets.db")
	db, err := sqlite.NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := sqlite.NewStore(db)

	// Setup Application
	app := NewApplication()
	app.fs = afero.NewMemMapFs() // Use in-memory FS to avoid side effects

	handler := app.createAPIHandler(store)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	t.Run("CreateSecret_And_Retrieve_Leaked_Value", func(t *testing.T) {
		// 1. Create a Secret with a sensitive value
		secretID := "sensitive-secret-123"
		secretValue := "SUPER_SECRET_VALUE_DO_NOT_SHARE"
		body := map[string]interface{}{
			"id":   secretID,
			"name": "My Secret",
			"key":  "my_secret_key",
			"value": secretValue,
		}
		bodyBytes, _ := json.Marshal(body)

		resp, err := http.Post(ts.URL+"/secrets", "application/json", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// 2. Retrieve the Secret Detail via API
		resp, err = http.Get(ts.URL + "/secrets/" + secretID)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// 3. Check if the value is redacted
		assert.Equal(t, "[REDACTED]", result["value"], "Secret value should be redacted in detail view")
	})

	t.Run("ListSecrets_Redaction_Check", func(t *testing.T) {
		// Verify List Secrets DOES redact (as seen in code)
		resp, err := http.Get(ts.URL + "/secrets")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var results []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&results)
		require.NoError(t, err)

		found := false
		for _, s := range results {
			if s["id"] == "sensitive-secret-123" {
				found = true
				assert.Equal(t, "[REDACTED]", s["value"], "ListSecrets should redact values")
			}
		}
		assert.True(t, found, "Created secret should be in list")
	})
}
