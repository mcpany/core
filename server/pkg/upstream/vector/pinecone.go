// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package vector provides vector database upstream implementations.
package vector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// PineconeClient represents the public PineconeClient entity.
//
// Summary: Defines the structured data model representing a client.
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
type PineconeClient struct {
	config  *configv1.PineconeVectorDB
	client  *http.Client
	baseURL string
}

// NewPineconeClient serves as a public interface for interacting with NewPineconeClient.
//
// Summary: Constructs and returns an initialized pinecone client ready for consumption.
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
func NewPineconeClient(config *configv1.PineconeVectorDB) (*PineconeClient, error) {
	if config.GetApiKey() == "" {
		return nil, fmt.Errorf("api_key is required for Pinecone")
	}

	// Determine Base URL
	// If host is provided, use it.
	// Otherwise construct it.
	baseURL := config.GetHost()
	if baseURL == "" {
		if config.GetIndexName() == "" || config.GetProjectId() == "" || config.GetEnvironment() == "" {
			return nil, fmt.Errorf("host OR (index_name, project_id, environment) must be provided")
		}
		// Legacy/Standard format: https://index-name-project-id.svc.environment.pinecone.io
		// But Pinecone serverless URLs are different.
		// Safe fallback is to require Host if users are on serverless?
		// Or try to construct:
		baseURL = fmt.Sprintf("https://%s-%s.svc.%s.pinecone.io", config.GetIndexName(), config.GetProjectId(), config.GetEnvironment())
	}

	return &PineconeClient{
		config:  config,
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: baseURL,
	}, nil
}

//nolint:unparam // method is always POST for now, but keeping it flexible for future
func (c *PineconeClient) doRequest(ctx context.Context, method, path string, body interface{}) (map[string]interface{}, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(data)
	}

	u, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Api-Key", c.config.GetApiKey())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	// errcheck: ensure body is closed
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("pinecone request failed: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return result, nil
}

// Query serves as a public interface for interacting with Query.
//
// Summary: Query the  appropriately based on current system conditions.
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
func (c *PineconeClient) Query(ctx context.Context, vector []float32, topK int64, filter map[string]interface{}, namespace string) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"vector":          vector,
		"topK":            topK,
		"includeMetadata": true,
		"includeValues":   false,
	}
	if filter != nil {
		req["filter"] = filter
	}
	if namespace != "" {
		req["namespace"] = namespace
	}

	return c.doRequest(ctx, "POST", "/query", req)
}

// Upsert serves as a public interface for interacting with Upsert.
//
// Summary: Upsert the  appropriately based on current system conditions.
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
func (c *PineconeClient) Upsert(ctx context.Context, vectors []map[string]interface{}, namespace string) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"vectors": vectors,
	}
	if namespace != "" {
		req["namespace"] = namespace
	}

	return c.doRequest(ctx, "POST", "/vectors/upsert", req)
}

// Delete serves as a public interface for interacting with Delete.
//
// Summary: Delete the  appropriately based on current system conditions.
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
func (c *PineconeClient) Delete(ctx context.Context, ids []string, namespace string, filter map[string]interface{}) (map[string]interface{}, error) {
	req := map[string]interface{}{}
	if len(ids) > 0 {
		req["ids"] = ids
	}
	if filter != nil {
		req["filter"] = filter
	}
	// "deleteAll" can only be used if ids and filter are empty
	if len(ids) == 0 && filter == nil {
		req["deleteAll"] = true
	}
	if namespace != "" {
		req["namespace"] = namespace
	}

	return c.doRequest(ctx, "POST", "/vectors/delete", req)
}

// DescribeIndexStats serves as a public interface for interacting with DescribeIndexStats.
//
// Summary: Describe the index stats appropriately based on current system conditions.
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
func (c *PineconeClient) DescribeIndexStats(ctx context.Context, filter map[string]interface{}) (map[string]interface{}, error) {
	req := map[string]interface{}{}
	if filter != nil {
		req["filter"] = filter
	}
	// DescribeIndexStats is usually a POST for Pinecone with optional filter
	return c.doRequest(ctx, "POST", "/describe_index_stats", req)
}
