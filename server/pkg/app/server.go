// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package app provides the main application logic.
package app

import (
	"context"
	"crypto/subtle"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mcpany/core/server/pkg/alerts"
	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/catalog"
	"github.com/mcpany/core/server/pkg/config"
	"github.com/mcpany/core/server/pkg/discovery"
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
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/storage/postgres"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/mcpany/core/server/pkg/telemetry"
	"github.com/mcpany/core/server/pkg/tokenizer"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream/factory"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/mcpany/core/server/pkg/webhooks"
	"github.com/mcpany/core/server/pkg/worker"
	"github.com/pmezard/go-difflib/difflib"

	// config_v1 "github.com/mcpany/core/proto/config/v1".
	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/afero"
	gogrpc "google.golang.org/grpc"
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
	// Sanitize the filename to prevent reflected XSS and ensure safe filesystem usage
	safeFilename := util.SanitizeFilename(header.Filename)
	w.Header().Set("Content-Type", "text/plain")
	_, _ = fmt.Fprintf(w, "File '%s' uploaded successfully (size: %d bytes)", html.EscapeString(safeFilename), written)
}

// RunOptions configuration for starting the MCP Any application.
//
// Summary: Options for configuring the application runtime.
//
// Fields:
//   - Ctx: context.Context. The context for the application.
//   - Fs: afero.Fs. The filesystem interface.
//   - Stdio: bool. Whether to run in stdio mode (for CLI/one-off usage).
//   - JSONRPCPort: string. The port for the JSON-RPC/HTTP server.
//   - GRPCPort: string. The port for the gRPC registration server.
//   - ConfigPaths: []string. Paths to configuration files.
//   - APIKey: string. The master API key for the server.
//   - ShutdownTimeout: time.Duration. The timeout for graceful shutdown.
//   - TLSCert: string. Path to the TLS certificate file.
//   - TLSKey: string. Path to the TLS private key file.
//   - TLSClientCA: string. Path to the TLS client CA certificate file (for mTLS).
//   - DBPath: string. Path to the SQLite database file.
type RunOptions struct {
	Ctx             context.Context
	Fs              afero.Fs
	Stdio           bool
	JSONRPCPort     string
	GRPCPort        string
	ConfigPaths     []string
	APIKey          string
	ShutdownTimeout time.Duration
	TLSCert         string
	TLSKey          string
	TLSClientCA     string
	DBPath          string
}

// Runner defines the interface for running the application.
//
// Summary: Interface for application execution and management.
type Runner interface {
	// Run starts the application with the given options.
	//
	// Summary: Starts the application.
	//
	// Parameters:
	//   - opts: RunOptions. The configuration for running.
	//
	// Returns:
	//   - error: An error if startup or execution fails.
	Run(opts RunOptions) error

	// ReloadConfig reloads the application configuration.
	//
	// Summary: Triggers a configuration reload.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the operation.
	//   - fs: afero.Fs. The filesystem.
	//   - configPaths: []string. Paths to configuration files.
	//
	// Returns:
	//   - error: An error if reload fails.
	ReloadConfig(ctx context.Context, fs afero.Fs, configPaths []string) error
}

