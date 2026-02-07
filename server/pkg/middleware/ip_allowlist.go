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
// It is thread-safe and supports dynamic updates.
type IPAllowlistMiddleware struct {
	mu            sync.RWMutex
	allowedIPNets []*net.IPNet
}

// NewIPAllowlistMiddleware creates a new IPAllowlistMiddleware.
//
// Summary: Initializes a new IP allowlist middleware with the provided CIDRs or IP addresses.
//
// Parameters:
//   - allowedCIDRs: []string. A list of allowed CIDR blocks or IP addresses.
//
// Returns:
//   - *IPAllowlistMiddleware: The initialized middleware.
//   - error: An error if any of the CIDRs/IPs are invalid.
func NewIPAllowlistMiddleware(allowedCIDRs []string) (*IPAllowlistMiddleware, error) {
	m := &IPAllowlistMiddleware{}
	if err := m.Update(allowedCIDRs); err != nil {
		return nil, err
	}
	return m, nil
}

// Update updates the allowlist with new CIDRs/IPs.
//
// Summary: Dynamically updates the list of allowed IPs.
//
// Parameters:
//   - allowedCIDRs: []string. A list of allowed CIDR blocks or IP addresses.
//
// Returns:
//   - error: An error if any of the CIDRs/IPs are invalid.
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
// Summary: Verifies if the remote address matches any of the allowed networks.
//
// Parameters:
//   - remoteAddr: string. The remote address (IP or IP:Port).
//
// Returns:
//   - bool: True if allowed, false otherwise.
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
// Summary: Middleware that blocks requests from unauthorized IP addresses.
//
// Parameters:
//   - next: http.Handler. The next handler in the chain.
//
// Returns:
//   - http.Handler: The wrapped handler with IP filtering.
func (m *IPAllowlistMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.Allow(r.RemoteAddr) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
