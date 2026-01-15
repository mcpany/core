// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package app provides the main application logic.
package app

import (
	"bytes"
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
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	pb_admin "github.com/mcpany/core/proto/admin/v1"
	v1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/server/pkg/admin"
	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/gc"
	"github.com/mcpany/core/server/pkg/health"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/mcpserver"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/profile"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/storage/postgres"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/mcpany/core/server/pkg/telemetry"
	"github.com/mcpany/core/server/pkg/tokenizer"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/mcpany/core/server/pkg/worker"
	otelgrpc "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	// config_v1 "github.com/mcpany/core/proto/config/v1".
	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const authMiddlewareName = "auth"

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

	// Clean up any temporary files created by ParseMultipartForm
	if r.MultipartForm != nil {
		defer func() {
			if err := r.MultipartForm.RemoveAll(); err != nil {
				logging.GetLogger().Error("Failed to remove multipart form files", "error", err)
			}
		}()
	}

	// Consume the file content without writing to disk.
	// We discard the content to avoid disk usage and potential residue.
	written, err := io.Copy(io.Discard, file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	// Respond with the file name and size
	// Sanitize the filename to prevent reflected XSS
	w.Header().Set("Content-Type", "text/plain")
	_, _ = fmt.Fprintf(w, "File '%s' uploaded successfully (size: %d bytes)", html.EscapeString(header.Filename), written)
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
	// It returns	// Run starts the application with the given configuration.
	Run(
		ctx context.Context,
		fs afero.Fs,
		stdio bool,
		jsonrpcPort, grpcPort string,
		configPaths []string,
		apiKey string,
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
	//   - ctx: The context for the reload operation.
	//   - fs: The filesystem interface for reading configuration files.
	//   - configPaths: A slice of paths to configuration files to reload.
	//
	// Returns:
	//   - An error if the configuration reload fails.
	ReloadConfig(ctx context.Context, fs afero.Fs, configPaths []string) error
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
	ServiceRegistry  serviceregistry.ServiceRegistryInterface
	TopologyManager  *topology.Manager
	UpstreamFactory  factory.Factory
	configFiles      map[string]string
	fs               afero.Fs
	configPaths      []string
	Storage          storage.Storage
	TemplateManager  *TemplateManager
	busProvider      *bus.Provider

	// Middleware handles for dynamic updates
	standardMiddlewares *middleware.StandardMiddlewares
	// Settings Manager for global settings (dynamic updates)
	SettingsManager *GlobalSettingsManager
	// Profile Manager for dynamic profile updates
	ProfileManager *profile.Manager
	// Auth Manager (stored here for access in runServerMode, though it is also passed to serviceregistry)
	// We need to keep a reference to update it on reload.
	AuthManager *auth.Manager
	// Middlewares that need manual updates
	ipMiddleware   *middleware.IPAllowlistMiddleware
	corsMiddleware *middleware.HTTPCORSMiddleware

	startupCh   chan struct{}
	startupOnce sync.Once
	configMu    sync.Mutex
	// Store explicit API Key passed via CLI args
	explicitAPIKey string

	// lastReloadErr stores the error from the last configuration reload.
	// It is protected by configMu.
	lastReloadErr error
	// lastReloadTime stores the time of the last configuration reload attempt.
	// It is protected by configMu.
	lastReloadTime time.Time
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
		UpstreamFactory: factory.NewUpstreamServiceFactory(pool.NewManager(), nil),
		configFiles:     make(map[string]string),
		startupCh:       make(chan struct{}),
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
//
//nolint:gocyclo // Run is the main entry point and setup function, expected to be complex
func (a *Application) Run(
	ctx context.Context,
	fs afero.Fs,
	stdio bool,
	jsonrpcPort, grpcPort string,
	configPaths []string,
	apiKey string,
	shutdownTimeout time.Duration,
) error {
	log := logging.GetLogger()
	fs, err := setup(fs)
	if err != nil {
		return fmt.Errorf("failed to setup filesystem: %w", err)
	}
	a.fs = fs
	a.configPaths = configPaths
	a.explicitAPIKey = apiKey

	// Telemetry initialization moved after config loading


	log.Info("Starting MCP Any Service...")

	// Load initial services from config files and Storage
	var storageStore config.Store
	var storageCloser func() error

	// Default to SQLite if not specified or explicitly sqlite
	dbDriver := config.GlobalSettings().GetDbDriver()
	switch dbDriver {
	case "", "sqlite":
		dbPath := config.GlobalSettings().DBPath()
		if dbPath == "" {
			dbPath = "mcpany.db"
		}
		sqliteDB, err := sqlite.NewDB(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize sqlite db: %w", err)
		}
		storageCloser = sqliteDB.Close
		storageStore = sqlite.NewStore(sqliteDB)
	case "postgres":
		dsn := config.GlobalSettings().GetDbDsn()
		if dsn == "" {
			return fmt.Errorf("postgres driver selected but db_dsn is empty")
		}
		pgDB, err := postgres.NewDB(dsn)
		if err != nil {
			return fmt.Errorf("failed to initialize postgres db: %w", err)
		}
		storageCloser = func() error { return pgDB.Close() }
		storageStore = postgres.NewStore(pgDB)
	default:
		return fmt.Errorf("unsupported db driver: %s", dbDriver)
	}
	defer func() {
		if storageCloser != nil {
			_ = storageCloser()
		}
	}()

	var stores []config.Store
	if len(configPaths) > 0 {
		// Use strict FileStore to fail fast on configuration errors (Track 1: Friction Fighter)
		stores = append(stores, config.NewFileStore(fs, configPaths))
	}
	stores = append(stores, storageStore)
	multiStore := config.NewMultiStore(stores...)

	var cfg *config_v1.McpAnyServerConfig
	cfg, err = config.LoadServices(ctx, multiStore, "server")
	if err != nil {
		return fmt.Errorf("failed to load services from config: %w", err)
	}
	if cfg == nil {
		cfg = &config_v1.McpAnyServerConfig{}
	}


	// Initialize Telemetry with loaded config
	shutdownTelemetry, err := telemetry.InitTelemetry(ctx, appconsts.Name, appconsts.Version, cfg.GetGlobalSettings().GetTelemetry(), os.Stderr)
	if err != nil {
		// Log error but don't fail startup just for telemetry if we want resilience,
		// but typically we might want to know.
		log.Error("Failed to initialize telemetry", "error", err)
	} else {
		defer func() {
			if err := shutdownTelemetry(context.Background()); err != nil {
				log.Error("Failed to shutdown telemetry", "error", err)
			}
		}()
	}

	// Initialize Settings Manager
	a.SettingsManager = NewGlobalSettingsManager(
		apiKey,
		cfg.GetGlobalSettings().GetAllowedIps(),
		// Logic for origins default moved to inside NewGlobalSettingsManager or updated here
		nil,
	)
	a.SettingsManager.Update(cfg.GetGlobalSettings(), apiKey)

	busConfig := cfg.GetGlobalSettings().GetMessageBus()
	busProvider, err := bus.NewProvider(busConfig)
	if err != nil {
		return fmt.Errorf("failed to create bus provider: %w", err)
	}
	a.busProvider = busProvider

	poolManager := pool.NewManager()
	if gs := cfg.GetGlobalSettings(); gs != nil {
		validation.SetAllowedPaths(gs.GetAllowedFilePaths())
	}
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager, cfg.GetGlobalSettings())
	a.ToolManager = tool.NewManager(busProvider)
	// Add Tool Metrics Middleware
	a.ToolManager.AddMiddleware(middleware.NewToolMetricsMiddleware(tokenizer.NewSimpleTokenizer()))
	// Add Resilience Middleware
	a.ToolManager.AddMiddleware(middleware.NewResilienceMiddleware(a.ToolManager))

	a.PromptManager = prompt.NewManager()
	a.TemplateManager = NewTemplateManager("data") // Use "data" directory for now
	a.ResourceManager = resource.NewManager()

	// Initialize auth manager
	authManager := auth.NewManager()
	authManager.SetUsers(cfg.GetUsers())

	// Cast storageStore to storage.Storage
	if s, ok := storageStore.(storage.Storage); ok {
		authManager.SetStorage(s)
	} else {
		// This should theoretically not happen if storageStore is properly initialized from sqlite/postgres
		log.Warn("storageStore does not implement storage.Storage, interactive OAuth will be disabled")
	}

	// Use SetAPIKey from config if available
	if a.SettingsManager.GetAPIKey() != "" {
		authManager.SetAPIKey(a.SettingsManager.GetAPIKey())
	}
	// Note: previous code checked cfg.GetGlobalSettings().GetApiKeyParamName() but that might be inside Authentication config?
	// GlobalSettings usually has Authentication field.
	// Let's rely on SettingsManager or check cfg.GetGlobalSettings().GetAuthentication() if needed
	// For API Key param name, it is likely in Authentication message if configured.
	// But AuthManager uses APIKeyAuthenticator which takes config.APIKeyAuth.
	// The explicit API key (CLI) overrides or sets a simple key.
	// We'll leave the complex check out for now unless it was critical for something else.

	// Register auth manager
	a.AuthManager = authManager

	// Initialize Profile Manager and set profile definitions
	// GetProfileDefinitions returns nil if not set, handled by Update
	var profileDefinitions []*config_v1.ProfileDefinition
	if cfg.GetGlobalSettings() != nil {
		profileDefinitions = cfg.GetGlobalSettings().GetProfileDefinitions()
	} else {
		profileDefinitions = config.GlobalSettings().GetProfileDefinitions()
	}
	a.ProfileManager = profile.NewManager(profileDefinitions)

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
	a.ServiceRegistry = serviceRegistry

	// New message bus and workers
	upstreamWorker := worker.NewUpstreamWorker(busProvider, a.ToolManager)
	registrationWorker := worker.NewServiceRegistrationWorker(busProvider, serviceRegistry)

	// Create a context for workers that we can cancel on shutdown
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()

	// Start background workers
	upstreamWorker.Start(workerCtx)
	registrationWorker.Start(workerCtx)

	// If we're using an in-memory bus, start the in-process worker
	if busConfig == nil || busConfig.GetInMemory() != nil {
		workerCfg := &worker.Config{
			MaxWorkers:   10,
			MaxQueueSize: 100,
		}
		inProcessWorker := worker.New(busProvider, workerCfg)
		inProcessWorker.Start(workerCtx)
		defer inProcessWorker.Stop()
	}

	// Initialize and start Global GC Worker
	gcSettings := cfg.GetGlobalSettings().GetGcSettings()
	if gcSettings != nil && gcSettings.GetEnabled() {
		interval, _ := time.ParseDuration(gcSettings.GetInterval())
		ttl, _ := time.ParseDuration(gcSettings.GetTtl())

		gpPaths := gcSettings.GetPaths()
		// Always include the bundle directory if it's set in env (which we did for config)
		// Or we can rely on config.
		// For now, respect config exactly.

		gcWorker := gc.New(gc.Config{
			Enabled:  true,
			Interval: interval,
			TTL:      ttl,
			Paths:    gpPaths,
		})
		gcWorker.Start(workerCtx)
	}

	// Initialize Topology Manager
	a.TopologyManager = topology.NewManager(serviceRegistry, a.ToolManager)

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
		workerCancel()
		upstreamWorker.Stop()
		registrationWorker.Stop()
		return fmt.Errorf("failed to create mcp server: %w", err)
	}

	mcpSrv.SetReloadFunc(func(ctx context.Context) error {
		return a.ReloadConfig(ctx, fs, configPaths)
	})

	a.ToolManager.SetMCPServer(mcpSrv)

	if cfg.GetUpstreamServices() != nil {
		// Publish registration requests to the bus for each service
		registrationBus, err := bus.GetBus[*bus.ServiceRegistrationRequest](
			busProvider,
			"service_registration_requests",
		)
		if err != nil {
			workerCancel()
			upstreamWorker.Stop()
			registrationWorker.Stop()
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
	standardMiddlewares, err := middleware.InitStandardMiddlewares(mcpSrv.AuthManager(), a.ToolManager, cfg.GetGlobalSettings().GetAudit(), cachingMiddleware, cfg.GetGlobalSettings().GetRateLimit(), cfg.GetGlobalSettings().GetDlp())
	if err != nil {
		workerCancel()
		upstreamWorker.Stop()
		registrationWorker.Stop()
		return fmt.Errorf("failed to init standard middlewares: %w", err)
	}
	a.standardMiddlewares = standardMiddlewares
	if standardMiddlewares.Cleanup != nil {
		defer func() {
			if err := standardMiddlewares.Cleanup(); err != nil {
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
			{Name: proto.String(authMiddlewareName), Priority: proto.Int32(20)},
			{Name: proto.String("logging"), Priority: proto.Int32(30)},
			{Name: proto.String("audit"), Priority: proto.Int32(40)},
			{Name: proto.String("dlp"), Priority: proto.Int32(42)},
			{Name: proto.String("global_ratelimit"), Priority: proto.Int32(45)},
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

	// 🛡️ Sentinel Security Update: Enforce Auth Middleware for HTTP/Server Mode
	// If NOT running in stdio mode (which implies server mode), ensure "auth" middleware is present
	// unless explicitly disabled by user config.
	// If the user provided a list of middlewares but forgot "auth", we inject it.
	if !stdio {
		authPresent := false
		for _, m := range middlewares {
			if m.GetName() == authMiddlewareName {
				authPresent = true
				break
			}
		}

		if !authPresent {
			logging.GetLogger().Warn("Auth middleware missing from configuration. Injecting it for security.")
			// Add auth middleware with default priority 20
			authMiddlewareConfig := &config_v1.Middleware{
				Name:     proto.String(authMiddlewareName),
				Priority: proto.Int32(20),
			}
			middlewares = append(middlewares, authMiddlewareConfig)
		}
	}

	// Apply middlewares
	// Registry returns sorted list based on priority (low to high).
	// If priority 0 is first, it wraps the rest?
	// If we iterate:
	// M1(M2(M3(...)))
	// M1 is priority 0.
	// If running in stdio mode, we must remove the auth middleware as it requires
	// an HTTP request availability in the context, which is not present in stdio.
	// Stdio mode implies local access (shell), so we trust the user.
	if stdio {
		var filtered []*config_v1.Middleware
		for _, m := range middlewares {
			if m.GetName() != authMiddlewareName {
				filtered = append(filtered, m)
			}
		}
		middlewares = filtered
	}

	chain := middleware.GetMCPMiddlewares(middlewares)
	for _, m := range chain {
		logging.GetLogger().Info("Adding middleware", "count", len(chain))
		mcpSrv.Server().AddReceivingMiddleware(m)
	}

	// Add Topology Middleware (Always Active)
	mcpSrv.Server().AddReceivingMiddleware(a.TopologyManager.Middleware)

	// Add Prometheus Metrics Middleware (Always Active)
	// We use SimpleTokenizer for low-overhead token counting
	mcpSrv.Server().AddReceivingMiddleware(middleware.PrometheusMetricsMiddleware(tokenizer.NewSimpleTokenizer()))

	if stdio {
		err := a.runStdioModeFunc(ctx, mcpSrv)
		workerCancel()
		upstreamWorker.Stop()
		registrationWorker.Stop()
		return err
	}

	bindAddress := jsonrpcPort
	if cfg.GetGlobalSettings().GetMcpListenAddress() != "" {
		bindAddress = cfg.GetGlobalSettings().GetMcpListenAddress()
	}

	// Use storageStore which is initialized as either sqlite or postgres
	// We need to assert it to storage.Storage. Both implement it.
	// But stores[...] is config.Store. storageStore is config.Store.
	// However, we know storageStore implements storage.Storage because we created it as such.

	// Wait, storageStore is declared as config.Store in my previous edit.
	// I should cast it or change its type declaration.
	// Let's change declaration in previous step, but since I can't undo easily without reset,
	// I'll cast it here.
	s, ok := storageStore.(storage.Storage)
	if !ok {
		// Should not happen if code is correct
		return fmt.Errorf("storage store does not implement storage.Storage")
	}
	a.Storage = s

	a.Storage = s

	// Signal startup complete
	startupCallback := func() {
		a.startupOnce.Do(func() {
			close(a.startupCh)
		})
	}

	// Start servers
	if err := a.runServerMode(
		ctx,
		mcpSrv,
		busProvider,
		bindAddress,
		grpcPort,
		shutdownTimeout,
		cfg.GetGlobalSettings(),
		cachingMiddleware,
		s,
		serviceRegistry,
		startupCallback,
	); err != nil {
		workerCancel()
		upstreamWorker.Stop()
		registrationWorker.Stop()
		return err
	}

	// Stop workers
	workerCancel()
	upstreamWorker.Stop()
	registrationWorker.Stop()

	return nil
}

// ReloadConfig reloads the configuration from the given paths and updates the
// services.
func (a *Application) ReloadConfig(ctx context.Context, fs afero.Fs, configPaths []string) error {
	log := logging.GetLogger()
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			log.Error("ReloadConfig panicked", "panic", r)
		}
		log.Info("ReloadConfig completed", "duration", time.Since(start))
	}()

	a.configMu.Lock()
	defer a.configMu.Unlock()

	log.Info("Reloading configuration...")
	metrics.IncrCounter([]string{"config", "reload", "total"}, 1)

	cfg, err := a.loadConfig(ctx, fs, configPaths)
	a.lastReloadTime = time.Now()
	a.lastReloadErr = err
	if err != nil {
		metrics.IncrCounter([]string{"config", "reload", "errors"}, 1)
		return fmt.Errorf("failed to load services from config: %w", err)
	}

	// Update global settings
	a.updateGlobalSettings(cfg)

	// Update Users (Dynamic!)
	if a.AuthManager != nil {
		a.AuthManager.SetUsers(cfg.GetUsers())
		log.Info("Updated users configuration")
	}

	// Update profiles on reload
	a.ToolManager.SetProfiles(
		cfg.GetGlobalSettings().GetProfiles(),
		cfg.GetGlobalSettings().GetProfileDefinitions(),
	)

	// Update Profile Manager (Dynamic!)
	if a.ProfileManager != nil {
		a.ProfileManager.Update(cfg.GetGlobalSettings().GetProfileDefinitions())
		log.Info("Updated profile definitions configuration")
	}

	// Reconcile services (add/remove/update)
	a.reconcileServices(ctx, cfg)
	return nil
}

func (a *Application) loadConfig(ctx context.Context, fs afero.Fs, configPaths []string) (*config_v1.McpAnyServerConfig, error) {
	var stores []config.Store
	if len(configPaths) > 0 {
		stores = append(stores, config.NewFileStore(fs, configPaths))
	}
	if a.Storage != nil {
		stores = append(stores, a.Storage)
	}

	store := config.NewMultiStore(stores...)
	return config.LoadServices(ctx, store, "server")
}

func (a *Application) updateGlobalSettings(cfg *config_v1.McpAnyServerConfig) {
	log := logging.GetLogger()
	if a.SettingsManager != nil {
		a.SettingsManager.Update(cfg.GetGlobalSettings(), a.explicitAPIKey)
	}

	// Update dynamic middlewares
	if a.ipMiddleware != nil {
		if err := a.ipMiddleware.Update(a.SettingsManager.GetAllowedIPs()); err != nil {
			log.Error("Failed to update IP allowlist", "error", err)
		}
	}
	if a.corsMiddleware != nil {
		a.corsMiddleware.Update(a.SettingsManager.GetAllowedOrigins())
	}

	if a.standardMiddlewares != nil {
		if a.standardMiddlewares.Audit != nil {
			if err := a.standardMiddlewares.Audit.UpdateConfig(cfg.GetGlobalSettings().GetAudit()); err != nil {
				log.Error("Failed to update audit middleware config", "error", err)
			}
		}
		if a.standardMiddlewares.GlobalRateLimit != nil {
			a.standardMiddlewares.GlobalRateLimit.UpdateConfig(cfg.GetGlobalSettings().GetRateLimit())
		}
	}
}

// reconcileServices reconciles the service registry with the new configuration.
func (a *Application) reconcileServices(ctx context.Context, cfg *config_v1.McpAnyServerConfig) {
	log := logging.GetLogger()
	// Get current active services
	currentServicesMap := make(map[string]*config_v1.UpstreamServiceConfig)
	if a.ServiceRegistry != nil {
		services, err := a.ServiceRegistry.GetAllServices()
		if err == nil {
			for _, s := range services {
				currentServicesMap[s.GetName()] = s
			}
		}
	}

	// Map new services by name for easy lookup
	newServices := make(map[string]*config_v1.UpstreamServiceConfig)
	if cfg.GetUpstreamServices() != nil {
		for _, svc := range cfg.GetUpstreamServices() {
			if !svc.GetDisable() {
				newServices[svc.GetName()] = svc
			}
		}
	}

	// Identify removed services
	for name := range currentServicesMap {
		if _, exists := newServices[name]; !exists {
			log.Info("Removing service", "service", name)
			if a.ServiceRegistry != nil {
				if err := a.ServiceRegistry.UnregisterService(ctx, name); err != nil {
					log.Error("Failed to unregister service", "service", name, "error", err)
				}
			}
		}
	}

	// Identify added or updated services
	for name, newSvc := range newServices {
		oldConfig, exists := currentServicesMap[name]
		needsUpdate := false

		if !exists {
			log.Info("Adding new service", "service", name)
			needsUpdate = true
		} else {
			// Compare configs
			newSvcCopy := proto.Clone(newSvc).(*config_v1.UpstreamServiceConfig)
			if newSvcCopy.GetId() == "" {
				newSvcCopy.Id = oldConfig.Id
			}
			if newSvcCopy.GetSanitizedName() == "" {
				newSvcCopy.SanitizedName = oldConfig.SanitizedName
			}

			if !proto.Equal(oldConfig, newSvcCopy) {
				log.Info("Updating service", "service", name)
				needsUpdate = true
				if a.ServiceRegistry != nil {
					if err := a.ServiceRegistry.UnregisterService(ctx, name); err != nil {
						log.Error("Failed to unregister service for update", "service", name, "error", err)
					}
				}
			}
		}

		if needsUpdate {
			switch {
			case a.busProvider != nil:
				// Async registration via bus to support retries
				registrationBus, err := bus.GetBus[*bus.ServiceRegistrationRequest](
					a.busProvider,
					bus.ServiceRegistrationRequestTopic,
				)
				if err != nil {
					log.Error("Failed to get registration bus during reload", "error", err)
					continue
				}
				regReq := &bus.ServiceRegistrationRequest{Config: newSvc}
				if err := registrationBus.Publish(context.Background(), "request", regReq); err != nil {
					log.Error("Failed to publish registration request during reload", "error", err)
				} else {
					log.Info("Queued service for registration update", "service", name)
				}
			case a.ServiceRegistry != nil:
				// Fallback to sync registration if bus is not available (e.g. tests without full init)
				_, _, _, err := a.ServiceRegistry.RegisterService(context.Background(), newSvc)
				if err != nil {
					log.Error("Failed to register upstream service", "service", name, "error", err)
					continue
				}
			default:
				log.Warn("ServiceRegistry is nil, cannot register service", "service", name)
			}
		} else {
			log.Debug("Service unchanged", "service", name)
		}
	}

	log.Info("Reload complete", "tools_count", len(a.ToolManager.ListTools()))
}

// WaitForStartup waits for the application to be fully initialized.
// It returns nil if startup completes, or context error if context is canceled.
func (a *Application) WaitForStartup(ctx context.Context) error {
	select {
	case <-a.startupCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
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

// configHealthCheck checks the status of the configuration.
func (a *Application) configHealthCheck(_ context.Context) health.CheckResult {
	a.configMu.Lock()
	defer a.configMu.Unlock()

	if a.lastReloadErr != nil {
		return health.CheckResult{
			Status:  "degraded",
			Message: a.lastReloadErr.Error(),
			Latency: time.Since(a.lastReloadTime).String(),
		}
	}

	status := "ok"
	if a.lastReloadTime.IsZero() {
		status = "unknown"
	}

	return health.CheckResult{
		Status:  status,
		Latency: time.Since(a.lastReloadTime).String(),
	}
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
	globalSettings *config_v1.GlobalSettings,
	cachingMiddleware *middleware.CachingMiddleware,
	store storage.Storage,
	serviceRegistry *serviceregistry.ServiceRegistry,
	startupCallback func(),
) error {
	ipMiddleware, err := middleware.NewIPAllowlistMiddleware(a.SettingsManager.GetAllowedIPs())
	if err != nil {
		return fmt.Errorf("failed to create IP allowlist middleware: %w", err)
	}
	a.ipMiddleware = ipMiddleware

	// localCtx is used to manage the lifecycle of the servers started in this function.
	// It's canceled when this function returns, ensuring that all servers are shut down.
	localCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	defer cancel()

	errChan := make(chan error, 2)
	readyChan := make(chan struct{}, 2)
	expectedReady := 0
	var wg sync.WaitGroup

	rawHTTPHandler := mcp.NewStreamableHTTPHandler(func(_ *http.Request) *mcp.Server {
		return mcpSrv.Server()
	}, nil)

	// Wrap the HTTP handler with OpenTelemetry instrumentation
	httpHandler := otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "http.request", r) //nolint:revive,staticcheck // matching legacy usage in auth.go
		rawHTTPHandler.ServeHTTP(w, r.WithContext(ctx))
	}), "server-request")

	// Check if auth middleware is disabled in config
	var authDisabled bool

	// Use passed globalSettings for middleware config check
	if globalSettings != nil {
		for _, m := range globalSettings.GetMiddlewares() {
			if m.GetName() == authMiddlewareName && m.GetDisabled() {
				authDisabled = true
				break
			}
		}
	} else {
		// Fallback to singleton if nil (should not happen in normal Run)
		for _, m := range config.GlobalSettings().Middlewares() {
			if m.GetName() == authMiddlewareName && m.GetDisabled() {
				authDisabled = true
				break
			}
		}
	}

	var authMiddleware func(http.Handler) http.Handler
	if authDisabled {
		logging.GetLogger().Warn("Auth middleware is disabled by config! Enforcing private-IP-only access for safety.")
		// Even if auth is disabled, we enforce private-IP-only access to prevent public exposure.
		authMiddleware = a.createAuthMiddleware(true)
	} else {
		authMiddleware = a.createAuthMiddleware(false)
	}

	mux := http.NewServeMux()

	// UI Handler
	uiFS := http.FileServer(http.Dir("./ui"))
	mux.Handle("/ui/", http.StripPrefix("/ui", uiFS))

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

		// Dynamic User Lookup
		user, ok := a.AuthManager.GetUser(uid)
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

		// 1. Profile Auth - REMOVED (Per-profile ingress auth is no longer supported in this refactor)
		var isAuthenticated bool
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
			apiKey := a.SettingsManager.GetAPIKey()
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
		// Dynamic Profile Lookup
		if def, ok := a.ProfileManager.GetProfileDefinition(profileID); ok && len(def.RequiredRoles) > 0 {
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
				// Don't leak required roles to the client
				logging.GetLogger().Warn("Forbidden access to profile", "profile", profileID, "user", uid, "required_roles", def.RequiredRoles, "user_roles", user.GetRoles())
				http.Error(w, "Forbidden", http.StatusForbidden)
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

	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "OK")
	}))
	mux.Handle("/metrics", authMiddleware(metrics.Handler()))
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

	// Apply Global Rate Limit: 20 RPS with a burst of 50.
	// This helps prevent basic DoS attacks on all HTTP endpoints, including /upload.
	// We enable trustProxy if MCPANY_TRUST_PROXY is set, to handle load balancers correctly.
	trustProxy := os.Getenv("MCPANY_TRUST_PROXY") == util.TrueStr
	rateLimiter := middleware.NewHTTPRateLimitMiddleware(20, 50, middleware.WithTrustProxy(trustProxy))

	// Apply CORS Middleware
	corsMiddleware := middleware.NewHTTPCORSMiddleware(a.SettingsManager.GetAllowedOrigins())
	a.corsMiddleware = corsMiddleware

	// Middleware order: SecurityHeaders -> CORS -> JSONRPCCompliance -> IPAllowList -> RateLimit -> Mux
	// We wrap everything with a debug logger to see what's coming in
	handler := middleware.HTTPSecurityHeadersMiddleware(
		corsMiddleware.Handler(
			middleware.JSONRPCComplianceMiddleware(
				ipMiddleware.Handler(
					rateLimiter.Handler(mux),
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

	adminServer := admin.NewServer(cachingMiddleware, a.ToolManager, store)
	pb_admin.RegisterAdminServiceServer(grpcServer, adminServer)

	// Initialize gRPC-Web wrapper even if gRPC port is not exposed
	wrappedGrpc = grpcweb.WrapServer(grpcServer,
		grpcweb.WithOriginFunc(func(_ string) bool { return true }),
		grpcweb.WithAllowedRequestHeaders([]string{"*"}),
	)

	if grpcBindAddress != "" {
		if !strings.Contains(grpcBindAddress, ":") {
			grpcBindAddress = ":" + grpcBindAddress
		}
		lis, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", grpcBindAddress)
		if err != nil {
			errChan <- fmt.Errorf("gRPC server failed to listen: %w", err)
		} else {
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
	// Register Root Handler with gRPC-Web support
	mux.Handle("/", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc != nil && wrappedGrpc.IsGrpcWebRequest(r) {
			wrappedGrpc.ServeHTTP(w, r)
			return
		}
		httpHandler.ServeHTTP(w, r)
	})))

	httpLis, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", httpBindAddress)
	if err != nil {
		errChan <- fmt.Errorf("HTTP server failed to listen: %w", err)
	} else {
		expectedReady++
		startHTTPServer(localCtx, &wg, errChan, readyChan, "MCP Any HTTP", httpLis, handler, shutdownTimeout)
	}

	// Wait for servers to be ready
	timeout := time.NewTimer(10 * time.Second) // Reasonable timeout for binding ports
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

// createAuthMiddleware creates the authentication middleware.
func (a *Application) createAuthMiddleware(forcePrivateIPOnly bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := util.ExtractIP(r.RemoteAddr)
			ctx := util.ContextWithRemoteIP(r.Context(), ip)
			r = r.WithContext(ctx)

			apiKey := a.SettingsManager.GetAPIKey()

			if !forcePrivateIPOnly && apiKey != "" {
				// Check X-API-Key or Authorization header
				requestKey := r.Header.Get("X-API-Key")
				if requestKey == "" {
					authHeader := r.Header.Get("Authorization")
					if strings.HasPrefix(authHeader, "Bearer ") {
						requestKey = strings.TrimPrefix(authHeader, "Bearer ")
					}
				}

				if subtle.ConstantTimeCompare([]byte(requestKey), []byte(apiKey)) != 1 {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			} else {
				// Sentinel Security: If no API key is configured, enforce localhost-only access.
				// This prevents accidental exposure of the server to the public internet (RCE risk).
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					// Fallback if RemoteAddr is weird, assume host is the string itself
					host = r.RemoteAddr
				}

				// Check if the request is from a loopback address
				ip := net.ParseIP(host)
				if !util.IsPrivateIP(ip) {
					logging.GetLogger().Warn("Blocked public internet request because no API Key is configured", "remote_addr", r.RemoteAddr)
					http.Error(w, "Forbidden: Public access requires an API Key to be configured", http.StatusForbidden)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
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
	readyChan chan<- struct{},
	name string,
	lis net.Listener,
	shutdownTimeout time.Duration,
	server *gogrpc.Server,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverLog := logging.GetLogger().With("server", name)

		if server == nil {
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
				server.GracefulStop()
			}()

			timer := time.NewTimer(shutdownTimeout)
			defer timer.Stop()
			select {
			case <-stopped:
				// Successful graceful shutdown.
			case <-timer.C:
				// Graceful shutdown timed out.
				serverLog.Warn("Graceful shutdown timed out, forcing stop.")
				server.Stop()
			}
		}()

		serverLog.Info("gRPC server listening", "port", lis.Addr().String())
		if readyChan != nil {
			readyChan <- struct{}{}
		}
		if err := server.Serve(lis); err != nil && err != gogrpc.ErrServerStopped {
			errChan <- fmt.Errorf("[%s] server failed to serve: %w", name, err)
			cancel() // Signal shutdown goroutine to exit
		}
		<-shutdownComplete
		serverLog.Info("Server shut down.")
	}()
}

// startHTTPServer starts an HTTP server in a new goroutine. It handles graceful
// shutdown when the context is canceled.
//
// ctx is the context for managing the server's lifecycle.
// wg is a WaitGroup to signal when the server has shut down.
// errChan is a channel for reporting errors during startup.
// name is a descriptive name for the server, used in logging.
// lis is the net.Listener on which the server will listen.
// handler is the HTTP handler for processing requests.
func startHTTPServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	errChan chan<- error,
	readyChan chan<- struct{},
	name string,
	lis net.Listener,
	handler http.Handler,
	shutdownTimeout time.Duration,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverLog := logging.GetLogger().With("server", name)
		// Listener passed in is already listening
		if readyChan != nil {
			readyChan <- struct{}{}
		}
		// We don't close lis here because http.Server.Serve closes it,
		// or we might want to close it on shutdown?
		// http.Server.Serve docs: "Serve accepts incoming connections on the Listener l... always returns a non-nil error."
		// It does NOT say it closes the listener.
		// However, Shutdown() closes the listener?
		// "Shutdown gracefully shuts down the server without interrupting any active connections... Shutdown works by first closing all open listeners..."
		// So Server.Shutdown closes it.
		// BUT if we error out before Shutdown, we should close it?
		// Let's rely on Server.Serve or Shutdown closing it, or defer close if not?
		// Ideally we defer close if Serve returns error other than ErrServerClosed.
		// Use a flag to check if Shutdown was called?
		// Or just defer Close() ignoring error? http.Server might close it too, double close is fine for net.Listener usually.
		// But let's check stdlib behavior.
		// Safe bet: defer lis.Close() at the top.
		// If Shutdown is called, it closes it. Double close is harmless for TCP listeners.
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
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       120 * time.Second,
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
