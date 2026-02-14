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
// Implementations should search for specific services (e.g., Docker containers,
// local processes) and return their configurations as UpstreamServiceConfig.
type Provider interface {
	// Name returns the unique name of the discovery provider.
	//
	// Returns:
	//   - string: The provider name.
	Name() string
	// Discover attempts to find services and return their configurations.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the discovery operation.
	//
	// Returns:
	//   - []*configv1.UpstreamServiceConfig: A list of discovered service configurations.
	//   - error: An error if discovery fails significantly (partial failures should ideally be logged).
	Discover(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error)
}

// OllamaProvider discovers local Ollama instances.
//
// It attempts to connect to the configured endpoint (defaulting to localhost:11434)
// and, if successful, returns a configuration to proxy requests to it.
type OllamaProvider struct {
	// Endpoint is the URL of the Ollama server (e.g., "http://localhost:11434").
	Endpoint   string
	client     *http.Client
	clientOnce sync.Once
}

// Name returns the name of the provider.
//
// Summary: Retrieves the provider name.
//
// Returns:
//   - string: "ollama".
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// Discover attempts to find local Ollama instances and return them as tools.
//
// It checks if the Ollama endpoint is reachable. If so, it returns an
// HTTP upstream service configuration pointing to the /v1 OpenAI-compatible endpoint.
//
// Summary: Discovers local Ollama services.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: A list containing the Ollama service config if found.
//   - error: An error if the check fails.
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
