// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
)

// handleResourceRead handles reading a specific resource.
//
// Summary: Reads a specific resource.
//
// Parameters:
//   - None.
//
// Returns:
//   - http.HandlerFunc: The handler.
//
// Errors:
//   - Writes HTTP 405 if the method is not GET.
//   - Writes HTTP 400 if the URI parameter is missing.
//   - Writes HTTP 500 if the resource read fails internally.
//
// Side Effects:
//   - Reads from the ResourceManager.
func (a *Application) handleResourceRead() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		uri := r.URL.Query().Get("uri")
		if uri == "" {
			http.Error(w, "uri required", http.StatusBadRequest)
			return
		}

		res, ok := a.ResourceManager.GetResource(uri)
		if !ok {
			http.NotFound(w, r)
			return
		}

		result, err := res.Read(r.Context())
		if err != nil {
			logging.GetLogger().Error("failed to read resource", "uri", uri, "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}

// handlePromptExecute handles requests to execute a specific prompt.
//
// Summary: Executes a specific prompt.
//
// Parameters:
//   - None.
//
// Returns:
//   - http.HandlerFunc: The configured handler for prompt execution.
//
// Errors:
//   - Writes HTTP 405 if the method is not POST.
//   - Writes HTTP 404 if the prompt is not found or path is invalid.
//   - Writes HTTP 400 for bad JSON requests.
//   - Writes HTTP 500 if the execution fails.
//
// Side Effects:
//   - Executes the underlying prompt logic, which might log errors or interact with upstream services.
func (a *Application) handlePromptExecute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Path: /prompts/{name}/execute
		path := strings.TrimPrefix(r.URL.Path, "/prompts/")
		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}
		name := parts[0]
		action := parts[1]

		if action != "execute" {
			http.NotFound(w, r)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read body as RawMessage
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}

		prompt, ok := a.PromptManager.GetPrompt(name)
		if !ok {
			http.NotFound(w, r)
			return
		}

		result, err := prompt.Get(r.Context(), json.RawMessage(body))
		if err != nil {
			logging.GetLogger().Error("failed to execute prompt", "name", name, "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}
