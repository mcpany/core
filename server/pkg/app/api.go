// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// readBodyWithLimit reads the request body with a limit and returns the bytes.
// If the body exceeds the limit, it writes an error response and returns nil, error.
func readBodyWithLimit(w http.ResponseWriter, r *http.Request, limit int64) ([]byte, error) {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return nil, err
		}
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return nil, err
	}
	return body, nil
}

// createAPIHandler creates a http.Handler for the config API.
func (a *Application) createAPIHandler(store storage.Storage) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/services", a.handleServices(store))
	mux.HandleFunc("/services/", a.handleServiceDetail(store))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	// Doctor API
	doctor := health.NewDoctor()
	doctor.AddCheck("configuration", a.configHealthCheck)
	mux.Handle("/doctor", doctor.Handler())

	mux.HandleFunc("/settings", a.handleSettings(store))
	mux.HandleFunc("/debug/auth-test", a.handleAuthTest())

	mux.HandleFunc("/tools", a.handleTools())
	mux.HandleFunc("/execute", a.handleExecute())

	mux.HandleFunc("/prompts", a.handlePrompts())
	mux.HandleFunc("/prompts/", a.handlePromptExecute()) // Handles /prompts/{name}/execute

	mux.HandleFunc("/resources", a.handleResources())
	mux.HandleFunc("/resources/read", a.handleResourceRead())

	mux.HandleFunc("/secrets", a.handleSecrets(store))
	mux.HandleFunc("/secrets/", a.handleSecretDetail(store))

	mux.HandleFunc("/topology", a.handleTopology())

	mux.HandleFunc("/templates", a.handleTemplates())
	mux.HandleFunc("/templates/", a.handleTemplateDetail())

	mux.HandleFunc("/profiles", a.handleProfiles(store))
	mux.HandleFunc("/profiles/", a.handleProfileDetail(store))

	mux.HandleFunc("/collections", a.handleCollections(store))
	mux.HandleFunc("/collections/", a.handleCollectionDetail(store))

	// Users
	mux.HandleFunc("/users", a.handleUsers(store))
	mux.HandleFunc("/users/", a.handleUserDetail(store))

	// Credentials
	mux.HandleFunc("/credentials", a.listCredentialsHandler)
	mux.HandleFunc("/credentials/", func(w http.ResponseWriter, r *http.Request) {
		// Manual dispatch for detail vs specific
		// listCredentialsHandler handles GET /credentials (handled above)
		// create is POST /credentials (handled below)
		// Detail methods use path suffix
		if r.Method == http.MethodPost {
			a.createCredentialHandler(w, r)
			return
		}
		// Check if it's a detail request
		path := strings.TrimPrefix(r.URL.Path, "/credentials/")
		if path != "" {
			switch r.Method {
			case http.MethodGet:
				a.getCredentialHandler(w, r)
			case http.MethodPut:
				a.updateCredentialHandler(w, r)
			case http.MethodDelete:
				a.deleteCredentialHandler(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		http.NotFound(w, r)
	})

	// Auth (OAuth)
	mux.HandleFunc("/auth/oauth/initiate", a.handleInitiateOAuth)
	mux.HandleFunc("/auth/oauth/callback", a.handleOAuthCallback)

	mux.HandleFunc("/ws/logs", a.handleLogsWS())

	return mux
}

func (a *Application) handleServices(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			var services []*configv1.UpstreamServiceConfig
			var err error
			if a.ServiceRegistry != nil {
				services, err = a.ServiceRegistry.GetAllServices()
			} else {
				// Fallback to store if registry not initialized (though it should be)
				services, err = store.ListServices(r.Context())
			}
			if err != nil {
				logging.GetLogger().Error("failed to list services", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			var buf []byte
			buf = append(buf, '[')
			opts := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: false}
			// Sort services for consistent output
			// (Optional but good for tests)

			for i, svc := range services {
				if i > 0 {
					buf = append(buf, ',')
				}
				b, err := opts.Marshal(svc)
				if err != nil {
					logging.GetLogger().Error("failed to marshal service", "error", err)
					continue
				}

				// Inject runtime error information if available
				// We unmarshal the JSON bytes to a map, inject the error field, and marshal back.
				// This is a trade-off for not modifying the proto definition for a transient status.
				var jsonMap map[string]any
				if err := json.Unmarshal(b, &jsonMap); err == nil && a.ServiceRegistry != nil {
					if svcID := svc.GetId(); svcID != "" {
						if errMsg, ok := a.ServiceRegistry.GetServiceError(svcID); ok {
							jsonMap["last_error"] = errMsg
						}
					}
					// Also check sanitize name if ID lookup fails (or both?)
					if svc.GetId() == "" && svc.GetSanitizedName() != "" {
						if errMsg, ok := a.ServiceRegistry.GetServiceError(svc.GetSanitizedName()); ok {
							jsonMap["last_error"] = errMsg
						}
					}

					// Marshal back to JSON
					if enrichedBytes, err := json.Marshal(jsonMap); err == nil {
						b = enrichedBytes
					}
				}

				buf = append(buf, b...)
			}
			buf = append(buf, ']')
			_, _ = w.Write(buf)

		case http.MethodPost:
			var svc configv1.UpstreamServiceConfig
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
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

			if isUnsafeConfig(&svc) && os.Getenv("MCPANY_ALLOW_UNSAFE_CONFIG") != util.TrueStr {
				logging.GetLogger().Warn("Blocked unsafe service creation via API", "service", svc.GetName())
				http.Error(w, "Creation of local command execution services (stdio/command_line) is disabled for security reasons. Configure them via file instead.", http.StatusBadRequest)
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
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after save", "error", err)
			}

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("{}"))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleServiceDetail(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/services/")
		parts := strings.Split(path, "/")
		if len(parts) < 1 || parts[0] == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}
		name := parts[0]

		if len(parts) == 2 && parts[1] == "status" {
			a.handleServiceStatus(w, r, name, store)
			return
		}

		if len(parts) == 2 && parts[1] == "check" {
			a.handleServiceCheck(w, r, name)
			return
		}

		if len(parts) > 1 {
			http.NotFound(w, r)
			return
		}

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
			var svc configv1.UpstreamServiceConfig
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
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

			if isUnsafeConfig(&svc) && os.Getenv("MCPANY_ALLOW_UNSAFE_CONFIG") != util.TrueStr {
				logging.GetLogger().Warn("Blocked unsafe service update via API", "service", name)
				http.Error(w, "Configuration of local command execution services (stdio/command_line) is disabled for security reasons. Configure them via file instead.", http.StatusBadRequest)
				return
			}

			if err := store.SaveService(r.Context(), &svc); err != nil {
				logging.GetLogger().Error("failed to save service", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after update", "error", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))
		case http.MethodDelete:
			if err := store.DeleteService(r.Context(), name); err != nil {
				logging.GetLogger().Error("failed to delete service", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after delete", "error", err)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleServiceStatus(w http.ResponseWriter, r *http.Request, name string, store storage.Storage) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	svc, err := store.GetService(r.Context(), name)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if svc == nil {
		http.NotFound(w, r)
		return
	}

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
		"metrics": map[string]any{},
	})
}

func (a *Application) handleSettings(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			settings, err := store.GetGlobalSettings(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to get global settings", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if settings == nil {
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
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}
			if err := protojson.Unmarshal(body, &settings); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := store.SaveGlobalSettings(r.Context(), &settings); err != nil {
				logging.GetLogger().Error("failed to save global settings", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after settings save", "error", err)
			}

			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleTools() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}

func (a *Application) handleExecute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req tool.ExecutionRequest
		// Limit execution request body to 5MB (tools might have large arguments)
		body, err := readBodyWithLimit(w, r, 5*1024*1024)
		if err != nil {
			return
		}

		if err := json.Unmarshal(body, &req); err != nil {
			logging.GetLogger().Error("failed to decode execution request", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

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
	}
}

func (a *Application) handlePrompts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			prompts := a.PromptManager.ListPrompts()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(prompts)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleResources() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			resources := a.ResourceManager.ListResources()
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resources)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleSecrets(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			secrets, err := store.ListSecrets(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to list secrets", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			// Redact sensitive values
			for _, s := range secrets {
				s.Value = proto.String("[REDACTED]")
			}
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
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
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}
			if err := protojson.Unmarshal(body, &secret); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if secret.GetId() == "" {
				secret.Id = proto.String(uuid.New().String())
			}
			if err := store.SaveSecret(r.Context(), &secret); err != nil {
				logging.GetLogger().Error("failed to save secret", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleSecretDetail(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/secrets/")
		if id == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			secret, err := store.GetSecret(r.Context(), id)
			if err != nil {
				logging.GetLogger().Error("failed to get secret", "id", id, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if secret == nil {
				http.NotFound(w, r)
				return
			}
			// Redact sensitive value
			secret.Value = proto.String("[REDACTED]")

			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
			b, _ := opts.Marshal(secret)
			_, _ = w.Write(b)

		case http.MethodDelete:
			if err := store.DeleteSecret(r.Context(), id); err != nil {
				logging.GetLogger().Error("failed to delete secret", "id", id, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleProfiles(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			profiles, err := store.ListProfiles(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to list profiles", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
			var buf []byte
			buf = append(buf, '[')
			for i, p := range profiles {
				if i > 0 {
					buf = append(buf, ',')
				}
				b, _ := opts.Marshal(p)
				buf = append(buf, b...)
			}
			buf = append(buf, ']')
			_, _ = w.Write(buf)

		case http.MethodPost:
			var profile configv1.ProfileDefinition
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}
			if err := protojson.Unmarshal(body, &profile); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if profile.GetName() == "" {
				http.Error(w, "name is required", http.StatusBadRequest)
				return
			}
			if err := store.SaveProfile(r.Context(), &profile); err != nil {
				logging.GetLogger().Error("failed to save profile", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			// Trigger reload
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after profile save", "error", err)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("{}"))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleProfileDetail(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/profiles/")
		if name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}

		if strings.HasSuffix(name, "/export") {
			name = strings.TrimSuffix(name, "/export")
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			profile, err := store.GetProfile(r.Context(), name)
			if err != nil {
				logging.GetLogger().Error("failed to get profile for export", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if profile == nil {
				http.NotFound(w, r)
				return
			}
			exportProfile := proto.Clone(profile).(*configv1.ProfileDefinition)
			config.StripSecretsFromProfile(exportProfile)
			w.Header().Set("Content-Type", "application/json")
			// Force download? Maybe 'Content-Disposition: attachment; filename="profile.json"'
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.json\"", name))
			opts := protojson.MarshalOptions{UseProtoNames: true, Multiline: true, Indent: "  "}
			b, _ := opts.Marshal(exportProfile)
			_, _ = w.Write(b)
			return
		}

		switch r.Method {
		case http.MethodGet:
			profile, err := store.GetProfile(r.Context(), name)
			if err != nil {
				logging.GetLogger().Error("failed to get profile", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if profile == nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
			b, _ := opts.Marshal(profile)
			_, _ = w.Write(b)

		case http.MethodPut:
			var profile configv1.ProfileDefinition
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}
			if err := protojson.Unmarshal(body, &profile); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			profile.Name = proto.String(name) // Force name match

			if err := store.SaveProfile(r.Context(), &profile); err != nil {
				logging.GetLogger().Error("failed to save profile", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after profile update", "error", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))

		case http.MethodDelete:
			if err := store.DeleteProfile(r.Context(), name); err != nil {
				logging.GetLogger().Error("failed to delete profile", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after profile delete", "error", err)
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleCollections(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			collections, err := store.ListServiceCollections(r.Context())
			if err != nil {
				logging.GetLogger().Error("failed to list collections", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
			var buf []byte
			buf = append(buf, '[')
			for i, c := range collections {
				if i > 0 {
					buf = append(buf, ',')
				}
				b, _ := opts.Marshal(c)
				buf = append(buf, b...)
			}
			buf = append(buf, ']')
			_, _ = w.Write(buf)

		case http.MethodPost:
			var collection configv1.Collection
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}
			if err := protojson.Unmarshal(body, &collection); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if collection.GetName() == "" {
				http.Error(w, "name is required", http.StatusBadRequest)
				return
			}
			if err := store.SaveServiceCollection(r.Context(), &collection); err != nil {
				logging.GetLogger().Error("failed to save collection", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("{}"))
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleCollectionDetail(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/collections/")
		if name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}

		if strings.HasSuffix(name, "/export") {
			name = strings.TrimSuffix(name, "/export")
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			collection, err := store.GetServiceCollection(r.Context(), name)
			if err != nil {
				logging.GetLogger().Error("failed to get collection for export", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if collection == nil {
				http.NotFound(w, r)
				return
			}
			exportCollection := proto.Clone(collection).(*configv1.Collection)
			config.StripSecretsFromCollection(exportCollection)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.json\"", name))
			opts := protojson.MarshalOptions{UseProtoNames: true, Multiline: true, Indent: "  "}
			b, _ := opts.Marshal(exportCollection)
			_, _ = w.Write(b)
			return
		}

		if strings.HasSuffix(name, "/apply") {
			name = strings.TrimSuffix(name, "/apply")
			a.handleCollectionApply(w, r, name, store)
			return
		}

		switch r.Method {
		case http.MethodGet:
			collection, err := store.GetServiceCollection(r.Context(), name)
			if err != nil {
				logging.GetLogger().Error("failed to get collection", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if collection == nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
			b, _ := opts.Marshal(collection)
			_, _ = w.Write(b)

		case http.MethodPut:
			var collection configv1.Collection
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}
			if err := protojson.Unmarshal(body, &collection); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			collection.Name = proto.String(name) // Force name match

			if err := store.SaveServiceCollection(r.Context(), &collection); err != nil {
				logging.GetLogger().Error("failed to save collection", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))

		case http.MethodDelete:
			if err := store.DeleteServiceCollection(r.Context(), name); err != nil {
				logging.GetLogger().Error("failed to delete collection", "name", name, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleCollectionApply(w http.ResponseWriter, r *http.Request, name string, store storage.Storage) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	collection, err := store.GetServiceCollection(r.Context(), name)
	if err != nil {
		logging.GetLogger().Error("failed to get collection for apply", "name", name, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if collection == nil {
		http.NotFound(w, r)
		return
	}

	// Apply services
	for _, rawSvc := range collection.Services {
		svc := proto.Clone(rawSvc).(*configv1.UpstreamServiceConfig)
		// We should probably check if service already exists?
		// "Upsert" logic ideally.
		// And we need to validate it.
		if err := config.ValidateOrError(r.Context(), svc); err != nil {
			logging.GetLogger().Error("invalid service in collection", "service", svc.GetName(), "error", err)
			continue // Skip invalid? Or error out?
		}

		if isUnsafeConfig(svc) && os.Getenv("MCPANY_ALLOW_UNSAFE_CONFIG") != util.TrueStr {
			logging.GetLogger().Warn("Skipping unsafe service in collection apply", "service", svc.GetName())
			continue
		}

		if err := store.SaveService(r.Context(), svc); err != nil {
			logging.GetLogger().Error("failed to save service from collection", "service", svc.GetName(), "error", err)
			// Continue or abort?
			// Maybe best effort?
		}
	}

	// Trigger reload
	if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
		logging.GetLogger().Error("failed to reload config after collection apply", "error", err)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{}"))
}

func isUnsafeConfig(service *configv1.UpstreamServiceConfig) bool {
	if mcp := service.GetMcpService(); mcp != nil {
		connType := mcp.WhichConnectionType()
		if connType == configv1.McpUpstreamService_StdioConnection_case ||
			connType == configv1.McpUpstreamService_BundleConnection_case {
			return true
		}
	}
	if service.GetCommandLineService() != nil {
		return true
	}
	return false
}

func (a *Application) handleServiceCheck(w http.ResponseWriter, r *http.Request, name string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if a.ServiceRegistry == nil {
		http.Error(w, "Service Registry not available", http.StatusInternalServerError)
		return
	}

	err := a.ServiceRegistry.CheckServiceHealth(r.Context(), name)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Service is healthy",
	})
}
