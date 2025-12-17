// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/mcpany/core/pkg/logging"
)

// IPAllowlistMiddleware creates an HTTP middleware that restricts access
// to a list of allowed IP addresses or CIDR blocks.
//
// If the allowlist is empty, all requests are allowed.
func IPAllowlistMiddleware(allowedIPs []string) func(http.Handler) http.Handler {
	if len(allowedIPs) == 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	var ipNets []*net.IPNet
	for _, ipStr := range allowedIPs {
		if !strings.Contains(ipStr, "/") {
			// Assume single IP, convert to /32 or /128
			ip := net.ParseIP(ipStr)
			if ip == nil {
				logging.GetLogger().Warn("Invalid IP address in allowlist", "ip", ipStr)
				continue
			}
			if ip.To4() != nil {
				ipStr += "/32"
			} else {
				ipStr += "/128"
			}
		}
		_, ipNet, err := net.ParseCIDR(ipStr)
		if err != nil {
			logging.GetLogger().Warn("Invalid CIDR in allowlist", "cidr", ipStr, "error", err)
			continue
		}
		ipNets = append(ipNets, ipNet)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				// Fallback if RemoteAddr doesn't have port (unlikely for HTTP)
				host = r.RemoteAddr
			}

			// Handle [IPv6] format
			host = strings.TrimPrefix(host, "[")
			host = strings.TrimSuffix(host, "]")

			ip := net.ParseIP(host)
			if ip == nil {
				logging.GetLogger().Warn("Could not parse remote IP", "remote_addr", r.RemoteAddr)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			allowed := false
			for _, ipNet := range ipNets {
				if ipNet.Contains(ip) {
					allowed = true
					break
				}
			}

			if !allowed {
				logging.GetLogger().Warn("Access denied", "remote_ip", host)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
