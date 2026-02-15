// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
)

// listSecretsHandler returns all secrets (masked).
func (a *Application) listSecretsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}
	ctx := r.Context()
	secrets, err := a.Storage.ListSecrets(ctx)
	if err != nil {
		writeError(w, err)
		return
	}
	// Sanitize secrets
	for _, s := range secrets {
		sanitizeSecret(s)
	}
	resp := configv1.SecretList_builder{
		Secrets: secrets,
	}.Build()
	writeJSON(w, http.StatusOK, resp)
}

// getSecretHandler returns a secret by ID (masked).
func (a *Application) getSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}
	// Extract ID from path. path is /api/v1/secrets/:id
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// /api/v1/secrets/:id -> ["api", "v1", "secrets", "id"]
	if len(pathParts) < 4 {
		writeError(w, fmt.Errorf("id is required"))
		return
	}
	id := pathParts[3]

	ctx := r.Context()
	secret, err := a.Storage.GetSecret(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}
	if secret == nil {
		writeError(w, fmt.Errorf("secret not found: %s", id))
		return
	}
	sanitizeSecret(secret)
	writeJSON(w, http.StatusOK, secret)
}

// createSecretHandler creates or updates a secret.
func (a *Application) createSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}

	// Authorization: Only admins can manage secrets
	if !auth.NewRBACEnforcer().HasRoleInContext(r.Context(), "admin") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	ctx := r.Context()
	var secret configv1.Secret

	body, err := readBodyWithLimit(w, r, 1048576)
	if err != nil {
		return
	}

	if err := protojson.Unmarshal(body, &secret); err != nil {
		writeError(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if secret.GetName() == "" || secret.GetKey() == "" {
		writeError(w, fmt.Errorf("name and key are required"))
		return
	}

	if secret.GetId() == "" {
		slug, _ := util.SanitizeID([]string{secret.GetName()}, false, 50, 4)
		secret.SetId(slug)
	}

	if secret.GetCreatedAt() == "" {
		secret.SetCreatedAt(time.Now().Format(time.RFC3339))
	}

	if err := a.Storage.SaveSecret(ctx, &secret); err != nil {
		writeError(w, err)
		return
	}

	sanitizeSecret(&secret)
	writeJSON(w, http.StatusCreated, &secret)
}

// updateSecretHandler updates a secret.
func (a *Application) updateSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}

	if !auth.NewRBACEnforcer().HasRoleInContext(r.Context(), "admin") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		writeError(w, fmt.Errorf("id is required"))
		return
	}
	id := pathParts[3]

	ctx := r.Context()
	var secret configv1.Secret

	body, err := readBodyWithLimit(w, r, 1048576)
	if err != nil {
		return
	}

	if err := protojson.Unmarshal(body, &secret); err != nil {
		writeError(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Force ID match
	secret.SetId(id)
	if secret.GetName() == "" {
		secret.SetName(id)
	}

	if err := a.Storage.SaveSecret(ctx, &secret); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{})
}

// deleteSecretHandler deletes a secret.
func (a *Application) deleteSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}

	if !auth.NewRBACEnforcer().HasRoleInContext(r.Context(), "admin") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 4 {
		writeError(w, fmt.Errorf("id is required"))
		return
	}
	id := pathParts[3]

	ctx := r.Context()
	if err := a.Storage.DeleteSecret(ctx, id); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// revealSecretHandler reveals a secret value.
func (a *Application) revealSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}

	if !auth.NewRBACEnforcer().HasRoleInContext(r.Context(), "admin") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 5 || pathParts[4] != "reveal" {
		writeError(w, fmt.Errorf("invalid path"))
		return
	}
	id := pathParts[3]

	ctx := r.Context()
	secret, err := a.Storage.GetSecret(ctx, id)
	if err != nil {
		writeError(w, err)
		return
	}
	if secret == nil {
		writeError(w, fmt.Errorf("secret not found: %s", id))
		return
	}

	// Log access
	user, _ := auth.UserFromContext(r.Context())
	logging.GetLogger().Info("Secret revealed", "id", id, "user", user)

	// Update last used
	secret.SetLastUsed(time.Now().Format(time.RFC3339))
	if err := a.Storage.SaveSecret(ctx, secret); err != nil {
		fmt.Printf("Failed to update last used for secret %s: %v\n", id, err)
	}

	writeJSON(w, http.StatusOK, map[string]string{"value": secret.GetValue()})
}

func sanitizeSecret(s *configv1.Secret) {
	if s == nil {
		return
	}
	if s.GetValue() != "" {
		s.SetValue("********")
	}
}
