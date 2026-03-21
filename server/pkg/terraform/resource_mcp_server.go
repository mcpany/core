// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package terraform provides a Terraform provider skeleton.
package terraform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ResourceMCPServer represents the configuration schema for an MCP Server resource
// This would map to hashicorp/terraform-plugin-sdk in a real provider.
//
// Summary: Represents a ResourceMCPServer.
type ResourceMCPServer struct {
	Name    string `json:"name"`
	Port    int    `json:"port"`
	Enabled bool   `json:"enabled"`
}

// Schema returns the Terraform schema definition (Mock). Returns the result.
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
// Summary: Executes Schema operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
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

// Create mimics the Create operation of a Terraform resource. Returns an error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - serverURL (string): The URL of the MCP server API.
//   - resource (*ResourceMCPServer): The resource to create.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the HTTP request fails or the status is not created/ok.
//
// Side Effects:
//   - None
//
// Summary: Initializes Create operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.
func Create(ctx context.Context, serverURL string, resource *ResourceMCPServer) error {
	payload, err := json.Marshal(resource)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", serverURL+"/api/v1/servers", bytes.NewBuffer(payload)) //nolint:noctx,gosec
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create server: status %d", resp.StatusCode)
	}
	return nil
}

// Read mimics the Read operation. Returns the result or an error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - serverURL (string): The URL of the MCP server API.
//   - name (string): The name of the resource.
//
// Returns:
//   - *ResourceMCPServer: The resulting *ResourceMCPServer.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the HTTP request fails or the status is not ok.
//
// Side Effects:
//   - None
//
// Summary: Retrieves Read operation.
//
// Parameters:
//   - TODO: Document parameters.
//
// Returns:
//   - TODO: Document returns.
//
// Errors:
//   - TODO: Document errors.
//
// Side Effects:
//   - None.
func Read(ctx context.Context, serverURL string, name string) (*ResourceMCPServer, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", serverURL+"/api/v1/servers/"+name, nil) //nolint:noctx,gosec
	if err != nil {
		return nil, err
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // not found
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to read server: status %d", resp.StatusCode)
	}
	var res ResourceMCPServer
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return &res, nil
}
