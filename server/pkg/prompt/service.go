// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Service handles the business logic for the prompts feature. It provides.
//
// Summary: handles the business logic for the prompts feature. It provides.
type Service struct {
	promptManager ManagerInterface
	mcpServer     *mcp.Server
}

// NewService creates and returns a new Service instance.
//
// Summary: creates and returns a new Service instance.
//
// Parameters:
//   - promptManager: ManagerInterface. The promptManager.
//
// Returns:
//   - *Service: The *Service.
func NewService(promptManager ManagerInterface) *Service {
	s := &Service{
		promptManager: promptManager,
	}
	// s.promptManager.OnListChanged(s.onPromptListChanged)
	return s
}

// SetMCPServer sets the MCP server instance for the service.
//
// Summary: sets the MCP server instance for the service.
//
// Parameters:
//   - mcpServer: *mcp.Server. The mcpServer.
//
// Returns:
//   None.
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

// ListPrompts handles the "prompts/list" MCP request. It retrieves the list of.
//
// Summary: handles the "prompts/list" MCP request. It retrieves the list of.
//
// Parameters:
//   - _: context.Context. The _.
//   - _: *mcp.ListPromptsRequest. The _.
//
// Returns:
//   - *mcp.ListPromptsResult: The *mcp.ListPromptsResult.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

// GetPrompt handles the "prompts/get" MCP request. It retrieves a specific.
//
// Summary: handles the "prompts/get" MCP request. It retrieves a specific.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - req: *mcp.GetPromptRequest. The req.
//
// Returns:
//   - *mcp.GetPromptResult: The *mcp.GetPromptResult.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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
