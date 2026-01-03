// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/pkg/storage/sqlite"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAPIHandler_LargeBody(t *testing.T) {
	// Setup SQLite DB
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := sqlite.NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	store := sqlite.NewStore(db)

	// Setup Application
	app := NewApplication()
	app.fs = afero.NewMemMapFs()

	handler := app.createAPIHandler(store)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	// Create a large body (2MB)
	largeBody := make([]byte, 2*1024*1024)
	// Fill with some valid-ish JSON structure if needed, but for size limit check,
	// just the size matters. However, io.ReadAll will read it all before protojson fails on invalid JSON.
	// If we want to test that it fails *because* of size, we should expect 413 or similar.
	// But current implementation just reads it.

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/services", bytes.NewReader(largeBody))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Before fix: it might return 400 because the body is not valid JSON, but it READS the whole body.
	// After fix: it should reject reading the body or fail with a specific error if we check it.
	// The key is that `http.MaxBytesReader` will prevent reading more than N bytes.

	// For now, we just assert that we CAN send it and get a response (it doesn't crash).
	// With the fix, we want to ensure we get a specific behavior, but detecting "read too much"
	// is easiest by checking if the server protected itself.

	// In a real exploit, this would consume memory.
	// Here we just verify the test runs.
	// Now that we've implemented the fix, we expect 413 Payload Too Large
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}
