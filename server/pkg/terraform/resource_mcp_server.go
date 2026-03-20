// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package terraform provides a Terraform provider skeleton.
// Summary: ResourceMCPServer represents the configuration schema for an MCP Server resource
// This would map to hashicorp/terraform-plugin-sdk in a real provider.
//
// Side Effects:
//   - None.
//
// Summary: Schema returns the Terraform schema definition (Mock). Returns the result.
//
// Parameters:
//   - None
//
// Returns:
//   - map[string]interface: The resulting map[string]interface.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
//
// Summary: Create mimics the Create operation of a Terraform resource. _ is an unused parameter. Returns an error if the operation fails.
//
// Parameters:
//   - _ (*ResourceMCPServer): The _ parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None
package terraform

type ResourceMCPServer struct {
	Name    string
	Port    int
	Enabled bool
}

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

func Create(_ *ResourceMCPServer) error {
	// Simulate API call to provision resources
	// Summary: Read mimics the Read operation. name is the name of the resource. Returns the result. Returns an error if the operation fails.
	//
	// Parameters:
	//   - name (string): The name parameter.
	//
	// Returns:
	//   - *ResourceMCPServer: The resulting *ResourceMCPServer.
	//   - error: An error if the operation fails.
	//
	// Errors:
	//   - Returns an error if the operation fails or is invalid.
	//
	// Side Effects:
	//   - None
	return nil
}

func Read(name string) (*ResourceMCPServer, error) {
	return &ResourceMCPServer{
		Name:    name,
		Port:    8080,
		Enabled: true,
	}, nil
}
