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
//
// Summary: gRPC credentials adapter for upstream authentication.
type PerRPCCredentials struct {
	authenticator UpstreamAuthenticator
}

// NewPerRPCCredentials creates a new gRPC PerRPCCredentials from an UpstreamAuthenticator.
//
// Summary: Initializes NewPerRPCCredentials operation.
//
// Parameters:
//   - authenticator: UpstreamAuthenticator. The upstream authenticator to be used for generating gRPC request metadata.
//
// Returns:
//   - credentials.PerRPCCredentials: The gRPC credentials object.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewPerRPCCredentials(authenticator UpstreamAuthenticator) credentials.PerRPCCredentials {
	if authenticator == nil {
		return nil
	}
	return &PerRPCCredentials{authenticator: authenticator}
}

// GetRequestMetadata retrieves the authentication metadata for an outgoing gRPC request.
//
// Summary: Retrieves GetRequestMetadata operation.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - uri: ...string. Optional URIs.
//
// Returns:
//   - map[string]string: The metadata map.
//   - error: An error if metadata retrieval fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
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

// RequireTransportSecurity indicates whether a secure transport is required.
//
// Summary: Executes RequireTransportSecurity operation.
//
// Returns:
//   - bool: True if transport security is required, false otherwise.
//
// Parameters:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (c *PerRPCCredentials) RequireTransportSecurity() bool {
	// This should be true if TLS is enabled for the gRPC connection.
	// For now, returning false to align with the current insecure setup.
	return false
}
