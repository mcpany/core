// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"text/template"
	"time"

	"github.com/PaesslerAG/jsonpath"
)

// HTTPEmbeddingProvider represents the public HTTPEmbeddingProvider entity.
//
// Summary: Defines the structured data model representing a embedding provider.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type HTTPEmbeddingProvider struct {
	url              string
	headers          map[string]string
	bodyTemplate     *template.Template
	responseJSONPath string
	client           *http.Client
}

// NewHTTPEmbeddingProvider serves as a public interface for interacting with NewHTTPEmbeddingProvider.
//
// Summary: Constructs and returns an initialized http embedding provider ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewHTTPEmbeddingProvider(url string, headers map[string]string, bodyTemplateStr, responseJSONPath string) (*HTTPEmbeddingProvider, error) {
	if url == "" {
		return nil, fmt.Errorf("url is required")
	}

	tmpl, err := template.New("body").Parse(bodyTemplateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid body template: %w", err)
	}

	return &HTTPEmbeddingProvider{
		url:              url,
		headers:          headers,
		bodyTemplate:     tmpl,
		responseJSONPath: responseJSONPath,
		client:           &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Embed serves as a public interface for interacting with Embed.
//
// Summary: Embed the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (p *HTTPEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	// Simple template replacement.
	// We assume formatting is handled by the caller or configuration?
	// Actually, bodyTemplate is a go template.
	var bodyBuffer bytes.Buffer
	if err := p.bodyTemplate.Execute(&bodyBuffer, map[string]string{"input": text}); err != nil {
		return nil, fmt.Errorf("failed to execute body template: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.url, &bodyBuffer)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range p.headers {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse JSON
	var v interface{}
	if err := json.Unmarshal(respBody, &v); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response json: %w", err)
	}

	// Extract using JSONPath
	result, err := jsonpath.Get(p.responseJSONPath, v)
	if err != nil {
		return nil, fmt.Errorf("jsonpath extraction failed: %w", err)
	}

	// Convert result to []float32
	// jsonpath might return []interface{} where items are float64
	interfaceSlice, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("jsonpath result is not an array: %T", result)
	}

	embedding := make([]float32, len(interfaceSlice))
	for i, item := range interfaceSlice {
		if f, ok := item.(float64); ok {
			embedding[i] = float32(f)
		} else {
			return nil, fmt.Errorf("embedding element %d is not a number: %T", i, item)
		}
	}

	if len(embedding) == 0 {
		return nil, fmt.Errorf("empty embedding returned")
	}

	return embedding, nil
}
