// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// SSOMiddleware serves as a public interface for interacting with SSOMiddleware.
//
// Summary: Sso the middleware appropriately based on current system conditions.
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
func SSOMiddleware(config *configv1.SSOConfig) func(http.Handler) http.Handler {
	// Reusable HTTP client for IDP requests
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config == nil || !config.GetEnabled() {
				next.ServeHTTP(w, r)
				return
			}

			// 1. Check for Identity Header (Trusted Proxy pattern)
			// In a real production environment, this header MUST be stripped
			// by the edge proxy before reaching the application to prevent spoofing.
			userID := r.Header.Get("X-MCP-Identity")
			if userID != "" {
				r.Header.Set("X-User-ID", userID)
				next.ServeHTTP(w, r)
				return
			}

			// 2. Check for Bearer Token
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				// Validate token by calling the IDP /userinfo endpoint
				idpURL := config.GetIdpUrl()
				if idpURL != "" {
					userInfoURL := strings.TrimRight(idpURL, "/") + "/userinfo"
					reqInfo, err := http.NewRequestWithContext(r.Context(), "GET", userInfoURL, nil)
					if err == nil {
						reqInfo.Header.Set("Authorization", authHeader)
						resp, err := client.Do(reqInfo)
						if err == nil {
							defer resp.Body.Close()
							if resp.StatusCode == http.StatusOK {
								var userInfo struct {
									Sub   string `json:"sub"`
									Email string `json:"email"`
								}
								if err := json.NewDecoder(resp.Body).Decode(&userInfo); err == nil {
									uid := userInfo.Sub
									if uid == "" {
										uid = userInfo.Email
									}
									if uid != "" {
										r.Header.Set("X-User-ID", uid)
										next.ServeHTTP(w, r)
										return
									}
								}
							}
						}
					}
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			resp := map[string]string{
				"error":     "Authentication required",
				"login_url": config.GetIdpUrl() + "/login",
			}
			json.NewEncoder(w).Encode(resp)
		})
	}
}
