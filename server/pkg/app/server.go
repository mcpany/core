// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package app provides the main application logic.
package app

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/mcpany/core/pkg/admin"
	"github.com/mcpany/core/pkg/appconsts"
	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/config"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/mcpserver"
	"github.com/mcpany/core/pkg/metrics"
	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/serviceregistry"
	"github.com/mcpany/core/pkg/storage/sqlite"
	"github.com/mcpany/core/pkg/telemetry"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	"github.com/mcpany/core/pkg/upstream/factory"
	"github.com/mcpany/core/pkg/worker"
	pb_admin "github.com/mcpany/core/proto/admin/v1"
	v1 "github.com/mcpany/core/proto/api/v1"
	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var healthCheckClient = &http.Client{
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func (a *Application) uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit the request body size to 10MB to prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed to get file from form", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	// Create a temporary file to store the uploaded content
	tmpfile, err := os.CreateTemp("", "upload-*.txt")
	if err != nil {
		http.Error(w, "failed to create temporary file", http.StatusInternalServerError)
		return
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }() // clean up

	// Copy the uploaded file to the temporary file
	if _, err := io.Copy(tmpfile, file); err != nil {
		http.Error(w, "failed to copy file", http.StatusInternalServerError)
		return
	}

	// Respond with the file name and size
	// Sanitize the filename to prevent reflected XSS
	w.Header().Set("Content-Type", "text/plain")
	_, _ = fmt.Fprintf(w, "File '%s' uploaded successfully (size: %d bytes)", html.EscapeString(header.Filename), header.Size)
}

// Runner defines the interface for running the MCP Any application. It abstracts
// the application's entry point, allowing for different implementations or mocks
// for testing purposes.
type Runner interface {
	// Run starts the MCP Any application with the given context, filesystem, and
	// configuration. It is the primary entry point for the server.
	//
	// ctx is the context that controls the application's lifecycle.
	// fs is the filesystem interface for reading configurations.
	// stdio specifies whether to run in standard I/O mode.
	// jsonrpcPort is the port for the JSON-RPC server.
	// grpcPort is the port for the gRPC registration server.
	// configPaths is a slice of paths to configuration files.
	//
	// It returns an error if the application fails to start or run.
	Run(
		ctx context.Context,
		fs afero.Fs,
		stdio bool,
		jsonrpcPort, grpcPort string,
		configPaths []string,
		shutdownTimeout time.Duration,
	) error

	// ReloadConfig reloads the application configuration from the provided file system
	// and paths. It updates the internal state of the application, such as
	// service registries and managers, to reflect changes in the configuration files.
	//
	// Parameters:
	//   - fs: The filesystem interface for reading configuration files.
	//   - configPaths: A slice of paths to configuration files to reload.
	//
	// Returns:
	//   - An error if the configuration reload fails.
	ReloadConfig(fs afero.Fs, configPaths []string) error
}

// Application is the main application struct, holding the dependencies and
// logic for the MCP Any server. It encapsulates the components required to run
// the server, such as the stdio mode handler, and provides the main `Run`
// method that starts the application.
type Application struct {
	runStdioModeFunc func(ctx context.Context, mcpSrv *mcpserver.Server) error
	PromptManager    prompt.ManagerInterface
	ToolManager      tool.ManagerInterface
	ResourceManager  resource.ManagerInterface
	UpstreamFactory  factory.Factory
	configFiles      map[string]string
	fs               afero.Fs
	configPaths      []string
}

// NewApplication creates a new Application with default dependencies.
// It initializes the application with the standard implementation of the stdio
// mode runner, making it ready to be configured and started.
//
// Returns a new instance of the Application, ready to be run.
func NewApplication() *Application {
	busProvider, _ := bus.NewProvider(nil)
	return &Application{
		runStdioModeFunc: runStdioMode,
		PromptManager:    prompt.NewManager(),
		ToolManager:      tool.NewManager(busProvider),

		ResourceManager: resource.NewManager(),
		UpstreamFactory: factory.NewUpstreamServiceFactory(pool.NewManager()),
		configFiles:     make(map[string]string),
	}
}

