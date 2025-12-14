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

package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/mcpany/core/pkg/logging"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestDebugMiddleware(t *testing.T) {
	var logOutput bytes.Buffer
	logging.Init(slog.LevelDebug, &logOutput)

	mw := DebugMiddleware()
	handler := mw(func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.ListToolsResult{
			Tools: []*mcp.Tool{
				{Name: "test-tool"},
			},
		}, nil
	})

	req := &mcp.ListToolsRequest{}
	_, err := handler(context.Background(), "tools/list", req)
	assert.NoError(t, err)

	logStr := logOutput.String()
	assert.Contains(t, logStr, "method=tools/list")
	assert.Contains(t, logStr, "request")
	assert.Contains(t, logStr, "response")
	assert.Contains(t, logStr, "test-tool")
}

func TestDebugMiddleware_NoLoggingWhenDisabled(t *testing.T) {
	var logOutput bytes.Buffer
	logging.Init(slog.LevelInfo, &logOutput)

	mw := DebugMiddleware()
	handler := mw(func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.ListToolsResult{
			Tools: []*mcp.Tool{
				{Name: "test-tool"},
			},
		}, nil
	})

	req := &mcp.ListToolsRequest{}
	_, err := handler(context.Background(), "tools/list", req)
	assert.NoError(t, err)

	assert.False(t, strings.Contains(logOutput.String(), "MCP Request"))
	assert.False(t, strings.Contains(logOutput.String(), "MCP Response"))
}
