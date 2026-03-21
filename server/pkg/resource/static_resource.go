// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Summary: StaticResource implements the Resource interface for resources that are defined statically in the configuration (e.g. pointing to a URL). Represents a StaticResource.
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
type StaticResource struct {
	resource      *mcp.Resource
	serviceID     string
	httpClient    *http.Client
	staticContent *configv1.StaticResource
}

// Summary: NewStaticResource creates a new instance of StaticResource. Initializes a static resource.
//
// Parameters:
//   - def (*configv1.ResourceDefinition): The def parameter.
//   - serviceID (string): The serviceID parameter.
//
// Returns:
//   - *StaticResource: The resulting *StaticResource.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewStaticResource(def *configv1.ResourceDefinition, serviceID string) *StaticResource {
	return &StaticResource{
		resource: &mcp.Resource{
			URI:         def.GetUri(),
			Name:        def.GetName(),
			Description: def.GetDescription(),
			MIMEType:    def.GetMimeType(),
			Size:        def.GetSize(),
		},
		serviceID:     serviceID,
		httpClient:    util.NewSafeHTTPClient(),
		staticContent: def.GetStatic(),
	}
}

// Summary: Resource returns the MCP representation of the resource. Retrieves the MCP resource metadata.
//
// Parameters:
//   - None.
//
// Returns:
//   - *mcp.Resource: The resulting *mcp.Resource.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (r *StaticResource) Resource() *mcp.Resource {
	return r.resource
}

// Summary: Service returns the ID of the service that provides this resource. Retrieves the service ID.
//
// Parameters:
//   - None.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (r *StaticResource) Service() string {
	return r.serviceID
}

// Summary: Read retrieves the content of the resource by fetching the URI. Fetches the resource content.
//
// Parameters:
//   - ctx (context.Context): The ctx parameter.
//
// Returns:
//   - *mcp.ReadResourceResult: The resulting *mcp.ReadResourceResult.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (r *StaticResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	if r.staticContent != nil {
		var blob []byte
		switch r.staticContent.WhichContentType() {
		case configv1.StaticResource_TextContent_case:
			blob = []byte(r.staticContent.GetTextContent())
		case configv1.StaticResource_BinaryContent_case:
			blob = r.staticContent.GetBinaryContent()
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      r.resource.URI,
					Blob:     blob,
					MIMEType: r.resource.MIMEType,
				},
			},
		}, nil
	}

	// Simple HTTP get for now
	// We might want to use a shared client or pool if available, but for now default http client.
	req, err := http.NewRequestWithContext(ctx, "GET", r.resource.URI, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resource: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch resource, status: %d", resp.StatusCode)
	}

	// Limit the size of the resource to prevent DoS attacks (OOM).
	// Default to 10MB if not specified.
	const defaultMaxResourceSize = 10 * 1024 * 1024 // 10MB
	limit := int64(defaultMaxResourceSize)
	if r.resource.Size > 0 {
		limit = r.resource.Size
	}

	var reader io.Reader
	if limit == math.MaxInt64 {
		// If limit is MaxInt64, limit+1 would overflow to negative, causing LimitReader to read 0 bytes.
		// In this case, we read everything without a limit reader (or rely on io.ReadAll reading everything).
		reader = resp.Body
	} else {
		// Read up to limit + 1 to detect if the resource exceeds the limit
		reader = io.LimitReader(resp.Body, limit+1)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read resource body: %w", err)
	}

	if int64(len(data)) > limit {
		return nil, fmt.Errorf("resource size exceeds limit of %d bytes", limit)
	}

	mimeType := r.resource.MIMEType
	if mimeType == "" {
		mimeType = resp.Header.Get("Content-Type")
	}

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      r.resource.URI,
				Blob:     data,
				MIMEType: mimeType,
			},
		},
	}, nil
}

// Summary: Subscribe is not yet implemented for static resources. Subscribes to resource updates (Not Implemented).
//
// Parameters:
//   - _ (context.Context): The _ parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
func (r *StaticResource) Subscribe(_ context.Context) error {
	return fmt.Errorf("subscribing to static resources is not yet implemented")
}
