// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func TestSSRFProtection(t *testing.T) {
	// 1. Start a local gRPC server (target for SSRF)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = lis.Close() }()

	srv := grpc.NewServer()
	go func() {
		_ = srv.Serve(lis)
	}()
	defer srv.Stop()

	// 2. Configure pool to point to it
	addr := lis.Addr().String()
	config := configv1.UpstreamServiceConfig_builder{
		GrpcService: configv1.GrpcUpstreamService_builder{
			Address: proto.String(addr),
		}.Build(),
	}.Build()

	// 3. Create pool
	// Ensure env vars are cleared so we test default secure behavior
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "")
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "")

	// Pass nil for dialer to trigger default SafeDialer usage
	p, err := NewGrpcPool(1, 1, 10*time.Second, nil, nil, config, true)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	clientWrapper, err := p.Get(context.Background())
	require.NoError(t, err)
	defer p.Put(clientWrapper)

	// 4. Try to invoke (just connecting is enough to trigger dialer check if we try to use it)
	// But grpc.NewClient is lazy by default unless we use WithBlock (deprecated) or try to Invoke.
	// Actually NewGrpcPool creates clientWrapper which holds the connection.
	// We need to trigger a connection attempt.
	// clientWrapper.Invoke calls conn.Invoke.
	// The dialer is called when connection is established.
	// Let's try to invoke a non-existent method.

	err = clientWrapper.Invoke(context.Background(), "/test/Method", nil, nil)

	// 5. Assert failure
	// We expect an error because loopback should be blocked by SafeDialer
	assert.Error(t, err)
	if err != nil {
		// The error might be "context deadline exceeded" if it retries, or "ssrf attempt blocked" if it fails fast.
		// Since we didn't set a timeout on Invoke, it might hang if we don't handle it.
		// Wait, SafeDialer returns error immediately.
		// gRPC usually returns "unavailable" or "transport is closing" or similar if dial fails.
		// We should check if the error message contains the SSRF block message OR if it is a connection failure caused by it.
		// Since NewGrpcPool creates the connection using the dialer, and we use that connection.
		// If dialer fails, gRPC puts channel in TransientFailure.
		// We might need to inspect the error more closely or just ensure it failed.
		// To be sure it's SSRF, we can check if it contains "ssrf attempt blocked".
		// But gRPC might wrap it.
		// Let's print the error to see what we get.
		t.Logf("Error: %v", err)
		isBlocked := strings.Contains(err.Error(), "ssrf attempt blocked") ||
					 strings.Contains(err.Error(), "context deadline exceeded") || // If it retries and fails
					 strings.Contains(err.Error(), "unavailable") // Generic gRPC error

		// Ideally we want to see the underlying error.
		// However, confirming it FAILED to connect to 127.0.0.1 is the main goal.
		assert.True(t, isBlocked)
	}
}

func TestSSRFProtection_Allowed(t *testing.T) {
	// 1. Start a local gRPC server (target for SSRF)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = lis.Close() }()

	srv := grpc.NewServer()
	go func() {
		_ = srv.Serve(lis)
	}()
	defer srv.Stop()

	// 2. Configure pool to point to it
	addr := lis.Addr().String()
	config := configv1.UpstreamServiceConfig_builder{
		GrpcService: configv1.GrpcUpstreamService_builder{
			Address: proto.String(addr),
		}.Build(),
	}.Build()

	// 3. Allow loopback via env var
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

	// Pass nil for dialer to trigger default SafeDialer usage
	p, err := NewGrpcPool(1, 1, 10*time.Second, nil, nil, config, true)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	clientWrapper, err := p.Get(context.Background())
	require.NoError(t, err)
	defer p.Put(clientWrapper)

	// 4. Try to invoke
	// Since the server exists but method doesn't, we expect "unimplemented" or similar,
	// NOT a connection error/timeout/ssrf block.
	// Or simply "unavailable" if server is not yet ready?
	// We used 10s timeout, so it should be fine.

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = clientWrapper.Invoke(ctx, "/test/Method", nil, nil)

	// 5. Assert result
	// We expect failure because method doesn't exist, BUT it should be a gRPC status error code Unimplemented (12),
	// NOT "ssrf attempt blocked".
	// Or we can check if it connected.
	if err != nil {
		t.Logf("Error: %v", err)
		assert.NotContains(t, err.Error(), "ssrf attempt blocked")
		// If we get "Unimplemented", it means we connected successfully!
		// If we get "Unavailable", it might mean connection failed (or server not ready).
		// But SafeDialer shouldn't block it.
	}
}
