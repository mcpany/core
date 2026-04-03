// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/mcpany/core/server/pkg/logging"
)

// IPAllowlistMiddleware represents the public IPAllowlistMiddleware entity.
//
// Summary: Defines the structured data model representing a allowlist middleware.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type IPAllowlistMiddleware struct {
	mu            sync.RWMutex
	allowedIPNets []*net.IPNet
}

// NewIPAllowlistMiddleware serves as a public interface for interacting with NewIPAllowlistMiddleware.
//
// Summary: Constructs and returns an initialized ip allowlist middleware ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewIPAllowlistMiddleware(allowedCIDRs []string) (*IPAllowlistMiddleware, error) {
	m := &IPAllowlistMiddleware{}
	if err := m.Update(allowedCIDRs); err != nil {
		return nil, err
	}
	return m, nil
}

// Update serves as a public interface for interacting with Update.
//
// Summary: Update the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *IPAllowlistMiddleware) Update(allowedCIDRs []string) error {
	nets := make([]*net.IPNet, 0, len(allowedCIDRs))
	for _, cidr := range allowedCIDRs {
		// Try parsing as CIDR first
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			nets = append(nets, ipNet)
			continue
		}

		// If not CIDR, try as single IP
		ip := net.ParseIP(cidr)
		if ip == nil {
			return fmt.Errorf("invalid IP or CIDR: %s", cidr)
		}

		// Convert single IP to /32 or /128
		mask := net.CIDRMask(32, 32)
		if ip.To4() == nil {
			mask = net.CIDRMask(128, 128)
		}
		nets = append(nets, &net.IPNet{IP: ip, Mask: mask})
	}

	m.mu.Lock()
	m.allowedIPNets = nets
	m.mu.Unlock()
	return nil
}

// Allow serves as a public interface for interacting with Allow.
//
// Summary: Allow the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *IPAllowlistMiddleware) Allow(remoteAddr string) bool {
	m.mu.RLock()
	nets := m.allowedIPNets
	m.mu.RUnlock()

	if len(nets) == 0 {
		return true
	}

	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}

	// Handle bracketed IPv6 if port was missing and brackets remained
	if len(host) > 0 && host[0] == '[' && host[len(host)-1] == ']' {
		host = host[1 : len(host)-1]
	}

	ip := net.ParseIP(host)
	if ip == nil {
		logging.GetLogger().Warn("Failed to parse remote IP", "remote_addr", remoteAddr)
		return false
	}

	for _, ipNet := range nets {
		if ipNet.Contains(ip) {
			return true
		}
	}
	logging.GetLogger().Warn("Access denied", "remote_ip", ip.String())
	return false
}

// Handler serves as a public interface for interacting with Handler.
//
// Summary: Handler the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *IPAllowlistMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.Allow(r.RemoteAddr) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
