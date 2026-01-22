package app

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadFile_TruncationBug(t *testing.T) {
	app := NewApplication()

	// Reproduce the bug: SanitizeFilename produces invalid UTF-8 when truncating multibyte characters
	// Create a long filename that causes truncation in the middle of a multibyte character.
	// 254 'a's + '好' (3 bytes).
	// Length = 257.
	// Truncate at 255 -> 254 'a's + 1st byte of '好'.
	prefix := strings.Repeat("a", 254)
	longName := prefix + "好" + ".txt"

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", longName)
	require.NoError(t, err)
	_, err = part.Write([]byte("content"))
	require.NoError(t, err)
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	app.uploadFile(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check response body for validity
	respBody := w.Body.String()

	// The response format is: "File '%s' uploaded successfully (size: %d bytes)"
	// We extract the filename part.
	// Note: html.EscapeString is called on the filename.

	// If the filename is invalid UTF-8, html.EscapeString might produce invalid UTF-8 or replacement chars.
	// But `SanitizeFilename` definitely produced invalid UTF-8 string before calling html.EscapeString.
	// Go strings are arbitrary bytes, so it "works" until you try to decode it.

	// Check if the response body contains valid UTF-8.
	// If SanitizeFilename returned invalid UTF-8, and html.EscapeString preserved it (or didn't fix it),
	// then the response might be invalid UTF-8.
	// Actually html.EscapeString iterates string. If it sees invalid UTF-8, it replaces with Replacement Char (U+FFFD).
	// So validity might be "fixed" by html.EscapeString by replacing with .

	// Let's check if we see the replacement character, OR if we can verify the behavior of SanitizeFilename directly via unit test (which we did).
	// But for E2E, finding  is a sign that something went wrong with encoding/truncation.

	// However, simply checking ValidString might pass if html.EscapeString fixed it.
	// Ideally we want to ensure the filename is truncated CLEANLY (no partial chars).
	// So we should NOT see  (U+FFFD) at the end of the filename.

	assert.True(t, utf8.ValidString(respBody), "Response body should be valid UTF-8")

	// If invalid truncation happened, the last character of the filename in the response
	// (before ' uploaded successfully') might be the replacement char if html.EscapeString fixed it.
	// Or it might be garbage bytes if it didn't.

	// Let's rely on the unit test for strict verification, but here we can check that we don't have garbage.
	// Actually, if I call `SanitizeFilename` directly in the test (like the unit test), it's better.
	// But this is "E2E-like" test.

	if !utf8.ValidString(respBody) {
		t.Logf("Response body is INVALID UTF-8. Bytes: %v", []byte(respBody))
	} else {
        t.Logf("Response body is valid UTF-8")
    }

	// The unit test failure confirmed that SanitizeFilename returns invalid UTF-8.
	// The E2E test failure condition:
	// If I modify SanitizeFilename to NOT produce invalid UTF-8, then this test should assert that.

	// If `html.EscapeString` replaces invalid bytes with , then checking for  is a way to detect the bug.
	// A correct implementation would truncate before the multibyte char, so no .

	if strings.Contains(respBody, "\ufffd") {
		t.Errorf("Response contains replacement character (U+FFFD), indicating invalid truncation: %s", respBody)
	}
}
