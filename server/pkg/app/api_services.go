package app

import (
	"encoding/json"
	"fmt"
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
)

func (a *Application) handleServices(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			a.handleListServices(w, r, store)
		case http.MethodPost:
			a.handleCreateService(w, r, store)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleListServices(w http.ResponseWriter, r *http.Request, store storage.Storage) {
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

			// Inject Tool Count
			if a.ToolManager != nil {
				tools := a.ToolManager.ListTools()
				count := 0
				svcID := svc.GetId()
				// Fallback to name if ID is empty or not matching (though tools should use ID)
				sanitizedName := svc.GetSanitizedName()

				for _, t := range tools {
					tSvcID := t.Tool().GetServiceId()
					if tSvcID != "" && (tSvcID == svcID || tSvcID == sanitizedName) {
						count++
					}
				}
				jsonMap["tool_count"] = count
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
}

func (a *Application) handleCreateService(w http.ResponseWriter, r *http.Request, store storage.Storage) {
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

	// Sentinel Security: Block unsafe configurations unless admin or explicitly allowed
	if isUnsafeConfig(&svc) {
		allow := false
		if os.Getenv("MCPANY_ALLOW_UNSAFE_CONFIG") == util.TrueStr {
			allow = true
		} else if auth.NewRBACEnforcer().HasRoleInContext(r.Context(), "admin") {
			allow = true
		}

		if !allow {
			logging.GetLogger().Warn("Blocked unsafe service creation via API", "service", svc.GetName())
			http.Error(w, "Creation of unsafe services (filesystem/sql/stdio/command_line) is restricted to admins. Configure them via file or ensure you have admin privileges.", http.StatusForbidden)
			return
		}
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
}

func (a *Application) handleServiceValidate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var svc configv1.UpstreamServiceConfig
		body, err := readBodyWithLimit(w, r, 1048576)
		if err != nil {
			return
		}
		if err := protojson.Unmarshal(body, &svc); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 1. Static Validation
		if err := config.ValidateOrError(r.Context(), &svc); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"valid":   false,
				"error":   err.Error(),
				"details": "Static validation failed",
			})
			return
		}

		// 2. Connectivity / Health Check
		var checkErr error
		var checkDetails string

		// HTTP & GraphQL
		if httpSvc := svc.GetHttpService(); httpSvc != nil {
			checkErr = checkURLReachability(r.Context(), httpSvc.GetAddress())
			checkDetails = "HTTP reachability check failed"
		} else if gqlSvc := svc.GetGraphqlService(); gqlSvc != nil {
			checkErr = checkURLReachability(r.Context(), gqlSvc.GetAddress())
			checkDetails = "GraphQL reachability check failed"
		} else if fsSvc := svc.GetFilesystemService(); fsSvc != nil {
			// Filesystem check
			for _, path := range fsSvc.GetRootPaths() {
				if err := checkFilesystemAccess(path); err != nil {
					checkErr = err
					checkDetails = fmt.Sprintf("Filesystem path check failed for %s", path)
					break
				}
			}
		} else if cmdSvc := svc.GetCommandLineService(); cmdSvc != nil {
			// Command check
			checkErr = checkCommandAvailability(cmdSvc.GetCommand(), cmdSvc.GetWorkingDirectory())
			checkDetails = "Command availability check failed"
		} else if mcpSvc := svc.GetMcpService(); mcpSvc != nil {
			// MCP Remote check (if stdio, check command; if http, check url)
			switch mcpSvc.WhichConnectionType() {
			case configv1.McpUpstreamService_StdioConnection_case:
				stdio := mcpSvc.GetStdioConnection()
				if stdio != nil {
					checkErr = checkCommandAvailability(stdio.GetCommand(), stdio.GetWorkingDirectory())
					checkDetails = "MCP Stdio command check failed"
				}
			case configv1.McpUpstreamService_HttpConnection_case:
				httpConn := mcpSvc.GetHttpConnection()
				if httpConn != nil {
					checkErr = checkURLReachability(r.Context(), httpConn.GetHttpAddress())
					checkDetails = "MCP HTTP reachability check failed"
				}
			}
		}

		if checkErr != nil {
			w.Header().Set("Content-Type", "application/json")
			// Return 200 OK but with valid=false to distinguish from malformed request
			_ = json.NewEncoder(w).Encode(map[string]any{
				"valid":   false,
				"error":   checkErr.Error(),
				"details": checkDetails,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"valid": true,
		})
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

		if len(parts) == 2 && parts[1] == "restart" {
			a.handleServiceRestart(w, r, name, store)
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
			svc.SetName(name) // Force name match

			// Validate service configuration before saving
			if err := config.ValidateOrError(r.Context(), &svc); err != nil {
				http.Error(w, "invalid service configuration: "+err.Error(), http.StatusBadRequest)
				return
			}

			// Sentinel Security: Block unsafe configurations unless admin or explicitly allowed
			if isUnsafeConfig(&svc) {
				allow := false
				if os.Getenv("MCPANY_ALLOW_UNSAFE_CONFIG") == util.TrueStr {
					allow = true
				} else if auth.NewRBACEnforcer().HasRoleInContext(r.Context(), "admin") {
					allow = true
				}

				if !allow {
					logging.GetLogger().Warn("Blocked unsafe service update via API", "service", name)
					http.Error(w, "Configuration of unsafe services (filesystem/sql/stdio/command_line) is restricted to admins. Configure them via file or ensure you have admin privileges.", http.StatusForbidden)
					return
				}
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

func (a *Application) handleServiceRestart(w http.ResponseWriter, r *http.Request, name string, store storage.Storage) {
	if r.Method != http.MethodPost {
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

	if a.ServiceRegistry != nil {
		// Unregister to force stop
		if err := a.ServiceRegistry.UnregisterService(r.Context(), name); err != nil {
			logging.GetLogger().Error("failed to unregister service during restart", "name", name, "error", err)
			// Continue to reload, as it might just be not running or already stopped
		}
	}

	// Trigger reload to re-register
	if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
		logging.GetLogger().Error("failed to reload config after restart", "error", err)
		http.Error(w, "Failed to restart service: "+err.Error(), http.StatusInternalServerError)
		return
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
	if service.GetFilesystemService() != nil {
		return true
	}
	if service.GetSqlService() != nil {
		return true
	}
	return false
}
