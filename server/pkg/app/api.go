package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/middleware"
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
//
// Summary: Creates the main API handler mux.
//
// Parameters:
//   - store: storage.Storage. The storage backend.
//
// Returns:
//   - http.Handler: The configured handler.
func (a *Application) createAPIHandler(store storage.Storage) http.Handler {
	mux := http.NewServeMux()

	// Apply Login Rate Limit: 1 RPS with a burst of 5.
	trustProxy := os.Getenv("MCPANY_TRUST_PROXY") == util.TrueStr
	loginRateLimiter := middleware.NewHTTPRateLimitMiddleware(1, 5, middleware.WithTrustProxy(trustProxy))

	mux.HandleFunc("/services", a.handleServices(store))
	mux.HandleFunc("/services/validate", a.handleServiceValidate())
	mux.HandleFunc("/services/", a.handleServiceDetail(store))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	// Doctor API
	doctor := health.NewDoctor()
	doctor.AddCheck("configuration", a.configHealthCheck)
	doctor.AddCheck("filesystem", a.filesystemHealthCheck)
	mux.Handle("/doctor", doctor.Handler())
	mux.HandleFunc("/system/status", a.handleSystemStatus)
	mux.HandleFunc("/audit/export", a.handleAuditExport)
	mux.HandleFunc("/validate", a.handleValidate())

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
	mux.HandleFunc("/dashboard/metrics", a.handleDashboardMetrics())
	mux.HandleFunc("/dashboard/traffic", a.handleDashboardTraffic())
	mux.HandleFunc("/dashboard/top-tools", a.handleDashboardTopTools())
	mux.HandleFunc("/dashboard/tool-failures", a.handleDashboardToolFailures())
	mux.HandleFunc("/dashboard/tool-usage", a.handleDashboardToolUsage())
	mux.HandleFunc("/dashboard/health", a.handleDashboardHealth())

	mux.HandleFunc("/templates", a.handleTemplates())
	mux.HandleFunc("/templates/", a.handleTemplateDetail())

	mux.HandleFunc("/profiles", a.handleProfiles(store))
	mux.HandleFunc("/profiles/", a.handleProfileDetail(store))

	// Stacks (Aliases for Collections with YAML support)
	mux.HandleFunc("/stacks/", a.handleStackConfig(store))

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
	mux.Handle("/auth/login", loginRateLimiter.Handler(http.HandlerFunc(a.handleLogin)))
	mux.HandleFunc("/auth/oauth/initiate", a.handleInitiateOAuth)
	mux.HandleFunc("/auth/oauth/callback", a.handleOAuthCallback)

	mux.HandleFunc("/alerts", a.handleAlerts())
	mux.HandleFunc("/alerts/webhook", a.handleAlertWebhook())
	mux.HandleFunc("/alerts/rules", a.handleAlertRules())
	mux.HandleFunc("/alerts/rules/", a.handleAlertRuleDetail())
	mux.HandleFunc("/alerts/", a.handleAlertDetail())

	mux.HandleFunc("/traces", a.handleTraces())
	mux.HandleFunc("/ws/logs", a.handleLogsWS())
	mux.HandleFunc("/ws/traces", a.handleTracesWS())

	return mux
}

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

func checkURLReachability(ctx context.Context, urlStr string) error {
	client := util.NewSafeHTTPClient()
	client.Timeout = 5 * time.Second

	// Try HEAD first
	req, err := http.NewRequestWithContext(ctx, "HEAD", urlStr, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		// Fallback to GET if HEAD is not supported (Method Not Allowed) or fails
		req, err = http.NewRequestWithContext(ctx, "GET", urlStr, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		resp, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to reach %s: %w", urlStr, err)
		}
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 && resp.StatusCode != http.StatusMethodNotAllowed && resp.StatusCode != http.StatusUnauthorized {
		// We treat 401/403 as "reachable but requires auth", which is fine for basic connectivity check (auth check is deeper).
		// But 404 or 500 might indicate issues.
		// Actually, for validation, maybe we should be strict?
		// Let's just warn if it's 5xx. 404 might be valid if it's a base URL.
		if resp.StatusCode >= 500 {
			return fmt.Errorf("server returned error status: %s", resp.Status)
		}
	}
	return nil
}

func checkFilesystemAccess(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("failed to access path: %w", err)
	}
	// We allow both files and directories, so existence is sufficient validation for now.
	return nil
}