// Run starts the MCP Any server and all its components. It initializes the core
// services, loads configurations from the provided paths, starts background
// workers for handling service registration and upstream service communication,
// and launches the gRPC and JSON-RPC servers.
//
// The server's lifecycle is managed by the provided context. A graceful
// shutdown is initiated when the context is canceled.
//
// Parameters:
//   - ctx: The context for managing the application's lifecycle.
//   - fs: The filesystem interface for reading configuration files.
//   - stdio: A boolean indicating whether to run in standard I/O mode.
//   - jsonrpcPort: The port for the JSON-RPC server.
//   - grpcPort: The port for the gRPC registration server. An empty string
//     disables the gRPC server.
//   - configPaths: A slice of paths to service configuration files.
//   - shutdownTimeout: The duration to wait for a graceful shutdown before
//     forcing termination.
//
// Returns an error if any part of the startup or execution fails.
func (a *Application) Run(
	ctx context.Context,
	fs afero.Fs,
	stdio bool,
	jsonrpcPort, grpcPort string,
	configPaths []string,
	shutdownTimeout time.Duration,
) error {
	log := logging.GetLogger()
	fs, err := setup(fs)
	if err != nil {
		return fmt.Errorf("failed to setup filesystem: %w", err)
	}
	a.fs = fs
	a.configPaths = configPaths

	shutdownTracer, err := telemetry.InitTracer(ctx, appconsts.Name, appconsts.Version, os.Stderr)
	if err != nil {
		log.Error("Failed to initialize tracer", "error", err)
	} else {
		defer func() {
			if err := shutdownTracer(context.Background()); err != nil {
				log.Error("Failed to shutdown tracer", "error", err)
			}
		}()
	}

	log.Info("Starting MCP Any Service...")

	// Load initial services from config files and SQLite
	dbPath := config.GlobalSettings().DBPath()
	if dbPath == "" {
		dbPath = "mcpany.db"
	}
	sqliteDB, err := sqlite.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize sqlite db: %w", err)
	}
	defer func() { _ = sqliteDB.Close() }()
	sqliteStore := sqlite.NewStore(sqliteDB)

	var stores []config.Store
	if len(configPaths) > 0 {
		stores = append(stores, config.NewFileStore(fs, configPaths))
	}
	stores = append(stores, sqliteStore)
	multiStore := config.NewMultiStore(stores...)

	var cfg *config_v1.McpAnyServerConfig
	cfg, err = config.LoadServices(multiStore, "server")
	if err != nil {
		return fmt.Errorf("failed to load services from config: %w", err)
	}
	if cfg == nil {
		cfg = &config_v1.McpAnyServerConfig{}
	}

	busConfig := cfg.GetGlobalSettings().GetMessageBus()
	busProvider, err := bus.NewProvider(busConfig)
	if err != nil {
		return fmt.Errorf("failed to create bus provider: %w", err)
	}
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager)
	a.ToolManager = tool.NewManager(busProvider)
	// Add Tool Metrics Middleware
	a.ToolManager.AddMiddleware(middleware.NewToolMetricsMiddleware())

	a.PromptManager = prompt.NewManager()
	a.ResourceManager = resource.NewManager()
	authManager := auth.NewManager()
	if cfg.GetGlobalSettings().GetApiKey() != "" {
		authManager.SetAPIKey(cfg.GetGlobalSettings().GetApiKey())
	}

	// Set profiles for tool filtering
	a.ToolManager.SetProfiles(
		cfg.GetGlobalSettings().GetProfiles(),
		cfg.GetGlobalSettings().GetProfileDefinitions(),
	)

	serviceRegistry := serviceregistry.New(
		upstreamFactory,
		a.ToolManager,
		a.PromptManager,
		a.ResourceManager,
		authManager,
	)

	// New message bus and workers
	upstreamWorker := worker.NewUpstreamWorker(busProvider, a.ToolManager)
	registrationWorker := worker.NewServiceRegistrationWorker(busProvider, serviceRegistry)

	// Start background workers
	upstreamWorker.Start(ctx)
	registrationWorker.Start(ctx)

	// If we're using an in-memory bus, start the in-process worker
	if busConfig == nil || busConfig.GetInMemory() != nil {
		workerCfg := &worker.Config{
			MaxWorkers:   10,
			MaxQueueSize: 100,
		}
		inProcessWorker := worker.New(busProvider, workerCfg)
		inProcessWorker.Start(ctx)
		defer inProcessWorker.Stop()
	}

	// Initialize servers with the message bus
	mcpSrv, err := mcpserver.NewServer(
		ctx,
		a.ToolManager,
		a.PromptManager,
		a.ResourceManager,
		authManager,
		serviceRegistry,
		busProvider,
		config.GlobalSettings().IsDebug(),
	)
	if err != nil {
		return fmt.Errorf("failed to create mcp server: %w", err)
	}

	mcpSrv.SetReloadFunc(func() error {
		return a.ReloadConfig(fs, configPaths)
	})

	a.ToolManager.SetMCPServer(mcpSrv)

	if cfg.GetUpstreamServices() != nil {
		// Publish registration requests to the bus for each service
		registrationBus, err := bus.GetBus[*bus.ServiceRegistrationRequest](
			busProvider,
			"service_registration_requests",
		)
		if err != nil {
			return fmt.Errorf("failed to get registration bus: %w", err)
		}
		for _, serviceConfig := range cfg.GetUpstreamServices() {
			if serviceConfig.GetDisable() {
				log.Info("Skipping disabled service", "service", serviceConfig.GetName())
				continue
			}
			log.Info(
				"Queueing service for registration from config",
				"service",
				serviceConfig.GetName(),
			)
			regReq := &bus.ServiceRegistrationRequest{Config: serviceConfig}
			// We don't need a correlation ID since we are not waiting for a response here
			if err := registrationBus.Publish(ctx, "request", regReq); err != nil {
				log.Error("Failed to publish registration request", "error", err)
			}
		}
	} else {
		log.Info("No services found in config, skipping service registration.")
	}

	// Initialize standard middlewares in registry
	cachingMiddleware := middleware.NewCachingMiddleware(a.ToolManager)
	auditCleanup, err := middleware.InitStandardMiddlewares(mcpSrv.AuthManager(), a.ToolManager, cfg.GetGlobalSettings().GetAudit(), cachingMiddleware)
	if err != nil {
		return fmt.Errorf("failed to init standard middlewares: %w", err)
	}
	if auditCleanup != nil {
		defer func() {
			if err := auditCleanup(); err != nil {
				log.Error("Failed to close audit middleware", "error", err)
			}
		}()
	}

	// Get configured middlewares
	middlewares := config.GlobalSettings().Middlewares()
	if len(middlewares) == 0 {
		// Default chain if none configured
		middlewares = []*config_v1.Middleware{
			{Name: proto.String("debug"), Priority: proto.Int32(10)},
			{Name: proto.String("auth"), Priority: proto.Int32(20)},
			{Name: proto.String("logging"), Priority: proto.Int32(30)},
			{Name: proto.String("audit"), Priority: proto.Int32(40)},
			{Name: proto.String("call_policy"), Priority: proto.Int32(50)},
			{Name: proto.String("caching"), Priority: proto.Int32(60)},
			{Name: proto.String("ratelimit"), Priority: proto.Int32(70)},
			// CORS is typically 0 or negative to be outermost, but AddReceivingMiddleware adds in order.
			// The SDK executes them in reverse order of addition?
			// Wait, mcp.Server implementation:
			// "Middleware is called in the order it was added." -> First added = First called?
			// Usually middleware "wraps" the handler. first(second(handler)).
			// If I add A then B.
			// Chain = A(B(handler)).
			// Helper `AddReceivingMiddleware` usually appends.
			// Let's assume standard "wrap" logic.
			// We want CORS outer.
			{Name: proto.String("cors"), Priority: proto.Int32(0)},
		}
	}

	// Apply middlewares
	// Registry returns sorted list based on priority (low to high).
	// If priority 0 is first, it wraps the rest?
	// If we iterate:
	// M1(M2(M3(...)))
	// M1 is priority 0.
	chain := middleware.GetMCPMiddlewares(middlewares)
	for _, m := range chain {
		mcpSrv.Server().AddReceivingMiddleware(m)
	}

	if stdio {
		return a.runStdioModeFunc(ctx, mcpSrv)
	}

	bindAddress := jsonrpcPort
	if cfg.GetGlobalSettings().GetMcpListenAddress() != "" {
		bindAddress = cfg.GetGlobalSettings().GetMcpListenAddress()
	}

	var allowedIPs []string
	if cfg.GetGlobalSettings() != nil {
		allowedIPs = cfg.GetGlobalSettings().GetAllowedIps()
	}

	maxBodySize := cfg.GetGlobalSettings().GetMaxRequestBodySize()
	if maxBodySize <= 0 {
		maxBodySize = 5 << 20 // 5MB default
	}

	return a.runServerMode(ctx, mcpSrv, busProvider, bindAddress, grpcPort, shutdownTimeout, cfg.GetUsers(), cfg.GetGlobalSettings().GetProfileDefinitions(), allowedIPs, cachingMiddleware, sqliteStore, maxBodySize)
}

