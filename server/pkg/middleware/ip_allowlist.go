// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"fmt"
	"net"
	"net/http"

	"github.com/mcpany/core/pkg/logging"
)

// IPAllowlistMiddleware restricts access to allowed IP addresses.
type IPAllowlistMiddleware struct {
	allowedIPNets []*net.IPNet
}

// NewIPAllowlistMiddleware creates a new IPAllowlistMiddleware.
// allowedCIDRs is the allowedCIDRs.
// Returns the result, an error.
func NewIPAllowlistMiddleware(allowedCIDRs []string) (*IPAllowlistMiddleware, error) {
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
			return nil, fmt.Errorf("invalid IP or CIDR: %s", cidr)
		}

		// Convert single IP to /32 or /128
		mask := net.CIDRMask(32, 32)
		if ip.To4() == nil {
			mask = net.CIDRMask(128, 128)
		}
		nets = append(nets, &net.IPNet{IP: ip, Mask: mask})
	}

	return &IPAllowlistMiddleware{
		allowedIPNets: nets,
	}, nil
}

// Allow checks if the given remote address is allowed.
// remoteAddr should be in the form "IP" or "IP:Port".
// Returns the result.
func (m *IPAllowlistMiddleware) Allow(remoteAddr string) bool {
	if len(m.allowedIPNets) == 0 {
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

	for _, ipNet := range m.allowedIPNets {
		if ipNet.Contains(ip) {
			return true
		}
	}
	logging.GetLogger().Warn("Access denied", "remote_ip", ip.String())
	return false
}

// Handler returns an HTTP handler that enforces the allowlist.
// next is the next.
func (m *IPAllowlistMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.Allow(r.RemoteAddr) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
