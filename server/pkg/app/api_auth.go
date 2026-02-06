package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mcpany/core/server/pkg/auth"
)

// handleInitiateOAuth handles the request to initiate an OAuth2 flow.
// POST /auth/oauth/initiate
// Body: {"service_id": "github", "redirect_url": "..."}.
func (a *Application) handleInitiateOAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ServiceID    string `json:"service_id"`
		CredentialID string `json:"credential_id"`
		RedirectURL  string `json:"redirect_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if (req.ServiceID == "" && req.CredentialID == "") || req.RedirectURL == "" {
		http.Error(w, "service_id (or credential_id) and redirect_url are required", http.StatusBadRequest)
		return
	}

	userID, ok := auth.UserFromContext(r.Context())
	if !ok {
		// Ideally this endpoint is protected by auth middleware.
		// If not, we might need to assume a default user or error.
		// For now error.
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	url, state, err := a.AuthManager.InitiateOAuth(r.Context(), userID, req.ServiceID, req.CredentialID, req.RedirectURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to initiate oauth: %v", err), http.StatusInternalServerError)
		return
	}

	resp := struct {
		AuthorizationURL string `json:"authorization_url"`
		State            string `json:"state"`
	}{
		AuthorizationURL: url,
		State:            state,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// handleOAuthCallback handles the OAuth2 callback.
// POST /auth/oauth/callback
// Body: {"service_id": "github", "code": "...", "redirect_url": "..."}
// Note: Usually callbacks are GET requests to the frontend, which then POST code to backend.
func (a *Application) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ServiceID    string `json:"service_id"`
		CredentialID string `json:"credential_id"`
		Code         string `json:"code"`
		RedirectURL  string `json:"redirect_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if (req.ServiceID == "" && req.CredentialID == "") || req.Code == "" || req.RedirectURL == "" {
		http.Error(w, "service_id (or credential_id), code, and redirect_url are required", http.StatusBadRequest)
		return
	}

	userID, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := a.AuthManager.HandleOAuthCallback(r.Context(), userID, req.ServiceID, req.CredentialID, req.Code, req.RedirectURL); err != nil {
		http.Error(w, fmt.Sprintf("failed to handle callback: %v", err), http.StatusInternalServerError)
		return
	}

	// Helper to send success response
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "success"}`))
}
