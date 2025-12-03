// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"context"
	"net/http"
	"testing"
	"time"

	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func NewTestPoolManager(t *testing.T) *pool.Manager {
	t.Helper()
	pm := pool.NewManager()
	httpPool, err := pool.New(
		func(ctx context.Context) (*client.HttpClientWrapper, error) {
			return &client.HttpClientWrapper{Client: &http.Client{Timeout: 5 * time.Second}}, nil
		},
		1,
		10,
		int(1*time.Minute),
		false,
	)
	require.NoError(t, err)
	pm.Register("test-service", httpPool)
	return pm
}

type MockAuthenticator struct {
	AuthenticateFunc func(req *http.Request) error
}

func (m *MockAuthenticator) Authenticate(req *http.Request) error {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(req)
	}
	return nil
}

func NewTempFs() (afero.Fs, func(), error) {
	fs := afero.NewOsFs()
	tempDir, err := afero.TempDir(fs, "", "mcpany-test-")
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		_ = fs.RemoveAll(tempDir)
	}
	return afero.NewBasePathFs(fs, tempDir), cleanup, nil
}

func NewTestMCPServer(
	ctx context.Context,
	toolManager tool.ToolManagerInterface,
	promptManager prompt.PromptManagerInterface,
	resourceManager resource.ResourceManagerInterface,
	busProvider *bus.BusProvider,
) (*mcpserver.Server, error) {
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(nil, toolManager, promptManager, resourceManager, authManager)
	return mcpserver.NewServer(ctx, toolManager, promptManager, resourceManager, authManager, serviceRegistry, busProvider, false)
}

func CreateTestUploadRequest(t *testing.T, method, fileName, fileContent string) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()
	var body bytes.Buffer
	var writer *multipart.Writer
	if fileContent != "" {
		writer = multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", fileName)
		require.NoError(t, err)
		_, err = io.WriteString(part, fileContent)
		require.NoError(t, err)
		writer.Close()
	}

	req := httptest.NewRequest(method, "/upload", &body)
	if writer != nil {
		req.Header.Set("Content-Type", writer.FormDataContentType())
	}
	return req, httptest.NewRecorder()
}
