/*
 * Copyright 2024 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package middleware

import (
	"net"
	"net/http"

	"github.com/mcpany/core/pkg/logging"
)

// IPAllowlist is a middleware that checks if the request's IP address is in the allowlist.
type IPAllowlist struct {
	allowedNetworks []*net.IPNet
}

// NewIPAllowlist creates a new IPAllowlist middleware.
func NewIPAllowlist(allowedIPs []string) (*IPAllowlist, error) {
	var allowedNetworks []*net.IPNet
	for _, ipStr := range allowedIPs {
		_, ipNet, err := net.ParseCIDR(ipStr)
		if err != nil {
			ip := net.ParseIP(ipStr)
			if ip == nil {
				return nil, err
			}
			ipNet = &net.IPNet{IP: ip, Mask: net.CIDRMask(len(ip)*8, len(ip)*8)}
		}
		allowedNetworks = append(allowedNetworks, ipNet)
	}
	return &IPAllowlist{allowedNetworks: allowedNetworks}, nil
}

// Handler is the middleware handler.
func (i *IPAllowlist) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logging.GetLogger()
		ipStr, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Error("Failed to split host and port", "error", err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		ip := net.ParseIP(ipStr)
		if ip == nil {
			log.Error("Failed to parse IP address", "ip", ipStr)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		allowed := false
		for _, network := range i.allowedNetworks {
			if network.Contains(ip) {
				allowed = true
				break
			}
		}

		if !allowed {
			log.Warn("IP address not in allowlist", "ip", ipStr)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
