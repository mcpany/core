package util

import (
	"context"
	"net"
	"testing"
)

// MockSafeDialerResolver implements IPResolver
type MockSafeDialerResolver struct {
	ips []net.IP
	err error
}

func (m *MockSafeDialerResolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	return m.ips, m.err
}

// MockSafeDialerDialer implements NetDialer
type MockSafeDialerDialer struct {
	conn net.Conn
	err  error
}

func (m *MockSafeDialerDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return m.conn, m.err
}

func TestSafeDialer_UnspecifiedAddress_Bypass(t *testing.T) {
	// Scenario: AllowPrivate=true, AllowLoopback=false.
	// Dialing 0.0.0.0 (Unspecified) should be blocked because it resolves to localhost.

	resolver := &MockSafeDialerResolver{
		ips: []net.IP{net.ParseIP("0.0.0.0")},
	}
	dialer := &MockSafeDialerDialer{
		conn: &net.TCPConn{},
	}

	sd := NewSafeDialer()
	sd.AllowPrivate = true
	sd.AllowLoopback = false
	sd.Resolver = resolver
	sd.Dialer = dialer

	_, err := sd.DialContext(context.Background(), "tcp", "0.0.0.0:80")

	if err == nil {
		t.Errorf("Expected error (block) when dialing 0.0.0.0 with AllowLoopback=false, but got success")
	} else {
		// Verify error message if possible, but mainly we want an error
		t.Logf("Blocked as expected: %v", err)
	}
}

func TestSafeDialer_UnspecifiedAddress_Allowed(t *testing.T) {
	// Scenario: AllowPrivate=true, AllowLoopback=true.
	// Dialing 0.0.0.0 should be allowed.

	resolver := &MockSafeDialerResolver{
		ips: []net.IP{net.ParseIP("0.0.0.0")},
	}
	dialer := &MockSafeDialerDialer{
		conn: &net.TCPConn{},
	}

	sd := NewSafeDialer()
	sd.AllowPrivate = true
	sd.AllowLoopback = true
	sd.Resolver = resolver
	sd.Dialer = dialer

	_, err := sd.DialContext(context.Background(), "tcp", "0.0.0.0:80")

	if err != nil {
		t.Errorf("Expected success when dialing 0.0.0.0 with AllowLoopback=true, but got error: %v", err)
	}
}

func TestSafeDialer_UnspecifiedIPv6_Bypass(t *testing.T) {
	// Scenario: AllowPrivate=true, AllowLoopback=false.
	// Dialing :: (Unspecified IPv6) should be blocked.

	resolver := &MockSafeDialerResolver{
		ips: []net.IP{net.ParseIP("::")},
	}
	dialer := &MockSafeDialerDialer{
		conn: &net.TCPConn{},
	}

	sd := NewSafeDialer()
	sd.AllowPrivate = true
	sd.AllowLoopback = false
	sd.Resolver = resolver
	sd.Dialer = dialer

	_, err := sd.DialContext(context.Background(), "tcp6", "[::]:80")

	if err == nil {
		t.Errorf("Expected error (block) when dialing :: with AllowLoopback=false, but got success")
	}
}
