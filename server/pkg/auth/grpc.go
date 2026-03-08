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

// PerRPCCredentials adapts an UpstreamAuthenticator to the gRPC credentials.PerRPCCredentials interface. It allows applying upstream authentication headers to outgoing gRPC requests.
//
// Summary: PerRPCCredentials adapts an UpstreamAuthenticator to the gRPC credentials.PerRPCCredentials interface. It allows applying upstream authentication headers to outgoing gRPC requests.
//
// Fields:
//   - Contains the configuration and state properties required for PerRPCCredentials functionality.
type PerRPCCredentials struct {
	authenticator UpstreamAuthenticator
}

// NewPerRPCCredentials creates a new gRPC PerRPCCredentials from an UpstreamAuthenticator. It returns nil if the provided authenticator is nil. authenticator is the upstream authenticator to be used for generating gRPC request metadata.
//
// Parameters:
//   - authenticator (UpstreamAuthenticator): The authenticator parameter.
//
// Returns:
//   - credentials.PerRPCCredentials: The resulting credentials.PerRPCCredentials.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func NewPerRPCCredentials(authenticator UpstreamAuthenticator) credentials.PerRPCCredentials {
	if authenticator == nil {
		return nil
	}
	return &PerRPCCredentials{authenticator: authenticator}
}

// GetRequestMetadata retrieves the authentication metadata for an outgoing gRPC request. It uses the wrapped UpstreamAuthenticator to generate the necessary headers and transforms them into gRPC metadata. ctx is the context for the request. uri is the URI of the gRPC service being called.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - _ (...string): The _ parameter.
//
// Returns:
//   - map[string]string: The resulting map[string]string.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None
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

// RequireTransportSecurity indicates whether a secure transport (e.g., TLS) is required for the credentials. This implementation returns false, but should be updated if TLS is enabled for the gRPC connection.
//
// Parameters:
//   - None
//
// Returns:
//   - bool: True if successful, false otherwise.
//
// Errors:
//   - None
//
// Side Effects:
//   - None
func (c *PerRPCCredentials) RequireTransportSecurity() bool {
	// This should be true if TLS is enabled for the gRPC connection.
	// For now, returning false to align with the current insecure setup.
	return false
}
