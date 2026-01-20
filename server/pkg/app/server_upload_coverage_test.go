// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUploadFile_Coverage(t *testing.T) {
	app := NewApplication()

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/upload", nil)
		w := httptest.NewRecorder()

		app.uploadFile(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("Missing File", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close() // Empty form

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		app.uploadFile(w, req)

		resp := w.Result()
		// If file is missing, FormFile returns error
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}
