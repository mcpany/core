// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"google.golang.org/protobuf/encoding/protojson"
)

const userLibraryServiceName = "user-library"

func (a *Application) handleCreatePrompt(w http.ResponseWriter, r *http.Request, store storage.Storage) {
	var prompt configv1.PromptDefinition
	body, err := readBodyWithLimit(w, r, 1048576)
	if err != nil {
		return
	}
	if err := protojson.Unmarshal(body, &prompt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if prompt.GetName() == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	// Load user-library service
	svc, err := store.GetService(r.Context(), userLibraryServiceName)
	if err != nil {
		logging.GetLogger().Error("failed to get user-library service", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if svc == nil {
		// Should not happen if initialized correctly, but handle gracefully
		http.Error(w, "User library service not found. Please restart server to initialize.", http.StatusInternalServerError)
		return
	}

	// Check for duplicates
	for _, p := range svc.GetPrompts() {
		if p.GetName() == prompt.GetName() {
			http.Error(w, "prompt with this name already exists", http.StatusConflict)
			return
		}
	}

	// Add prompt
	newPrompts := append(svc.GetPrompts(), &prompt)
	svc.SetPrompts(newPrompts)

	// Save service
	if err := store.SaveService(r.Context(), svc); err != nil {
		logging.GetLogger().Error("failed to save user-library service", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Reload config to update PromptManager
	if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
		logging.GetLogger().Error("failed to reload config after prompt create", "error", err)
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte("{}"))
}

func (a *Application) handleUpdatePrompt(w http.ResponseWriter, r *http.Request, name string, store storage.Storage) {
	var prompt configv1.PromptDefinition
	body, err := readBodyWithLimit(w, r, 1048576)
	if err != nil {
		return
	}
	if err := protojson.Unmarshal(body, &prompt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Force name match
	prompt.SetName(name)

	// Load user-library service
	svc, err := store.GetService(r.Context(), userLibraryServiceName)
	if err != nil {
		logging.GetLogger().Error("failed to get user-library service", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if svc == nil {
		http.Error(w, "User library service not found", http.StatusInternalServerError)
		return
	}

	found := false
	prompts := svc.GetPrompts()
	for i, p := range prompts {
		if p.GetName() == name {
			prompts[i] = &prompt
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Prompt not found in user library (cannot edit system prompts)", http.StatusNotFound)
		return
	}

	// Ensure modified slice is set back
	svc.SetPrompts(prompts)

	if err := store.SaveService(r.Context(), svc); err != nil {
		logging.GetLogger().Error("failed to save user-library service", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
		logging.GetLogger().Error("failed to reload config after prompt update", "error", err)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{}"))
}

func (a *Application) handleDeletePrompt(w http.ResponseWriter, r *http.Request, name string, store storage.Storage) {
	// Load user-library service
	svc, err := store.GetService(r.Context(), userLibraryServiceName)
	if err != nil {
		logging.GetLogger().Error("failed to get user-library service", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if svc == nil {
		http.Error(w, "User library service not found", http.StatusInternalServerError)
		return
	}

	found := false
	currentPrompts := svc.GetPrompts()
	newPrompts := make([]*configv1.PromptDefinition, 0, len(currentPrompts))
	for _, p := range currentPrompts {
		if p.GetName() == name {
			found = true
			continue
		}
		newPrompts = append(newPrompts, p)
	}

	if !found {
		http.Error(w, "Prompt not found in user library", http.StatusNotFound)
		return
	}

	svc.SetPrompts(newPrompts)

	if err := store.SaveService(r.Context(), svc); err != nil {
		logging.GetLogger().Error("failed to save user-library service", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
		logging.GetLogger().Error("failed to reload config after prompt delete", "error", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *Application) handlePromptExecute(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/prompts/")
		parts := strings.Split(path, "/")
		if len(parts) < 2 || parts[1] != "execute" {
			http.NotFound(w, r)
			return
		}
		name := parts[0]

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read args from body
		var req struct {
			Arguments map[string]string `json:"arguments"`
		}
		body, err := readBodyWithLimit(w, r, 1048576)
		if err != nil {
			return
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		prompt, ok := a.PromptManager.GetPrompt(name)
		if !ok {
			http.Error(w, "prompt not found", http.StatusNotFound)
			return
		}

		argsBytes, err := json.Marshal(req.Arguments)
		if err != nil {
			http.Error(w, "failed to marshal arguments", http.StatusInternalServerError)
			return
		}

		result, err := prompt.Get(r.Context(), argsBytes)
		if err != nil {
			logging.GetLogger().Error("failed to execute prompt", "name", name, "error", err)
			http.Error(w, fmt.Sprintf("failed to execute prompt: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}

func (a *Application) handlePromptsDispatch(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if it's /prompts or /prompts/NAME
		if r.URL.Path == "/prompts" || r.URL.Path == "/prompts/" {
			switch r.Method {
			case http.MethodGet:
				// List all prompts (from PromptManager)
				prompts := a.PromptManager.ListPrompts()
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(prompts)
			case http.MethodPost:
				// Create new prompt
				a.handleCreatePrompt(w, r, store)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// /prompts/NAME
		path := strings.TrimPrefix(r.URL.Path, "/prompts/")
		parts := strings.Split(path, "/")
		name := parts[0]

		if len(parts) > 1 {
			// Sub-resources like /prompts/NAME/execute
			if parts[1] == "execute" {
				a.handlePromptExecute(store)(w, r)
				return
			}
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodPut:
			a.handleUpdatePrompt(w, r, name, store)
		case http.MethodDelete:
			a.handleDeletePrompt(w, r, name, store)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
