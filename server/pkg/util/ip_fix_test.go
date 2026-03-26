// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"net"
	"testing"
)

// TestIsPrivateNetworkIP_LoopbackIPv6 ...
// Summary: TestIsPrivateNetworkIP_LoopbackIPv6
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	ip := net.ParseIP("::1")
	if IsPrivateNetworkIP(ip) {
		t.Errorf("IsPrivateNetworkIP(::1) = true; want false (loopback is not private network)")
	}
}

type reproMockDialer struct {
	called bool
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
	m.called = true
	// Simulate success or failure doesn't matter, we just check if it was called.
	// But returning nil conn might cause panic if caller uses it.
	// DialContext in SafeDialer returns (net.Conn, error).
	// If we return nil, nil, SafeDialer returns nil, nil.
	return nil, nil
}

// TestSafeDialer_LoopbackIPv6_Allowed ...
// Summary: TestSafeDialer_LoopbackIPv6_Allowed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Setup SafeDialer with AllowLoopback=true, AllowPrivate=false
	md := &reproMockDialer{}
	dialer := &SafeDialer{
		AllowLoopback: true,
		AllowPrivate:  false,
		Dialer:        md,
	}

	// Try to dial ::1
	// We use [::1] literal to ensure it resolves to IPv6 loopback
	_, err := dialer.DialContext(context.Background(), "tcp", "[::1]:80")

	// Expectation: No error about private IP.
	// If the bug exists, it will fail with "ssrf attempt blocked: host ::1 resolved to private ip ::1"

	if err != nil {
		t.Fatalf("DialContext failed: %v", err)
	}

	if !md.called {
		t.Error("Inner dialer was not called, meaning it was blocked")
	}
}
