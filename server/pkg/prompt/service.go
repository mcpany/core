// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Service handles the business logic for the prompts feature.
//
// Summary: Provides prompt management services.
//
// Description:
// It provides methods for listing available prompts and retrieving a specific prompt by
// name.
type Service struct {
	promptManager ManagerInterface
	mcpServer     *mcp.Server
}

// NewService creates and returns a new Service instance.
//
// Summary: Initializes a new Service.
//
// Parameters:
//   - promptManager: ManagerInterface. The manager to use for prompt operations.
//
// Returns:
//   - *Service: The initialized Service.
func NewService(promptManager ManagerInterface) *Service {
	s := &Service{
		promptManager: promptManager,
	}
	// s.promptManager.OnListChanged(s.onPromptListChanged)
	return s
}

// SetMCPServer sets the MCP server instance for the service.
//
// Summary: Configures the MCP server reference.
//
// Parameters:
//   - mcpServer: *mcp.Server. The MCP server instance.
//
// Side Effects:
//   - Updates the internal reference and the prompt manager's provider.
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
// Summary: Lists available prompts.
//
// Parameters:
//   - _ : context.Context. The context (unused).
//   - _ : *mcp.ListPromptsRequest. The request (unused).
//
// Returns:
//   - *mcp.ListPromptsResult: The list of prompts.
//   - error: Always nil in this implementation.
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
// Summary: Retrieves and executes a specific prompt.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *mcp.GetPromptRequest. The request containing the prompt name and arguments.
//
// Returns:
//   - *mcp.GetPromptResult: The result of the prompt execution.
//   - error: An error if the prompt is not found or execution fails.
//
// Throws/Errors:
//   - ErrPromptNotFound: If the prompt does not exist.
//   - Error if argument marshaling fails.
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