// ReloadConfig reloads the configuration from the given paths and updates the
// services.
func (a *Application) ReloadConfig(fs afero.Fs, configPaths []string) error {
	log := logging.GetLogger()
	log.Info("Reloading configuration...")
	metrics.IncrCounter([]string{"config", "reload", "total"}, 1)
	store := config.NewFileStore(fs, configPaths)
	cfg, err := config.LoadServices(store, "server")
	if err != nil {
		metrics.IncrCounter([]string{"config", "reload", "errors"}, 1)
		return fmt.Errorf("failed to load services from config: %w", err)
	}

	// Update profiles on reload
	a.ToolManager.SetProfiles(
		cfg.GetGlobalSettings().GetProfiles(),
		cfg.GetGlobalSettings().GetProfileDefinitions(),
	)

	// Clear existing services
	// We iterate over all currently registered services and clear them.
	// This handles cases where a service was removed or renamed in the new config.
	for _, svcInfo := range a.ToolManager.ListServices() {
		serviceID := svcInfo.Name
		a.ToolManager.ClearToolsForService(serviceID)
		a.ResourceManager.ClearResourcesForService(serviceID)
		a.PromptManager.ClearPromptsForService(serviceID)
	}

	if cfg.GetUpstreamServices() != nil {
		for _, serviceConfig := range cfg.GetUpstreamServices() {
			if serviceConfig.GetDisable() {
				log.Info("Skipping disabled service", "service", serviceConfig.GetName())
				continue
			}

			// Reload tools, prompts, and resources
			upstream, err := a.UpstreamFactory.NewUpstream(serviceConfig)
			if err != nil {
				log.Error("Failed to get upstream service", "error", err)
				continue
			}
			if upstream != nil {
				_, _, _, err = upstream.Register(context.Background(), serviceConfig, a.ToolManager, a.PromptManager, a.ResourceManager, false)
				if err != nil {
					log.Error("Failed to register upstream service", "error", err)
					continue
				}
			}
		}
	}
	log.Info("Tools", "tools", a.ToolManager.ListTools())
	log.Info("Prompts", "prompts", a.PromptManager.ListPrompts())
	log.Info("Resources", "resources", a.ResourceManager.ListResources())
	return nil
}

