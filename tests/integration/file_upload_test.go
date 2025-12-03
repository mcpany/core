// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
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
	// Start the server
	server := StartInProcessMCPANYServer(t, "FileUploadTest")
	defer server.CleanupFunc()

	// Create a dummy file to upload
	file, err := os.CreateTemp("", "upload-*.txt")
	require.NoError(t, err)
	defer os.Remove(file.Name())

	_, err = file.WriteString("hello world")
	require.NoError(t, err)
	file.Close()

	// Create a new multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a new form file
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	require.NoError(t, err)

	// Open the dummy file
	file, err = os.Open(file.Name())
	require.NoError(t, err)
	defer file.Close()

	// Copy the file to the form file
	_, err = io.Copy(part, file)
	require.NoError(t, err)
	writer.Close()

	// Create a new request
	req, err := http.NewRequest("POST", server.JSONRPCEndpoint+"/upload", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	if resp != nil {
		defer resp.Body.Close()
	}

	// Check the response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
