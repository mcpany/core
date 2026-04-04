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

// SSOMiddleware creates a new SSO middleware.
//
// Summary: Creates a new SSO middleware.
//
// Parameters:
//   - config (*configv1.SSOConfig): Parameter.
//
// Returns:
//   - func(http.Handler) http.Handler: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

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
