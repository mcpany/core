/*
 * Copyright 2025 Author(s) of MCP Any
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

package mcpserver

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	router := NewRouter()

	t.Run("register and get handler", func(t *testing.T) {
		expectedResult := &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "test result"}},
		}
		handler := func(ctx context.Context, req mcp.Request) (mcp.Result, error) {
			return expectedResult, nil
		}
		router.Register("test/method", handler)

		retrievedHandler, ok := router.GetHandler("test/method")
		assert.True(t, ok, "handler should be found")
		assert.NotNil(t, retrievedHandler, "handler should not be nil")

		res, err := retrievedHandler(context.Background(), nil)
		assert.NoError(t, err)
		assert.Equal(t, expectedResult, res)
	})

	t.Run("get non-existent handler", func(t *testing.T) {
		_, ok := router.GetHandler("non/existent")
		assert.False(t, ok, "handler should not be found")
	})
}
