// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"sigs.k8s.io/yaml"
)

// ValidateRequest represents the request body for the validation endpoint.
type ValidateRequest struct {
	Content string `json:"content"`
	Format  string `json:"format"` // "json" or "yaml"
}

// ValidateResponse represents the response body for the validation endpoint.
type ValidateResponse struct {
	Valid   bool   `json:"valid"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// handleValidate returns a handler for validating configuration snippets.
func (a *Application) handleValidate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ValidateRequest
		body, err := readBodyWithLimit(w, r, 1048576) // 1MB limit
		if err != nil {
			return
		}

		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		if req.Content == "" {
			http.Error(w, "content is required", http.StatusBadRequest)
			return
		}

		// 1. Unmarshal to JSON map for schema validation (catches type errors and unknown fields)
		var rawMap map[string]interface{}
		var jsonBytes []byte

		// Try to detect format if not provided
		format := req.Format
		if format == "" {
			if json.Valid([]byte(req.Content)) {
				format = "json"
			} else {
				format = "yaml"
			}
		}

		if format == "json" {
			jsonBytes = []byte(req.Content)
		} else {
			var err error
			jsonBytes, err = yaml.YAMLToJSON([]byte(req.Content))
			if err != nil {
				a.writeValidateResponse(w, false, fmt.Sprintf("Failed to parse YAML: %v", err))
				return
			}
		}

		if err := json.Unmarshal(jsonBytes, &rawMap); err != nil {
			a.writeValidateResponse(w, false, fmt.Sprintf("Failed to parse %s: %v", format, err))
			return
		}

		// 2. Schema Validation (Good for basic structure)
		if err := config.ValidateConfigAgainstSchema(rawMap); err != nil {
			a.writeValidateResponse(w, false, err.Error())
			return
		}

		// 3. Proto Validation (Required fields, custom logic)
		cfg := configv1.McpAnyServerConfig_builder{}.Build()
		engine, _ := config.NewEngine("config." + format) // Dummy path for engine selection
		if err := engine.Unmarshal([]byte(req.Content), cfg); err != nil {
			// This might catch things the schema didn't, or provide better messages
			a.writeValidateResponse(w, false, "Refined validation failed: "+err.Error())
			return
		}

		// Run deep validation
		// Security: Skip filesystem and secret checks to prevent Oracle attacks (Information Leakage).
		ctx := context.WithValue(r.Context(), config.SkipSecretValidationKey, true)
		ctx = context.WithValue(ctx, config.SkipFilesystemCheckKey, true)

		if errs := config.Validate(ctx, cfg, config.Server); len(errs) > 0 {
			var errMsg string
			for i, e := range errs {
				if i > 0 {
					errMsg += "; "
				}
				errMsg += e.Error()
			}
			a.writeValidateResponse(w, false, "Deep validation failed: "+errMsg)
			return
		}

		a.writeValidateResponse(w, true, "Configuration is valid")
	}
}

func (a *Application) writeValidateResponse(w http.ResponseWriter, valid bool, msg string) {
	resp := ValidateResponse{
		Valid:   valid,
		Message: msg,
	}
	if !valid {
		resp.Error = msg
	}

	w.Header().Set("Content-Type", "application/json")
	if !valid {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_ = json.NewEncoder(w).Encode(resp)
}
