// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"net"
	"testing"
)

// MockSafeDialerResolver implements IPResolver
// MockSafeDialerResolver implements IPResolver
// Summary: MockSafeDialerResolver
	ips []net.IP
	err error
}

// LookupIP ...
// Summary: LookupIP
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.ips, m.err
}

// MockSafeDialerDialer implements NetDialer
// MockSafeDialerDialer implements NetDialer
// Summary: MockSafeDialerDialer
	conn net.Conn
	err  error
}

// DialContext ...
// Summary: DialContext
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.conn, m.err
}

// TestSafeDialer_UnspecifiedAddress_Bypass ...
// Summary: TestSafeDialer_UnspecifiedAddress_Bypass
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestSafeDialer_UnspecifiedAddress_Allowed ...
// Summary: TestSafeDialer_UnspecifiedAddress_Allowed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestSafeDialer_UnspecifiedIPv6_Bypass ...
// Summary: TestSafeDialer_UnspecifiedIPv6_Bypass
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
