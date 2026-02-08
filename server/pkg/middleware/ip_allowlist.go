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

// IPAllowlistMiddleware restricts access to allowed IP addresses.
//
// Summary: restricts access to allowed IP addresses.
type IPAllowlistMiddleware struct {
	mu            sync.RWMutex
	allowedIPNets []*net.IPNet
}

// NewIPAllowlistMiddleware creates a new IPAllowlistMiddleware.
//
// Summary: creates a new IPAllowlistMiddleware.
//
// Parameters:
//   - allowedCIDRs: []string. The allowedCIDRs.
//
// Returns:
//   - *IPAllowlistMiddleware: The *IPAllowlistMiddleware.
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func NewIPAllowlistMiddleware(allowedCIDRs []string) (*IPAllowlistMiddleware, error) {
	m := &IPAllowlistMiddleware{}
	if err := m.Update(allowedCIDRs); err != nil {
		return nil, err
	}
	return m, nil
}

// Update updates the allowlist with new CIDRs/IPs.
//
// Summary: updates the allowlist with new CIDRs/IPs.
//
// Parameters:
//   - allowedCIDRs: []string. The allowedCIDRs.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

// Allow checks if the given remote address is allowed.
//
// Summary: checks if the given remote address is allowed.
//
// Parameters:
//   - remoteAddr: string. The remoteAddr.
//
// Returns:
//   - bool: The bool.
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

// Handler returns an HTTP handler that enforces the allowlist.
//
// Summary: returns an HTTP handler that enforces the allowlist.
//
// Parameters:
//   - next: http.Handler. The next.
//
// Returns:
//   - http.Handler: The http.Handler.
func (m *IPAllowlistMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.Allow(r.RemoteAddr) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
