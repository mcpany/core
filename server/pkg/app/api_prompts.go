// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"net/http"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// handleCreatePrompt handles creating a new prompt in the user-library service.
func (a *Application) handleCreatePrompt(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		ctx := r.Context()
		svc, err := store.GetService(ctx, "user-library")
		if err != nil {
			logging.GetLogger().Error("failed to get user-library service", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if svc == nil {
			http.Error(w, "user-library service not found", http.StatusInternalServerError)
			return
		}

		// We modify the CommandLineService.Prompts slice
		cmdSvc := svc.GetCommandLineService()
		if cmdSvc == nil {
			// Init if missing
			cmdSvc = configv1.CommandLineUpstreamService_builder{
				Command: proto.String("true"),
			}.Build()
			svc.SetCommandLineService(cmdSvc)
		}

		// Check duplicate
		prompts := cmdSvc.GetPrompts()
		for _, p := range prompts {
			if p.GetName() == prompt.GetName() {
				http.Error(w, "prompt with this name already exists", http.StatusConflict)
				return
			}
		}

		prompts = append(prompts, &prompt)
		cmdSvc.SetPrompts(prompts)
		svc.SetCommandLineService(cmdSvc) // Ensure updates propagated

		if err := store.SaveService(ctx, svc); err != nil {
			logging.GetLogger().Error("failed to save service", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Trigger reload
		if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
			logging.GetLogger().Error("failed to reload config after prompt create", "error", err)
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("{}"))
	}
}

// handleUpdatePrompt handles updating a prompt in the user-library service.
func (a *Application) handleUpdatePrompt(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/prompts/")
		if name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}

		var prompt configv1.PromptDefinition
		body, err := readBodyWithLimit(w, r, 1048576)
		if err != nil {
			return
		}
		if err := protojson.Unmarshal(body, &prompt); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		prompt.SetName(name)

		ctx := r.Context()
		svc, err := store.GetService(ctx, "user-library")
		if err != nil {
			logging.GetLogger().Error("failed to get user-library service", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if svc == nil {
			http.Error(w, "user-library service not found", http.StatusInternalServerError)
			return
		}

		cmdSvc := svc.GetCommandLineService()
		if cmdSvc == nil {
			http.Error(w, "user-library invalid state", http.StatusInternalServerError)
			return
		}

		prompts := cmdSvc.GetPrompts()
		found := false
		for i, p := range prompts {
			if p.GetName() == name {
				prompts[i] = &prompt
				found = true
				break
			}
		}

		if !found {
			http.NotFound(w, r)
			return
		}

		cmdSvc.SetPrompts(prompts)
		svc.SetCommandLineService(cmdSvc)

		if err := store.SaveService(ctx, svc); err != nil {
			logging.GetLogger().Error("failed to save service", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
			logging.GetLogger().Error("failed to reload config after prompt update", "error", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}
}

// handleDeletePrompt handles deleting a prompt from the user-library service.
func (a *Application) handleDeletePrompt(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/prompts/")
		if name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		svc, err := store.GetService(ctx, "user-library")
		if err != nil {
			logging.GetLogger().Error("failed to get user-library service", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if svc == nil {
			http.Error(w, "user-library service not found", http.StatusInternalServerError)
			return
		}

		cmdSvc := svc.GetCommandLineService()
		if cmdSvc == nil {
			http.Error(w, "user-library invalid state", http.StatusInternalServerError)
			return
		}

		prompts := cmdSvc.GetPrompts()
		found := false
		var newPrompts []*configv1.PromptDefinition
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
		svc.SetCommandLineService(cmdSvc)

		if err := store.SaveService(ctx, svc); err != nil {
			logging.GetLogger().Error("failed to save service", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err := a.ReloadConfig(ctx, a.fs, a.configPaths); err != nil {
			logging.GetLogger().Error("failed to reload config after prompt delete", "error", err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}
}
