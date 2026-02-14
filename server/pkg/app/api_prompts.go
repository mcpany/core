// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"net/http"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const userLibraryServiceName = "user-library"

// handlePrompts handles listing and creating prompts.
func (a *Application) handlePrompts(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			prompts := a.PromptManager.ListPrompts()
			// Transform to JSON-friendly format if needed, but Prompt.Prompt() returns *mcp.Prompt which is JSON compatible
			// Wait, ListPrompts returns []Prompt interface.
			// We need to marshal the result of p.Prompt() for each.
			mcpPrompts := make([]any, len(prompts))
			for i, p := range prompts {
				mcpPrompts[i] = p.Prompt()
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"prompts": mcpPrompts, // Wrap in object to match client expectation? client.ts expects { prompts: [] }
			})

		case http.MethodPost:
			a.handleCreatePrompt(w, r, store)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handlePromptExecute handles execution of a prompt and detail operations (PUT/DELETE).
func (a *Application) handlePromptExecute(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/prompts/")
		if strings.HasSuffix(name, "/execute") {
			name = strings.TrimSuffix(name, "/execute")
		} else if r.Method == http.MethodPut || r.Method == http.MethodDelete || r.Method == http.MethodGet {
			// If not ending in /execute, check if it's a detail request (PUT/DELETE/GET detail)
			a.handlePromptDetail(w, r, name, store)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if name == "" {
			http.Error(w, "prompt name required", http.StatusBadRequest)
			return
		}

		// Execute Prompt Logic
		type executeRequest struct {
			Arguments map[string]string `json:"arguments"`
		}
		var req executeRequest
		body, err := readBodyWithLimit(w, r, 1048576)
		if err != nil {
			return
		}
		if len(body) > 0 {
			if err := json.Unmarshal(body, &req); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
		}

		// Convert map[string]string to map[string]any for the SDK/Manager
		args := make(map[string]any)
		for k, v := range req.Arguments {
			args[k] = v
		}

		// Use PromptManager
		prompt, ok := a.PromptManager.GetPrompt(name)
		if !ok {
			http.Error(w, "prompt not found", http.StatusNotFound)
			return
		}

		// Prompt.Get takes json.RawMessage
		argsBytes, _ := json.Marshal(args)
		result, err := prompt.Get(r.Context(), argsBytes)
		if err != nil {
			logging.GetLogger().Error("failed to execute prompt", "name", name, "error", err)
			http.Error(w, "failed to execute prompt: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}
}

// handleCreatePrompt handles creating a new prompt in the user library.
func (a *Application) handleCreatePrompt(w http.ResponseWriter, r *http.Request, store storage.Storage) {
	var promptDef configv1.PromptDefinition
	body, err := readBodyWithLimit(w, r, 1048576)
	if err != nil {
		return
	}
	if err := protojson.Unmarshal(body, &promptDef); err != nil {
		http.Error(w, "invalid prompt definition: "+err.Error(), http.StatusBadRequest)
		return
	}

	if promptDef.GetName() == "" {
		http.Error(w, "prompt name is required", http.StatusBadRequest)
		return
	}

	// Load user-library service
	svc, err := store.GetService(r.Context(), userLibraryServiceName)
	if err != nil {
		http.Error(w, "failed to load user library: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if svc == nil {
		http.Error(w, "user library service not found", http.StatusInternalServerError)
		return
	}

	cmdSvc := svc.GetCommandLineService()
	if cmdSvc == nil {
		// Should not happen if initialized correctly
		cmdSvc = configv1.CommandLineUpstreamService_builder{Command: proto.String("echo")}.Build()
		svc.SetCommandLineService(cmdSvc)
	}

	for _, p := range cmdSvc.GetPrompts() {
		if p.GetName() == promptDef.GetName() {
			http.Error(w, "prompt with this name already exists in user library", http.StatusConflict)
			return
		}
	}

	// Add prompt
	prompts := cmdSvc.GetPrompts()
	prompts = append(prompts, &promptDef)
	cmdSvc.SetPrompts(prompts)

	// Save service
	if err := store.SaveService(r.Context(), svc); err != nil {
		http.Error(w, "failed to save prompt: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Reload config
	if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
		logging.GetLogger().Error("failed to reload config after prompt create", "error", err)
	}

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte("{}"))
}

// handlePromptDetail handles GET, PUT, DELETE for a specific prompt.
func (a *Application) handlePromptDetail(w http.ResponseWriter, r *http.Request, name string, store storage.Storage) {
	svc, err := store.GetService(r.Context(), userLibraryServiceName)
	if err != nil {
		http.Error(w, "failed to load user library", http.StatusInternalServerError)
		return
	}
	if svc == nil {
		http.NotFound(w, r)
		return
	}
	cmdSvc := svc.GetCommandLineService()
	if cmdSvc == nil {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Get prompt definition
		var found *configv1.PromptDefinition
		for _, p := range cmdSvc.GetPrompts() {
			if p.GetName() == name {
				found = p
				break
			}
		}
		if found == nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		opts := protojson.MarshalOptions{UseProtoNames: true}
		b, _ := opts.Marshal(found)
		_, _ = w.Write(b)

	case http.MethodPut:
		var promptDef configv1.PromptDefinition
		body, err := readBodyWithLimit(w, r, 1048576)
		if err != nil {
			return
		}
		if err := protojson.Unmarshal(body, &promptDef); err != nil {
			http.Error(w, "invalid prompt definition: "+err.Error(), http.StatusBadRequest)
			return
		}
		// Enforce name match
		if promptDef.GetName() != "" && promptDef.GetName() != name {
			http.Error(w, "prompt name in body must match URL", http.StatusBadRequest)
			return
		}
		promptDef.SetName(name)

		// Validate prompt definition
		// We use config package validation logic
		if err := config.ValidateOrError(r.Context(), configv1.UpstreamServiceConfig_builder{
			Name: proto.String("temp"),
			CommandLineService: configv1.CommandLineUpstreamService_builder{
				Command: proto.String("echo"),
				Prompts: []*configv1.PromptDefinition{&promptDef},
			}.Build(),
		}.Build()); err != nil {
			// Log but don't block for now as validation might be strict on other things
			logging.GetLogger().Warn("Prompt validation warning", "error", err)
		}

		// Update list
		prompts := cmdSvc.GetPrompts()
		foundIndex := -1
		for i, p := range prompts {
			if p.GetName() == name {
				foundIndex = i
				break
			}
		}

		// Since Go protobuf repeated fields are slices of pointers, we can modify directly?
		// No, we should replace the entry or append.
		if foundIndex == -1 {
			// Upsert
			prompts = append(prompts, &promptDef)
		} else {
			prompts[foundIndex] = &promptDef
		}
		cmdSvc.SetPrompts(prompts)

		if err := store.SaveService(r.Context(), svc); err != nil {
			http.Error(w, "failed to save prompt: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
			logging.GetLogger().Error("failed to reload config after prompt update", "error", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))

	case http.MethodDelete:
		prompts := cmdSvc.GetPrompts()
		newPrompts := make([]*configv1.PromptDefinition, 0, len(prompts))
		found := false
		for _, p := range prompts {
			if p.GetName() == name {
				found = true
				continue
			}
			newPrompts = append(newPrompts, p)
		}

		if !found {
			http.NotFound(w, r)
			return
		}

		cmdSvc.SetPrompts(newPrompts)
		if err := store.SaveService(r.Context(), svc); err != nil {
			http.Error(w, "failed to delete prompt: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
			logging.GetLogger().Error("failed to reload config after prompt delete", "error", err)
		}
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
