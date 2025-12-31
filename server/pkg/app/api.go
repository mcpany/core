// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/mcpany/core/pkg/config"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/storage"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// createAPIHandler creates a http.Handler for the config API.
func (a *Application) createAPIHandler(store storage.Storage) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			services, err := store.ListServices(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to list services", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			var buf []byte
			buf = append(buf, '[')
			opts := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: false}
			for i, svc := range services {
				if i > 0 {
					buf = append(buf, ',')
				}
				b, err := opts.Marshal(svc)
				if err != nil {
					logging.GetLogger().Error("failed to marshal service", "error", err)
					continue
				}
				buf = append(buf, b...)
			}
			buf = append(buf, ']')
			_, _ = w.Write(buf)

		case http.MethodPost:
			// Limit request body to 1MB
			r.Body = http.MaxBytesReader(w, r.Body, 1048576)
			var svc configv1.UpstreamServiceConfig
			body, err := io.ReadAll(r.Body)
			if err != nil {
				var maxBytesErr *http.MaxBytesError
				if errors.As(err, &maxBytesErr) {
					http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
					return
				}
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			if err := protojson.Unmarshal(body, &svc); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if svc.GetName() == "" {
				http.Error(w, "name is required", http.StatusBadRequest)
				return
			}

			// Validate service configuration before saving
			if err := config.ValidateOrError(r.Context(), &svc); err != nil {
				http.Error(w, "invalid service configuration: "+err.Error(), http.StatusBadRequest)
				return
			}

			// Auto-generate ID if missing? Store handles it if we pass empty ID (fallback to name).
			// But creating UUID here might be better? For now name fallback is fine.

			if err := store.SaveService(r.Context(), &svc); err != nil {
				logging.GetLogger().Error("failed to save service", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Trigger reload
			if err := a.ReloadConfig(a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after save", "error", err)
				// Don't fail the request, but log error
			}

			w.WriteHeader(http.StatusCreated)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/services/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/services/")
		parts := strings.Split(path, "/")
		if len(parts) < 1 || parts[0] == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}
		name := parts[0]

		if len(parts) == 2 && parts[1] == "status" {
			// Get Service Status
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}

			// For now, check if service exists in store and return active if so.
			// TODO: Hook into real metrics/health check.
			svc, err := store.GetService(name)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if svc == nil {
				http.NotFound(w, r)
				return
			}

			// Mock status for now, or check real connection if possible.
			// We can check if it's in a.ToolManager.ListServices() which implies it's loaded.
			loaded := false
			for _, info := range a.ToolManager.ListServices() {
				if info.Name == name {
					loaded = true
					break
				}
			}

			status := "Inactive"
			if loaded {
				status = "Active"
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"name":    name,
				"status":  status,
				"metrics": map[string]any{}, // TODO: Populate with real metrics
			})
			return
		}

		if len(parts) > 1 {
			http.NotFound(w, r)
			return
		}

		// ... existing logic for /services/{name} ...
		switch r.Method {
		case http.MethodGet:
			svc, err := store.GetService(r.Context(), name)
			if err != nil {
				logging.GetLogger().Error("failed to get service", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if svc == nil {
				http.NotFound(w, r)
				return
			}
			opts := protojson.MarshalOptions{UseProtoNames: true}
			b, _ := opts.Marshal(svc)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(b)
		case http.MethodPut:
			// Limit request body to 1MB
			r.Body = http.MaxBytesReader(w, r.Body, 1048576)
			var svc configv1.UpstreamServiceConfig
			body, err := io.ReadAll(r.Body)
			if err != nil {
				var maxBytesErr *http.MaxBytesError
				if errors.As(err, &maxBytesErr) {
					http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
					return
				}
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			if err := protojson.Unmarshal(body, &svc); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			svc.Name = proto.String(name) // Force name match

			// Validate service configuration before saving
			if err := config.ValidateOrError(r.Context(), &svc); err != nil {
				http.Error(w, "invalid service configuration: "+err.Error(), http.StatusBadRequest)
				return
			}

			if err := store.SaveService(r.Context(), &svc); err != nil {
				logging.GetLogger().Error("failed to save service", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if err := a.ReloadConfig(a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after update", "error", err)
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			if err := store.DeleteService(r.Context(), name); err != nil {
				logging.GetLogger().Error("failed to delete service", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if err := a.ReloadConfig(a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after delete", "error", err)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	mux.HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			settings, err := store.GetGlobalSettings()
			if err != nil {
				logging.GetLogger().Error("failed to get global settings", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if settings == nil {
				// Return default empty settings if not found
				settings = &configv1.GlobalSettings{}
			}
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}
			b, err := opts.Marshal(settings)
			if err != nil {
				logging.GetLogger().Error("failed to marshal settings", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			_, _ = w.Write(b)

		case http.MethodPost:
			var settings configv1.GlobalSettings
			body, _ := io.ReadAll(r.Body)
			if err := protojson.Unmarshal(body, &settings); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			// Hardcode ID to 1? Or just pass it to store which handles it.
			if err := store.SaveGlobalSettings(&settings); err != nil {
				logging.GetLogger().Error("failed to save global settings", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Trigger reload
			if err := a.ReloadConfig(a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after settings save", "error", err)
				// Don't fail the request, but log
			}

			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Tools
	mux.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tools := a.ToolManager.ListTools()
			var toolList []*mcp.Tool
			for _, t := range tools {
				toolList = append(toolList, t.MCPTool())
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(toolList)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Execute Tool
	mux.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req tool.ExecutionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// If ToolInputs is not set but Arguments is, marshal Arguments to ToolInputs
		// The tool manager expects ToolInputs (json.RawMessage)
		if len(req.ToolInputs) == 0 && len(req.Arguments) > 0 {
			b, err := json.Marshal(req.Arguments)
			if err != nil {
				http.Error(w, "failed to marshal arguments", http.StatusBadRequest)
				return
			}
			req.ToolInputs = b
		}

		result, err := a.ToolManager.ExecuteTool(r.Context(), &req)
		if err != nil {
			logging.GetLogger().Error("failed to execute tool", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	})

	// Prompts
	mux.HandleFunc("/prompts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			prompts := a.PromptManager.ListPrompts()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(prompts)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Resources
	mux.HandleFunc("/resources", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resources := a.ResourceManager.ListResources()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resources)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Secrets
	mux.HandleFunc("/secrets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			secrets, err := store.ListSecrets()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// Redact sensitive values
			for _, s := range secrets {
				s.Value = proto.String("[REDACTED]")
			}
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
			// Client expects array, let's just write array manually or use Gateway behavior if possible.
			// protojson doesn't support marshaling slice directly.
			// Let's construct a JSON array manually using the partials.
			var buf []byte
			buf = append(buf, '[')
			for i, s := range secrets {
				if i > 0 {
					buf = append(buf, ',')
				}
				b, _ := opts.Marshal(s)
				buf = append(buf, b...)
			}
			buf = append(buf, ']')
			_, _ = w.Write(buf)

		case http.MethodPost:
			var secret configv1.Secret
			body, _ := io.ReadAll(r.Body)
			if err := protojson.Unmarshal(body, &secret); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if secret.GetId() == "" {
				// Generate ID if missing (or use UUID)
				// For now error if missing
				// http.Error(w, "id is required", http.StatusBadRequest)
				// actually client might send without ID for new
				// let's verify if store handles it? No store checks ID.
				// Client should likely generate ID or we do it here.
				// For now assume client sends ID.
			}
			if err := store.SaveSecret(&secret); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/secrets/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/secrets/")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodDelete:
			if err := store.DeleteSecret(id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
