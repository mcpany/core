package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleUploadSkillAsset_DoS(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{
		SkillManager: manager,
	}
	handler := app.handleUploadSkillAsset()

	// 11 MB Payload
	// Using a smaller buffer but claiming a larger size via Content-Length won't work with bytes.Reader
	// We need to actually send bytes or use a mock reader.
	// To avoid OOMing the test runner, let's use a mocked reader that errors after reading > 10MB?
	// Or just use 11MB. 11MB is small enough for modern machines.
	largeBody := bytes.Repeat([]byte("A"), 11*1024*1024)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/skills/test-skill/assets?path=test.txt", bytes.NewReader(largeBody))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Expect 413 Payload Too Large
	// Currently, without the fix, this will likely succeed (200) or fail with 500 if save fails.
	// We assert that it fails with 413 (standard for MaxBytesReader)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected 413 Payload Too Large, got %d", w.Code)
	}
}

func TestHandleUploadSkillAsset_PathParsing(t *testing.T) {
	manager, _ := setupSkillManagerForHTTPTest(t)
	app := &Application{
		SkillManager: manager,
	}
	handler := app.handleUploadSkillAsset()

	testCases := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{"Double Slash", "/api/v1/skills//assets?path=test.txt", http.StatusBadRequest},
		{"Missing Name", "/api/v1/skills//assets?path=test.txt", http.StatusBadRequest},
		{"Short Path", "/api/v1/skills/assets?path=test.txt", http.StatusBadRequest},
		// Trailing slash might result in empty last element with split
		{"Trailing Slash", "/api/v1/skills/test-skill/assets/?path=test.txt", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tc.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, tc.expectedStatus, w.Code, "Path: %s", tc.path)
		})
	}
}
