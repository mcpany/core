// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestProfileRBACMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		profiles      []*configv1.ProfileDefinition
		userRoles     []string
		expectedCode  int
		expectedBody  string
	}{
		{
			name:          "No profiles, allow all",
			profiles:      []*configv1.ProfileDefinition{},
			userRoles:     []string{},
			expectedCode:  http.StatusOK,
		},
		{
			name: "Profile with no requirements (Public), allow all",
			profiles: []*configv1.ProfileDefinition{
				configv1.ProfileDefinition_builder{
					Name:          proto.String("public"),
					RequiredRoles: []string{},
				}.Build(),
			},
			userRoles:    []string{},
			expectedCode: http.StatusOK,
		},
		{
			name: "Mixed Public and Private, allow public user",
			profiles: []*configv1.ProfileDefinition{
				configv1.ProfileDefinition_builder{
					Name:          proto.String("public"),
					RequiredRoles: []string{},
				}.Build(),
				configv1.ProfileDefinition_builder{
					Name:          proto.String("admin-only"),
					RequiredRoles: []string{"admin"},
				}.Build(),
			},
			userRoles:    []string{}, // No roles
			expectedCode: http.StatusOK,
		},
		{
			name: "Requirement exists, user has matching role",
			profiles: []*configv1.ProfileDefinition{
				configv1.ProfileDefinition_builder{
					Name:          proto.String("admin-only"),
					RequiredRoles: []string{"admin"},
				}.Build(),
			},
			userRoles:    []string{"admin"},
			expectedCode: http.StatusOK,
		},
		{
			name: "Requirement exists, user has non-matching role",
			profiles: []*configv1.ProfileDefinition{
				configv1.ProfileDefinition_builder{
					Name:          proto.String("admin-only"),
					RequiredRoles: []string{"admin"},
				}.Build(),
			},
			userRoles:    []string{"viewer"},
			expectedCode: http.StatusForbidden,
		},
		{
			name: "Multiple requirements, user matches one",
			profiles: []*configv1.ProfileDefinition{
				configv1.ProfileDefinition_builder{
					Name:          proto.String("mixed"),
					RequiredRoles: []string{"admin", "editor"},
				}.Build(),
			},
			userRoles:    []string{"editor"},
			expectedCode: http.StatusOK,
		},
		{
			name: "Requirement exists, user has no roles",
			profiles: []*configv1.ProfileDefinition{
				configv1.ProfileDefinition_builder{
					Name:          proto.String("admin-only"),
					RequiredRoles: []string{"admin"},
				}.Build(),
			},
			userRoles:    []string{},
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := ProfileRBACMiddleware(tt.profiles)
			wrapped := middleware(handler)

			req := httptest.NewRequest("GET", "/", nil)
			ctx := req.Context()
			if len(tt.userRoles) > 0 {
				ctx = auth.ContextWithRoles(ctx, tt.userRoles)
			}
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			wrapped.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
		})
	}
}
