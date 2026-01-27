/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// Package discovery implements auto-discovery providers for tools.
package discovery

import (
	"context"
	"fmt"
	"net/http"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// Provider defines the interface for auto-discovering local services.
type Provider interface {
	// Name returns the name of the discovery provider.
	Name() string
	// Discover attempts to find services and return their configurations.
	Discover(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error)
}

// OllamaProvider discovers local Ollama instances.
// OllamaProvider is a provider that discovers local Ollama instances.
type OllamaProvider struct {
	Endpoint string // e.g., "http://localhost:11434"
}

// Name returns the name of the provider.
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// Discover attempts to find local Ollama instances and return them as tools.
func (p *OllamaProvider) Discover(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", p.Endpoint+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama not found at %s: %w", p.Endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	// If reachable, return a template config for Ollama as an OpenAI-compatible service
	// Note: Ollama has an OpenAI compatible /v1 endpoint now.
	return []*configv1.UpstreamServiceConfig{
		{
			Name:    proto.String("Local Ollama"),
			Version: proto.String("v1"),
			ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
				HttpService: &configv1.HttpUpstreamService{
					Address: proto.String(p.Endpoint + "/v1"),
				},
			},
			Tags: []string{"local-llm", "ollama", "openai-compatible"},
		},
	}, nil
}
