// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckConnection_Coverage(t *testing.T) {
	// Allow loopback for this test as we are testing connection checks on local listener
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

	// Start a listener
	ln, err := net.Listen("tcp", "127.0.0.1:24001")
	if err != nil {
		ln, err = net.Listen("tcp", "127.0.0.1:0") // Fallback
	}
	require.NoError(t, err)
	defer ln.Close()

	addr := ln.Addr().String()

	// 1. Success case with host:port
	err = CheckConnection(context.Background(), addr)
	assert.NoError(t, err)

	// 2. Success case with scheme
	err = CheckConnection(context.Background(), "http://"+addr)
	assert.NoError(t, err)

	// 3. Failure case
	// Pick a port that is likely closed.
	err = CheckConnection(context.Background(), "127.0.0.1:54321")
	assert.Error(t, err)

	// 4. Invalid address
	err = CheckConnection(context.Background(), "invalid:address:port")
	assert.Error(t, err)

	// 5. Invalid URL
	err = CheckConnection(context.Background(), "http://[::1]:namedport") // named port is invalid in URL
	assert.Error(t, err)
}

func TestSafeDialer_Coverage(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "false")

	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// By default, SafeDialer blocks loopback
	client := NewSafeHTTPClient()
	_, err := client.Get(ts.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "loopback")

	// Allow loopback via env
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	client = NewSafeHTTPClient()
	resp, err := client.Get(ts.URL)
	assert.NoError(t, err)
	resp.Body.Close()

	// Direct SafeDialContext usage
	// Should fail for loopback if default
	dialer := NewSafeDialer()
	// Split ts.URL (http://127.0.0.1:port)
	u, _ :=  ts.Listener.Addr().(*net.TCPAddr)
	addr := u.String()

	_, err = dialer.DialContext(context.Background(), "tcp", addr)
	assert.Error(t, err)
}

func TestCheckConnection_NoPort(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

	// Mocking CheckConnection's behavior for no port is hard because it defaults to 80.
	// But we can check that it fails or succeeds depending on port 80 accessibility.
	// Usually 127.0.0.1:80 is closed.

	err := CheckConnection(context.Background(), "127.0.0.1")
	// If it fails, that's fine. If it succeeds, that's also fine (if something is running).
	// But we want to exercise the code path:
	// host = address, port = "80"
	// So we assert that it doesn't panic.
	_ = err
}

// Mocks for SafeDialer tests

type spyResolver struct {
	ips         []net.IP
	err         error
	lastNetwork string
}

func (s *spyResolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	s.lastNetwork = network
	// Simulate behavior of a real resolver: if asked for ip4, only return IPv4.
	var filtered []net.IP
	for _, ip := range s.ips {
		isV4 := ip.To4() != nil
		if network == "ip4" && !isV4 {
			continue
		}
		if network == "ip6" && isV4 {
			continue
		}
		filtered = append(filtered, ip)
	}
	return filtered, s.err
}

type mockDialer struct {
	conn net.Conn
	err  error
}

func (m *mockDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return m.conn, m.err
}

func TestSafeDialer_Bug_MixedFamilies(t *testing.T) {
	// Scenario: Host resolves to a private IPv6 (loopback) and a public IPv4.
	// We dial with "tcp4", so we should ignore the IPv6 address and succeed with IPv4.

	privateIPv6 := net.ParseIP("::1") // Loopback, blocked by default
	publicIPv4 := net.ParseIP("8.8.8.8")

	resolver := &spyResolver{
		ips: []net.IP{privateIPv6, publicIPv4},
	}

	dialer := &mockDialer{
		conn: &net.TCPConn{}, // dummy
	}

	sd := NewSafeDialer()
	sd.Resolver = resolver
	sd.Dialer = dialer
	// Defaults: AllowLoopback=false, AllowPrivate=false.

	// Action: Dial "tcp4". We only want IPv4.
	// Since IPv4 (8.8.8.8) is public, this SHOULD succeed.
	// But currently, SafeDialer looks up "ip" (both), receives both from our spy (if logic is flawed) or resolver,
	// and checks ALL returned IPs.
	_, err := sd.DialContext(context.Background(), "tcp4", "example.com:80")

	if err != nil {
		t.Fatalf("SafeDialer.DialContext(tcp4) failed unexpectedly: %v", err)
	}

	if resolver.lastNetwork != "ip4" {
		t.Errorf("Expected resolver network to be 'ip4', got '%s'", resolver.lastNetwork)
	}
}

func TestSafeDialer_NetworkMapping(t *testing.T) {
	tests := []struct {
		network         string
		expectedLookup  string
	}{
		{"tcp", "ip"},
		{"tcp4", "ip4"},
		{"tcp6", "ip6"},
		{"udp", "ip"},
		{"udp4", "ip4"},
		{"udp6", "ip6"},
		{"ip", "ip"},
		{"ip4", "ip4"},
		{"ip6", "ip6"},
	}

	for _, tt := range tests {
		t.Run(tt.network, func(t *testing.T) {
			resolver := &spyResolver{
				ips: []net.IP{net.ParseIP("127.0.0.1")}, // dummy
			}
			dialer := &mockDialer{conn: &net.TCPConn{}}
			sd := NewSafeDialer()
			sd.Resolver = resolver
			sd.Dialer = dialer
			sd.AllowLoopback = true // Allow loopback so we don't block

			_, _ = sd.DialContext(context.Background(), tt.network, "127.0.0.1:80")

			if resolver.lastNetwork != tt.expectedLookup {
				t.Errorf("For network '%s', expected lookup '%s', got '%s'", tt.network, tt.expectedLookup, resolver.lastNetwork)
			}
		})
	}
}
