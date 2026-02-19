// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

// handleStackConfig manages stack (collection) configuration in YAML format.
//
// Summary: Gets or sets stack configuration.
//
// Parameters:
//   - store: storage.Storage. The storage interface.
//
// Returns:
//   - http.HandlerFunc: The HTTP handler function.
func (a *Application) handleStackConfig(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only supporting GET and POST for now (POST updates/saves)
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Path: /api/v1/stacks/{id}/config
		// We need to extract ID. API routing in api.go usually strips prefix?
		// api.go: mux.HandleFunc("/stacks/", a.handleStackConfig(store)) -- NOTE: I need to update api.go to route correctly.
		// If api.go routes "/stacks/" -> this handler, we need to parse.
		// Assuming standard pattern: /api/v1/stacks/{id}/config
		path := strings.TrimPrefix(r.URL.Path, "/stacks/")
		parts := strings.Split(path, "/")
		if len(parts) < 2 || parts[1] != "config" {
			http.NotFound(w, r)
			return
		}
		stackID := parts[0]

		if r.Method == http.MethodGet {
			a.getStackConfig(w, r, store, stackID)
		} else {
			a.saveStackConfig(w, r, store, stackID)
		}
	}
}

// getStackConfig retrieves the configuration of a stack as YAML.
//
// Summary: Retrieves stack configuration.
//
// Parameters:
//   - w: http.ResponseWriter. The response writer.
//   - r: *http.Request. The HTTP request.
//   - store: storage.Storage. The storage interface.
//   - stackID: string. The ID of the stack.
func (a *Application) getStackConfig(w http.ResponseWriter, r *http.Request, store storage.Storage, stackID string) {
	collection, err := store.GetServiceCollection(r.Context(), stackID)
	if err != nil {
		logging.GetLogger().Error("failed to get stack config", "id", stackID, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if collection == nil {
		// Return 404
		http.NotFound(w, r)
		return
	}

	// Convert to YAML
	// First Proto -> JSON
	opts := protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}
	jsonBytes, err := opts.Marshal(collection)
	if err != nil {
		logging.GetLogger().Error("failed to marshal stack to json", "id", stackID, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// JSON -> Map -> YAML
	var jsonObj map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonObj); err != nil {
		logging.GetLogger().Error("failed to unmarshal intermediate json", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Clean up empty fields if necessary, or let YAML handle it.
	// We might want to remove "name" if it's redundant with ID, but keeping it is fine.

	yamlBytes, err := yaml.Marshal(jsonObj)
	if err != nil {
		logging.GetLogger().Error("failed to marshal stack to yaml", "id", stackID, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain") // UI expects raw text? Client says text/yaml or plain.
	_, _ = w.Write(yamlBytes)
}

// saveStackConfig updates the configuration of a stack from YAML.
//
// Summary: Updates stack configuration.
//
// Parameters:
//   - w: http.ResponseWriter. The response writer.
//   - r: *http.Request. The HTTP request.
//   - store: storage.Storage. The storage interface.
//   - stackID: string. The ID of the stack.
func (a *Application) saveStackConfig(w http.ResponseWriter, r *http.Request, store storage.Storage, stackID string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	// YAML -> Map -> JSON -> Proto
	var yamlObj map[string]interface{}
	if err := yaml.Unmarshal(body, &yamlObj); err != nil {
		http.Error(w, "Invalid YAML: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate structure roughly?
	// We rely on config.validate?

	jsonBytes, err := json.Marshal(yamlObj)
	if err != nil {
		http.Error(w, "Failed to convert YAML to JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	var collection configv1.Collection
	if err := protojson.Unmarshal(jsonBytes, &collection); err != nil {
		http.Error(w, "Invalid Configuration format: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Ensure ID matches path?
	if collection.GetName() == "" {
		collection.SetName(stackID)
	} else if collection.GetName() != stackID {
		// Allow name change? No, ID in path should be authoritative for the update target.
		// But "Name" is the primary key in Store usually?
		// store.SaveServiceCollection uses Name as ID usually? or we have ID field?
		// Collection proto has `optional string name`.
		// existing logic uses Name.
		if collection.GetName() != stackID {
			http.Error(w, "Stack name in config must match URL path", http.StatusBadRequest)
			return
		}
	}

	// Validate services inside
	for _, svc := range collection.GetServices() {
		if err := config.ValidateOrError(r.Context(), svc); err != nil {
			http.Error(w, "Invalid service in stack: "+err.Error(), http.StatusBadRequest)
			return
		}
		// Security check
		if isUnsafeConfig(svc) && !a.isUnsafeAllowed(r) { // Need to export/access checks
			http.Error(w, "Unsafe service configuration not allowed", http.StatusForbidden)
			return
		}
	}

	if err := store.SaveServiceCollection(r.Context(), &collection); err != nil {
		logging.GetLogger().Error("failed to save stack", "id", stackID, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Reload config to apply changes
	if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
		logging.GetLogger().Error("failed to reload config after stack save", "error", err)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("{}"))
}

// isUnsafeAllowed checks if the current request context allows unsafe configurations.
//
// Summary: Checks permission for unsafe config.
//
// Parameters:
//   - r: *http.Request. The HTTP request.
//
// Returns:
//   - bool: True if allowed, false otherwise.
func (a *Application) isUnsafeAllowed(r *http.Request) bool {
	if os.Getenv("MCPANY_ALLOW_UNSAFE_CONFIG") == util.TrueStr {
		return true
	}
	if auth.NewRBACEnforcer().HasRoleInContext(r.Context(), "admin") {
		return true
	}
	return false
}
