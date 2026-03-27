// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package terraform provides a Terraform provider skeleton.
package terraform

// ResourceMCPServer represents the configuration schema for an MCP Server resource
// This would map to hashicorp/terraform-plugin-sdk in a real provider.
//
// Summary: Represents a ResourceMCPServer.
type ResourceMCPServer struct {
	Name    string
	Port    int
	Enabled bool
}

// Schema provides schema functionality.
//
// Summary: Schema.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
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

// Create provides create functionality.
//
// Summary: Create.
//
// Parameters.
//   - _: The parameter.
//
// Returns.
//   - result: The result.
func Create(_ *ResourceMCPServer) error {
	// Simulate API call to provision resources
	return nil
}

// Read provides read functionality.
//
// Summary: Read.
//
// Parameters.
//   - name: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func Read(name string) (*ResourceMCPServer, error) {
	return &ResourceMCPServer{
		Name:    name,
		Port:    8080,
		Enabled: true,
	}, nil
}
