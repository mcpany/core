// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/mock/gomock"
)

// MockResource implements resource.Resource
type errorResource struct{}

func (e *errorResource) Resource() *mcp.Resource { return &mcp.Resource{URI: "error://test"} }
func (e *errorResource) Service() string         { return "test" }
func (e *errorResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	return nil, errors.New("read failed")
}
func (e *errorResource) Subscribe(ctx context.Context) error { return nil }

// MockPrompt implements prompt.Prompt
type errorPrompt struct{}

func (e *errorPrompt) Prompt() *mcp.Prompt { return &mcp.Prompt{Name: "error-prompt"} }
func (e *errorPrompt) Service() string     { return "test" }
func (e *errorPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	return nil, errors.New("get failed")
}

func TestHandleResourceReadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockResManager := resource.NewMockManagerInterface(ctrl)

	// Setup app with mock manager
	app, _ := setupCoverageTestApp()
	app.ResourceManager = mockResManager

	// Setup expectation
	mockResManager.EXPECT().GetResource("error://test").Return(&errorResource{}, true)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/resources/read?uri=error://test", nil)
	w := httptest.NewRecorder()

	app.handleResourceRead().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}

	if strings.TrimSpace(w.Body.String()) != http.StatusText(http.StatusInternalServerError) {
		t.Errorf("Expected body '%s', got '%s'", http.StatusText(http.StatusInternalServerError), w.Body.String())
	}
}

func TestHandlePromptExecuteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPromptManager := prompt.NewMockManagerInterface(ctrl)

	// Setup app with mock manager
	app, _ := setupCoverageTestApp()
	app.PromptManager = mockPromptManager

	// Setup expectation
	mockPromptManager.EXPECT().GetPrompt("error-prompt").Return(&errorPrompt{}, true)

	// Make request
	req := httptest.NewRequest(http.MethodPost, "/prompts/error-prompt/execute", nil)
	w := httptest.NewRecorder()

	app.handlePromptExecute().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}

	if strings.TrimSpace(w.Body.String()) != http.StatusText(http.StatusInternalServerError) {
		t.Errorf("Expected body '%s', got '%s'", http.StatusText(http.StatusInternalServerError), w.Body.String())
	}
}