func checkCommandAvailability(command string, workDir string) error {
	if command == "" {
		return fmt.Errorf("command is empty")
	}

	// If absolute path, check existence
	if filepath.IsAbs(command) {
		if _, err := os.Stat(command); err != nil {
			return fmt.Errorf("executable not found at %s", command)
		}
	} else {
		// Look in PATH
		if _, err := exec.LookPath(command); err != nil {
			return fmt.Errorf("command %s not found in PATH", command)
		}
	}

	// Check working directory if provided
	if workDir != "" {
		info, err := os.Stat(workDir)
		if err != nil {
			return fmt.Errorf("working directory not found: %s", workDir)
		}
		if !info.IsDir() {
			return fmt.Errorf("working directory path is not a directory: %s", workDir)
		}
	}

	return nil
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
				settings = configv1.GlobalSettings_builder{}.Build()
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
				s.SetValue("[REDACTED]")
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
				http.Error(w, "id is required", http.StatusBadRequest)
				return
			}

			// Validate skipped as config.ValidateOrError expects UpstreamServiceConfig

			if err := store.SaveSecret(r.Context(), &secret); err != nil {
				logging.GetLogger().Error("failed to save secret", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Reload
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after secret save", "error", err)
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
		path := strings.TrimPrefix(r.URL.Path, "/secrets/")
		if path == "" {
			http.Error(w, "id required", http.StatusBadRequest)
			return
		}

		if strings.HasSuffix(path, "/reveal") {
			id := strings.TrimSuffix(path, "/reveal")
			if id == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			a.handleSecretReveal(w, r, id, store)
			return
		}

		switch r.Method {
		case http.MethodGet:
			secret, err := store.GetSecret(r.Context(), path)
			if err != nil {
				logging.GetLogger().Error("failed to get secret", "id", path, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if secret == nil {
				http.NotFound(w, r)
				return
			}
			// Redact
			secret.SetValue("[REDACTED]")
			w.Header().Set("Content-Type", "application/json")
			opts := protojson.MarshalOptions{UseProtoNames: true}
			b, _ := opts.Marshal(secret)
			_, _ = w.Write(b)

		case http.MethodPut:
			var secret configv1.Secret
			body, err := readBodyWithLimit(w, r, 1048576)
			if err != nil {
				return
			}
			if err := protojson.Unmarshal(body, &secret); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if secret.GetName() == "" && secret.GetId() != "" {
				secret.SetName(secret.GetId())
			}


			// Force ID
			secret.SetId(path)

			if err := store.SaveSecret(r.Context(), &secret); err != nil {
				logging.GetLogger().Error("failed to save secret", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after secret update", "error", err)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{}"))

		case http.MethodDelete:
			if err := store.DeleteSecret(r.Context(), path); err != nil {
				logging.GetLogger().Error("failed to delete secret", "id", path, "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after secret delete", "error", err)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (a *Application) handleSecretReveal(w http.ResponseWriter, r *http.Request, id string, store storage.Storage) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	secret, err := store.GetSecret(r.Context(), id)
	if err != nil {
		logging.GetLogger().Error("failed to get secret for reveal", "id", id, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if secret == nil {
		http.NotFound(w, r)
		return
	}

	// Log the access (Audit)
	user, _ := auth.UserFromContext(r.Context())
	logging.GetLogger().Info("Secret revealed", "id", id, "user", user)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"value": secret.GetValue(),
	})
}

func (a *Application) handleProfiles(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Mix of config defined and DB defined?
			// Profiles are in GlobalSettings.
			// Currently GlobalSettings are single object.
			// But DB can store user profiles separately?
			// The handler seems to treat them as separate entities, but config stores them in GlobalSettings.ProfileDefinitions.
			// Storage methods for Profiles might map to GlobalSettings mutation.

			// Assuming Store.ListProfiles exists (it usually extracts from GlobalSettings)
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
			// ProfileDefinition uses Name as identifier, no ID field.

			if err := store.SaveProfile(r.Context(), &profile); err != nil {
				logging.GetLogger().Error("failed to save profile", "error", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			// Trigger reload
			if err := a.ReloadConfig(r.Context(), a.fs, a.configPaths); err != nil {
				logging.GetLogger().Error("failed to reload config after profile save", "error", err)
			}
			w.WriteHeader(http.StatusOK)
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
			profile.SetName(name) // Force name match

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
			collection.SetName(name) // Force name match

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
	for _, rawSvc := range collection.GetServices() {
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
	if service.GetFilesystemService() != nil {
		return true
	}
	if service.GetSqlService() != nil {
		return true
	}
	return false
}
