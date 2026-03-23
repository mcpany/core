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
	"time"
)

// httpClient is a shared HTTP client with a timeout for the terraform provider.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// ResourceMCPServer represents the configuration schema for an MCP Server resource
// This would map to hashicorp/terraform-plugin-sdk in a real provider.
//
// Summary: Represents a ResourceMCPServer.
type ResourceMCPServer struct {
	Name    string `json:"name"`
	Port    int    `json:"port"`
	Enabled bool   `json:"enabled"`
}

// Schema returns the Terraform schema definition.
//
// Summary: Executes Schema operation.
//
// Parameters:
//   - None
//
// Returns:
//   - map[string]interface{}: The resulting schema map.
//
// Errors:
//   - None
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
// Summary: Initializes Create operation.
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
//   - Makes an HTTP POST request to the serverURL.
func Create(ctx context.Context, serverURL string, resource *ResourceMCPServer) error {
	payload, err := json.Marshal(resource)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", serverURL+"/api/v1/servers", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create server: status %d", resp.StatusCode)
	}
	return nil
}

// Read mimics the Read operation.
//
// Summary: Retrieves Read operation.
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
//   - Makes an HTTP GET request to the serverURL.
func Read(ctx context.Context, serverURL string, name string) (*ResourceMCPServer, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", serverURL+"/api/v1/servers/"+name, nil)
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // not found.
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to read server: status %d", resp.StatusCode)
	}
	var res ResourceMCPServer
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}
	return &res, nil
}
