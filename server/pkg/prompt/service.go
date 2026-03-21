// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Summary: Service handles the business logic for the prompts feature. It provides methods for listing available prompts and retrieving a specific prompt by name. Represents a Service.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Service struct {
	promptManager ManagerInterface
	mcpServer     *mcp.Server
}

// Summary: NewService creates and returns a new Service instance. Initializes a new Prompt Service.
//
// Parameters:
//   - promptManager (ManagerInterface): The promptManager parameter.
//
// Returns:
//   - *Service: The resulting *Service.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewService(promptManager ManagerInterface) *Service {
	s := &Service{
		promptManager: promptManager,
	}
	// s.promptManager.OnListChanged(s.onPromptListChanged)
	return s
}

// Summary: SetMCPServer sets the MCP server instance for the service. Configures the underlying MCP server.
//
// Parameters:
//   - mcpServer (*mcp.Server): The mcpServer parameter.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Summary: ListPrompts handles the "prompts/list" MCP request. Lists all available prompts.
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//   - _ (*mcp.ListPromptsRequest): The _ parameter.
//
// Returns:
//   - *mcp.ListPromptsResult: The resulting *mcp.ListPromptsResult.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: GetPrompt handles the "prompts/get" MCP request. Retrieves and executes a specific prompt.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//   - req (*mcp.GetPromptRequest): The req parameter.
//
// Returns:
//   - *mcp.GetPromptResult: The resulting *mcp.GetPromptResult.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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
