// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	pb_admin "github.com/mcpany/core/proto/admin/v1"
	v1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/admin"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/util"
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	// config_v1 "github.com/mcpany/core/proto/config/v1".
	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/api/rest"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func (a *Application) runServerMode(
	ctx context.Context,
	mcpSrv *mcpserver.Server,
	bus *bus.Provider,
	bindAddress, grpcPort string,
	shutdownTimeout time.Duration,
	globalSettings *config_v1.GlobalSettings,
	cachingMiddleware *middleware.CachingMiddleware,
	standardMiddlewares *middleware.StandardMiddlewares,
	store storage.Storage,
	serviceRegistry *serviceregistry.ServiceRegistry,
	startupCallback func(),
	tlsCert, tlsKey, tlsClientCA string,
) error {
	ipMiddleware, err := middleware.NewIPAllowlistMiddleware(a.SettingsManager.GetAllowedIPs())
	if err != nil {
		return fmt.Errorf("failed to create IP allowlist middleware: %w", err)
	}

	a.configMu.Lock()
	a.ipMiddleware = ipMiddleware
	a.configMu.Unlock()

	// localCtx is used to manage the lifecycle of the servers started in this function.
	// It's canceled when this function returns, ensuring that all servers are shut down.
	localCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errChan := make(chan error, 2)
	readyChan := make(chan struct{}, 2)
	expectedReady := 0
	var wg sync.WaitGroup

	rawHTTPHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return mcpSrv.Server()
	}, nil)

	// Wrap the HTTP handler with OpenTelemetry instrumentation
	// Note: We don't inject HTTPRequestContextKey here anymore because we do it globally
	// in the HTTPRequestContextMiddleware.
	httpHandler := otelhttp.NewHandler(rawHTTPHandler, "server-request")

	// Check if auth middleware is disabled in config
	var authDisabled bool

	// Use passed globalSettings for middleware config check
	if globalSettings != nil {
		logging.GetLogger().Info("DEBUG: GlobalSettings middlewares", "count", len(globalSettings.GetMiddlewares()), "middlewares", globalSettings.GetMiddlewares())
		for _, m := range globalSettings.GetMiddlewares() {
			if m.GetName() == authMiddlewareName && m.GetDisabled() {
				// Only disable if API Key is NOT set.
				// If API Key is present, we enforce auth regardless of this flag to prevent accidental exposure.
				if a.SettingsManager.GetAPIKey() == "" {
					authDisabled = true
				} else {
					logging.GetLogger().Warn("Auth middleware disabled in config but API Key is present. IGNORING disable flag to enforce security.", "api_key_present", true)
				}
				break
			}
		}
	}
	// Note: We don't fall back to config.GlobalSettings() singleton here because it
	// might be modified by other tests in the same package, leading to flaky tests.
	// If globalSettings is nil, authDisabled remains false (enabled).

	// Trust Proxy Config
	trustProxy := os.Getenv("MCPANY_TRUST_PROXY") == util.TrueStr

	var authMiddleware func(http.Handler) http.Handler
	if authDisabled {
		logging.GetLogger().Warn("Auth middleware is disabled by config! Enforcing private-IP-only access for safety.")
		// Even if auth is disabled, we enforce private-IP-only access to prevent public exposure.
		authMiddleware = a.createAuthMiddleware(true, trustProxy)
	} else {
		authMiddleware = a.createAuthMiddleware(false, trustProxy)
	}

	mux := http.NewServeMux()

	// UI Handler
	// We prioritize serving from build directories (./ui/out, ./ui/dist).
	// If only ./ui exists, we check if it contains source code (package.json) and block it if so.
	// We use a.fs (Afero) to allow testing/mocking.
	var uiPath string
	var uiFS http.FileSystem

	if _, err := a.fs.Stat("./ui/out"); err == nil {
		uiPath = "./ui/out"
	} else if _, err := a.fs.Stat("./ui/dist"); err == nil {
		uiPath = "./ui/dist"
	} else if _, err := a.fs.Stat("./ui"); err == nil {
		// Check for package.json to detect source code
		if _, err := a.fs.Stat("./ui/package.json"); err == nil {
			logging.GetLogger().Warn("UI directory ./ui contains package.json. Refusing to serve source code for security.", "path", "./ui")
		} else {
			uiPath = "./ui"
		}
	} else {
		logging.GetLogger().Info("No UI directory found (./ui/out, ./ui/dist, ./ui). UI will not be served.")
	}

	if uiPath != "" {
		// Use Afero's httpFs adapter
		// We create a BasePathFs to restrict access to the UI directory
		baseFs := afero.NewBasePathFs(a.fs, uiPath)
		uiFS = afero.NewHttpFs(baseFs)

		// File server with Cache-Control headers
		fileServer := http.FileServer(uiFS)
		cachingFileServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add Cache-Control headers
			// For immutable assets (usually hashed), we can cache for a long time.
			// Next.js puts static assets in _next/static.
			if strings.Contains(r.URL.Path, "_next/static/") || strings.Contains(r.URL.Path, "static/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else if strings.HasSuffix(r.URL.Path, ".html") || r.URL.Path == "/" {
				// HTML files should not be cached (or short cache) to ensure updates are seen
				w.Header().Set("Cache-Control", "no-cache")
			}
			fileServer.ServeHTTP(w, r)
		})

		// Apply Gzip compression
		handler := middleware.GzipCompressionMiddleware(cachingFileServer)
		mux.Handle("/ui/", http.StripPrefix("/ui", handler))
	}

	// Handle root path with gRPC-Web support
	// We defer the decision to the wrapper or the httpHandler
	// But we need wrappedGrpc to be ready.
	// Since we are moving gRPC init before this, we can use a closure.
	// However, we haven't moved it yet in this execution flow relative to lines 1179.
	// So we need to do the setup HERE or move this Handler registration DOWN?
	// Moving mux.Handle("/", ...) down is safer.

	// API Routes for Configuration Management
	// Protected by auth middleware
	apiHandler := http.StripPrefix("/api/v1", a.createAPIHandler(store))
	mux.Handle("/api/v1/", authMiddleware(apiHandler))

	// Topology API is now handled by apiHandler via api.go

	// Catalog API
	// Expose via REST for UI
	mux.Handle("/api/v1/catalog/services", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		catalogServer := rest.NewCatalogServer(a.CatalogManager)
		resp, err := catalogServer.ListServices(r.Context(), &v1.ListCatalogServicesRequest{})
		if err != nil {
			logging.GetLogger().Error("Failed to list catalog services", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		b, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(resp)
		if err != nil {
			logging.GetLogger().Error("Failed to marshal catalog services response", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(b)
	})))

	logging.GetLogger().Info("DEBUG: Registering /mcp/u/ handler")
	// Multi-user handler
	// pattern: /mcp/u/{uid}/profile/{profile_id}
	// We use a prefix match via stripping.
	// NOTE: We manually handle the path parsing because we support subpaths like /sse or /messages
	mux.HandleFunc("/mcp/u/", func(w http.ResponseWriter, r *http.Request) {
		// Fix data race: Shadow ctx to prevent modifying the captured variable from outer scope.
		// Use request context as the base.
		ctx := r.Context()

		// Expected path: /mcp/u/{uid}/profile/{profileId}/...
		parts := strings.Split(r.URL.Path, "/")
		// parts[0] = ""
		// parts[1] = "mcp"
		// parts[2] = "u"
		// parts[3] = {uid}
		// parts[4] = "profile"
		// parts[5] = {profileId}
		if len(parts) < 6 || parts[4] != "profile" {
			http.NotFound(w, r)
			return
		}
		uid := parts[3]
		profileID := parts[5]

		// Dynamic User Lookup
		user, ok := a.AuthManager.GetUser(uid)
		// SECURITY: Do NOT return "User not found" yet to prevent user enumeration via status codes.
		// We must attempt authentication first. If authentication fails, we return Unauthorized.
		// If authentication succeeds (e.g. global admin), THEN we can reveal that the user is missing.

		// Authentication Logic with Priority:
		// 1. Profile Auth - REMOVED
		// 2. User Auth
		// 3. Global Auth

		var isAuthenticated bool

		// 2. User Auth
		// Only attempt if user exists and has auth configured
		if ok && user.GetAuthentication() != nil {
			if err := auth.ValidateAuthentication(ctx, user.GetAuthentication(), r); err == nil {
				isAuthenticated = true
			} else {
				// User auth configured but failed.
				// We do NOT return immediately, we might fall through to global auth?
				// Typically, if User Auth is present, it is required.
				// However, if we return 401 here immediately, we distinguish between "User exists + bad pass" (401)
				// and "User does not exist" -> Fallthrough -> Global Auth -> "No Global Auth" -> 403/404?
				// To prevent enumeration, we must ensure the behavior is consistent.
				// If we fail here, we return 401.
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// 3. Global Auth (Fallback if User Auth not used or User not found)
		if !isAuthenticated {
			// If User Auth was attempted and failed, we already returned 401 above.
			// So we are here only if User Auth was NOT attempted (User missing OR User has no auth config).

			apiKey := a.SettingsManager.GetAPIKey()
			if apiKey != "" {
				// Manual check for global key
				requestKey := r.Header.Get("X-API-Key")
				if requestKey == "" {
					authHeader := r.Header.Get("Authorization")
					if strings.HasPrefix(authHeader, "Bearer ") {
						requestKey = strings.TrimPrefix(authHeader, "Bearer ")
					}
				}

				if subtle.ConstantTimeCompare([]byte(requestKey), []byte(apiKey)) == 1 {
					isAuthenticated = true
				} else {
					// Global auth configured but failed.
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			} else {
				// No Global Auth configured.
				// If User Auth was NOT attempted (e.g. user missing), and no Global Auth,
				// we fall into "No Auth Configured" logic.

				// Sentinel Security: Enforce private network access if no auth is configured.
				ip := util.GetClientIP(r, trustProxy)
				if !util.IsPrivateIP(net.ParseIP(ip)) {
					logging.GetLogger().Warn("Blocked public internet request to /mcp/u/ because no API Key is configured", "remote_addr", r.RemoteAddr, "client_ip", ip)
					// We return Unauthorized (401) instead of Forbidden (403) to match the User Auth failure case,
					// reducing information leakage about user existence vs configuration state.
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
				isAuthenticated = true
			}
		}

		if !isAuthenticated {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Authentication passed. Now we can check if user exists.
		if !ok {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Profile access check
		hasAccess := false
		for _, pid := range user.GetProfileIds() {
			if pid == profileID {
				hasAccess = true
				break
			}
		}
		if !hasAccess {
			http.Error(w, "Forbidden: User does not have access to this profile", http.StatusForbidden)
			return
		}

		// RBAC Check: Check if profile requires specific roles
		// Dynamic Profile Lookup
		if def, ok := a.ProfileManager.GetProfileDefinition(profileID); ok && len(def.GetRequiredRoles()) > 0 {
			hasRole := false
			// Check if user has any of the required roles
			for _, requiredRole := range def.GetRequiredRoles() {
				for _, userRole := range user.GetRoles() {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}
			if !hasRole {
				// Don't leak required roles to the client
				logging.GetLogger().Warn("Forbidden access to profile", "profile", profileID, "required_roles", def.GetRequiredRoles(), "user_roles", user.GetRoles())
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
		}

		// Inject context
		// We use r.Context() to ensure we start fresh, discarding any partial auth context from above.
		// Note: We assign to the shadowed ctx variable.
		ctx = auth.ContextWithUser(ctx, uid)
		ctx = auth.ContextWithProfileID(ctx, profileID)
		ctx = auth.ContextWithRoles(ctx, user.GetRoles())

		// Strip the prefix so the underlying handler sees the relative path
		prefix := fmt.Sprintf("/mcp/u/%s/profile/%s", uid, profileID)
		delegate := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size to 5MB to prevent DoS attacks via large payloads.
			// This applies to both the stateless JSON-RPC handler and the underlying MCP handler.
			r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

			logging.GetLogger().Info("Delegate Handler", "method", r.Method, "path", r.URL.Path)
			// Support stateless JSON-RPC for simple clients
			if r.Method == http.MethodPost && (r.URL.Path == "/" || r.URL.Path == "") {
				var req struct {
					JSONRPC string          `json:"jsonrpc"`
					ID      any             `json:"id"`
					Method  string          `json:"method"`
					Params  json.RawMessage `json:"params"`
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					// http.MaxBytesReader returns an error if the limit is exceeded.
					// We should log it and return an appropriate error.
					logging.GetLogger().Error("Failed to read request body", "error", err)
					if strings.Contains(err.Error(), "request body too large") {
						http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
						return
					}
					http.Error(w, "Failed to read request body", http.StatusInternalServerError)
					return
				}

				r.Body = io.NopCloser(bytes.NewBuffer(body)) // Restore body just in case
				if err := json.Unmarshal(body, &req); err != nil {
					http.Error(w, "Invalid JSON", http.StatusBadRequest)
					return
				}

				if req.Method == "tools/list" {
					tools := mcpSrv.ListTools()
					var responseTools []map[string]any
					for _, t := range tools {
						v1Tool := t.Tool()
						serviceID := v1Tool.GetServiceId()
						_, ok := mcpSrv.GetServiceInfo(serviceID)
						if !ok {
							continue
						}

						// Check profiles
						if profileID != "" {
							if !mcpSrv.ToolManager().IsServiceAllowed(serviceID, profileID) {
								continue
							}
						}

						responseTools = append(responseTools, map[string]any{
							"name":        v1Tool.GetName(),
							"description": v1Tool.GetDescription(),
						})
					}

					// Ensure we return an empty list if no tools are found/allowed, not nil?
					// JSON encoding nil slice as null is usually fine, but empty list [] is better for clients.
					if responseTools == nil {
						responseTools = []map[string]any{}
					}

					resp := map[string]any{
						"jsonrpc": "2.0",
						"id":      req.ID,
						"result": map[string]any{
							"tools": responseTools,
						},
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(resp)
					return
				}

				// Add logging to see unsupported methods
				logging.GetLogger().Info("Unsupported stateless method", "method", req.Method)
				http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
				return
			}
			httpHandler.ServeHTTP(w, r)
		})
		http.StripPrefix(prefix, delegate).ServeHTTP(w, r.WithContext(ctx))
	})

	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "OK")
	})
	mux.Handle("/healthz", healthHandler)
	mux.Handle("/health", healthHandler)
	mux.Handle("/metrics", authMiddleware(metrics.Handler()))
	mux.Handle("/api/v1/alignment/status", authMiddleware(a.handleActiveIntentAlignment()))
	mux.Handle("/upload", authMiddleware(http.HandlerFunc(a.uploadFile)))

	// OIDC Routes
	var oidcConfig *config_v1.OIDCConfig
	if globalSettings != nil {
		oidcConfig = globalSettings.GetOidc()
	} else {
		oidcConfig = config.GlobalSettings().GetOidc()
	}

	if oidcConfig != nil {
		provider, err := auth.NewOIDCProvider(localCtx, auth.OIDCConfig{
			Issuer:       oidcConfig.GetIssuer(),
			ClientID:     oidcConfig.GetClientId(),
			ClientSecret: oidcConfig.GetClientSecret(),
			RedirectURL:  oidcConfig.GetRedirectUrl(),
		})
		if err != nil {
			logging.GetLogger().Error("Failed to initialize OIDC provider", "error", err)
		} else {
			mux.HandleFunc("/auth/login", provider.HandleLogin)
			mux.HandleFunc("/auth/callback", provider.HandleCallback)
		}
	}

	// OAuth API Routes
	mux.Handle("/auth/oauth/initiate", authMiddleware(http.HandlerFunc(a.handleInitiateOAuth)))
	mux.Handle("/auth/oauth/callback", authMiddleware(http.HandlerFunc(a.handleOAuthCallback)))

	// Secrets API
	mux.Handle("/api/v1/secrets", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			a.listSecretsHandler(w, r)
		case http.MethodPost:
			a.createSecretHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/api/v1/secrets/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for /reveal
		if strings.HasSuffix(r.URL.Path, "/reveal") {
			a.revealSecretHandler(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			a.getSecretHandler(w, r)
		case http.MethodDelete:
			a.deleteSecretHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Credentials API
	// Note: Standard mux doesn't handle methods nicely, so we route by path and check method in handler.
	// We route /credentials to list (GET) and create (POST)
	// We route /credentials/ to get/update/delete (with ID)
	mux.Handle("/credentials", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			a.listCredentialsHandler(w, r)
		case http.MethodPost:
			a.createCredentialHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	// mux.Handle("/api/v1/skills", authMiddleware(a.handleListSkills())) // Replaced by gRPC Gateway
	// mux.Handle("/api/v1/skills/create", authMiddleware(a.handleCreateSkill())) // Replaced by gRPC Gateway

	// Register Config Validation Endpoint
	mux.Handle("/api/v1/config/validate", authMiddleware(http.HandlerFunc(rest.ValidateConfigHandler)))

	// Asset upload is handled later in the gRPC gateway block to support fallback

	// Wait, we need to handle assets specifically.
	// Let's use a more specific path for assets if possible, or ensure we fallback to gwmux?
	// Mux doesn't fallback easily.
	// Better: Register /v1/skills/{name}/assets manual handler if possible?
	// ServeMux doesn't support wildcards.
	// So we must handle /v1/skills/ and forward non-asset requests?
	// But `gwmux` is NOT a simple handler we can call easily from here without re-entering the stack.
	// ACTUALLY: gwmux is served via `mux.Handle("/v1/", ...)`
	// If I register `mux.Handle("/v1/skills/", ...)` it takes precedence.
	// So I MUST handle standard skill requests here too if I do that.
	// OR I can use a different prefix for assets? No, API spec.
	// OR I pass standard requests to `gwmux`?
	// `gwmux.ServeHTTP(w, r)`!
	// Yes, I can use `gwmux` as fallback.

	mux.Handle("/credentials/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This handles /credentials/:id
		switch r.Method {
		case http.MethodGet:
			a.getCredentialHandler(w, r)
		case http.MethodPut:
			a.updateCredentialHandler(w, r)
		case http.MethodDelete:
			a.deleteCredentialHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))
	mux.Handle("/debug/auth-test", authMiddleware(http.HandlerFunc(a.testAuthHandler)))
	mux.Handle("/api/v1/debug/seed_traffic", authMiddleware(a.handleDebugSeedTraffic()))
	mux.Handle("/api/v1/debug/seed", authMiddleware(a.handleDebugSeed()))
	mux.Handle("/api/v1/debug/traces", authMiddleware(a.handleDebugSeedTraces()))

	// User Preferences
	mux.Handle("/api/v1/user/preferences", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			a.handleGetUserPreferences(w, r)
		case http.MethodPost:
			a.handleUpdateUserPreferences(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Register Debugger API if enabled
	if standardMiddlewares != nil && standardMiddlewares.Debugger != nil {
		mux.Handle("/debug/entries", authMiddleware(standardMiddlewares.Debugger.APIHandler()))
	}

	// Register Recursive Context Manager
	if standardMiddlewares != nil && standardMiddlewares.RecursiveContext == nil {
		standardMiddlewares.RecursiveContext = middleware.NewRecursiveContextManager()
	}
	if standardMiddlewares != nil {
		mux.Handle("/context/session", authMiddleware(standardMiddlewares.RecursiveContext.APIHandler()))
		mux.Handle("/context/session/", authMiddleware(standardMiddlewares.RecursiveContext.APIHandler()))
	}

	httpBindAddress := bindAddress
	if httpBindAddress == "" {
		if envAddr := os.Getenv("MCPANY_DEFAULT_HTTP_ADDR"); envAddr != "" {
			httpBindAddress = envAddr
		} else {
			httpBindAddress = ":8070"
		}
	} else if !strings.Contains(httpBindAddress, ":") {
		httpBindAddress = ":" + httpBindAddress
	}

	// Apply Global Rate Limit: 20 RPS with a burst of 50.
	// This helps prevent basic DoS attacks on all HTTP endpoints, including /upload.
	// We enable trustProxy if MCPANY_TRUST_PROXY is set, to handle load balancers correctly.
	// trustProxy is already defined above
	rateLimiter := middleware.NewHTTPRateLimitMiddleware(20, 50, middleware.WithTrustProxy(trustProxy))

	// Apply CORS Middleware
	corsMiddleware := middleware.NewHTTPCORSMiddleware(a.SettingsManager.GetAllowedOrigins())
	a.configMu.Lock()
	a.corsMiddleware = corsMiddleware
	a.configMu.Unlock()

	// Apply CSRF Middleware
	csrfMiddleware := middleware.NewCSRFMiddleware(a.SettingsManager.GetAllowedOrigins())
	a.configMu.Lock()
	a.csrfMiddleware = csrfMiddleware
	a.configMu.Unlock()

	// Prepare final handler (Mux wrapped with Content Optimizer and Debugger)
	var finalHandler http.Handler = mux

	if standardMiddlewares != nil {
		// Context Optimizer (inner)
		if standardMiddlewares.ContextOptimizer != nil {
			finalHandler = standardMiddlewares.ContextOptimizer.Handler(finalHandler)
		}
		// Debugger (outer to capture optimized response)
		if standardMiddlewares.Debugger != nil {
			finalHandler = standardMiddlewares.Debugger.Handler(finalHandler)
		}
		// Recursive Context
		if standardMiddlewares.RecursiveContext != nil {
			finalHandler = standardMiddlewares.RecursiveContext.HandleContext(finalHandler)
		}
	}

	// Middleware order: SecurityHeaders -> CORS -> CSRF -> JSONRPCCompliance -> Recovery -> IPAllowList -> RateLimit -> (Debugger -> Optimizer -> Mux)
	// We wrap everything with a debug logger to see what's coming in
	handler := middleware.HTTPSecurityHeadersMiddleware(
		corsMiddleware.Handler(
			csrfMiddleware.Handler(
				middleware.JSONRPCComplianceMiddleware(
					middleware.RecoveryMiddleware(
						a.HTTPRequestContextMiddleware(
							ipMiddleware.Handler(
								rateLimiter.Handler(finalHandler),
							),
						),
					),
				),
			),
		),
	)

	// gRPC Server Setup
	var grpcServer *gogrpc.Server
	var wrappedGrpc *grpcweb.WrappedGrpcServer

	grpcBindAddress := grpcPort

	// Initialize gRPC Interceptors
	grpcUnaryInterceptor := func(ctx context.Context, req interface{}, _ *gogrpc.UnaryServerInfo, handler gogrpc.UnaryHandler) (interface{}, error) {
		if p, ok := peer.FromContext(ctx); ok {
			ip := util.ExtractIP(p.Addr.String())
			ctx = util.ContextWithRemoteIP(ctx, ip)

			if !ipMiddleware.Allow(p.Addr.String()) {
				return nil, status.Error(codes.PermissionDenied, "IP not allowed")
			}
		}
		return handler(ctx, req)
	}
	grpcStreamInterceptor := func(srv interface{}, ss gogrpc.ServerStream, _ *gogrpc.StreamServerInfo, handler gogrpc.StreamHandler) error {
		if p, ok := peer.FromContext(ss.Context()); ok {
			ip := util.ExtractIP(p.Addr.String())
			// Wrapper to modify context for stream
			wrappedStream := &util.WrappedServerStream{
				ServerStream: ss,
				Ctx:          util.ContextWithRemoteIP(ss.Context(), ip),
			}
			if !ipMiddleware.Allow(p.Addr.String()) {
				return status.Error(codes.PermissionDenied, "IP not allowed")
			}
			return handler(srv, wrappedStream)
		}
		return handler(srv, ss)
	}
	grpcOpts := []gogrpc.ServerOption{
		gogrpc.UnaryInterceptor(grpcUnaryInterceptor),
		gogrpc.StreamInterceptor(grpcStreamInterceptor),
		gogrpc.StatsHandler(&metrics.GrpcStatsHandler{Wrapped: otelgrpc.NewServerHandler()}),
	}

	grpcServer = gogrpc.NewServer(grpcOpts...)
	reflection.Register(grpcServer)

	// Register Services
	registrationServer, err := mcpserver.NewRegistrationServer(bus, a.AuthManager)
	if err != nil {
		return fmt.Errorf("failed to create API server: %w", err)
	}
	v1.RegisterRegistrationServiceServer(grpcServer, registrationServer)

	adminServer := admin.NewServer(cachingMiddleware, a.ToolManager, serviceRegistry, store, a.DiscoveryManager, a.GetAuditMiddleware)
	pb_admin.RegisterAdminServiceServer(grpcServer, adminServer)

	// Register Skill Service
	v1.RegisterSkillServiceServer(grpcServer, NewSkillServiceServer(a.SkillManager))

	// Initialize gRPC-Web wrapper even if gRPC port is not exposed
	wrappedGrpc = grpcweb.WrapServer(grpcServer,
		grpcweb.WithOriginFunc(func(_ string) bool { return true }),
		grpcweb.WithAllowedRequestHeaders([]string{"*"}),
	)

	if grpcBindAddress != "" {
		if !strings.Contains(grpcBindAddress, ":") {
			grpcBindAddress = ":" + grpcBindAddress
		}
		lis, err := util.ListenWithRetry(ctx, "tcp", grpcBindAddress)
		if err != nil {
			errChan <- wrapBindError(err, "gRPC", grpcBindAddress, "--grpc-port")
		} else {
			if addr, ok := lis.Addr().(*net.TCPAddr); ok {
				a.BoundGRPCPort.Store(int32(addr.Port)) //nolint:gosec // Port fits in int32

				// Register gRPC Gateway with the bound port
				gwmux := runtime.NewServeMux()
				opts := []gogrpc.DialOption{gogrpc.WithTransportCredentials(insecure.NewCredentials())}
				endpoint := fmt.Sprintf("127.0.0.1:%d", a.BoundGRPCPort.Load())

				if err := v1.RegisterRegistrationServiceHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
					errChan <- fmt.Errorf("failed to register gateway: %w", err)
				} else if err := v1.RegisterSkillServiceHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
					errChan <- fmt.Errorf("failed to register skill gateway: %w", err)
				} else if err := pb_admin.RegisterAdminServiceHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
					errChan <- fmt.Errorf("failed to register admin gateway: %w", err)
				} else {
					// Consolidated handler for /v1/ to support both gRPC Gateway and Asset Uploads
					mux.Handle("/v1/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if strings.HasSuffix(r.URL.Path, "/assets") {
							a.handleUploadSkillAsset()(w, r)
						} else {
							gwmux.ServeHTTP(w, r)
						}
					})))
				}
			}
			expectedReady++
			startGrpcServer(
				localCtx,
				&wg,
				errChan,
				readyChan,
				"Registration",
				lis,
				shutdownTimeout,
				grpcServer,
			)
		}
	}

	// Register Root Handler with gRPC-Web support
	mux.Handle("/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc != nil && wrappedGrpc.IsGrpcWebRequest(r) {
			wrappedGrpc.ServeHTTP(w, r)
			return
		}

		// UI Routing for root path
		if r.URL.Path == "/" && uiPath != "" {
			http.ServeFile(w, r, filepath.Join(uiPath, "index.html"))
			return
		}

		// Fallback to JSON-RPC handler (for API calls at root or SSE)
		httpHandler.ServeHTTP(w, r)
	})))

	var httpLis net.Listener

	if tlsCert != "" && tlsKey != "" {
		cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
		if err != nil {
			return fmt.Errorf("failed to load TLS key pair: %w", err)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}
		if tlsClientCA != "" {
			caCert, err := os.ReadFile(tlsClientCA)
			if err != nil {
				return fmt.Errorf("failed to read TLS client CA: %w", err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.ClientCAs = caCertPool
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}

		logging.GetLogger().Info("Enabling TLS for HTTP server", "mtls_enabled", tlsConfig.ClientAuth == tls.RequireAndVerifyClientCert)

		// Use standard Listen and then wrap with TLS
		l, err := util.ListenWithRetry(ctx, "tcp", httpBindAddress)
		if err != nil {
			// Handle error
			errChan <- wrapBindError(err, "HTTP", httpBindAddress, "--json-rpc-port")
		} else {
			httpLis = tls.NewListener(l, tlsConfig)
		}
	} else {
		l, err := util.ListenWithRetry(ctx, "tcp", httpBindAddress)
		if err != nil {
			// Handle error
			errChan <- wrapBindError(err, "HTTP", httpBindAddress, "--json-rpc-port")
		} else {
			httpLis = l
		}
	}

	if httpLis != nil {
		if addr, ok := httpLis.Addr().(*net.TCPAddr); ok {
			a.BoundHTTPPort.Store(int32(addr.Port)) //nolint:gosec // Port fits in int32
		}
		expectedReady++
		// Handle active connection tracking
		connState := func(_ net.Conn, state http.ConnState) {
			switch state {
			case http.StateNew:
				atomic.AddInt32(&a.activeConnections, 1)
			case http.StateClosed, http.StateHijacked:
				atomic.AddInt32(&a.activeConnections, -1)
			}
		}

		startHTTPServer(localCtx, &wg, errChan, readyChan, "MCP Any HTTP", httpLis, handler, shutdownTimeout, connState)
	}

	// Wait for servers to be ready
	timeout := time.NewTimer(30 * time.Second) // Reasonable timeout for binding ports, increased for slow CI
	defer timeout.Stop()

	for i := 0; i < expectedReady; i++ {
		select {
		case <-readyChan:
			// One server is ready
		case err := <-errChan:
			return fmt.Errorf("failed to start a server: %w", err)
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			return fmt.Errorf("timed out waiting for servers to be ready")
		}
	}

	if startupCallback != nil {
		startupCallback()
	}

	var startupErr error
	select {
	case err := <-errChan:
		startupErr = fmt.Errorf("failed to start a server: %w", err)
		logging.GetLogger().Error("Server startup failed, initiating shutdown...", "error", startupErr)
		// A server failed to start, so we need to trigger a shutdown of any other
		// servers that may have started successfully.
		cancel()
	case <-localCtx.Done():
		logging.GetLogger().Info("Received shutdown signal, shutting down gracefully...")
	}

	// N.B. We wait for the servers to shut down regardless of whether there was a
	// startup error or a shutdown signal.
	logging.GetLogger().Info("Waiting for HTTP and gRPC servers to shut down...")
	wg.Wait()
	logging.GetLogger().Info("All servers have shut down.")

	// Shutdown all upstreams
	if serviceRegistry != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()
		if err := serviceRegistry.Close(shutdownCtx); err != nil {
			logging.GetLogger().Error("Failed to shutdown services", "error", err)
		}
	}

	return startupErr
}