// Application is the main application struct, holding the dependencies and logic for the MCP Any server.
//
// Summary: The main application container.
//
// Fields:
//   - PromptManager: prompt.ManagerInterface. Manages AI prompts.
//   - ToolManager: tool.ManagerInterface. Manages tools and execution.
//   - ResourceManager: resource.ManagerInterface. Manages resources (files, data).
//   - ServiceRegistry: serviceregistry.ServiceRegistryInterface. Manages upstream service connections.
//   - TopologyManager: *topology.Manager. Manages the topology of the server.
//   - UpstreamFactory: factory.Factory. Creates upstream service clients.
//   - Storage: storage.Storage. Persistent storage interface.
//   - TemplateManager: *TemplateManager. Manages templates.
//   - SkillManager: *skill.Manager. Manages agent skills.
//   - AlertsManager: *alerts.Manager. Manages system alerts.
//   - DiscoveryManager: *discovery.Manager. Manages auto-discovery of services.
//   - SettingsManager: *GlobalSettingsManager. Manages dynamic global settings.
//   - ProfileManager: *profile.Manager. Manages user profiles.
//   - AuthManager: *auth.Manager. Manages authentication and authorization.
//   - RegistrationRetryDelay: time.Duration. Delay between service registration retries.
//   - MetricsGatherer: prometheus.Gatherer. Interface for gathering metrics.
//   - BoundHTTPPort: atomic.Int32. The actual bound HTTP port.
//   - BoundGRPCPort: atomic.Int32. The actual bound gRPC port.
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
	// Store explicit API Key passed via CLI args
	explicitAPIKey string

	// SkillManager manages agent skills
	SkillManager *skill.Manager

	// AlertsManager manages system alerts
	AlertsManager *alerts.Manager

	// WebhooksManager manages outbound webhooks
	WebhooksManager *webhooks.Manager

	// DiscoveryManager manages auto-discovery providers
	DiscoveryManager *discovery.Manager

	// CatalogManager manages dynamic service catalog
	CatalogManager *catalog.Manager

	// lastReloadErr stores the error from the last configuration reload.
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
	csrfMiddleware *middleware.CSRFMiddleware

	busProvider *bus.Provider

	startupCh   chan struct{}
	startupOnce sync.Once
	configMu    sync.Mutex

	// lastReloadErr stores the error from the last configuration reload.

	// lastReloadErr stores the error from the last configuration reload.
	// It is protected by configMu.
	lastReloadErr error
	// lastReloadTime stores the time of the last configuration reload attempt.
	// It is protected by configMu.
	lastReloadTime time.Time

	// lastGoodConfig stores the raw content of the last successfully loaded configuration files.
	// Map key is the file path, value is the file content.
	// It is protected by configMu.
	lastGoodConfig map[string]string

	// configDiff stores the diff between the last good config and the failed config.
	// It is protected by configMu.
	configDiff string

	// BoundHTTPPort stores the actual port the HTTP server is listening on.
	BoundHTTPPort atomic.Int32
	// BoundGRPCPort stores the actual port the gRPC server is listening on.
	BoundGRPCPort atomic.Int32

	// startTime is the time the application started.
	startTime time.Time
	// activeConnections tracks the number of active HTTP connections.
	activeConnections int32

	// RegistrationRetryDelay allows configuring the retry delay for service registration.
	// If 0, it defaults to 5 seconds (in the worker).
	RegistrationRetryDelay time.Duration

	// MetricsGatherer is the interface for gathering metrics.
	// Defaults to prometheus.DefaultGatherer.
	MetricsGatherer prometheus.Gatherer

	// statsCache for dashboard
	statsCacheMu sync.RWMutex
	statsCache   map[string]statsCacheEntry
}

type statsCacheEntry struct {
	Data      any
	ExpiresAt time.Time
}

// NewApplication creates a new Application with default dependencies.
//
// Summary: Initializes a new Application instance.
//
// Returns:
//   - (*Application): The initialized application.
func NewApplication() *Application {
	busProvider, _ := bus.NewProvider(nil)
	return &Application{
		runStdioModeFunc: runStdioMode,
		PromptManager:    prompt.NewManager(),
		ToolManager:      tool.NewManager(busProvider),
		AlertsManager:    alerts.NewManager(),
		WebhooksManager:  webhooks.NewManager(),
		CatalogManager:   catalog.NewManager(afero.NewOsFs(), "marketplace/catalog"), // Default path, can be overridden

		ResourceManager: resource.NewManager(),
		UpstreamFactory: factory.NewUpstreamServiceFactory(pool.NewManager(), nil),
		configFiles:     make(map[string]string),
		startupCh:       make(chan struct{}),
		startTime:       time.Now(),
		MetricsGatherer: prometheus.DefaultGatherer,
		statsCache:      make(map[string]statsCacheEntry),
	}
}

