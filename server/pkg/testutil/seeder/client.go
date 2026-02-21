// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package seeder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Client is a helper for calling the seed endpoint.
type Client struct {
	BaseURL string
	Client  *http.Client
	APIKey  string
}

// NewClient creates a new seeder client.
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		Client:  &http.Client{},
		APIKey:  apiKey,
	}
}

// SeedRequest mirrors the server's SeedRequest struct.
type SeedRequest struct {
	ServicesRaw    []json.RawMessage `json:"upstream_services"`
	CredentialsRaw []json.RawMessage `json:"credentials"`
	SecretsRaw     []json.RawMessage `json:"secrets"`
	ProfilesRaw    []json.RawMessage `json:"profiles"`
	UsersRaw       []json.RawMessage `json:"users"`
	SettingsRaw    json.RawMessage   `json:"global_settings,omitempty"`
}

// Seed sends a seed request to the server.
func (c *Client) Seed(req SeedRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal seed request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/api/v1/debug/seed", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("X-API-Key", c.APIKey)
	}

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send seed request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("seed request failed with status: %s", resp.Status)
	}

	return nil
}
