// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileUpload(t *testing.T) {
	// Start the server using subprocess to ensure isolation
	server := StartMCPANYServer(t, "FileUploadTest")
	defer server.CleanupFunc()

	// Create a dummy file to upload
	file, err := os.CreateTemp("", "upload-*.txt")
	require.NoError(t, err)
	defer func() { _ = os.Remove(file.Name()) }()

	_, err = file.WriteString("hello world")
	require.NoError(t, err)
	_ = file.Close()

	// Create a new multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a new form file
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	require.NoError(t, err)

	// Open the dummy file
	file, err = os.Open(file.Name())
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	// Copy the file to the form file
	_, err = io.Copy(part, file)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	// Create a new request
	// Use the JSONRPC Endpoint base but hit /upload
	// NOTE: StartMCPANYServer provides JSONRPCEndpoint as "http://ip:port".
	req, err := http.NewRequest("POST", server.JSONRPCEndpoint+"/upload", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Add Authorization if API Key is present
	if server.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+server.APIKey)
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}

	// Check the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