// setup initializes the filesystem for the server. It ensures that a valid
// afero.Fs is available, returning an error if a nil filesystem is provided.
//
// fs is the filesystem to be validated.
//
// It returns a non-nil afero.Fs or an error if the provided filesystem is nil.
func setup(fs afero.Fs) (afero.Fs, error) {
	log := logging.GetLogger()
	if fs == nil {
		log.Error(
			"setup called with nil afero.Fs. This is not allowed; an explicit afero.Fs must be provided.",
		)
		return nil, fmt.Errorf("filesystem not provided")
	}
	return fs, nil
}

// runStdioMode starts the server in standard I/O mode, which is useful for
// debugging and simple, single-client scenarios. It uses the standard input
// and output as the transport layer.
//
// ctx is the context for managing the server's lifecycle.
// mcpSrv is the MCP server instance to run.
//
// It returns an error if the server fails to run in stdio mode.
func runStdioMode(ctx context.Context, mcpSrv *mcpserver.Server) error {
	log := logging.GetLogger()
	log.Info("Starting in stdio mode")
	return mcpSrv.Server().Run(ctx, &mcp.StdioTransport{})
}

// HealthCheck performs a health check against a running server by sending an
// HTTP GET request to its /healthz endpoint. This is useful for monitoring and
// ensuring the server is operational.
//
// The function constructs the health check URL from the provided address and
// sends an HTTP GET request. It expects a 200 OK status code for a successful
// health check.
//
// Parameters:
//   - out: The writer to which the success message will be written.
//   - addr: The address (host:port) on which the server is running.
//
// Returns nil if the server is healthy (i.e., responds with a 200 OK), or an
// error if the health check fails for any reason (e.g., connection error,
// non-200 status code).
func HealthCheck(out io.Writer, addr string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return HealthCheckWithContext(ctx, out, addr)
}

