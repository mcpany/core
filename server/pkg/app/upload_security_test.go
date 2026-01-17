// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUploadFile_Security(t *testing.T) {
	app := NewApplication()

	t.Run("Reflected XSS", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		// Use a filename with special characters that need escaping, but avoid <script>
		// if the environment strips it. & and " are standard HTML special chars.
		filename := "test&file\"name.txt"
		part, err := writer.CreateFormFile("file", filename)
		if err != nil {
			t.Fatal(err)
		}
		part.Write([]byte("content"))
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		app.uploadFile(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK, got %d", resp.StatusCode)
		}

		responseBody := w.Body.String()

		// Check that characters are sanitized
		// & and " should be replaced by _
		sanitizedFilename := "test_file_name.txt"
		if strings.Contains(responseBody, filename) {
			t.Errorf("Reflected XSS/Sanitization failure: Response contains raw filename: %s", responseBody)
		}

		if !strings.Contains(responseBody, sanitizedFilename) {
			t.Errorf("Expected sanitized filename '%s' not found. Body: %s", sanitizedFilename, responseBody)
		}
	})

	t.Run("Size Limit", func(t *testing.T) {
		// 11MB payload
		size := 11 * 1024 * 1024
		largeData := make([]byte, size)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "large.txt")
		if err != nil {
			t.Fatal(err)
		}
		_, err = part.Write(largeData)
		if err != nil {
			t.Fatal(err)
		}
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		app.uploadFile(w, req)

		resp := w.Result()

		// Vulnerability check: Should NOT succeed for 11MB file if limit is 10MB
		// Current implementation has no limit, so this will return 200 OK and fail the test.
		if resp.StatusCode == http.StatusOK {
			t.Errorf("Size limit vulnerability detected: Uploaded %d bytes successfully (expected rejection)", size)
		}
	})
}
