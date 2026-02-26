// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package terraform provides a Terraform provider skeleton.
package terraform

// ResourceMCPServer represents the configuration schema for an MCP Server resource
// This would map to hashicorp/terraform-plugin-sdk in a real provider.
type ResourceMCPServer struct {
	Name    string
	Port    int
	Enabled bool
}

// Schema returns the Terraform schema definition (Mock).
//
// Summary: Returns the Terraform schema definition for the MCP server resource.
//
// Parameters:
//   - None.
//
// Returns:
//   - map[string]interface{}: The schema definition map.
//
// Side Effects:
//   - None.
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
	}
}

// Create mimics the Create operation of a Terraform resource.
//
// Summary: Provisions a new MCP server resource (Mock).
//
// Parameters:
//   - _ (*ResourceMCPServer): The resource configuration (unused in mock).
//
// Returns:
//   - error: An error if the provisioning fails.
//
// Side Effects:
//   - None (Mock implementation).
func Create(_ *ResourceMCPServer) error {
	// Simulate API call to provision resources
	return nil
}

// Read mimics the Read operation.
//
// Summary: Reads the state of an existing MCP server resource (Mock).
//
// Parameters:
//   - name (string): The name of the resource to read.
//
// Returns:
//   - *ResourceMCPServer: The resource state.
//   - error: An error if the read operation fails.
//
// Side Effects:
//   - None (Mock implementation).
func Read(name string) (*ResourceMCPServer, error) {
	return &ResourceMCPServer{
		Name:    name,
		Port:    8080,
		Enabled: true,
	}, nil
}
