// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"net"

	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
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
// authenticator is the upstream authenticator to be used for generating gRPC
// request metadata.
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
// ctx is the context for the request.
// uri is the URI of the gRPC service being called.
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
func (c *PerRPCCredentials) RequireTransportSecurity() bool {
	// This should be true if TLS is enabled for the gRPC connection.
	// For now, returning false to align with the current insecure setup.
	return false
}

// NewUnaryServerInterceptor creates a unary server interceptor for authentication.
func NewUnaryServerInterceptor(authManager *Manager, trustProxy bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx, err := validateGRPCAuth(ctx, authManager, trustProxy)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

// NewStreamServerInterceptor creates a stream server interceptor for authentication.
func NewStreamServerInterceptor(authManager *Manager, trustProxy bool) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx, err := validateGRPCAuth(ss.Context(), authManager, trustProxy)
		if err != nil {
			return err
		}
		// Wrapper for stream to return new context
		wrapped := &util.WrappedServerStream{
			ServerStream: ss,
			Ctx:          newCtx,
		}
		return handler(srv, wrapped)
	}
}

// validateGRPCAuth performs the authentication check for gRPC requests.
func validateGRPCAuth(ctx context.Context, authManager *Manager, trustProxy bool) (context.Context, error) {
	// 1. Extract credentials from metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	// Create a dummy HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "/", nil)
	if err != nil {
		return ctx, status.Error(codes.Internal, "failed to create request for auth")
	}

	// Populate headers from metadata
	// Common auth headers
	if val := md.Get("x-api-key"); len(val) > 0 {
		req.Header.Set("X-API-Key", val[0])
	}
	if val := md.Get("authorization"); len(val) > 0 {
		req.Header.Set("Authorization", val[0])
	}

	// 2. Authenticate using AuthManager
	// We pass empty serviceID to trigger global auth checks (API Key or Basic Auth)
	if newCtx, err := authManager.Authenticate(ctx, "", req); err == nil {
		return newCtx, nil
	}

	// 3. Fallback: Check if we allow localhost access when NO auth is configured
	if !authManager.HasGlobalAuth() {
		p, ok := peer.FromContext(ctx)
		if !ok {
			return ctx, status.Error(codes.Unauthenticated, "unauthorized: no peer info")
		}

		// Check IP
		ipStr := util.ExtractIP(p.Addr.String())
		// If trustProxy is true, we might need to look at X-Forwarded-For equivalent in metadata?
		// But for gRPC, usually we look at direct connection or X-Forwarded-For if set by gateway.
		// util.ExtractIP handles stripping port.
		// NOTE: gRPC Gateway sets X-Forwarded-For.
		if trustProxy {
			if vals := md.Get("x-forwarded-for"); len(vals) > 0 {
				// Use the first IP in the list (client IP)
				ips := strings.Split(vals[0], ",")
				if len(ips) > 0 {
					ipStr = strings.TrimSpace(ips[0])
				}
			}
		}

		ip := net.ParseIP(ipStr)
		if util.IsPrivateIP(ip) {
			return ctx, nil
		}
	}

	return ctx, status.Error(codes.Unauthenticated, "unauthorized")
}
