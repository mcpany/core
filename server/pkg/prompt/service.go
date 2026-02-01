// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Service handles the business logic for the prompts feature. It provides
// methods for listing available prompts and retrieving a specific prompt by
// name.
type Service struct {
	promptManager ManagerInterface
	mcpServer     *mcp.Server
}

// NewService creates and returns a new Service instance.
//
// Parameters:
//   - promptManager: ManagerInterface. The prompt manager to use.
//
// Returns:
//   - *Service: The newly created Service instance.
func NewService(promptManager ManagerInterface) *Service {
	s := &Service{
		promptManager: promptManager,
	}
	// s.promptManager.OnListChanged(s.onPromptListChanged)
	return s
}

// SetMCPServer sets the MCP server instance for the service.
//
// Parameters:
//   - mcpServer: *mcp.Server. The MCP server instance.
func (s *Service) SetMCPServer(mcpServer *mcp.Server) {
	s.mcpServer = mcpServer
	s.promptManager.SetMCPServer(NewMCPServerProvider(mcpServer))
}

// onPromptListChanged notifies clients that the prompt list has changed.
// Currently this is a no-op as the go-sdk does not expose a public Notify method
// for PromptListChanged.
// func (s *Service) onPromptListChanged() {
//    // Waiting for SDK support for public notification triggering
//	  // log.Warn("Prompt list changed notification not sent (SDK limitation)")
// }

// ListPrompts handles the "prompts/list" MCP request.
//
// It retrieves the list of available prompts from the Manager, converts them to the MCP format,
// and returns them to the client.
//
// Parameters:
//   - _ : context.Context. The execution context.
//   - _ : *mcp.ListPromptsRequest. The list prompts request.
//
// Returns:
//   - *mcp.ListPromptsResult: The list of available prompts.
//   - error: An error if retrieval fails (unlikely in current implementation).
func (s *Service) ListPrompts(
	_ context.Context,
	_ *mcp.ListPromptsRequest,
) (*mcp.ListPromptsResult, error) {
	prompts := s.promptManager.ListPrompts()
	mcpPrompts := make([]*mcp.Prompt, len(prompts))
	for i, p := range prompts {
		mcpPrompts[i] = p.Prompt()
	}
	return &mcp.ListPromptsResult{
		Prompts: mcpPrompts,
	}, nil
}

// GetPrompt handles the "prompts/get" MCP request.
//
// It retrieves a specific prompt by name from the Manager and executes it with the provided
// arguments, returning the result.
//
// Parameters:
//   - ctx: context.Context. The execution context.
//   - req: *mcp.GetPromptRequest. The get prompt request containing the prompt name and arguments.
//
// Returns:
//   - *mcp.GetPromptResult: The result of the prompt execution.
//   - error: An error if the prompt is not found or execution fails.
//
// Errors:
//   - Returns ErrPromptNotFound if the prompt does not exist.
//   - Returns error if argument marshaling or prompt execution fails.
func (s *Service) GetPrompt(
	ctx context.Context,
	req *mcp.GetPromptRequest,
) (*mcp.GetPromptResult, error) {
	p, ok := s.promptManager.GetPrompt(req.Params.Name)
	if !ok {
		return nil, ErrPromptNotFound
	}

	argsBytes, err := json.Marshal(req.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prompt arguments: %w", err)
	}

	return p.Get(ctx, argsBytes)
}
