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
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Service handles the business logic for the prompts feature. It provides
// methods for listing available prompts and retrieving a specific prompt by
// name.
type Service struct {
	promptManager PromptManagerInterface
	mcpServer     *mcp.Server
}

// NewService creates and returns a new Service instance.
func NewService(promptManager PromptManagerInterface) *Service {
	s := &Service{
		promptManager: promptManager,
	}
	// s.promptManager.OnListChanged(s.onPromptListChanged)
	return s
}

// SetMCPServer sets the MCP server instance for the service.
func (s *Service) SetMCPServer(mcpServer *mcp.Server) {
	s.mcpServer = mcpServer
}

// TODO: Re-enable this method when the go-sdk supports notifications.
// See: https://github.com/modelcontextprotocol/go-sdk/issues/123
// func (s *Service) onPromptListChanged() {
// 	if s.mcpServer != nil {
// 		s.mcpServer.Notify("notifications/prompts/list_changed", nil)
// 	}
// }

// ListPrompts handles the "prompts/list" MCP request. It retrieves the list of
// available prompts from the PromptManager, converts them to the MCP format, and
// returns them to the client.
func (s *Service) ListPrompts(
	_ context.Context,
	_ *mcp.ListPromptsRequest,
) (*mcp.ListPromptsResult, error) {
	prompts := s.promptManager.ListPrompts()
	mcpPrompts := make([]*mcp.Prompt, len(prompts))
	for i, p := range prompts {
		mcpPrompts[i] = &mcp.Prompt{
			Name:        p.Name,
			Description: p.Description,
		}
	}
	return &mcp.ListPromptsResult{
		Prompts: mcpPrompts,
	}, nil
}

var (
	// ErrPromptNotFound is returned when a prompt is not found.
	ErrPromptNotFound = errors.New("prompt not found")
)