// HealthCheckWithContext performs a health check against a running server by
// sending an HTTP GET request to its /healthz endpoint. This is useful for
// monitoring and ensuring the server is operational.
//
// The function constructs the health check URL from the provided address and
// sends an HTTP GET request. It expects a 200 OK status code for a successful
// health check.
//
// Parameters:
//   - ctx: The context for managing the health check's lifecycle.
//   - out: The writer to which the success message will be written.
//   - addr: The address (host:port) on which the server is running.
//
// Returns nil if the server is healthy (i.e., responds with a 200 OK), or an
// error if the health check fails for any reason (e.g., connection error,
// non-200 status code).
func HealthCheckWithContext(
	ctx context.Context,
	out io.Writer,
	addr string,
) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("http://%s/healthz", addr),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request for health check: %w", err)
	}

	resp, err := healthCheckClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// We must read the body and close it to ensure the underlying connection can be reused.
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status code: %d", resp.StatusCode)
	}

	_, _ = fmt.Fprintln(out, "Health check successful: server is running and healthy.")
	return nil
}

// runServerMode runs the server in the standard HTTP and gRPC server mode. It
// starts the HTTP server for JSON-RPC and the gRPC server for service
// registration, and handles graceful shutdown.
//
// ctx is the context for managing the server's lifecycle.
// mcpSrv is the MCP server instance.
// bus is the message bus for inter-component communication.
// jsonrpcPort is the port for the JSON-RPC server.
// grpcPort is the port for the gRPC registration server.
//
// It returns an error if any of the servers fail to start or run.

