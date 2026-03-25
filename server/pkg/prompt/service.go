// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
// onPromptListChanged notifies clients that the prompt list has changed.
// Currently this is a no-op as the go-sdk does not expose a public Notify method
// for PromptListChanged.
// func (s *Service) onPromptListChanged() {
//    // Waiting for SDK support for public notification triggering
//	  // log.Warn("Prompt list changed notification not sent (SDK limitation)")
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// }
// ListPrompts handles the "prompts/list" MCP request.
//
// Summary: Lists all available prompts.
// Parameters:
//   - standard arguments based on function signature.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - req: *mcp.ListPromptsRequest. The request object.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Returns:
//   - *mcp.ListPromptsResult: The list of prompts.
// GetPrompt handles the "prompts/get" MCP request.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
