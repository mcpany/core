// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package rest provides REST API handlers for the server.
package rest

import (
	"encoding/json"
	"net/http"

	"github.com/mcpany/core/server/pkg/config"
	"gopkg.in/yaml.v3"
)

// ValidateConfigHandler handles requests to validate configuration.
func ValidateConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit request body size to 5MB to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	var req ValidateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// 1. Unmarshal into generic map to validate against JSON schema
	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal([]byte(req.Content), &rawConfig); err != nil {
		// Do not return the specific error to avoid leaking internal details or structure
		respondWithValidationErrors(w, []string{"Invalid YAML/JSON format"})
		return
	}

	// 2. Validate against JSON Schema
	var errors []string
	if err := config.ValidateConfigAgainstSchema(rawConfig); err != nil {
		// Basic error formatting. Ideally, schema validation returns a list of errors.
		// For now, we wrap the single error.
		errors = append(errors, err.Error())
	}

	// 3. (Optional) Additional Semantic Validation using configv1.McpAnyServerConfig
	// This captures things that JSON schema might miss (custom Go validation logic)
	// For now, we rely primarily on Schema validation as requested by the task.

	if len(errors) > 0 {
		respondWithValidationErrors(w, errors)
		return
	}

	// Success
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ValidateConfigResponse{
		Valid: true,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func respondWithValidationErrors(w http.ResponseWriter, errors []string) {
	w.Header().Set("Content-Type", "application/json")
	// We return 200 OK because the *request* was successful, the *validation* result is false.
	// Returning 400 might imply the API usage was wrong.
	if err := json.NewEncoder(w).Encode(ValidateConfigResponse{
		Valid:  false,
		Errors: errors,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
