/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resource

import (
	"context"
	"fmt"
	"io"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// StaticResource implements the Resource interface for resources that are
// defined statically in the configuration (e.g. pointing to a URL).
type StaticResource struct {
	resource  *mcp.Resource
	serviceID string
}

// NewStaticResource creates a new instance of StaticResource.
func NewStaticResource(def *configv1.ResourceDefinition, serviceID string) *StaticResource {
	return &StaticResource{
		resource: &mcp.Resource{
			URI:         def.GetUri(),
			Name:        def.GetName(),
			Description: def.GetDescription(),
			MIMEType:    def.GetMimeType(),
			Size:        def.GetSize(),
		},
		serviceID: serviceID,
	}
}

// Resource returns the MCP representation of the resource.
func (r *StaticResource) Resource() *mcp.Resource {
	return r.resource
}

// Service returns the ID of the service that provides this resource.
func (r *StaticResource) Service() string {
	return r.serviceID
}

// Read retrieves the content of the resource by fetching the URI.
func (r *StaticResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	// Simple HTTP get for now
	// We might want to use a shared client or pool if available, but for now default http client.
	req, err := http.NewRequestWithContext(ctx, "GET", r.resource.URI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resource: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch resource, status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource body: %w", err)
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      r.resource.URI,
				Blob:     data,
				MIMEType: r.resource.MIMEType,
			},
		},
	}, nil
}

// Subscribe is not yet implemented for static resources.
func (r *StaticResource) Subscribe(_ context.Context) error {
	return fmt.Errorf("subscribing to static resources is not yet implemented")
}
