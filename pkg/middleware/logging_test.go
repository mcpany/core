/*
 * Copyright 2025 Author(s) of MCPXY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package middleware

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// memoryHandler is a simple slog.Handler that stores log messages in memory.
type memoryHandler struct {
	mu  sync.Mutex
	buf bytes.Buffer
	h   slog.Handler
}

func newMemoryHandler() *memoryHandler {
	mh := &memoryHandler{}
	mh.h = slog.NewTextHandler(&mh.buf, nil)
	return mh
}

func (h *memoryHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

func (h *memoryHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.h.Handle(ctx, r)
}

func (h *memoryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	return &memoryHandler{h: h.h.WithAttrs(attrs)}
}

func (h *memoryHandler) WithGroup(name string) slog.Handler {
	h.mu.Lock()
	defer h.mu.Unlock()
	return &memoryHandler{h: h.h.WithGroup(name)}
}

func (h *memoryHandler) String() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.buf.String()
}

func TestLoggingMiddleware(t *testing.T) {
	mh := newMemoryHandler()
	logger := slog.New(mh)

	t.Run("SuccessfulCall", func(t *testing.T) {
		mh.buf.Reset()

		expectedResult := &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: `{"status": "ok"}`},
			},
		}
		expectedErr := errors.New("some error")

		// Mock handler that will be wrapped by the middleware
		mockHandler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			assert.Equal(t, "test.method", method)
			// Simulate some work
			time.Sleep(10 * time.Millisecond)
			return expectedResult, expectedErr
		}

		loggingMiddleware := LoggingMiddleware(logger)
		wrappedHandler := loggingMiddleware(mockHandler)

		result, err := wrappedHandler(context.Background(), "test.method", &mcp.InitializeRequest{})

		// Assert that the result from the original handler is passed through
		assert.Equal(t, expectedResult, result)
		assert.Equal(t, expectedErr, err)

		// Assert that the logs were written
		logOutput := mh.String()
		require.True(t, strings.Contains(logOutput, "Request received"), "Log should contain 'Request received'")
		require.True(t, strings.Contains(logOutput, "method=test.method"), "Log should contain the method name")
		require.True(t, strings.Contains(logOutput, "Request completed"), "Log should contain 'Request completed'")
		require.True(t, strings.Contains(logOutput, "duration="), "Log should contain the duration")
	})

	t.Run("NilLogger", func(t *testing.T) {
		// This test ensures that the middleware falls back to the default logger when nil is passed.
		// As we can't easily capture the output of the global default logger without affecting other tests,
		// we will just ensure that the middleware still executes the next handler and returns its results.
		mockHandler := func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			return &mcp.CallToolResult{}, nil
		}

		loggingMiddleware := LoggingMiddleware(nil) // Pass nil to test the fallback
		wrappedHandler := loggingMiddleware(mockHandler)

		_, err := wrappedHandler(context.Background(), "test.method", &mcp.InitializeRequest{})
		assert.NoError(t, err, "The wrapped handler should execute without errors even with a nil logger")
	})
}
