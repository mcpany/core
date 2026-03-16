// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package terraform provides a Terraform provider skeleton.
package terraform

// ResourceMCPServer represents the configuration schema for an MCP Server resource
//
// Summary: ResourceMCPServer represents the configuration schema for an MCP Server resource
// Summary: ResourceMCPServer represents the configuration schema for an MCP Server resource
type ResourceMCPServer struct {
	Name    string
	Port    int
	Enabled bool
}
// Schema returns the Terraform schema definition (Mock). Returns the result.
//
// Summary: Schema returns the Terraform schema definition (Mock). Returns the result.
//
// Parameters:
//   - None.
//
// Returns:
//   - map[string]interface{}: The resulting text.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func Schema() map[string]interface{} {
	return map[string]interface{}{
		"name": map[string]interface{}{
			"type":        "TypeString",
			"required":    true,
			"description": "The name of the MCP server instance",
		},
		"port": map[string]interface{}{
			"type":        "TypeInt",
			"optional":    true,
			"default":     8080,
			"description": "Port to run the server on",
		},
		"enabled": map[string]interface{}{
			"type":        "TypeBool",
			"optional":    true,
			"default":     true,
			"description": "Whether the server is active",
		},
// Create mimics the Create operation of a Terraform resource. _ is an unused parameter. Returns an error if the operation fails.
//
// Summary: Create mimics the Create operation of a Terraform resource. _ is an unused parameter. Returns an error if the operation fails.
//
// Parameters:
//   - _ (*ResourceMCPServer): The provided _ data.
//
// Returns:
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Read mimics the Read operation. name is the name of the resource. Returns the result. Returns an error if the operation fails.
//
// Summary: Read mimics the Read operation. name is the name of the resource. Returns the result. Returns an error if the operation fails.
//
// Parameters:
//   - name (string): The human-readable or system name.
//
// Returns:
//   - *ResourceMCPServer: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func Read(name string) (*ResourceMCPServer, error) {
	return &ResourceMCPServer{
		Name:    name,
		Port:    8080,
		Enabled: true,
	}, nil
}
