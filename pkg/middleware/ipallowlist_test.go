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
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIPAllowlist(t *testing.T) {
	testCases := []struct {
		name           string
		allowedIPs     []string
		remoteAddr     string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Allowed IPv4",
			allowedIPs:     []string{"192.168.1.1"},
			remoteAddr:     "192.168.1.1:12345",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Disallowed IPv4",
			allowedIPs:     []string{"192.168.1.1"},
			remoteAddr:     "192.168.1.2:12345",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Allowed IPv4 CIDR",
			allowedIPs:     []string{"192.168.1.0/24"},
			remoteAddr:     "192.168.1.100:12345",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Disallowed IPv4 CIDR",
			allowedIPs:     []string{"192.168.1.0/24"},
			remoteAddr:     "192.168.2.1:12345",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Allowed IPv6",
			allowedIPs:     []string{"::1"},
			remoteAddr:     "[::1]:12345",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Disallowed IPv6",
			allowedIPs:     []string{"::1"},
			remoteAddr:     "[::2]:12345",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Allowed IPv6 CIDR",
			allowedIPs:     []string{"2001:db8::/32"},
			remoteAddr:     "[2001:db8::1]:12345",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Disallowed IPv6 CIDR",
			allowedIPs:     []string{"2001:db8::/32"},
			remoteAddr:     "[2001:db9::1]:12345",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "No Allowlist",
			allowedIPs:     []string{},
			remoteAddr:     "192.168.1.1:12345",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Malformed RemoteAddr",
			allowedIPs:     []string{"192.168.1.1"},
			remoteAddr:     "malformed",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:        "Invalid IP in Allowlist",
			allowedIPs:  []string{"invalid"},
			remoteAddr:  "192.168.1.1:12345",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			allowlist, err := NewIPAllowlist(tc.allowedIPs)
			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected an error, but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Failed to create IP allowlist: %v", err)
			}

			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tc.remoteAddr
			rr := httptest.NewRecorder()

			handler := allowlist.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tc.expectedStatus)
			}
		})
	}
}
