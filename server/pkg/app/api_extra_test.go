// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/stretchr/testify/assert"
)

func TestHandleResourceRead(t *testing.T) {
	app, _ := setupApiTestApp()

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/resources/read", nil)
		w := httptest.NewRecorder()
		mux := app.createAPIHandler(app.Storage)
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("MissingURI", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resources/read", nil)
		w := httptest.NewRecorder()
		mux := app.createAPIHandler(app.Storage)
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

    t.Run("ReadError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockResourceManager := resource.NewMockManagerInterface(ctrl)
		app.ResourceManager = mockResourceManager

		mockRes := resource.NewMockResourceInterface(ctrl)
		mockResourceManager.EXPECT().GetResource("mock://error").Return(mockRes, true)
		mockRes.EXPECT().Read(gomock.Any()).Return(nil, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/resources/read?uri=mock://error", nil)
		w := httptest.NewRecorder()

		mux := app.createAPIHandler(app.Storage)
		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandlePromptExecute(t *testing.T) {
	app, _ := setupApiTestApp()

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/prompts/test-prompt/execute", nil)
		w := httptest.NewRecorder()
		mux := app.createAPIHandler(app.Storage)
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("InvalidPath", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/prompts/test-prompt/other", nil)
		w := httptest.NewRecorder()
		mux := app.createAPIHandler(app.Storage)
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InvalidBody", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/prompts/test-prompt/execute", errReader{})
		w := httptest.NewRecorder()
		mux := app.createAPIHandler(app.Storage)
		mux.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

    t.Run("ExecuteError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPromptManager := prompt.NewMockManagerInterface(ctrl)
		app.PromptManager = mockPromptManager

		mockPrompt := prompt.NewMockPromptInterface(ctrl)
		mockPromptManager.EXPECT().GetPrompt("error-prompt").Return(mockPrompt, true)
        mockPrompt.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, assert.AnError)

		req := httptest.NewRequest(http.MethodPost, "/prompts/error-prompt/execute", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()

		mux := app.createAPIHandler(app.Storage)
		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

type errReader struct{}

func (errReader) Read(p []byte) (n int, err error) {
	return 0, assert.AnError
}
