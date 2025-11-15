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

package prompt

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

func TestProvider(t *testing.T) {
	t.Run("NewMCPServerProvider should create a new provider and return the correct server instance", func(t *testing.T) {
		mockMCPServer := &mcp.Server{}
		provider := NewMCPServerProvider(mockMCPServer)

		assert.NotNil(t, provider)
		assert.Equal(t, mockMCPServer, provider.Server())
	})
}
