// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/mcpany/core/pkg/logging"
)

// IPAllowlistMiddleware validates that the request comes from an allowed IP address.
type IPAllowlistMiddleware struct {
	allowedNets []*net.IPNet
	allowedIPs  []net.IP
}

// NewIPAllowlistMiddleware creates a new IPAllowlistMiddleware.
// It accepts a list of IP addresses or CIDR ranges.
// If the list is empty, all requests are allowed.
func NewIPAllowlistMiddleware(allowedList []string) (*IPAllowlistMiddleware, error) {
	var nets []*net.IPNet
	var ips []net.IP

	for _, s := range allowedList {
		if strings.Contains(s, "/") {
			_, n, err := net.ParseCIDR(s)
			if err != nil {
				return nil, fmt.Errorf("invalid CIDR %q: %w", s, err)
			}
			nets = append(nets, n)
		} else {
			ip := net.ParseIP(s)
			if ip == nil {
				return nil, fmt.Errorf("invalid IP address %q", s)
			}
			ips = append(ips, ip)
		}
	}

	return &IPAllowlistMiddleware{
		allowedNets: nets,
		allowedIPs:  ips,
	}, nil
}

// Handler returns an http.Handler that wraps the provided handler.
func (m *IPAllowlistMiddleware) Handler(next http.Handler) http.Handler {
	// If no restrictions are configured, allow all.
	if len(m.allowedNets) == 0 && len(m.allowedIPs) == 0 {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// If SplitHostPort fails, it might be because there is no port.
			// Try to use RemoteAddr as is.
			host = r.RemoteAddr
		}
		// Handle [::1] style IPv6 addresses if they lack port but have brackets?
		// SplitHostPort handles brackets. If it failed, it's likely just "127.0.0.1" or "::1" or malformed.
		// If it's "[::1]", SplitHostPort with no port will fail. We strip brackets manually if needed.
		host = strings.TrimPrefix(host, "[")
		host = strings.TrimSuffix(host, "]")

		ip := net.ParseIP(host)
		if ip == nil {
			logging.GetLogger().Warn("Could not parse remote IP", "remote_addr", r.RemoteAddr)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if !m.isAllowed(ip) {
			logging.GetLogger().Warn("Access denied", "ip", ip.String())
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *IPAllowlistMiddleware) isAllowed(ip net.IP) bool {
	for _, n := range m.allowedNets {
		if n.Contains(ip) {
			return true
		}
	}
	for _, allowedIP := range m.allowedIPs {
		if allowedIP.Equal(ip) {
			return true
		}
	}
	return false
}
