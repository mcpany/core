// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Helper to write JSON response.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if pm, ok := data.(proto.Message); ok {
		// Use protojson for proto messages to ensure correct field mapping
		marshaler := protojson.MarshalOptions{
			UseProtoNames: true, // Respect snake_case vs camelCase if needed, but standard is UseProtoNames=false -> camelCase
			// If we want to match how standard JSON encoding would have behaved (roughly), we should check.
			// However, our UI likely expects camelCase.
			// protojson defaults to camelCase.
			EmitUnpopulated: false,
		}
		b, err := marshaler.Marshal(pm)
		if err != nil {
			fmt.Printf("Failed to encode proto response: %v\n", err)
			return
		}
		if _, err := w.Write(b); err != nil {
			fmt.Printf("Failed to write response: %v\n", err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Printf("Failed to encode response: %v\n", err)
	}
}

// Helper to write error response.
func writeError(w http.ResponseWriter, err error) {
	// Attempt to map error to status code
	status := http.StatusInternalServerError
	var msg string

	errStr := err.Error()

	// Simple error checking for now, assuming util has some error types or just generic
	switch {
	case strings.Contains(errStr, "not found"):
		status = http.StatusNotFound
		msg = errStr
	case strings.Contains(errStr, "required") || strings.Contains(errStr, "invalid"):
		status = http.StatusBadRequest
		msg = errStr
	default:
		// For 500 errors, we don't leak details
		// We should log it though (but we don't have logger here easily unless we pass it or use global)
		// Assuming global logging
		// logging.GetLogger().Error("API Error", "error", err) // Cannot import due to cycle if logging depends on app? No, logging is in pkg/logging
		// But let's just sanitize.
		msg = "Internal Server Error"
	}

	writeJSON(w, status, map[string]string{"error": msg})
}

// listCredentialsHandler returns all credentials.
func (a *Application) listCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}
	ctx := r.Context()
	creds, err := a.Storage.ListCredentials(ctx)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, creds)
}

// getCredentialHandler returns a credential by ID.
func (a *Application) getCredentialHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}
	// Extract ID from path. Since we use generic mux, we assume path is /credentials/<id>
	// We handle trailing slash in main.go, but here we expect /credentials/ID
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		writeError(w, fmt.Errorf("id is required"))
		return
	}
	id := pathParts[len(pathParts)-1]

	ctx := r.Context()
	cred, err := a.Storage.GetCredential(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}
	if cred == nil {
		writeError(w, fmt.Errorf("credential not found: %s", id))
		return
	}
	writeJSON(w, http.StatusOK, cred)
}

// createCredentialHandler creates a new credential.
func (a *Application) createCredentialHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}
	ctx := r.Context()
	var cred configv1.Credential

	body, err := readBodyWithLimit(w, r, 1048576)
	if err != nil {
		return
	}

	unmarshaler := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := unmarshaler.Unmarshal(body, &cred); err != nil {
		writeError(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if cred.GetId() == "" {
		if cred.GetName() != "" {
			slug, _ := util.SanitizeID([]string{cred.GetName()}, false, 50, 4)
			cred.Id = proto.String(slug)
		} else {
			writeError(w, fmt.Errorf("id or name is required"))
			return
		}
	}

	if err := a.Storage.SaveCredential(ctx, &cred); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, &cred)
}

// updateCredentialHandler updates an existing credential.
func (a *Application) updateCredentialHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		writeError(w, fmt.Errorf("id is required"))
		return
	}
	id := pathParts[len(pathParts)-1]

	ctx := r.Context()
	var cred configv1.Credential

	body, err := readBodyWithLimit(w, r, 1048576)
	if err != nil {
		return
	}

	unmarshaler := protojson.UnmarshalOptions{DiscardUnknown: true}
	if err := unmarshaler.Unmarshal(body, &cred); err != nil {
		writeError(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Ensure ID matches path
	if cred.GetId() != "" && cred.GetId() != id {
		writeError(w, fmt.Errorf("id mismatch"))
		return
	}
	cred.Id = proto.String(id)

	if err := a.Storage.SaveCredential(ctx, &cred); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, &cred)
}

