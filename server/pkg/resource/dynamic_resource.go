// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package resource provides resource management functionality.
package resource

// NewDynamicResource creates a new instance of DynamicResource.
//
// Summary: Initializes a dynamic resource backed by a tool.
//
// Parameters:
//   - def: *configv1.ResourceDefinition. The resource definition.
//   - t: tool.Tool. The tool used to fetch the resource content.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Returns:
//   - *DynamicResource: The initialized dynamic resource.
//   - error: An error if validation fails.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func NewDynamicResource(def *configv1.ResourceDefinition, t tool.Tool) (*DynamicResource, error) {
	if def == nil {
		return nil, fmt.Errorf("resource definition is nil")
// Resource returns the MCP representation of the resource.
//
// Summary: Retrieves the MCP resource metadata.
//
// Returns:
//   - *mcp.Resource: The MCP resource definition.
// Read executes the associated tool to fetch the resource content.
//
// Summary: Fetches the resource content by executing the tool.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//
// Returns:
//   - *mcp.ReadResourceResult: The resource content.
//   - error: An error if the tool execution fails.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Side Effects:
//   - Executes the underlying tool, which may have its own side effects.
// Errors:
//   - triggers relevant error states on failure.
func (r *DynamicResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	// For now, we'll just execute the tool with no inputs.
	// In the future, we may need to pass inputs to the tool.
	result, err := r.tool.Execute(ctx, &tool.ExecutionRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to execute tool for dynamic resource: %w", err)
	}

	// The tool can return a string, a byte slice, or a map[string]interface{}.
	// We need to handle each of these cases.
	switch content := result.(type) {
	case string:
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      r.resource.URI,
					Text:     content,
					MIMEType: r.resource.MIMEType,
				},
			},
		}, nil
	case []byte:
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      r.resource.URI,
					Blob:     content,
					MIMEType: r.resource.MIMEType,
				},
			},
		}, nil
	case map[string]interface{}:
		// If the tool returns a map, we assume it's a JSON object.
		// We'll marshal it to a string and return it as text.
		data, err := json.Marshal(content)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tool result to JSON: %w", err)
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      r.resource.URI,
// Subscribe is not yet implemented for dynamic resources.
//
// Summary: Subscribes to resource updates (Not Implemented).
//
// Parameters:
//   - _: context.Context. Unused.
//
// Returns:
//   - error: Always returns an error indicating not implemented.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (r *DynamicResource) Subscribe(_ context.Context) error {
	return fmt.Errorf("subscribing to dynamic resources is not yet implemented")
}
