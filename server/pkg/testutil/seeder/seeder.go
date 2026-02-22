// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package seeder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// SeedData contains the data to be seeded.
type SeedData struct {
	Services    []*configv1.UpstreamServiceConfig
	Credentials []*configv1.Credential
	Secrets     []*configv1.Secret
	Profiles    []*configv1.ProfileDefinition
	Users       []*configv1.User
}

// Seed sends a seed request to the server.
func Seed(ctx context.Context, baseURL string, apiKey string, data *SeedData) error {
	reqBody := map[string][]json.RawMessage{
		"upstream_services": {},
		"credentials":       {},
		"secrets":           {},
		"profiles":          {},
		"users":             {},
	}

	opts := protojson.MarshalOptions{EmitUnpopulated: true}

	for _, s := range data.Services {
		b, err := opts.Marshal(s)
		if err != nil {
			return fmt.Errorf("failed to marshal service %s: %w", s.GetName(), err)
		}
		reqBody["upstream_services"] = append(reqBody["upstream_services"], b)
	}
	for _, c := range data.Credentials {
		b, err := opts.Marshal(c)
		if err != nil {
			return fmt.Errorf("failed to marshal credential %s: %w", c.GetId(), err)
		}
		reqBody["credentials"] = append(reqBody["credentials"], b)
	}
	for _, s := range data.Secrets {
		b, err := opts.Marshal(s)
		if err != nil {
			return fmt.Errorf("failed to marshal secret %s: %w", s.GetId(), err)
		}
		reqBody["secrets"] = append(reqBody["secrets"], b)
	}
	for _, p := range data.Profiles {
		b, err := opts.Marshal(p)
		if err != nil {
			return fmt.Errorf("failed to marshal profile %s: %w", p.GetName(), err)
		}
		reqBody["profiles"] = append(reqBody["profiles"], b)
	}
	for _, u := range data.Users {
		b, err := opts.Marshal(u)
		if err != nil {
			return fmt.Errorf("failed to marshal user %s: %w", u.GetId(), err)
		}
		reqBody["users"] = append(reqBody["users"], b)
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/api/v1/debug/seed", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send seed request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("seed request failed with status: %d", resp.StatusCode)
	}

	return nil
}