// deleteCredentialHandler deletes a credential.
func (a *Application) deleteCredentialHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		writeError(w, fmt.Errorf("id is required"))
		return
	}
	id := pathParts[len(pathParts)-1]

	ctx := r.Context()
	if err := a.Storage.DeleteCredential(ctx, id); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TestAuthRequest defines the payload for testing authentication.
type TestAuthRequest struct {
	// The credential to use (can be a reference ID or inline Credential).
	CredentialID string `json:"credential_id"`
	// OR inline authentication config
	Authentication *configv1.Authentication `json:"authentication"`
	// OR inline user token (for ad-hoc testing)
	UserToken *configv1.UserToken `json:"user_token"`

	// The URL to test against.
	TargetURL string `json:"target_url"`
	// HTTP Method (GET, POST, etc.)
	Method string `json:"method"`
}

// TestAuthResponse defines the response for testing authentication.
type TestAuthResponse struct {
	Status     int               `json:"status"`
	StatusText string            `json:"status_text"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Error      string            `json:"error,omitempty"`
}

// testAuthHandler tests authentication against a target URL.
func (a *Application) testAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}
	ctx := r.Context()
	var req TestAuthRequest

	body, err := readBodyWithLimit(w, r, 1048576)
	if err != nil {
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if req.TargetURL == "" {
		writeError(w, fmt.Errorf("target_url is required"))
		return
	}

	// Resolve Credentials
	var authConfig *configv1.Authentication
	var userToken *configv1.UserToken

	if req.CredentialID != "" {
		cred, err := a.Storage.GetCredential(ctx, req.CredentialID)
		if err != nil {
			writeError(w, err)
			return
		}
		if cred == nil {
			writeError(w, fmt.Errorf("credential not found: %s", req.CredentialID))
			return
		}
		authConfig = cred.Authentication
		userToken = cred.Token // Note: Field is 'Token' in proto
	} else {
		// Use inline
		authConfig = req.Authentication
		userToken = req.UserToken
	}

	prepareAndExecuteRequest(ctx, w, req, authConfig, userToken)
}

func prepareAndExecuteRequest(ctx context.Context, w http.ResponseWriter, req TestAuthRequest, authConfig *configv1.Authentication, userToken *configv1.UserToken) {
	// Prepare Request
	method := req.Method
	if method == "" {
		method = "GET"
	}

	// Use a clean http client
	client := &http.Client{
		// Timeout: util.DefaultTimeout, // Commented out used as per lint error
		Timeout: 30 * time.Second,
	}

	// Create Request
	httpReq, err := http.NewRequestWithContext(ctx, method, req.TargetURL, nil)
	if err != nil {
		writeError(w, fmt.Errorf("invalid target url: %v", err))
		return
	}

	// Apply Authentication
	applied := false
	if userToken != nil {
		// Prioritize UserToken (e.g. 3-legged OAuth or manually supplied token)
		if userToken.GetAccessToken() != "" {
			httpReq.Header.Set("Authorization", "Bearer "+userToken.GetAccessToken())
			applied = true
		}
	}

	if !applied && authConfig != nil {
		authenticator, err := auth.NewUpstreamAuthenticator(authConfig)
		if err != nil {
			writeError(w, fmt.Errorf("invalid auth config: %v", err))
			return
		}
		if authenticator != nil {
			if err := authenticator.Authenticate(httpReq); err != nil {
				writeError(w, fmt.Errorf("failed to apply authentication: %w", err))
				return
			}
			// applied = true // unused
		}
	}

	// Execute Request
	resp, err := client.Do(httpReq)
	if err != nil {
		writeJSON(w, http.StatusOK, TestAuthResponse{
			Error: fmt.Sprintf("Request failed: %v", err),
		})
		return
	}
	defer func() { _ = resp.Body.Close() }()

	// Read Body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		writeJSON(w, http.StatusOK, TestAuthResponse{
			Status:     resp.StatusCode,
			StatusText: resp.Status,
			Error:      fmt.Sprintf("Failed to read body: %v", err),
		})
		return
	}

	// Format Headers
	headers := make(map[string]string)
	for k, v := range resp.Header {
		headers[k] = strings.Join(v, ", ")
	}

	writeJSON(w, http.StatusOK, TestAuthResponse{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    headers,
		Body:       string(bodyBytes), // Limit size?
	})
}