// Run starts the MCP Any server and all its components.
//
// Summary: Executes the application.
//
// Parameters:
//   - opts (RunOptions): The runtime options.
//
// Returns:
//   - (error): An error if execution fails.
//
// Side Effects:
//   - Starts HTTP and gRPC servers.
//   - Initializes background workers.
//   - Loads configuration.
//
//nolint:gocyclo // Run is the main entry point and setup function, expected to be complex
func (a *Application) Run(opts RunOptions) error {
	log := logging.GetLogger()
	fs, err := setup(opts.Fs)
	if err != nil {
		return fmt.Errorf("failed to setup filesystem: %w", err)
	}
	a.fs = fs
	a.configPaths = opts.ConfigPaths
	a.explicitAPIKey = opts.APIKey
	log.Info("DEBUG: Run API Key", "key", opts.APIKey)

	// Telemetry initialization moved after config loading

	log.Info("Starting MCP Any Service...")

	// Load initial services from config files and Storage
	var storageStore config.Store
	var storageCloser func() error

	if a.Storage != nil {
		storageStore = a.Storage
	} else {
		// Default to SQLite if not specified or explicitly sqlite
		dbDriver := config.GlobalSettings().GetDbDriver()
		switch dbDriver {
		case "", "sqlite":
			dbPath := opts.DBPath
			if dbPath == "" {
				if _, ok := opts.Fs.(*afero.MemMapFs); ok {
					dbPath = ":memory:"
				}
			}
			if dbPath == "" {
				dbPath = config.GlobalSettings().DBPath()
			}
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
	}
	defer func() {
		if storageCloser != nil {
			_ = storageCloser()
		}
	}()

	// Determine config sources
	// Priority: File (if enabled/provided) > Database
	var stores []config.Store

	enableFileConfig := os.Getenv("MCPANY_ENABLE_FILE_CONFIG") == "true"
	if len(opts.ConfigPaths) > 0 {
		// Always load config files if they are explicitly provided
		log.Info("Loading config from files (highest priority)", "paths", opts.ConfigPaths)
		stores = append(stores, config.NewFileStore(fs, opts.ConfigPaths))
	} else if enableFileConfig {
		log.Info("File configuration enabled via env var, but no config paths provided.")
	}

	// Add database as a fallback/secondary source
	stores = append(stores, storageStore)

	multiStore := config.NewMultiStore(stores...)

	var cfg *config_v1.McpAnyServerConfig
	cfg, err = config.LoadServices(opts.Ctx, multiStore, "server")
	if err != nil {
		return fmt.Errorf("failed to load services from merged config: %w", err)
	}
	if cfg == nil {
		cfg = config_v1.McpAnyServerConfig_builder{}.Build()
	}

	// Initialize DB if empty (passing the loaded config to check if seeding is needed)
	if err := a.initializeDatabase(opts.Ctx, storageStore, cfg); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize Log Persistence (Hydration & Worker)
	if s, ok := storageStore.(storage.Storage); ok {
		if err := a.initializeLogPersistence(opts.Ctx, s); err != nil {
			log.Error("Failed to initialize log persistence", "error", err)
		}
		a.startLogPersistence(opts.Ctx, s)
	}
	a.lastReloadTime = time.Now()

	// Populate initial good config for diffing
	if len(opts.ConfigPaths) > 0 {
		a.lastGoodConfig, _ = a.readConfigFiles(fs, opts.ConfigPaths)
	}

	// Initialize Telemetry with loaded config
	shutdownTelemetry, err := telemetry.InitTelemetry(opts.Ctx, appconsts.Name, appconsts.Version, cfg.GetGlobalSettings().GetTelemetry(), os.Stderr)
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
		opts.APIKey,
		cfg.GetGlobalSettings().GetAllowedIps(),
		// Logic for origins default moved to inside NewGlobalSettingsManager or updated here
		nil,
	)
	a.SettingsManager.Update(cfg.GetGlobalSettings(), opts.APIKey)

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

	a.DiscoveryManager = discovery.NewManager()

	// Initialize Skill Manager
	skillManager, err := skill.NewManager("skills") // Use "skills" directory in CWD for now
	if err != nil {
		return fmt.Errorf("failed to initialize skill manager: %w", err)
	}
	a.SkillManager = skillManager

	// Initialize auth manager
	authManager := auth.NewManager()
	users := cfg.GetUsers()
	if s, ok := storageStore.(storage.Storage); ok {
		dbUsers, err := s.ListUsers(opts.Ctx)
		if err != nil {
			log.Error("failed to list users from storage", "error", err)
		} else {
			users = append(users, dbUsers...)
		}
	}
	authManager.SetUsers(users)

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
	var profileDefinitions []*config_v1.ProfileDefinition
	if cfg.GetGlobalSettings() != nil {
		profileDefinitions = cfg.GetGlobalSettings().GetProfileDefinitions()
	} else {
		profileDefinitions = config.GlobalSettings().GetProfileDefinitions()
	}

	// Ensure there is at least a "default" profile
	hasDefault := false
	for _, p := range profileDefinitions {
		if p.GetName() == "default" {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		log.Debug("Injected missing 'default' profile definition")
		profileDefinitions = append(profileDefinitions, config_v1.ProfileDefinition_builder{
			Name: proto.String("default"),
		}.Build())
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
	if a.RegistrationRetryDelay > 0 {
		registrationWorker.SetRetryDelay(a.RegistrationRetryDelay)
	}

	// Create a context for workers that we can cancel on shutdown
	workerCtx, workerCancel := context.WithCancel(opts.Ctx)
	defer workerCancel()

	// Start background workers
	upstreamWorker.Start(workerCtx)
	registrationWorker.Start(workerCtx)
	// Start periodic health checks (every 30 seconds)
	serviceRegistry.StartHealthChecks(workerCtx, 30*time.Second)

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
		opts.Ctx,
		a.ToolManager,
		a.PromptManager,
		a.ResourceManager,
		authManager,
		serviceRegistry,
		a.CatalogManager,
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
		return a.ReloadConfig(ctx, fs, opts.ConfigPaths)
	})

	// Register Skill resources
	if err := mcpserver.RegisterSkillResources(a.ResourceManager, a.SkillManager); err != nil {
		log.Error("Failed to register skill resources", "error", err)
		// Don't fail startup for this?
	}

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
			if err := registrationBus.Publish(opts.Ctx, "request", regReq); err != nil {
				log.Error("Failed to publish registration request", "error", err)
			}
		}
	} else {
		log.Info("No services found in config, skipping service registration.")
	}

	// Initialize standard middlewares in registry
	cachingMiddleware := middleware.NewCachingMiddleware(a.ToolManager)
	standardMiddlewares, err := middleware.InitStandardMiddlewares(
		mcpSrv.AuthManager(),
		a.ToolManager,
		cfg.GetGlobalSettings().GetAudit(),
		cachingMiddleware,
		cfg.GetGlobalSettings().GetRateLimit(),
		cfg.GetGlobalSettings().GetDlp(),
		cfg.GetGlobalSettings().GetContextOptimizer(),
		cfg.GetGlobalSettings().GetDebugger(),
		cfg.GetGlobalSettings().GetSmartRecovery(),
		nil,
	)
	if err != nil {
		workerCancel()
		upstreamWorker.Stop()
		registrationWorker.Stop()
		return fmt.Errorf("failed to init standard middlewares: %w", err)
	}

	// Auto-discovery of local services
	if cfg.GetGlobalSettings().GetAutoDiscoverLocal() {
		// Register default providers
		a.DiscoveryManager.RegisterProvider(&discovery.OllamaProvider{Endpoint: "http://localhost:11434"})
		a.DiscoveryManager.RegisterProvider(&discovery.OpenAPIProvider{Endpoint: "http://localhost:8080/openapi.json"})
		a.DiscoveryManager.RegisterProvider(&discovery.GRPCProvider{Endpoint: "localhost:50051"})
		a.DiscoveryManager.RegisterProvider(&discovery.GraphQLProvider{Endpoint: "http://localhost:8080/graphql"})

		discovered := a.DiscoveryManager.Run(opts.Ctx)
		for _, svc := range discovered {
			log.Info("Auto-discovered local service", "name", svc.GetName())
			// Use the getter for UpstreamServices
			cfg.SetUpstreamServices(append(cfg.GetUpstreamServices(), svc))
		}
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
	// We clone them to avoid modifying the singleton's underlying slice if we append/modify.
	middlewares := append([]*config_v1.Middleware(nil), config.GlobalSettings().Middlewares()...)
	if len(middlewares) == 0 {
		// Default chain if none configured
		middlewares = []*config_v1.Middleware{
			config_v1.Middleware_builder{
				Name:     proto.String("debug"),
				Priority: proto.Int32(10),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String(authMiddlewareName),
				Priority: proto.Int32(20),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String("logging"),
				Priority: proto.Int32(30),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String("audit"),
				Priority: proto.Int32(40),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String("dlp"),
				Priority: proto.Int32(42),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String("global_ratelimit"),
				Priority: proto.Int32(45),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String("call_policy"),
				Priority: proto.Int32(50),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String("caching"),
				Priority: proto.Int32(60),
			}.Build(),
			config_v1.Middleware_builder{
				Name:     proto.String("ratelimit"),
				Priority: proto.Int32(70),
			}.Build(),
			// CORS
			config_v1.Middleware_builder{
				Name:     proto.String("cors"),
				Priority: proto.Int32(0),
			}.Build(),
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
	// Stdio mode implies local access (shell), so we trust the user.
	if opts.Stdio {
		var filtered []*config_v1.Middleware
		for _, m := range middlewares {
			if m.GetName() != authMiddlewareName {
				filtered = append(filtered, m)
			}
		}
		middlewares = filtered
	} else {
		// Enforce auth middleware presence in non-stdio modes
		hasAuth := false
		for _, m := range middlewares {
			if m.GetName() == authMiddlewareName {
				hasAuth = true
				break
			}
		}
		if !hasAuth {
			logging.GetLogger().Warn("Auth middleware not found in config, injecting it")
			middlewares = append(middlewares, config_v1.Middleware_builder{
				Name:     proto.String(authMiddlewareName),
				Priority: proto.Int32(20), // Default priority
			}.Build())
		}
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

	if opts.Stdio {
		err := a.runStdioModeFunc(opts.Ctx, mcpSrv)
		workerCancel()
		upstreamWorker.Stop()
		registrationWorker.Stop()
		return err
	}

	bindAddress := opts.JSONRPCPort
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

	// Signal startup complete
	startupCallback := func() {
		a.startupOnce.Do(func() {
			close(a.startupCh)
		})
	}

	// Start servers
	if err := a.runServerMode(
		opts.Ctx,
		mcpSrv,
		busProvider,
		bindAddress,
		opts.GRPCPort,
		opts.ShutdownTimeout,
		cfg.GetGlobalSettings(),
		cachingMiddleware,
		standardMiddlewares,
		s,
		serviceRegistry,
		startupCallback,
		opts.TLSCert,
		opts.TLSKey,
		opts.TLSClientCA,
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
//
// Summary: Reloads application configuration from disk/storage.
//
// Parameters:
//   - ctx (context.Context): The context for the reload operation.
//   - fs (afero.Fs): The filesystem interface for reading configuration files.
//   - configPaths ([]string): A slice of paths to configuration files to reload.
//
// Returns:
//   - (error): An error if the configuration reload fails.
//
// Side Effects:
//   - Reads configuration files.
//   - Updates global settings, user auth, profiles, and service registry.
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

	// Read new config content first for diff generation
	newConfigRaw, readErr := a.readConfigFiles(fs, configPaths)
	if readErr != nil {
		log.Error("Failed to read config files for diff", "error", readErr)
	}

	cfg, err := a.loadConfig(ctx, fs, configPaths)
	a.lastReloadTime = time.Now()
	a.lastReloadErr = err
	if err != nil {
		metrics.IncrCounter([]string{"config", "reload", "errors"}, 1)
		// Generate Diff if we have previous good config and new config
		if newConfigRaw != nil && a.lastGoodConfig != nil {
			a.configDiff = a.generateConfigDiff(a.lastGoodConfig, newConfigRaw)
		}
		return fmt.Errorf("failed to load services from config: %w", err)
	}

	// Success: Update last good config and clear diff
	if newConfigRaw != nil {
		a.lastGoodConfig = newConfigRaw
		a.configDiff = ""
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

	if a.Storage != nil {
		stores = append(stores, a.Storage)
	}

	enableFileConfig := os.Getenv("MCPANY_ENABLE_FILE_CONFIG") == "true"
	if enableFileConfig && len(configPaths) > 0 {
		stores = append(stores, config.NewFileStore(fs, configPaths))
	}

	store := config.NewMultiStore(stores...)
	return config.LoadServices(ctx, store, "server")
}

func (a *Application) updateGlobalSettings(cfg *config_v1.McpAnyServerConfig) {
	log := logging.GetLogger()
	if a.SettingsManager != nil {
		a.SettingsManager.Update(cfg.GetGlobalSettings(), a.explicitAPIKey)
	}

	// Update log level
	if cfg.GetGlobalSettings().GetLogLevel() != 0 {
		newLevel := logging.ToSlogLevel(cfg.GetGlobalSettings().GetLogLevel())
		logging.SetLevel(newLevel)
		log.Info("Updated log level", "level", newLevel)
	}

	// Update Health Alerts
	if cfg.GetGlobalSettings().GetAlerts() != nil {
		health.SetGlobalAlertConfig(cfg.GetGlobalSettings().GetAlerts())
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
	if a.csrfMiddleware != nil {
		a.csrfMiddleware.Update(a.SettingsManager.GetAllowedOrigins())
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
//
//nolint:gocyclo // complexity is fine here
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

	// Auto-discovery of local services
	if cfg.GetGlobalSettings().GetAutoDiscoverLocal() {
		ollamaProvider := &discovery.OllamaProvider{Endpoint: "http://localhost:11434"}
		discovered, err := ollamaProvider.Discover(ctx)
		if err != nil {
			log.Warn("Failed to auto-discover local services during reload", "provider", ollamaProvider.Name(), "error", err)
		} else {
			for _, svc := range discovered {
				log.Info("Auto-discovered local service during reload", "name", svc.GetName())
				cfg.SetUpstreamServices(append(cfg.GetUpstreamServices(), svc))
			}
		}
	}

	// Map new services by name for easy lookup
	newServices := make(map[string]*config_v1.UpstreamServiceConfig)

	// Helper to deduplicate tools by name after proto.Merge appends them
	dedupTools := func(svc *config_v1.UpstreamServiceConfig) {
		if cmd := svc.GetCommandLineService(); cmd != nil {
			seen := make(map[string]bool)
			var deduplicated []*config_v1.ToolDefinition
			tools := cmd.GetTools()
			// Iterate backwards so the latter tools (from database override) take precedence
			for i := len(tools) - 1; i >= 0; i-- {
				t := tools[i]
				if !seen[t.GetName()] {
					seen[t.GetName()] = true
					// Prepend to keep original order with overrides
					deduplicated = append([]*config_v1.ToolDefinition{t}, deduplicated...)
				}
			}
			cmd.SetTools(deduplicated)
		}

		if http := svc.GetHttpService(); http != nil {
			seen := make(map[string]bool)
			var deduplicated []*config_v1.ToolDefinition
			tools := http.GetTools()
			for i := len(tools) - 1; i >= 0; i-- {
				t := tools[i]
				if !seen[t.GetName()] {
					seen[t.GetName()] = true
					deduplicated = append([]*config_v1.ToolDefinition{t}, deduplicated...)
				}
			}
			http.SetTools(deduplicated)
		}
	}
	if cfg.GetUpstreamServices() != nil {
		for _, svc := range cfg.GetUpstreamServices() {
			if !svc.GetDisable() {
				if existing, ok := newServices[svc.GetName()]; ok {
					proto.Merge(existing, svc)
					dedupTools(existing)
				} else {
					s := proto.Clone(svc).(*config_v1.UpstreamServiceConfig)
					dedupTools(s)
					newServices[svc.GetName()] = s
				}
			}
		}
	}
	if cfg.GetCollections() != nil {
		for _, collection := range cfg.GetCollections() {
			for _, svc := range collection.GetServices() {
				if svc.GetName() == "" {
					continue
				}
				if !svc.GetDisable() {
					if existing, ok := newServices[svc.GetName()]; ok {
						proto.Merge(existing, svc)
						dedupTools(existing)
					} else {
						s := proto.Clone(svc).(*config_v1.UpstreamServiceConfig)
						dedupTools(s)
						newServices[svc.GetName()] = s
					}
				}
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
				newSvcCopy.SetId(oldConfig.GetId())
			}
			if newSvcCopy.GetSanitizedName() == "" {
				newSvcCopy.SetSanitizedName(oldConfig.GetSanitizedName())
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

	// Update Auth Manager users
	users := cfg.GetUsers()
	if a.Storage != nil {
		dbUsers, err := a.Storage.ListUsers(ctx)
		if err != nil {
			log.Error("failed to list users from storage during reload", "error", err)
		} else {
			users = append(users, dbUsers...)
		}
	}
	if a.AuthManager != nil {
		a.AuthManager.SetUsers(users)
	}

	// Update Service Registry
	// We need to re-create services or update existing ones?
	// ServiceRegistry.UpdateServices?
	// Ideally we have a better way, but for now we might need to rely on individual updates or full re-init?
	// The ServiceRegistry holds state (tools, prompts).
	// If we just replace services, we might loose state.
	// But `UpdateServices` is not exposed on interface?
	// Actually `ServiceRegistry` is an interface.
	// Using `UpdateConfig` if available?
	// For now, let's assume `AuthManager` update is what we really needed for login.
	// Services are updated via bus or separate flow in real app usually.
	// But `server.go` logic for Reload needs to be checked.
	// For this task, updating AuthManager is sufficient for USER LOGIN.
}

// readConfigFiles reads the raw content of the configuration files.
// It handles directory walking similar to FileStore but only returns raw content.
func (a *Application) readConfigFiles(fs afero.Fs, paths []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, path := range paths {
		// Handle URL - skip for diffing for now as it requires network call and we only care about file changes mostly
		if strings.HasPrefix(strings.ToLower(path), "http://") || strings.HasPrefix(strings.ToLower(path), "https://") {
			continue
		}

		info, err := fs.Stat(path)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			err := afero.Walk(fs, path, func(p string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !fi.IsDir() {
					// Simple check for extension to match Config behavior roughly
					ext := strings.ToLower(filepath.Ext(p))
					if ext == ".yaml" || ext == ".yml" || ext == ".json" || ext == ".textproto" || ext == ".prototxt" {
						b, err := afero.ReadFile(fs, p)
						if err != nil {
							return err
						}
						result[p] = string(b)
					}
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			b, err := afero.ReadFile(fs, path)
			if err != nil {
				return nil, err
			}
			result[path] = string(b)
		}
	}
	return result, nil
}

// generateConfigDiff generates a unified diff between the old and new configuration files.
func (a *Application) generateConfigDiff(oldConfig, newConfig map[string]string) string {
	var diffs []string
	// Check for changed or new files
	for path, newContent := range newConfig {
		oldContent, exists := oldConfig[path]
		if !exists {
			// New file
			d, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(""),
				B:        difflib.SplitLines(newContent),
				FromFile: "/dev/null",
				ToFile:   path,
				Context:  3,
			})
			diffs = append(diffs, fmt.Sprintf("New file: %s\n%s", path, d))
		} else if oldContent != newContent {
			// Changed file
			d, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(oldContent),
				B:        difflib.SplitLines(newContent),
				FromFile: path + " (last known good)",
				ToFile:   path + " (current broken)",
				Context:  3,
			})
			diffs = append(diffs, d)
		}
	}
	// Check for deleted files
	for path, oldContent := range oldConfig {
		if _, exists := newConfig[path]; !exists {
			// Deleted file
			d, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
				A:        difflib.SplitLines(oldContent),
				B:        difflib.SplitLines(""),
				FromFile: path,
				ToFile:   "/dev/null",
				Context:  3,
			})
			diffs = append(diffs, fmt.Sprintf("Deleted file: %s\n%s", path, d))
		}
	}
	return strings.Join(diffs, "\n")
}

// WaitForStartup waits for the application to be fully initialized.
//
// Summary: Waits for application startup completion.
//
// It blocks until the startup process is complete or the context is canceled.
//
// Parameters:
//   - ctx (context.Context): The context to wait on.
//
// Returns:
//   - (error): nil if startup completes successfully, or a context error if canceled.
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
// Parameters:
//   - fs (afero.Fs): The filesystem to be validated.
//
// Returns:
//   - (afero.Fs): A non-nil afero.Fs.
//   - (error): An error if the provided filesystem is nil.
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
// Parameters:
//   - ctx (context.Context): The context for managing the server's lifecycle.
//   - mcpSrv (*mcpserver.Server): The MCP server instance to run.
//
// Returns:
//   - (error): An error if the server fails to run in stdio mode.
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
			Diff:    a.configDiff,
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

func (a *Application) filesystemHealthCheck(_ context.Context) health.CheckResult {
	if a.ServiceRegistry == nil {
		return health.CheckResult{Status: "ok"}
	}

	services, err := a.ServiceRegistry.GetAllServices()
	if err != nil {
		return health.CheckResult{
			Status:  "degraded",
			Message: fmt.Sprintf("failed to list services: %v", err),
		}
	}

	var issues []string
	start := time.Now()

	for _, svc := range services {
		fsSvc := svc.GetFilesystemService()
		if fsSvc == nil {
			continue
		}

		for virtualPath, localPath := range fsSvc.GetRootPaths() {
			if info, err := os.Stat(localPath); err != nil {
				issues = append(issues, fmt.Sprintf("service %q: root path %q (%s) is inaccessible: %v", svc.GetName(), virtualPath, localPath, err))
			} else if !info.IsDir() {
				issues = append(issues, fmt.Sprintf("service %q: root path %q (%s) is not a directory", svc.GetName(), virtualPath, localPath))
			}
		}
	}

	status := "ok"
	var message string
	if len(issues) > 0 {
		status = "degraded"
		message = strings.Join(issues, "; ")
	}

	return health.CheckResult{
		Status:  status,
		Message: message,
		Latency: time.Since(start).String(),
	}
}

// HealthCheck performs a health check against a running server.
//
// Summary: Checks the health of a running server.
//
// The function constructs the health check URL from the provided address and
// sends an HTTP GET request. It expects a 200 OK status code for a successful
// health check.
//
// Parameters:
//   - out (io.Writer): The writer to which the success message will be written.
//   - addr (string): The address (host:port) on which the server is running.
//   - timeout (time.Duration): The maximum duration to wait for the health check.
//
// Returns:
//   - (error): nil if healthy, or an error if the health check fails.
func HealthCheck(out io.Writer, addr string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return HealthCheckWithContext(ctx, out, addr)
}

// HealthCheckWithContext performs a health check against a running server with a context.
//
// Summary: Checks the health of a running server using a context.
//
// The function constructs the health check URL from the provided address and
// sends an HTTP GET request. It expects a 200 OK status code for a successful
// health check.
//
// Parameters:
//   - ctx (context.Context): The context for managing the health check's lifecycle.
//   - out (io.Writer): The writer to which the success message will be written.
//   - addr (string): The address (host:port) on which the server is running.
//
// Returns:
//   - (error): nil if healthy, or an error if the health check fails.
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
// Parameters:
//   - ctx (context.Context): The context for managing the server's lifecycle.
//   - mcpSrv (*mcpserver.Server): The MCP server instance.
//   - bus (*bus.Provider): The message bus for inter-component communication.
//   - bindAddress (string): The address for the HTTP/JSON-RPC server.
//   - grpcPort (string): The port for the gRPC registration server.
//   - shutdownTimeout (time.Duration): Duration to wait for graceful shutdown.
//   - globalSettings (*config_v1.GlobalSettings): Global configuration settings.
//   - cachingMiddleware (*middleware.CachingMiddleware): The caching middleware.
//   - standardMiddlewares (*middleware.StandardMiddlewares): The standard middleware chain.
//   - store (storage.Storage): The storage interface.
//   - serviceRegistry (*serviceregistry.ServiceRegistry): The service registry.
//   - startupCallback (func()): Callback function executed when servers are ready.
//   - tlsCert (string): Path to TLS certificate.
//   - tlsKey (string): Path to TLS key.
//   - tlsClientCA (string): Path to TLS Client CA.
//
// Returns:
//   - (error): An error if any of the servers fail to start or run.
//
//nolint:gocyclo

// createAuthMiddleware creates the authentication middleware.
func (a *Application) createAuthMiddleware(forcePrivateIPOnly bool, trustProxy bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow login endpoint without auth
			if r.URL.Path == "/api/v1/auth/login" {
				next.ServeHTTP(w, r)
				return
			}

			// Support passing Authorization via query parameter (essential for WebSockets)
			if r.Header.Get("Authorization") == "" {
				if authToken := r.URL.Query().Get("auth_token"); authToken != "" {
					// We assume Basic auth for now as that's what the UI uses for user login.
					// If the token doesn't start with "Basic " or "Bearer ", prepend "Basic ".
					if !strings.HasPrefix(authToken, "Basic ") && !strings.HasPrefix(authToken, "Bearer ") {
						r.Header.Set("Authorization", "Basic "+authToken)
					} else {
						r.Header.Set("Authorization", authToken)
					}
				}
			}

			ip := util.GetClientIP(r, trustProxy)
			ctx := util.ContextWithRemoteIP(r.Context(), ip)
			r = r.WithContext(ctx)
			apiKey := a.SettingsManager.GetAPIKey()
			requestKey := r.Header.Get("X-API-Key")
			logging.GetLogger().Info("DEBUG: AuthMiddleware details", "configured_key", apiKey, "request_key", requestKey, "path", r.URL.Path)
			authenticated := false

			// 1. Check Global API Key
			if apiKey != "" {
				requestKey := r.Header.Get("X-API-Key")
				if requestKey == "" {
					requestKey = r.URL.Query().Get("api_key")
				}
				if requestKey == "" {
					authHeader := r.Header.Get("Authorization")
					if strings.HasPrefix(authHeader, "Bearer ") {
						requestKey = strings.TrimPrefix(authHeader, "Bearer ")
					}
				}

				if subtle.ConstantTimeCompare([]byte(requestKey), []byte(apiKey)) == 1 {
					authenticated = true
					// Inject API Key into context if needed
					ctx = auth.ContextWithAPIKey(ctx, requestKey)
					// Global API Key grants Admin privileges (Root Access)
					ctx = auth.ContextWithRoles(ctx, []string{"admin"})
					// Also inject a placeholder user ID so that handlers expecting a user context don't fail
					ctx = auth.ContextWithUser(ctx, "system-admin")
				}
			}

			// 2. Check User Authentication (Basic Auth)
			if !authenticated {
				username, _, ok := r.BasicAuth()
				if ok && a.AuthManager != nil {
					if user, found := a.AuthManager.GetUser(username); found {
						if err := auth.ValidateAuthentication(ctx, user.GetAuthentication(), r); err == nil {
							authenticated = true
							ctx = auth.ContextWithUser(ctx, username)
							if len(user.GetRoles()) > 0 {
								ctx = auth.ContextWithRoles(ctx, user.GetRoles())
							}
						}
					}
				}
			}

			if authenticated {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if !forcePrivateIPOnly && apiKey != "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Sentinel Security: If no API key is configured (and no user auth succeeded), enforce localhost-only access.
			// This prevents accidental exposure of the server to the public internet (RCE risk).
			if apiKey == "" {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					// Fallback if RemoteAddr is weird, assume host is the string itself
					host = r.RemoteAddr
				}

				// Check if the request is from a loopback address
				ipAddr := net.ParseIP(host)
				if !util.IsPrivateIP(ipAddr) {
					logging.GetLogger().Warn("Blocked public internet request because no API Key is configured", "remote_addr", r.RemoteAddr)
					http.Error(w, "Forbidden: Public access requires an API Key to be configured", http.StatusForbidden)
					return
				}

				// Grant Admin privileges (Root Access) for local development/testing convenience
				// when running in insecure mode (private network, no API key).
				ctx = auth.ContextWithRoles(ctx, []string{"admin"})
				ctx = auth.ContextWithUser(ctx, "system-admin")
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// HTTPRequestContextMiddleware injects the HTTP request into the context.
//
// Summary: Middleware to add HTTP request to context.
//
// Parameters:
//   - next (http.Handler): The next handler.
//
// Returns:
//   - (http.Handler): The wrapped handler.
func (a *Application) HTTPRequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), middleware.HTTPRequestContextKey, r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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

// GetAuditMiddleware returns the current audit middleware.
//
// Summary: Returns the active audit middleware.
//
// Returns:
//   - *middleware.AuditMiddleware: The current audit middleware instance.
func (a *Application) GetAuditMiddleware() *middleware.AuditMiddleware {
	a.configMu.Lock()
	defer a.configMu.Unlock()
	if a.standardMiddlewares != nil {
		return a.standardMiddlewares.Audit
	}
	return nil
}

// wrapBindError checks if the error is a port conflict and returns a user-friendly error message.
func wrapBindError(err error, serverType, address, flag string) error {
	if strings.Contains(err.Error(), "address already in use") || strings.Contains(err.Error(), "bind: permission denied") {
		return fmt.Errorf("❌ %s server failed to listen on %s: %w\n\n💡 Tip: The port is already in use or restricted. Try using a different port:\n   mcpany run %s <new_port>", serverType, address, err, flag)
	}
	return fmt.Errorf("%s server failed to listen: %w", serverType, err)
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
	connState func(net.Conn, http.ConnState),
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
			ConnState: func(c net.Conn, state http.ConnState) {
				if connState != nil {
					connState(c, state)
				}
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
