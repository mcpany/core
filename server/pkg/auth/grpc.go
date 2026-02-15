// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/grpc/credentials"
)

// PerRPCCredentials adapts an UpstreamAuthenticator to the gRPC
// credentials.PerRPCCredentials interface. It allows applying upstream
// authentication headers to outgoing gRPC requests.
type PerRPCCredentials struct {
	authenticator UpstreamAuthenticator
}

// NewPerRPCCredentials creates a new gRPC PerRPCCredentials from an
// UpstreamAuthenticator. It returns nil if the provided authenticator is nil.
//
// Summary: Creates gRPC credentials from an upstream authenticator.
//
// Parameters:
//   - authenticator: UpstreamAuthenticator. The authenticator to use.
//
// Returns:
//   - credentials.PerRPCCredentials: The gRPC credentials implementation.
func NewPerRPCCredentials(authenticator UpstreamAuthenticator) credentials.PerRPCCredentials {
	if authenticator == nil {
		return nil
	}
	return &PerRPCCredentials{authenticator: authenticator}
}

// GetRequestMetadata retrieves the authentication metadata for an outgoing gRPC
// request. It uses the wrapped UpstreamAuthenticator to generate the necessary
// headers and transforms them into gRPC metadata.
//
// Summary: Generates gRPC metadata for authentication.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - _: ...string. The URI of the gRPC service (unused).
//
// Returns:
//   - map[string]string: The metadata map.
//   - error: An error if authentication fails.
func (c *PerRPCCredentials) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	if c.authenticator == nil {
		return nil, nil
	}

	// Create a dummy http.Request to leverage the existing Authenticate method.
	dummyReq, err := http.NewRequestWithContext(ctx, "POST", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create dummy request for grpc auth: %w", err)
	}

	if err := c.authenticator.Authenticate(dummyReq); err != nil {
		return nil, fmt.Errorf("failed to apply upstream authenticator for grpc: %w", err)
	}

	metadata := make(map[string]string)
	for key, values := range dummyReq.Header {
		// gRPC metadata keys are lowercased.
		metadata[strings.ToLower(key)] = strings.Join(values, ",")
	}

	return metadata, nil
}

// RequireTransportSecurity indicates whether a secure transport (e.g., TLS) is
// required for the credentials. This implementation returns false, but should be
// updated if TLS is enabled for the gRPC connection.
//
// Summary: Checks if transport security is required.
//
// Returns:
//   - bool: True if TLS is required, false otherwise.
func (c *PerRPCCredentials) RequireTransportSecurity() bool {
	// This should be true if TLS is enabled for the gRPC connection.
	// For now, returning false to align with the current insecure setup.
	return false
}
