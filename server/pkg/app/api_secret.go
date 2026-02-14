// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
)

// listSecretsHandler returns all secrets.
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
	// Sanitize secrets (mask values)
	for i := range secrets {
		secrets[i] = sanitizeSecret(secrets[i])
	}

	// Use SecretList proto message to ensure correct JSON serialization of fields
	// Using Builder for Opaque API
	resp := configv1.SecretList_builder{
		Secrets: secrets,
	}.Build()

	writeJSON(w, http.StatusOK, resp)
}

// createSecretHandler creates a new secret.
func (a *Application) createSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}
	ctx := r.Context()
	var secret configv1.Secret

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, fmt.Errorf("failed to read body: %w", err))
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	// Since Secret is a proto message, we use protojson
	if err := protojson.Unmarshal(body, &secret); err != nil {
		writeError(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Validation
	if secret.GetName() == "" {
		writeError(w, fmt.Errorf("name is required"))
		return
	}
	if secret.GetKey() == "" {
		writeError(w, fmt.Errorf("key is required"))
		return
	}
	if secret.GetValue() == "" {
		writeError(w, fmt.Errorf("value is required"))
		return
	}

	if secret.GetId() == "" {
		// Generate ID from name
		slug, _ := util.SanitizeID([]string{secret.GetName()}, false, 50, 4)
		secret.SetId(slug)
	}

	secret.SetCreatedAt(time.Now().Format(time.RFC3339))

	if err := a.Storage.SaveSecret(ctx, &secret); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, sanitizeSecret(&secret))
}

// getSecretHandler returns a secret by ID.
func (a *Application) getSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 { // Expect /secrets/:id
		writeError(w, fmt.Errorf("id is required"))
		return
	}
	id := pathParts[len(pathParts)-1]

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
	writeJSON(w, http.StatusOK, sanitizeSecret(secret))
}

// deleteSecretHandler deletes a secret.
func (a *Application) deleteSecretHandler(w http.ResponseWriter, r *http.Request) {
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
	if err := a.Storage.DeleteSecret(ctx, id); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// revealSecretHandler returns the unmasked secret value.
func (a *Application) revealSecretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, fmt.Errorf("method not allowed"))
		return
	}
	// Path: /api/v1/secrets/:id/reveal
	// parts: [api, v1, secrets, id, reveal]
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		writeError(w, fmt.Errorf("id is required"))
		return
	}
	// "reveal" is last, ID is second to last
	id := pathParts[len(pathParts)-2]

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

	// Return just the value in a JSON object, as expected by client.ts
	// revealSecret: async (id: string): Promise<{ value: string }>
	writeJSON(w, http.StatusOK, map[string]string{"value": secret.GetValue()})
}

// sanitizeSecret masks the sensitive value of a secret.
func sanitizeSecret(s *configv1.Secret) *configv1.Secret {
	if s == nil {
		return nil
	}
	// We modify in place assuming ownership of the object (returned from storage unmarshaler)
	s.SetValue("••••••••••••••••••••••••")
	return s
}
