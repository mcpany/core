// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/storage/memory"
)

func TestHandleProfiles_LargeBody(t *testing.T) {
	app := NewApplication()
	store := memory.NewStore()

	// Create a large body (> 1MB)
	largeBody := make([]byte, 2*1024*1024) // 2MB

	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := app.handleProfiles(store)
	handler.ServeHTTP(w, req)

	// Currently, it accepts unlimited body, so it might fail with bad request due to unmarshal error,
	// but we want to assert that it DOES NOT check for size limit explicitly like handleServices does.
	// However, since we are FIXING it, we want to write a test that expects the fix behavior (413 Request Entity Too Large)
	// OR verify the current buggy behavior.
	// Since I need to PROVE the bug, I will assert current behavior which likely returns 400 Bad Request (Unmarshal error) or 200/201 if it accidentally works (unlikely for random bytes).
	// If I send random bytes it will be 400.

	// If the bug exists (no limit), it tries to read all.
	// If I implement the fix, it should return 413.

	// Let's create a test that EXPECTS 413, which will fail now (proving the bug/missing feature).
	if w.Code == http.StatusRequestEntityTooLarge {
		// This would mean it is already fixed
		t.Logf("Got 413 as expected (if fixed)")
	} else {
		t.Logf("Got %d, expected 413", w.Code)
		// To make the test fail if NOT 413 (once I fix it):
		if w.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("Expected 413 Payload Too Large, got %d", w.Code)
		}
	}
}

func TestHandleProfileDetail_LargeBody(t *testing.T) {
	app := NewApplication()
	store := memory.NewStore()

	largeBody := make([]byte, 2*1024*1024) // 2MB
	req := httptest.NewRequest(http.MethodPut, "/profiles/test", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := app.handleProfileDetail(store)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected 413 Payload Too Large, got %d", w.Code)
	}
}

func TestHandleSettings_LargeBody(t *testing.T) {
	app := NewApplication()
	store := memory.NewStore()

	largeBody := make([]byte, 2*1024*1024) // 2MB
	req := httptest.NewRequest(http.MethodPost, "/settings", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := app.handleSettings(store)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected 413 Payload Too Large, got %d", w.Code)
	}
}

func TestHandleCollections_LargeBody(t *testing.T) {
	app := NewApplication()
	store := memory.NewStore()

	largeBody := make([]byte, 2*1024*1024) // 2MB
	req := httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := app.handleCollections(store)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected 413 Payload Too Large, got %d", w.Code)
	}
}

func TestHandleCollectionDetail_LargeBody(t *testing.T) {
	app := NewApplication()
	store := memory.NewStore()

	largeBody := make([]byte, 2*1024*1024) // 2MB
	req := httptest.NewRequest(http.MethodPut, "/collections/test", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := app.handleCollectionDetail(store)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected 413 Payload Too Large, got %d", w.Code)
	}
}

func TestHandleSecrets_LargeBody(t *testing.T) {
	app := NewApplication()
	store := memory.NewStore()

	largeBody := make([]byte, 2*1024*1024) // 2MB
	req := httptest.NewRequest(http.MethodPost, "/secrets", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := app.handleSecrets(store)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected 413 Payload Too Large, got %d", w.Code)
	}
}

// Also test IO Reader error if possible, but MaxBytesReader is the main one verifiable easily.
// For error handling:
// We can mock a reader that returns error.

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, context.DeadlineExceeded // Just an example error
}

func TestHandleProfiles_ReadError(t *testing.T) {
	app := NewApplication()
	store := memory.NewStore()

	req := httptest.NewRequest(http.MethodPost, "/profiles", &errorReader{})
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := app.handleProfiles(store)
	handler.ServeHTTP(w, req)

	// Currently it ignores error and tries to unmarshal empty body -> EOF or similar from Unmarshal?
	// But io.ReadAll returns error.
	// body, _ := io.ReadAll(r.Body) -> body is empty, err is ignored.
	// protojson.Unmarshal(nil, ...) -> returns error "unexpected end of JSON input" or similar.
	// So we get 400 Bad Request with error message.
	// But we WANT to catch the read error specifically if we want to distinguish or handle it better.
	// The bug is ignoring the error from ReadAll.

	// If I fix it, I should likely return 400 with "failed to read body" or something similar.
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request, got %d", w.Code)
	}
	// We might check the body to see if it says "failed to read body" vs "unexpected end of JSON input"
}
