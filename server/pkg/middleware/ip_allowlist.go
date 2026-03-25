// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware
// NewIPAllowlistMiddleware creates a new IPAllowlistMiddleware.
//
// Summary: Initializes the middleware with the initial list of allowed CIDRs.
//
// Parameters:
//   - allowedCIDRs: []string. A list of IP addresses or CIDR blocks to allow.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
//
// Returns:
// Update updates the allowlist with new CIDRs/IPs.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Summary: Dynamically updates the list of allowed IPs.
//
// Parameters:
//   - allowedCIDRs: []string. The new list of allowed IP addresses or CIDR blocks.
//
// Returns:
//   - error: An error if any of the provided CIDRs are invalid.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
// Allow checks if the given remote address is allowed.
//
// Summary: Checks if a remote address is in the allowed list.
//
// Parameters:
//   - remoteAddr: string. The remote address (IP or IP:Port).
//
// Returns:
//   - bool: True if allowed, false otherwise.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
// Handler returns an HTTP handler that enforces the allowlist.
//
// Summary: Returns an HTTP handler that blocks unauthorized IPs.
//
// Parameters:
//   - next: http.Handler. The next handler in the chain.
//
// Returns:
//   - http.Handler: The wrapped handler.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (m *IPAllowlistMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.Allow(r.RemoteAddr) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
