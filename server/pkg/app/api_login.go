// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mcpany/core/server/pkg/util/passhash"
)

// LoginRequest is the request body for login.
//
// Summary: is the request body for login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is the response body for login.
//
// Summary: is the response body for login.
type LoginResponse struct {
	Token string `json:"token"`
}

// handleLogin handles the login request.
// POST /api/v1/auth/login.
func (a *Application) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	// Find user by username (using ListUsers for now, or GetUser if ID==Username)
	// Since we set ID=Username for admin, we can try GetUser.
	// But to be generic, we might want to iterate users if ID matches.
	// We'll rely on our convention that ID is the username for simple auth managed users.

	ctx := r.Context()

	// This relies on the AuthManager having the users loaded, OR we query the storage.
	// The AuthManager loads users from config on startup.
	// If existing users were in DB, they are in AuthManager.users.

	// Check AuthManager first?
	// AuthManager stores users in memory map.

	// We don't expose GetUser from AuthManager directly, but we can try to "Authenticate"
	// using a mock credential? No, AuthManager.ValidateAuthentication takes a context/request.

	// Let's look up the user directly from AuthManager if possible, or Storage if better.
	// Accessing Storage is safer if we want to ensure persistence and freshness,
	// but AuthManager is what enforces auth really.
	// However, AuthManager.SetUsers loads from config.

	// Let's try to find the user in `a.AuthManager` if we can access it, or just use Storage.
	// Using Storage directly:

	user, err := a.Storage.GetUser(ctx, req.Username)
	if err != nil {
		// User not found or error
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if user == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check password
	authConfig := user.GetAuthentication()
	if authConfig == nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	basicAuth := authConfig.GetBasicAuth()
	if basicAuth == nil {
		// Only Basic Auth supported for this endpoint for now
		http.Error(w, "login not supported for this user type", http.StatusUnauthorized)
		return
	}

	// Check Hash
	if !passhash.CheckPassword(req.Password, basicAuth.GetPasswordHash()) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate Token
	// Basic Auth Token: base64(username:password)
	// NOTE: We are returning the raw password here in base64.
	// This is just standard Basic Auth, effectively.
	// The client will use this as `Authorization: Basic <token>`.
	// Ideally we would issue a session token, but the requirement is to use existing Basic Auth support.
	token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", req.Username, req.Password)))

	resp := LoginResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
