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
// Summary: Represents a PerRPCCredentials.
type PerRPCCredentials struct {
	authenticator UpstreamAuthenticator
}

// NewPerRPCCredentials creates a new per rpc credentials.
//
// Summary: Creates a new per rpc credentials.
//
// Parameters: - None.
//   - authenticator (UpstreamAuthenticator): The authenticator.
//
// Returns: - None.
//   - credentials.PerRPCCredentials: The result.
func NewPerRPCCredentials(authenticator UpstreamAuthenticator) credentials.PerRPCCredentials {
	if authenticator == nil {
		return nil
	}
	return &PerRPCCredentials{authenticator: authenticator}
}

// GetRequestMetadata retrieves the request metadata.
//
// Summary: Retrieves the request metadata.
//
// Parameters: - None.
//   - ctx (context.Context): The context for the request.
//   - _ (...string): Unused parameter.
//
// Returns: - None.
//   - map[string]string: The result.
//   - error: An error if the operation fails.
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

// RequireTransportSecurity requireTransportSecurity require transport security.
//
// Summary: RequireTransportSecurity require transport security.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - bool: The result.
func (c *PerRPCCredentials) RequireTransportSecurity() bool {
	// This should be true if TLS is enabled for the gRPC connection.
	// For now, returning false to align with the current insecure setup.
	return false
}
