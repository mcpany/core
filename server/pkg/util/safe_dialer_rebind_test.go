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
	r.callCount++
	return []net.IP{r.safeIP}, nil
}

type rebindMockDialer struct {
	dialedAddr string
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
	d.dialedAddr = address
	return &rebindMockConn{}, nil
}

type rebindMockConn struct{}

// Read ...
// Summary: Read
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Write ...
// Summary: Write
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Close ...
// Summary: Close
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// LocalAddr ...
// Summary: LocalAddr
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// RemoteAddr ...
// Summary: RemoteAddr
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// SetDeadline ...
// Summary: SetDeadline
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// SetReadDeadline ...
// Summary: SetReadDeadline
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// SetWriteDeadline ...
// Summary: SetWriteDeadline
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// TestSafeDialer_DNSRebindingProtection ...
// Summary: TestSafeDialer_DNSRebindingProtection
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