//
//nolint:gocyclo
func (a *Application) runServerMode(
	ctx context.Context,
	mcpSrv *mcpserver.Server,
	bus *bus.Provider,
	bindAddress, grpcPort string,
	shutdownTimeout time.Duration,
	users []*config_v1.User,
	profileDefinitions []*config_v1.ProfileDefinition,
	allowedIPs []string,
	cachingMiddleware *middleware.CachingMiddleware,
	store *sqlite.Store,
	maxRequestBodySize int64,
) error {
	ipMiddleware, err := middleware.NewIPAllowlistMiddleware(allowedIPs)
	if err != nil {
		return fmt.Errorf("failed to create IP allowlist middleware: %w", err)
	}

	// localCtx is used to manage the lifecycle of the servers started in this function.
	// It's canceled when this function returns, ensuring that all servers are shut down.
	localCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errChan := make(chan error, 2)
	var wg sync.WaitGroup

	httpHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return mcpSrv.Server()
	}, nil)

	apiKey := config.GlobalSettings().APIKey()
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := util.ExtractIP(r.RemoteAddr)
			ctx := util.ContextWithRemoteIP(r.Context(), ip)
			r = r.WithContext(ctx)

			if apiKey != "" {
				if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-API-Key")), []byte(apiKey)) != 1 {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}

	userMap := make(map[string]*config_v1.User)
	for _, u := range users {
		userMap[u.GetId()] = u
	}

	// Build Profile Definition Map for RBAC
	profileDefMap := make(map[string]*config_v1.ProfileDefinition)
	for _, def := range profileDefinitions {
		profileDefMap[def.GetName()] = def
	}

	mux := http.NewServeMux()

	// UI Handler
	uiFS := http.FileServer(http.Dir("./ui"))
	mux.Handle("/ui/", http.StripPrefix("/ui", uiFS))

	mux.Handle("/", authMiddleware(httpHandler))

	// API Routes for Configuration Management
	// Protected by auth middleware
	apiHandler := http.StripPrefix("/api/v1", a.createAPIHandler(store))
	mux.Handle("/api/v1/", authMiddleware(apiHandler))

	logging.GetLogger().Info("DEBUG: Registering /mcp/u/ handler")
	// Multi-user handler
	// pattern: /mcp/u/{uid}/profile/{profile_id}
	// We use a prefix match via stripping.
	// NOTE: We manually handle the path parsing because we support subpaths like /sse or /messages
	mux.HandleFunc("/mcp/u/", func(w http.ResponseWriter, r *http.Request) {
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

		user, ok := userMap[uid]
		if !ok {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Authentication Logic with Priority:
		// 1. Profile Authentication
		// 2. User Authentication
		// 3. Global Authentication
		// If a higher priority mechanism is configured, it MUST be satisfied. Lower priority checks are skipped.

		// Resolve the specific profile config from the user's allowed profiles
		// We need to find the profile config to check if it has an API Key.
		// The user object only has profile IDs.
		// We need to look up the profile definition.
		// Profile definitions are inside UpstreamServiceConfigs.
		// BUT, we are in the multi-user handler which routes to *any* service.
		// Wait, the profile is conceptual here?
		// The routing is /mcp/u/{uid}/profile/{profileId}
		// The `profileId` corresponds to a profile defined in ONE OR MORE upstream services.
		// Does a "Profile" exist independently?
		// In `config.proto`, `UpstreamServiceConfig` has `repeated Profile profiles`.
		// `Profile` has `id` and `api_key`.
		// Since a profile ID can be shared across services (e.g. "prod"), which `api_key` do we use?
		// If multiple services define "prod", do they share the same API key?
		// Assumption: If "prod" is defined in multiple places, they should logically share the key or we pick one.
		// Better approach: We iterate over all services to find the profile definition.

		var profileAuthConfig *config_v1.AuthenticationConfig
		// We need access to the full config or tool manager to look up profiles.
		// mcpSrv.ServiceRegistry() gives us registered services.

		services, err := mcpSrv.ServiceRegistry().GetAllServices()
		if err != nil {
			logging.GetLogger().Error("Failed to list services for profile auth check", "error", err)
		} else {
			for _, svc := range services {
				// svc is *config.UpstreamServiceConfig
				for _, p := range svc.GetProfiles() {
					if p.GetId() == profileID {
						profileAuthConfig = p.GetAuthentication()
						// We found the profile, break inner loop.
						// Should we break outer loop? Yes, assuming profile IDs are unique or we take first match.
						break
					}
				}
				if profileAuthConfig != nil {
					break
				}
			}
		}

		// Authentication Logic with Priority:
		// 1. Profile Authentication
		// 2. User Authentication
		// 3. Global Authentication

		isAuthenticated := false

		// 1. Profile Auth
		if profileAuthConfig != nil {
			if err := auth.ValidateAuthentication(r.Context(), profileAuthConfig, r); err == nil {
				isAuthenticated = true
			} else {
				// Profile auth configured but failed
				http.Error(w, "Unauthorized (Profile)", http.StatusUnauthorized)
				return
			}
		} else {
			// 2. User Auth
			if user.GetAuthentication() != nil {
				if err := auth.ValidateAuthentication(r.Context(), user.GetAuthentication(), r); err == nil {
					isAuthenticated = true
				} else {
					// User auth configured but failed
					http.Error(w, "Unauthorized (User)", http.StatusUnauthorized)
					return
				}
			} else {
				// 3. Global Auth
				// Global Auth is still a simple API Key string in GlobalSettings for now (based on current code).
				// We can continue to use it as is.
				if apiKey != "" {
					// Manual check for global key since it's a string, or wrap it.
					// We'll just do manual check to match existing behavior logic.
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
						// Global auth configured but failed
						http.Error(w, "Unauthorized (Global)", http.StatusUnauthorized)
						return
					}
				} else {
					// No auth configured at any level
					isAuthenticated = true
				}
			}
		}

		if !isAuthenticated {
			// Should be unreachable if logic covers all cases, but safety net
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		if def, ok := profileDefMap[profileID]; ok && len(def.RequiredRoles) > 0 {
			hasRole := false
			// Check if user has any of the required roles
			for _, requiredRole := range def.RequiredRoles {
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
				http.Error(w, fmt.Sprintf("Forbidden: Profile %s requires roles %v", profileID, def.RequiredRoles), http.StatusForbidden)
				return
			}
		}

		// Inject context
		ctx := auth.ContextWithUser(r.Context(), uid)
		ctx = auth.ContextWithProfileID(ctx, profileID)
		ctx = auth.ContextWithRoles(ctx, user.GetRoles())

		// Strip the prefix so the underlying handler sees the relative path
		prefix := fmt.Sprintf("/mcp/u/%s/profile/%s", uid, profileID)
		delegate := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logging.GetLogger().Info("Delegate Handler", "method", r.Method, "path", r.URL.Path)
			// Support stateless JSON-RPC for simple clients
			if handleStatelessJSONRPC(mcpSrv, w, r, maxRequestBodySize) {
				return
			}
			httpHandler.ServeHTTP(w, r)
		})
		http.StripPrefix(prefix, delegate).ServeHTTP(w, r.WithContext(ctx))
	})

	mux.Handle("/healthz", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "OK")
	})))
	mux.Handle("/metrics", authMiddleware(metrics.Handler()))
	mux.Handle("/upload", authMiddleware(http.HandlerFunc(a.uploadFile)))

	if grpcPort != "" {
		gwmux := runtime.NewServeMux()
		opts := []gogrpc.DialOption{gogrpc.WithTransportCredentials(insecure.NewCredentials())}
		endpoint := grpcPort
		if strings.HasPrefix(endpoint, ":") {
			endpoint = "localhost" + endpoint
		}
		if err := v1.RegisterRegistrationServiceHandlerFromEndpoint(ctx, gwmux, endpoint, opts); err != nil {
			return fmt.Errorf("failed to register gateway: %w", err)
		}
		mux.Handle("/v1/", authMiddleware(gwmux))
	}

	httpBindAddress := bindAddress
	if httpBindAddress == "" {
		httpBindAddress = "localhost:8070"
	} else if !strings.Contains(httpBindAddress, ":") {
		httpBindAddress = ":" + httpBindAddress
	}

	startHTTPServer(localCtx, &wg, errChan, "MCP Any HTTP", httpBindAddress, otelhttp.NewHandler(middleware.HTTPSecurityHeadersMiddleware(ipMiddleware.Handler(mux)), "mcp-server"), shutdownTimeout)

	grpcBindAddress := grpcPort
	if grpcBindAddress != "" {
		if !strings.Contains(grpcBindAddress, ":") {
			grpcBindAddress = ":" + grpcBindAddress
		}
		lis, err := net.Listen("tcp", grpcBindAddress)
		if err != nil {
			errChan <- fmt.Errorf("gRPC server failed to listen: %w", err)
		} else {
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
			}

			startGrpcServer(
				localCtx,
				&wg,
				errChan,
				"Registration",
				lis,
				shutdownTimeout,
				grpcOpts,
				func(s *gogrpc.Server) {
					registrationServer, err := mcpserver.NewRegistrationServer(bus)
					if err != nil {
						errChan <- fmt.Errorf("failed to create API server: %w", err)
						return
					}
					v1.RegisterRegistrationServiceServer(s, registrationServer)

					// Register Admin Service
					adminServer := admin.NewServer(cachingMiddleware, a.ToolManager)
					pb_admin.RegisterAdminServiceServer(s, adminServer)

					// config_v1.RegisterMcpAnyConfigServiceServer(s, mcpSrv.ConfigServer())
				},
			)
		}
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

	return startupErr
}

