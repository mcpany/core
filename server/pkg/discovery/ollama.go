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
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
)

// Provider defines the interface for auto-discovering local services.
//
// Summary: Represents a Provider.
type Provider interface {
	// Name returns the name of the discovery provider.
	Name() string
	// Discover attempts to find services and return their configurations.
	Discover(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error)
}

// OllamaProvider discovers local Ollama instances.
// OllamaProvider is a provider that discovers local Ollama instances.
//
// Summary: Represents a OllamaProvider.
type OllamaProvider struct {
	Endpoint   string // e.g., "http://localhost:11434"
	client     *http.Client
	clientOnce sync.Once
}

// Name returns the name of the provider.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - string: The resulting string.
//
// Errors: - None.
//   - None.
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes Name operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// Discover attempts to find local Ollama instances and return them as tools.
//
// Parameters: - None.
//   - ctx (context.Context): The context for the request.
//
// Returns: - None.
//   - []*configv1.UpstreamServiceConfig: The resulting []*configv1.UpstreamServiceConfig.
//   - error: An error if the operation fails.
//
// Errors: - None.
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects: - None.
//   - None.
//
// Summary: Executes Discover operation.
//
// Parameters: - None.
//
// Returns: - None.
//
// Errors: - None.
//
// Side Effects: - None.
//   - None.
func (p *OllamaProvider) Discover(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	// ⚡ BOLT: Reuse http.Client to avoid socket exhaustion.
	// Randomized Selection from Top 5 High-Impact Targets
	p.clientOnce.Do(func() {
		p.client = &http.Client{
			Timeout: 2 * time.Second,
		}
	})

	req, err := http.NewRequestWithContext(ctx, "GET", p.Endpoint+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
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
		configv1.UpstreamServiceConfig_builder{
			Name:    proto.String("Local Ollama"),
			Version: proto.String("v1"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: proto.String(p.Endpoint + "/v1"),
			}.Build(),
			Tags: []string{"local-llm", "ollama", "openai-compatible"},
		}.Build(),
	}, nil
}
