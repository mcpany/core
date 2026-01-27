// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"net"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockResolver implements IPResolver interface for testing
type MockResolver struct {
	IPs []net.IP
	Err error
}

func (m *MockResolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	return m.IPs, m.Err
}

func TestSafeDialer_BlocksMulticast(t *testing.T) {
	// 224.0.1.1 is Global Multicast (Internetwork Control Block)
	// It is NOT Link-Local.
	// It should be blocked by SafeDialer to prevent SSRF.

	dialer := NewSafeDialer()
	// Mock resolver to return multicast IP
	dialer.Resolver = &MockResolver{
		IPs: []net.IP{net.ParseIP("224.0.1.1")},
	}

	// Mock Dialer to avoid actual network call (though it would fail anyway)
	// But we want to ensure it fails at the *check* phase.
	// If checks pass, DialContext is called, which returns nil/error.
	// We want to see the specific SSRF error.

	ctx := context.Background()
	_, err := dialer.DialContext(ctx, "tcp", "example.com:80")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "ssrf attempt blocked", "Should block global multicast IP")
	assert.Contains(t, err.Error(), "224.0.1.1")
}