// startHTTPServer starts an HTTP server in a new goroutine. It handles graceful
// shutdown when the context is canceled.
//
// ctx is the context for managing the server's lifecycle.
// wg is a WaitGroup to signal when the server has shut down.
// errChan is a channel for reporting errors during startup.
// name is a descriptive name for the server, used in logging.
// addr is the address and port on which the server will listen.
// handler is the HTTP handler for processing requests.
func startHTTPServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	errChan chan<- error,
	name, addr string,
	handler http.Handler,
	shutdownTimeout time.Duration,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverLog := logging.GetLogger().With("server", name)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			errChan <- fmt.Errorf("[%s] server failed to listen: %w", name, err)
			return
		}
		defer func() { _ = lis.Close() }()

		serverLog = serverLog.With("port", lis.Addr().String())

		server := &http.Server{
			Handler: handler,
			BaseContext: func(_ net.Listener) context.Context {
				return ctx
			},
			ConnState: func(_ net.Conn, state http.ConnState) {
				switch state {
				case http.StateNew:
					metrics.IncrCounter([]string{"http", "connections", "opened", "total"}, 1)
				case http.StateClosed:
					metrics.IncrCounter([]string{"http", "connections", "closed", "total"}, 1)
				}
			},
			ReadHeaderTimeout: 10 * time.Second,
		}

		// localCtx is used to signal the shutdown goroutine to exit.
		localCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		shutdownComplete := make(chan struct{})
		go func() {
			defer close(shutdownComplete)
			select {
			case <-ctx.Done():
				// This is the normal shutdown path.
			case <-localCtx.Done():
				// This is the shutdown path for when the server fails to start.
			}
			shutdownCtx, cancelTimeout := context.WithTimeout(context.Background(), shutdownTimeout)
			defer cancelTimeout()
			serverLog.Info("Attempting to gracefully shut down server...")
			if err := server.Shutdown(shutdownCtx); err != nil {
				serverLog.Error("Shutdown error", "error", err)
			}
		}()

		serverLog.Info("HTTP server listening")
		if err := server.Serve(lis); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("[%s] server failed: %w", name, err)
			cancel() // Signal shutdown goroutine to exit
		}

		<-shutdownComplete
		serverLog.Info("Server shut down.")
	}()
}

