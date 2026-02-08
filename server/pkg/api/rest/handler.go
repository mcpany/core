// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package rest provides REST API handlers for the server.
package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"gopkg.in/yaml.v3"
)

// ValidateConfigHandler handles requests to validate configuration.
//
// Summary: Validates the provided configuration content against the JSON schema and internal logic.
//
// Parameters:
//   - w: http.ResponseWriter. The response writer to send the validation result.
//   - r: *http.Request. The HTTP request containing the configuration content in the body.
//
// Returns:
//   - JSON response with validation status and errors.
//
// Errors/Throws:
//   - Returns 405 Method Not Allowed if the request method is not POST.
//   - Returns 400 Bad Request if the request body is invalid or empty.
//   - Returns 500 Internal Server Error if there is an error encoding the response.
//
// Side Effects:
//   - Reads the request body.
//   - Writes to the response writer.
func ValidateConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Limit request body size to 5MB to prevent DoS
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	var req ValidateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondWithJSONError(w, http.StatusBadRequest, "Content is required")
		return
	}

	// 1. Unmarshal into generic map to validate against JSON schema
	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal([]byte(req.Content), &rawConfig); err != nil {
		respondWithValidationErrors(w, []string{fmt.Sprintf("Failed to parse YAML/JSON: %v", err)})
		return
	}

	// 2. Validate against JSON Schema
	var errors []string
	if err := config.ValidateConfigAgainstSchema(rawConfig); err != nil {
		// Basic error formatting. Ideally, schema validation returns a list of errors.
		// For now, we wrap the single error.
		errors = append(errors, err.Error())
	}

	// 3. Additional Semantic Validation using configv1.McpAnyServerConfig
	// This captures things that JSON schema might miss (custom Go validation logic like file existence)
	engine, err := config.NewEngine("config.yaml")
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to initialize config engine: %v", err))
	} else {
		// Skip schema validation in the engine since we already performed it above
		if configurable, ok := engine.(config.ConfigurableEngine); ok {
			configurable.SetSkipValidation(true)
		}

		cfg := configv1.McpAnyServerConfig_builder{}.Build()

		// ⚡ BOLT: Avoid double YAML parsing if engine supports map input.
		// Randomized Selection from Top 5 High-Impact Targets.
		var unmarshalErr error
		if structured, ok := engine.(config.StructuredEngine); ok {
			unmarshalErr = structured.UnmarshalFromMap(rawConfig, cfg, []byte(req.Content))
		} else {
			unmarshalErr = engine.Unmarshal([]byte(req.Content), cfg)
		}

		if unmarshalErr != nil {
			errors = append(errors, fmt.Sprintf("Failed to unmarshal config: %v", unmarshalErr))
		} else {
			// Run semantic validation (checks file existence, connectivity, etc.)
			// We skip secret validation (regex checks) to prevent oracle attacks where users probe secret values.
			ctx := context.WithValue(r.Context(), config.SkipSecretValidationKey, true)
			// We also skip filesystem existence checks to prevent filesystem enumeration attacks.
			ctx = context.WithValue(ctx, config.SkipFilesystemCheckKey, true)
			validationErrors := config.Validate(ctx, cfg, config.Server)
			for _, ve := range validationErrors {
				errors = append(errors, ve.Error())
			}
		}
	}

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

func respondWithJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ValidateConfigResponse{
		Valid:  false,
		Errors: []string{message},
	})
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
