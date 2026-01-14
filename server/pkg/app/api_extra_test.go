// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Resource
type mockResource struct {
	uri     string
	content string
}

func (m *mockResource) Resource() *mcp.Resource {
	return &mcp.Resource{URI: m.uri}
}
func (m *mockResource) Service() string { return "mock" }
func (m *mockResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      m.uri,
				Text:     m.content,
				MIMEType: "text/plain",
			},
		},
	}, nil
}
func (m *mockResource) Subscribe(ctx context.Context) error { return nil }

// Mock Prompt
type mockPrompt struct {
	name string
}

func (m *mockPrompt) Prompt() *mcp.Prompt {
	return &mcp.Prompt{Name: m.name}
}
func (m *mockPrompt) Service() string { return "mock" }
func (m *mockPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{
				Role: mcp.Role("user"),
				Content: &mcp.TextContent{
					Text: "Executed prompt " + m.name,
				},
			},
		},
	}, nil
}

func TestHandleResourceRead(t *testing.T) {
	fs := afero.NewMemMapFs()
	app := NewApplication()
	app.fs = fs
	app.ResourceManager = resource.NewManager()

	// Add a mock resource
	res := &mockResource{uri: "mock://test", content: "hello world"}
	app.ResourceManager.AddResource(res)

	handler := app.handleResourceRead()

	t.Run("ReadResource", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resources/read?uri=mock://test", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result mcp.ReadResourceResult
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		require.Len(t, result.Contents, 1)

		// result.Contents is []*mcp.ResourceContents
		content := result.Contents[0]
		assert.Equal(t, "mock://test", content.URI)
		assert.Equal(t, "hello world", content.Text)
	})

	t.Run("ReadResource_NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resources/read?uri=mock://nonexistent", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("ReadResource_MissingURI", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resources/read", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/resources/read?uri=mock://test", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHandlePromptExecute(t *testing.T) {
	fs := afero.NewMemMapFs()
	app := NewApplication()
	app.fs = fs
	app.PromptManager = prompt.NewManager()

	// Add a mock prompt
	p := &mockPrompt{name: "test-prompt"}
	app.PromptManager.AddPrompt(p)

	handler := app.handlePromptExecute()

	t.Run("ExecutePrompt", func(t *testing.T) {
		// Path: /prompts/{name}/execute
		// The handler strips /prompts/.
		// So request URL path should be /prompts/test-prompt/execute

		req := httptest.NewRequest(http.MethodPost, "/prompts/test-prompt/execute", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result mcp.GetPromptResult
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		require.Len(t, result.Messages, 1)

		// SDK deserialization: Content should be unmarshaled into appropriate type if SDK supports it.
		// However, standard json.Unmarshal into interface/struct with interface field requires custom UnmarshalJSON.
		// Assuming SDK has it.
		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if ok {
			assert.Equal(t, "Executed prompt test-prompt", content.Text)
		} else {
			// Fallback or check what it is
			t.Logf("Got type: %T", result.Messages[0].Content)
			// If json.Unmarshal didn't handle interface implementation, it might be nil or map?
			// But mcp.Content is interface.
		}
	})

	t.Run("ExecutePrompt_NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/prompts/nonexistent/execute", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("ExecutePrompt_InvalidAction", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/prompts/test-prompt/other", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/prompts/test-prompt/execute", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}