func handleStatelessJSONRPC(mcpSrv *mcpserver.Server, w http.ResponseWriter, r *http.Request, maxBodySize int64) bool {
	if !(r.Method == http.MethodPost && (r.URL.Path == "/" || r.URL.Path == "")) {
		return false
	}

	var req struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      any             `json:"id"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
	}

	// Limit request body to configured size to prevent DoS via large payloads
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		if strings.Contains(err.Error(), "too large") {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		}
		return true
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return true
	}

	if req.Method == "tools/list" {
		tools := mcpSrv.ListTools()
		var responseTools []map[string]any
		profileID, _ := auth.ProfileIDFromContext(r.Context())

		for _, t := range tools {
			v1Tool := t.Tool()
			serviceID := v1Tool.GetServiceId()
			info, ok := mcpSrv.GetServiceInfo(serviceID)
			if !ok {
				continue
			}

			// Check profiles
			allowed := false
			if len(info.Config.GetProfiles()) == 0 {
				allowed = true
			} else {
				// Check if current profileID matches any allowed profile
				for _, p := range info.Config.GetProfiles() {
					if p.GetId() == profileID {
						allowed = true
						break
					}
				}
			}

			if !allowed {
				continue
			}

			responseTools = append(responseTools, map[string]any{
				"name":        v1Tool.GetName(),
				"description": v1Tool.GetDescription(),
			})
		}

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
		return true
	}

	// Add logging to see unsupported methods
	logging.GetLogger().Info("Unsupported stateless method", "method", req.Method)
	http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
	return true
}

// startGrpcServer starts a gRPC server in a new goroutine. It handles graceful
// shutdown when the context is canceled.
//
// ctx is the context for managing the server's lifecycle.
// wg is a WaitGroup to signal when the server has shut down.
// errChan is a channel for reporting errors during startup.
// name is a descriptive name for the server, used in logging.
// lis is the net.Listener for the server.
// register is a function that registers the gRPC services with the server.
func startGrpcServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	errChan chan<- error,
	name string,
	lis net.Listener,
	shutdownTimeout time.Duration,
	opts []gogrpc.ServerOption,
	register func(*gogrpc.Server),
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverLog := logging.GetLogger().With("server", name)
		registerSafe := func() *gogrpc.Server {
			defer func() {
				if r := recover(); r != nil {
					serverLog.Error("Panic during gRPC service registration", "panic", r)
					errChan <- fmt.Errorf("[%s] panic during gRPC service registration: %v", name, r)
				}
			}()
			opts = append(opts, gogrpc.StatsHandler(&metrics.GrpcStatsHandler{}))
			grpcServer := gogrpc.NewServer(opts...)
			register(grpcServer)
			reflection.Register(grpcServer)
			return grpcServer
		}

		grpcServer := registerSafe()
		if grpcServer == nil {
			// A panic occurred during registration, and the error has been sent to the channel.
			// We can just return here.
			return
		}

		// localCtx is used to signal the shutdown goroutine to exit.
		localCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		shutdownComplete := make(chan struct{})
		go func() {
			defer close(shutdownComplete)
			select {
			case <-ctx.Done():
				// This is the normal shutdown path.
			case <-localCtx.Done():
				// This is the shutdown path for when the server fails to start.
			}

			serverLog.Info("Attempting to gracefully shut down server...")
			stopped := make(chan struct{})
			go func() {
				defer close(stopped)
				grpcServer.GracefulStop()
			}()

			timer := time.NewTimer(shutdownTimeout)
			defer timer.Stop()
			select {
			case <-stopped:
				// Successful graceful shutdown.
			case <-timer.C:
				// Graceful shutdown timed out.
				serverLog.Warn("Graceful shutdown timed out, forcing stop.")
				grpcServer.Stop()
			}
		}()

		serverLog.Info("gRPC server listening", "port", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil && err != gogrpc.ErrServerStopped {
			errChan <- fmt.Errorf("[%s] server failed to serve: %w", name, err)
			cancel() // Signal shutdown goroutine to exit
		}
		<-shutdownComplete
		serverLog.Info("Server shut down.")
	}()
}
