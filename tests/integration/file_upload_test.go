
package integration_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileUpload(t *testing.T) {
	// Create a dummy file to upload
	file, err := os.CreateTemp("", "upload-*.txt")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	_, err = file.WriteString("hello world")
	assert.NoError(t, err)
	file.Close()

	// Create a new multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a new form file
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	assert.NoError(t, err)

	// Open the dummy file
	file, err = os.Open(file.Name())
	assert.NoError(t, err)
	defer file.Close()

	// Copy the file to the form file
	_, err = io.Copy(part, file)
	assert.NoError(t, err)
	writer.Close()

	// Create a new request
	req, err := http.NewRequest("POST", "http://localhost:8081/upload", body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
	}

	// Check the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
