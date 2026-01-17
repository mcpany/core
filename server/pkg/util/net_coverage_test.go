// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"os"
	"testing"
)

// We don't import grpc here to avoid dependency issues if possible,
// but WrappedServerStream is in util package.
// If I use it, I need to know what it embeds.
// It embeds grpc.ServerStream.
// To use it in a struct literal, I don't strictly need to import grpc IF I don't field access the embedded field?
// No, I need it to be defined. But it is defined in the package.
// So `&WrappedServerStream{Ctx: ctx}` works.

func TestWrappedServerStream_Coverage(t *testing.T) {
	ctx := context.Background()
	ws := &WrappedServerStream{Ctx: ctx}
	if ws.Context() != ctx {
		t.Error("WrappedServerStream.Context() failed")
	}
}

func TestNewSafeHTTPClient_Coverage(t *testing.T) {
	// Defaults
	c := NewSafeHTTPClient()
	if c == nil {
		t.Fatal("NewSafeHTTPClient returned nil")
	}

	// Env vars
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	os.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")
	defer func() {
		os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
		os.Unsetenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES")
	}()

	c = NewSafeHTTPClient()
	if c == nil {
		t.Fatal("NewSafeHTTPClient returned nil with env vars")
	}
}

func TestCheckConnection_Coverage(t *testing.T) {
	// Should fail to connect to invalid address
	err := CheckConnection(context.Background(), "invalid-host-xyz:80")
	if err == nil {
		t.Error("CheckConnection should fail for invalid host")
	}

	// URL parsing
	err = CheckConnection(context.Background(), "http://invalid-host-xyz")
	if err == nil {
		t.Error("CheckConnection should fail for invalid host")
	}

	err = CheckConnection(context.Background(), "https://invalid-host-xyz")
	if err == nil {
		t.Error("CheckConnection should fail for invalid host")
	}

	// Invalid URL
	// "http://%42:80" -> parse error likely
	err = CheckConnection(context.Background(), "http://%42:80")
	if err == nil {
		t.Error("CheckConnection should fail for invalid url")
	}

	// SplitHostPort fails (no port, no scheme) -> assume port 80
	// "invalid-host" -> "invalid-host:80"
	err = CheckConnection(context.Background(), "invalid-host")
	if err == nil {
		t.Error("CheckConnection should fail")
	}
}
