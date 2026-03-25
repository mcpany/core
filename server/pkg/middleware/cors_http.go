// Copyright 2025 Author(s) of MCP Any
// NewHTTPCORSMiddleware creates a new HTTPCORSMiddleware.
//
// Summary: Initializes HTTP CORS middleware.
//
// If allowedOrigins is empty, it defaults to allowing nothing (or behaving like standard Same-Origin).
// To allow all, pass []string{"*"}.
//
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Parameters:
//   - allowedOrigins ([]string): The allowed origins.
//
// Returns:
// Update updates the allowed origins.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//
// Summary: Updates the allowed origins dynamically.
//
// Parameters:
//   - allowedOrigins ([]string): The new list of allowed origins.
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (m *HTTPCORSMiddleware) Update(allowedOrigins []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.updateInternal(allowedOrigins)
}

// updateInternal populates the internal map and flags.
// It must be called with the lock held or during initialization.
// ⚡ Bolt Optimization: Uses map for O(1) lookup instead of O(N) slice iteration.
func (m *HTTPCORSMiddleware) updateInternal(origins []string) {
	m.allowedOrigins = make(map[string]struct{}, len(origins))
// Handler wraps an http.Handler with CORS logic.
//
// Summary: Middleware to handle CORS headers.
//
// Parameters:
//   - next (http.Handler): The next handler in the chain.
//
// Returns:
//   - (http.Handler): The wrapped handler.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (m *HTTPCORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			// Not a CORS request
			next.ServeHTTP(w, r)
			return
		}

		m.mu.RLock()
		// Check for exact match first
		_, allowed := m.allowedOrigins[origin]
		wildcardAllowed := m.wildcardAllowed
		m.mu.RUnlock()

		if !allowed && !wildcardAllowed {
			// CORS check failed
			next.ServeHTTP(w, r)
			return
		}

		// Set CORS headers
		if allowed {
			// Exact match: Reflect origin and allow credentials
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Vary", "Origin")
		} else {
			// Wildcard match: Return "*" and NO credentials
			logging.GetLogger().Warn("CORS: Allowing wildcard origin", "origin", origin, "source", "HTTPCORSMiddleware")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			// No Access-Control-Allow-Credentials
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Requested-With, x-grpc-web, grpc-timeout, x-user-agent")
		w.Header().Set("Access-Control-Expose-Headers", "grpc-status, grpc-message, Date, Content-Length, Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
