// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockRebindingResolver struct {
	callCount int
	safeIP    net.IP
}

func (r *mockRebindingResolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	r.callCount++
	return []net.IP{r.safeIP}, nil
}

type rebindMockDialer struct {
	dialedAddr string
}

func (d *rebindMockDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	d.dialedAddr = address
	return &rebindMockConn{}, nil
}

type rebindMockConn struct{}

func (m *rebindMockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *rebindMockConn) Write(b []byte) (n int, err error)  { return 0, nil }
func (m *rebindMockConn) Close() error                       { return nil }
func (m *rebindMockConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (m *rebindMockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (m *rebindMockConn) SetDeadline(t time.Time) error      { return nil }
func (m *rebindMockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *rebindMockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestSafeDialer_DNSRebindingProtection(t *testing.T) {
	// This test verifies that SafeDialer resolves the IP once, validates it,
	// and then dials the *resolved IP* directly.
	// This prevents TOCTOU (Time-of-Check Time-of-Use) DNS rebinding attacks
	// where an attacker changes the DNS record between the check and the dial.

	safeIP := net.ParseIP("1.1.1.1")
	resolver := &mockRebindingResolver{
		safeIP: safeIP,
	}

	mockD := &rebindMockDialer{}

	safeDialer := NewSafeDialer()
	safeDialer.Resolver = resolver
	safeDialer.Dialer = mockD

	_, _ = safeDialer.DialContext(context.Background(), "tcp", "attacker.com:80")

	// Critical Check: The underlying dialer must receive the IP address, NOT the hostname.
	// If it received "attacker.com:80", the underlying dialer would resolve it again,
	// vulnerable to rebinding.
	expectedAddr := net.JoinHostPort(safeIP.String(), "80")
	assert.Equal(t, expectedAddr, mockD.dialedAddr, "SafeDialer should dial the resolved IP directly to prevent DNS rebinding")

	// Ensure resolution happened exactly once
	assert.Equal(t, 1, resolver.callCount)
}
