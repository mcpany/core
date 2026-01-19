// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// BinaryType defines the type of the binary being validated.
type BinaryType int

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
)

const (
	// Server represents the server binary.
	Server BinaryType = iota
	// Worker represents the worker binary.
	Worker
	// Client represents the client binary.
	Client
)

var (
	osStat       = os.Stat
	execLookPath = exec.LookPath
)

// ValidationError encapsulates a validation error for a specific service.
type ValidationError struct {
	ServiceName string
	Err         error
}

// Error returns the formatted error message.
//
// Returns the result.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("service %q: %v", e.ServiceName, e.Err)
}

// Validate inspects the given McpAnyServerConfig for correctness and consistency.
// It iterates through the list of upstream services, checking for valid
// service definitions, addresses, cache settings, and authentication
// configurations.
//
// Invalid services are not removed from the configuration; instead, a list of
// validation errors is returned.
//
// Parameters:
//
//	ctx: The context for the validation (used for secret resolution).
//	config: The server configuration to be validated.
//	binaryType: The type of binary (server, worker) which might affect validation rules.
//
// Returns:
//
//	[]ValidationError: A slice of ValidationErrors, which will be empty if the configuration is valid.
func Validate(ctx context.Context, config *configv1.McpAnyServerConfig, binaryType BinaryType) []ValidationError {
	var validationErrors []ValidationError
	serviceNames := make(map[string]bool)

	if gs := config.GetGlobalSettings(); gs != nil {
		if err := validateGlobalSettings(gs, binaryType); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				ServiceName: "global_settings",
				Err:         err,
			})
		}
	}

	userIDs := make(map[string]bool)
	for _, user := range config.GetUsers() {
		if user.GetId() == "" {
			validationErrors = append(validationErrors, ValidationError{
				ServiceName: "user",
				Err:         fmt.Errorf("user has empty id"),
			})
			continue
		}
		if userIDs[user.GetId()] {
			validationErrors = append(validationErrors, ValidationError{
				ServiceName: fmt.Sprintf("user:%s", user.GetId()),
				Err:         fmt.Errorf("duplicate user id"),
			})
		}
		userIDs[user.GetId()] = true

		if err := validateUser(ctx, user); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				ServiceName: fmt.Sprintf("user:%s", user.GetId()),
				Err:         err,
			})
		}
	}

	for _, service := range config.GetUpstreamServices() {
		if _, exists := serviceNames[service.GetName()]; exists {
			validationErrors = append(validationErrors, ValidationError{
				ServiceName: service.GetName(),
				Err:         fmt.Errorf("duplicate service name found"),
			})
		}
		serviceNames[service.GetName()] = true

		if err := validateUpstreamService(ctx, service); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				ServiceName: service.GetName(),
				Err:         err,
			})
		}
	}

	for _, collection := range config.GetCollections() {
		if err := validateCollection(ctx, collection); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				ServiceName: collection.GetName(),
				Err:         err,
			})
		}
	}

	return validationErrors
}

func validateCollection(ctx context.Context, collection *configv1.Collection) error {
	if collection.GetName() == "" {
		return fmt.Errorf("collection name is empty")
	}
	if collection.GetHttpUrl() == "" {
		if len(collection.GetServices()) == 0 && len(collection.GetSkills()) == 0 {
			return fmt.Errorf("collection must have either http_url or inline content (services/skills)")
		}
		// If content is present, HttpUrl is optional (inline collection)
	} else {
		// If HttpUrl is present, validate it
		if !validation.IsValidURL(collection.GetHttpUrl()) {
			return fmt.Errorf("invalid collection http_url: %s", collection.GetHttpUrl())
		}
		u, _ := url.Parse(collection.GetHttpUrl())
		if u.Scheme != schemeHTTP && u.Scheme != schemeHTTPS {
			return fmt.Errorf("invalid collection http_url scheme: %s", u.Scheme)
		}
	}

	if authConfig := collection.GetAuthentication(); authConfig != nil {
		if err := validateUpstreamAuthentication(ctx, authConfig); err != nil {
			return err
		}
	}
	return nil
}

func validateStdioArgs(command string, args []string, workingDir string) error {
	// Only check for common interpreters to avoid false positives on arbitrary flags
	baseCmd := filepath.Base(command)
	isInterpreter := false
	// List of common interpreters that take a script file as an argument
	interpreters := []string{"python", "python3", "node", "deno", "bun", "ruby", "perl", "bash", "sh", "zsh", "go"}
	for _, i := range interpreters {
		if baseCmd == i || strings.HasPrefix(baseCmd, i) { // e.g. python3.9
			isInterpreter = true
			break
		}
	}

	if !isInterpreter {
		return nil
	}

	isPython := strings.HasPrefix(baseCmd, "python")

	// Find the first non-flag argument
	for _, arg := range args {
		// Skip flags (start with -)
		if strings.HasPrefix(arg, "-") {
			// Special case for Python -m (module execution)
			// If we see -m, the next argument is a module name which might look like a file (e.g. http.server)
			// but shouldn't be validated as a file on disk.
			if isPython && arg == "-m" {
				return nil
			}
			continue
		}

		// Check for URLs (Deno/Bun/remote scripts)
		// These are valid arguments for some interpreters but not local files.
		if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
			return nil
		}

		// This is likely the script file.
		// We only validate it if it looks like a file (has an extension).
		// This avoids validating things like "install" or module names.
		ext := filepath.Ext(arg)
		if ext == "" {
			continue
		}

		// Check if file exists
		if err := validateFileExists(arg, workingDir); err != nil {
			return WrapActionableError(fmt.Sprintf("argument %q looks like a script file but does not exist", arg), err)
		}

		// We only check the FIRST script argument for interpreters.
		// Subsequent args are likely arguments to the script itself.
		return nil
	}
	return nil
}

func validateFileExists(path string, workingDir string) error {
	targetPath := path
	// If path is relative and workingDir is set, join them.
	// Note: If workingDir is empty, we check relative to CWD (process root), which is standard behavior.
	if !filepath.IsAbs(path) && workingDir != "" {
		targetPath = filepath.Join(workingDir, path)
	}

	info, err := osStat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			absPath, _ := filepath.Abs(targetPath)
			return &ActionableError{
				Err:        fmt.Errorf("file not found: %s", targetPath),
				Suggestion: fmt.Sprintf("Check if the file exists at %q (absolute: %q) and the server process has read permissions.", targetPath, absPath),
			}
		}
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("is a directory, expected a file")
	}
	return nil
}

func validateSecretMap(secrets map[string]*configv1.SecretValue) error {
	for key, secret := range secrets {
		if err := validateSecretValue(secret); err != nil {
			return fmt.Errorf("%q: %w", key, err)
		}
	}
	return nil
}

func validateSecretValue(secret *configv1.SecretValue) error {
	if secret == nil {
		return nil
	}
	switch secret.WhichValue() {
	case configv1.SecretValue_EnvironmentVariable_case:
		envVar := secret.GetEnvironmentVariable()
		if _, exists := os.LookupEnv(envVar); !exists {
			return &ActionableError{
				Err:        fmt.Errorf("environment variable %q is not set", envVar),
				Suggestion: fmt.Sprintf("Set the environment variable %q in your shell or .env file before starting the server.", envVar),
			}
		}
	case configv1.SecretValue_FilePath_case:
		if err := validation.IsAllowedPath(secret.GetFilePath()); err != nil {
			return fmt.Errorf("invalid secret file path %q: %w", secret.GetFilePath(), err)
		}
		// Validate that the file actually exists to fail fast
		if err := validation.FileExists(secret.GetFilePath()); err != nil {
			return fmt.Errorf("secret file %q does not exist: %w", secret.GetFilePath(), err)
		}
	case configv1.SecretValue_RemoteContent_case:
		remote := secret.GetRemoteContent()
		if remote.GetHttpUrl() == "" {
			return fmt.Errorf("remote secret has empty http_url")
		}
		if !validation.IsValidURL(remote.GetHttpUrl()) {
			return fmt.Errorf("remote secret has invalid http_url: %s", remote.GetHttpUrl())
		}
		u, _ := url.Parse(remote.GetHttpUrl())
		if u.Scheme != schemeHTTP && u.Scheme != schemeHTTPS {
			return fmt.Errorf("remote secret has invalid http_url scheme: %s", u.Scheme)
		}
	}
	return nil
}

func validateGlobalSettings(gs *configv1.GlobalSettings, binaryType BinaryType) error {
	switch binaryType {
	case Server:
		if gs.GetMcpListenAddress() != "" {
			if err := validation.IsValidBindAddress(gs.GetMcpListenAddress()); err != nil {
				return fmt.Errorf("invalid mcp_listen_address: %w", err)
			}
		}
	case Client:
		if gs.GetApiKey() != "" && len(gs.GetApiKey()) < 16 {
			return fmt.Errorf("API key must be at least 16 characters long")
		}
	}

	if bus := gs.GetMessageBus(); bus != nil {
		if redis := bus.GetRedis(); redis != nil {
			if redis.GetAddress() == "" {
				return fmt.Errorf("redis message bus address is empty")
			}
		}
	}

	if err := validateAuditConfig(gs.GetAudit()); err != nil {
		return fmt.Errorf("audit config error: %w", err)
	}

	if err := validateDLPConfig(gs.GetDlp()); err != nil {
		return fmt.Errorf("dlp config error: %w", err)
	}

	if err := validateGCSettings(gs.GetGcSettings()); err != nil {
		return fmt.Errorf("gc settings error: %w", err)
	}

	profileNames := make(map[string]bool)
	for _, profile := range gs.GetProfileDefinitions() {
		if profile.GetName() == "" {
			return fmt.Errorf("profile definition has empty name")
		}
		if profileNames[profile.GetName()] {
			return fmt.Errorf("duplicate profile definition name: %s", profile.GetName())
		}
		profileNames[profile.GetName()] = true
		if err := validateProfileDefinition(profile); err != nil {
			return fmt.Errorf("profile definition %q error: %w", profile.GetName(), err)
		}
	}

	return nil
}

// ValidateOrError validates a single upstream service configuration and returns an error if it's invalid.
//
// Parameters:
//
//	ctx: The context for the validation.
//	service: The upstream service configuration to validate.
//
// Returns:
//
//	error: An error if validation fails.
func ValidateOrError(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	return validateUpstreamService(ctx, service)
}

func validateUpstreamService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	if service.GetName() == "" {
		return fmt.Errorf("service name is empty")
	}

	if service.WhichServiceConfig() == configv1.UpstreamServiceConfig_ServiceConfig_not_set_case {
		return fmt.Errorf("service type not specified")
	}

	if err := validateServiceConfig(service); err != nil {
		return err
	}

	if service.GetCache() != nil {
		if service.GetCache().GetTtl() != nil && service.GetCache().GetTtl().GetSeconds() < 0 {
			return fmt.Errorf("invalid cache timeout: %v", service.GetCache().GetTtl().AsDuration())
		}
	}

	if authConfig := service.GetUpstreamAuth(); authConfig != nil {
		if err := validateAuthentication(ctx, authConfig); err != nil {
			return err
		}
	}
	return nil
}

// ... (other functions)

func validateAuthentication(ctx context.Context, authConfig *configv1.Authentication) error {
	switch authConfig.WhichAuthMethod() {
	case configv1.Authentication_ApiKey_case:
		return validateAPIKeyAuth(ctx, authConfig.GetApiKey())
	case configv1.Authentication_BearerToken_case:
		return validateBearerTokenAuth(ctx, authConfig.GetBearerToken())
	case configv1.Authentication_BasicAuth_case:
		return validateBasicAuth(ctx, authConfig.GetBasicAuth())
	case configv1.Authentication_Mtls_case:
		return validateMtlsAuth(authConfig.GetMtls())
	case configv1.Authentication_Oauth2_case:
		return validateOAuth2Auth(ctx, authConfig.GetOauth2())
	case configv1.Authentication_Oidc_case:
		return validateOIDCAuth(ctx, authConfig.GetOidc())
	case configv1.Authentication_TrustedHeader_case:
		return validateTrustedHeaderAuth(authConfig.GetTrustedHeader())
	}
	return nil
}

func validateServiceConfig(service *configv1.UpstreamServiceConfig) error {
	if httpService := service.GetHttpService(); httpService != nil {
		return validateHTTPService(httpService)
	} else if websocketService := service.GetWebsocketService(); websocketService != nil {
		return validateWebSocketService(websocketService)
	} else if grpcService := service.GetGrpcService(); grpcService != nil {
		return validateGrpcService(grpcService)
	} else if openapiService := service.GetOpenapiService(); openapiService != nil {
		return validateOpenAPIService(openapiService)
	} else if commandLineService := service.GetCommandLineService(); commandLineService != nil {
		return validateCommandLineService(commandLineService)
	} else if mcpService := service.GetMcpService(); mcpService != nil {
		return validateMcpService(mcpService)
	} else if sqlService := service.GetSqlService(); sqlService != nil {
		return validateSQLService(sqlService)
	} else if graphqlService := service.GetGraphqlService(); graphqlService != nil {
		return validateGraphQLService(graphqlService)
	} else if webrtcService := service.GetWebrtcService(); webrtcService != nil {
		return validateWebrtcService(webrtcService)
	}
	return nil
}

func validateHTTPService(httpService *configv1.HttpUpstreamService) error {
	if httpService.GetAddress() == "" {
		return &ActionableError{
			Err:        fmt.Errorf("http service has empty address"),
			Suggestion: "Set the 'address' field in the http_service configuration (e.g., 'http://localhost:8080').",
		}
	}
	if !validation.IsValidURL(httpService.GetAddress()) {
		return &ActionableError{
			Err:        fmt.Errorf("invalid http address: %s", httpService.GetAddress()),
			Suggestion: "Ensure the address is a valid URL (e.g., 'http://example.com').",
		}
	}
	u, _ := url.Parse(httpService.GetAddress())
	if u.Scheme != schemeHTTP && u.Scheme != schemeHTTPS {
		return &ActionableError{
			Err:        fmt.Errorf("invalid http address scheme: %s", u.Scheme),
			Suggestion: "Use 'http' or 'https' as the scheme (e.g., http://example.com).",
		}
	}

	for name, call := range httpService.GetCalls() {
		if err := validateSchema(call.GetInputSchema()); err != nil {
			return WrapActionableError(fmt.Sprintf("http call %q input_schema error", name), err)
		}
		if err := validateSchema(call.GetOutputSchema()); err != nil {
			return WrapActionableError(fmt.Sprintf("http call %q output_schema error", name), err)
		}
	}
	return nil
}

func validateWebSocketService(websocketService *configv1.WebsocketUpstreamService) error {
	if websocketService.GetAddress() == "" {
		return &ActionableError{
			Err:        fmt.Errorf("websocket service has empty address"),
			Suggestion: "Set the 'address' field in the websocket_service configuration (e.g., 'ws://localhost:8080').",
		}
	}
	if !validation.IsValidURL(websocketService.GetAddress()) {
		return &ActionableError{
			Err:        fmt.Errorf("invalid websocket address: %s", websocketService.GetAddress()),
			Suggestion: "Ensure the address is a valid URL (e.g., 'ws://example.com').",
		}
	}
	u, _ := url.Parse(websocketService.GetAddress())
	if u.Scheme != "ws" && u.Scheme != "wss" {
		return &ActionableError{
			Err:        fmt.Errorf("invalid websocket address scheme: %s", u.Scheme),
			Suggestion: "Use 'ws' or 'wss' as the scheme.",
		}
	}

	for name, call := range websocketService.GetCalls() {
		if err := validateSchema(call.GetInputSchema()); err != nil {
			return WrapActionableError(fmt.Sprintf("websocket call %q input_schema error", name), err)
		}
		if err := validateSchema(call.GetOutputSchema()); err != nil {
			return WrapActionableError(fmt.Sprintf("websocket call %q output_schema error", name), err)
		}
	}
	return nil
}

func validateGrpcService(grpcService *configv1.GrpcUpstreamService) error {
	if grpcService.GetAddress() == "" {
		return &ActionableError{
			Err:        fmt.Errorf("gRPC service has empty address"),
			Suggestion: "Set the 'address' field in the grpc_service configuration (e.g., 'localhost:50051').",
		}
	}

	if !grpcService.GetUseReflection() && len(grpcService.GetProtoDefinitions()) == 0 && len(grpcService.GetProtoCollection()) == 0 {
		return &ActionableError{
			Err:        fmt.Errorf("gRPC service requires proto definitions when reflection is disabled"),
			Suggestion: "Enable 'use_reflection: true' or provide 'proto_definitions' / 'proto_collection'.",
		}
	}

	for name, call := range grpcService.GetCalls() {
		if err := validateSchema(call.GetInputSchema()); err != nil {
			return WrapActionableError(fmt.Sprintf("grpc call %q input_schema error", name), err)
		}
		if err := validateSchema(call.GetOutputSchema()); err != nil {
			return WrapActionableError(fmt.Sprintf("grpc call %q output_schema error", name), err)
		}
	}
	return nil
}

func validateOpenAPIService(openapiService *configv1.OpenapiUpstreamService) error {
	if openapiService.GetAddress() == "" && openapiService.GetSpecContent() == "" && openapiService.GetSpecUrl() == "" {
		return &ActionableError{
			Err:        fmt.Errorf("openapi service must have either an address, spec content or spec url"),
			Suggestion: "Provide one of 'address', 'spec_content', or 'spec_url' in the openapi_service configuration.",
		}
	}
	if openapiService.GetAddress() != "" && !validation.IsValidURL(openapiService.GetAddress()) {
		return &ActionableError{
			Err:        fmt.Errorf("invalid openapi address: %s", openapiService.GetAddress()),
			Suggestion: "Ensure the address is a valid URL.",
		}
	}
	if openapiService.GetSpecUrl() != "" && !validation.IsValidURL(openapiService.GetSpecUrl()) {
		return &ActionableError{
			Err:        fmt.Errorf("invalid openapi spec_url: %s", openapiService.GetSpecUrl()),
			Suggestion: "Ensure the spec_url is a valid URL.",
		}
	}
	return nil
}

func validateCommandLineService(commandLineService *configv1.CommandLineUpstreamService) error {
	if commandLineService.GetCommand() == "" {
		return &ActionableError{
			Err:        fmt.Errorf("command_line_service has empty command"),
			Suggestion: "Set the 'command' field (e.g., './my-script.sh' or 'python my_script.py').",
		}
	}

	// Only validate command existence if not running in a container
	if commandLineService.GetContainerEnvironment().GetImage() == "" {
		if err := validateCommandExists(commandLineService.GetCommand(), commandLineService.GetWorkingDirectory()); err != nil {
			return WrapActionableError("command_line_service command validation failed", err)
		}
	}

	if err := validateContainerEnvironment(commandLineService.GetContainerEnvironment()); err != nil {
		return err
	}
	return nil
}

func validateContainerEnvironment(env *configv1.ContainerEnvironment) error {
	if env == nil {
		return nil
	}
	// We only validate volumes if an image is specified, as they are only used with Docker execution.
	if env.GetImage() != "" {
		for dest, src := range env.GetVolumes() {
			if dest == "" {
				return fmt.Errorf("container environment volume host path is empty")
			}
			if src == "" {
				return fmt.Errorf("container environment volume container path is empty")
			}
			// dest is the key (Host Path), src is the value (Container Path).
			// We must validate the Host Path (dest) to ensure it is secure.
			// It must be either relative to the CWD or in the allowed list.
			if err := validation.IsAllowedPath(dest); err != nil {
				return fmt.Errorf("container environment volume host path %q is not a secure path: %w", dest, err)
			}
		}
	}

	return nil
}

func validateMcpService(mcpService *configv1.McpUpstreamService) error {
	switch mcpService.WhichConnectionType() {
	case configv1.McpUpstreamService_HttpConnection_case:
		httpConn := mcpService.GetHttpConnection()
		if httpConn.GetHttpAddress() == "" {
			return &ActionableError{
				Err:        fmt.Errorf("mcp service with http_connection has empty http_address"),
				Suggestion: "Set the 'http_address' field (e.g., 'http://localhost:8080').",
			}
		}
		if !validation.IsValidURL(httpConn.GetHttpAddress()) {
			return &ActionableError{
				Err:        fmt.Errorf("mcp service with http_connection has invalid http_address: %s", httpConn.GetHttpAddress()),
				Suggestion: "Ensure the http_address is a valid URL.",
			}
		}
	case configv1.McpUpstreamService_StdioConnection_case:
		stdioConn := mcpService.GetStdioConnection()
		if len(stdioConn.GetCommand()) == 0 {
			return &ActionableError{
				Err:        fmt.Errorf("mcp service with stdio_connection has empty command"),
				Suggestion: "Set the 'command' field (e.g., 'npx', 'python', or path to executable).",
			}
		}

		// If running in Docker (container_image is set), we don't enforce host path/command restrictions
		if stdioConn.GetContainerImage() == "" {
			if err := validateCommandExists(stdioConn.GetCommand(), stdioConn.GetWorkingDirectory()); err != nil {
				return WrapActionableError("mcp service with stdio_connection command validation failed", err)
			}

			if err := validateStdioArgs(stdioConn.GetCommand(), stdioConn.GetArgs(), stdioConn.GetWorkingDirectory()); err != nil {
				return WrapActionableError("mcp service with stdio_connection argument validation failed", err)
			}

			if stdioConn.GetWorkingDirectory() != "" {
				if err := validation.IsAllowedPath(stdioConn.GetWorkingDirectory()); err != nil {
					return fmt.Errorf("mcp service with stdio_connection has insecure working_directory %q: %w", stdioConn.GetWorkingDirectory(), err)
				}
				if err := validateDirectoryExists(stdioConn.GetWorkingDirectory()); err != nil {
					return fmt.Errorf("mcp service with stdio_connection working_directory validation failed: %w", err)
				}
			}
		}

		if err := validateSecretMap(stdioConn.GetEnv()); err != nil {
			return fmt.Errorf("mcp service with stdio_connection has invalid secret environment variable: %w", err)
		}
	case configv1.McpUpstreamService_BundleConnection_case:
		bundleConn := mcpService.GetBundleConnection()
		if bundleConn.GetBundlePath() == "" {
			return fmt.Errorf("mcp service with bundle_connection has empty bundle_path")
		}
		if err := validation.IsAllowedPath(bundleConn.GetBundlePath()); err != nil {
			return fmt.Errorf("mcp service with bundle_connection has insecure bundle_path %q: %w", bundleConn.GetBundlePath(), err)
		}
		if err := validateSecretMap(bundleConn.GetEnv()); err != nil {
			return fmt.Errorf("mcp service with bundle_connection has invalid secret environment variable: %w", err)
		}
	default:
		return fmt.Errorf("mcp service has no connection_type")
	}

	for name, call := range mcpService.GetCalls() {
		if err := validateSchema(call.GetInputSchema()); err != nil {
			return WrapActionableError(fmt.Sprintf("mcp call %q input_schema error", name), err)
		}
		if err := validateSchema(call.GetOutputSchema()); err != nil {
			return WrapActionableError(fmt.Sprintf("mcp call %q output_schema error", name), err)
		}
	}
	return nil
}

func validateSQLService(sqlService *configv1.SqlUpstreamService) error {
	if sqlService.GetDriver() == "" {
		return &ActionableError{
			Err:        fmt.Errorf("sql service has empty driver"),
			Suggestion: "Set the 'driver' field (e.g., 'postgres', 'mysql', 'sqlite3').",
		}
	}
	if sqlService.GetDsn() == "" {
		return &ActionableError{
			Err:        fmt.Errorf("sql service has empty dsn"),
			Suggestion: "Set the 'dsn' (Data Source Name) field with connection details.",
		}
	}

	for name, call := range sqlService.GetCalls() {
		if call.GetQuery() == "" {
			return fmt.Errorf("sql call %q query is empty", name)
		}
		if err := validateSchema(call.GetInputSchema()); err != nil {
			return WrapActionableError(fmt.Sprintf("sql call %q input_schema error", name), err)
		}
		if err := validateSchema(call.GetOutputSchema()); err != nil {
			return WrapActionableError(fmt.Sprintf("sql call %q output_schema error", name), err)
		}
	}
	return nil
}

func validateGraphQLService(graphqlService *configv1.GraphQLUpstreamService) error {
	if graphqlService.GetAddress() == "" {
		return &ActionableError{
			Err:        fmt.Errorf("graphql service has empty address"),
			Suggestion: "Set the 'address' field in the graphql_service configuration.",
		}
	}
	if !validation.IsValidURL(graphqlService.GetAddress()) {
		return &ActionableError{
			Err:        fmt.Errorf("invalid graphql address: %s", graphqlService.GetAddress()),
			Suggestion: "Ensure the address is a valid URL.",
		}
	}
	u, _ := url.Parse(graphqlService.GetAddress())
	if u.Scheme != schemeHTTP && u.Scheme != schemeHTTPS {
		return &ActionableError{
			Err:        fmt.Errorf("invalid graphql address scheme: %s", u.Scheme),
			Suggestion: "Use 'http' or 'https' as the scheme.",
		}
	}
	return nil
}

func validateWebrtcService(webrtcService *configv1.WebrtcUpstreamService) error {
	if webrtcService.GetAddress() == "" {
		return &ActionableError{
			Err:        fmt.Errorf("webrtc service has empty address"),
			Suggestion: "Set the 'address' field in the webrtc_service configuration.",
		}
	}
	if !validation.IsValidURL(webrtcService.GetAddress()) {
		return &ActionableError{
			Err:        fmt.Errorf("invalid webrtc address: %s", webrtcService.GetAddress()),
			Suggestion: "Ensure the address is a valid URL.",
		}
	}
	u, _ := url.Parse(webrtcService.GetAddress())
	if u.Scheme != schemeHTTP && u.Scheme != schemeHTTPS {
		return &ActionableError{
			Err:        fmt.Errorf("invalid webrtc address scheme: %s", u.Scheme),
			Suggestion: "Use 'http' or 'https' as the scheme.",
		}
	}
	return nil
}

func validateUpstreamAuthentication(ctx context.Context, authConfig *configv1.Authentication) error {
	switch authConfig.WhichAuthMethod() {
	case configv1.Authentication_ApiKey_case:
		return validateAPIKeyAuth(ctx, authConfig.GetApiKey())
	case configv1.Authentication_BearerToken_case:
		return validateBearerTokenAuth(ctx, authConfig.GetBearerToken())
	case configv1.Authentication_BasicAuth_case:
		return validateBasicAuth(ctx, authConfig.GetBasicAuth())
	case configv1.Authentication_Mtls_case:
		return validateMtlsAuth(authConfig.GetMtls())
	case configv1.Authentication_Oauth2_case:
		return validateOAuth2Auth(ctx, authConfig.GetOauth2())
	}
	return nil
}

func validateAPIKeyAuth(ctx context.Context, apiKey *configv1.APIKeyAuth) error {
	if apiKey.GetParamName() == "" {
		return &ActionableError{
			Err:        fmt.Errorf("api key 'param_name' is empty"),
			Suggestion: "Set the 'param_name' (e.g., 'X-API-Key' or 'api_key').",
		}
	}
	// Value might be nil but usually we want to validate it resolves if present,
	// or ensure it IS present if required. For upstream, it's usually required.
	if apiKey.GetValue() == nil && apiKey.GetVerificationValue() == "" {
		// If both are missing, it's invalid?
		// VerificationValue is for incoming, Value (SecretValue) is for outgoing.
		// Detailed validation might depend on context (incoming vs outgoing),
		// but here we are validating "UpstreamServiceConfig" which implies outgoing.
		return &ActionableError{
			Err:        fmt.Errorf("api key 'value' is missing (required for outgoing auth)"),
			Suggestion: "Set the 'value' field using a secret (environment variable or file) for the API key.",
		}
	}
	if apiKey.GetValue() != nil {
		apiKeyValue, err := util.ResolveSecret(ctx, apiKey.GetValue())
		if err != nil {
			return fmt.Errorf("failed to resolve api key secret: %w", err)
		}
		if apiKeyValue == "" {
			return &ActionableError{
				Err:        fmt.Errorf("resolved api key value is empty"),
				Suggestion: "Check that the environment variable or file providing the API key is not empty.",
			}
		}
	}
	return nil
}

func validateBearerTokenAuth(ctx context.Context, bearerToken *configv1.BearerTokenAuth) error {
	tokenValue, err := util.ResolveSecret(ctx, bearerToken.GetToken())
	if err != nil {
		return fmt.Errorf("failed to resolve bearer token secret: %w", err)
	}
	if tokenValue == "" {
		return &ActionableError{
			Err:        fmt.Errorf("bearer token 'token' is empty"),
			Suggestion: "Check that the environment variable or file providing the Bearer token is not empty.",
		}
	}
	return nil
}

func validateBasicAuth(ctx context.Context, basicAuth *configv1.BasicAuth) error {
	if basicAuth.GetUsername() == "" {
		return &ActionableError{
			Err:        fmt.Errorf("basic auth 'username' is empty"),
			Suggestion: "Set the 'username' field.",
		}
	}
	passwordValue, err := util.ResolveSecret(ctx, basicAuth.GetPassword())
	if err != nil {
		return fmt.Errorf("failed to resolve basic auth password secret: %w", err)
	}
	if passwordValue == "" {
		return &ActionableError{
			Err:        fmt.Errorf("basic auth 'password' is empty"),
			Suggestion: "Check that the environment variable or file providing the password is not empty.",
		}
	}
	return nil
}

func validateOAuth2Auth(ctx context.Context, oauth *configv1.OAuth2Auth) error {
	if oauth.GetTokenUrl() == "" {
		if oauth.GetIssuerUrl() == "" {
			return &ActionableError{
				Err:        fmt.Errorf("oauth2 token_url is empty and no issuer_url provided"),
				Suggestion: "Provide 'token_url' OR 'issuer_url' to enable auto-discovery.",
			}
		}
		if !validation.IsValidURL(oauth.GetIssuerUrl()) {
			return fmt.Errorf("invalid oauth2 issuer_url: %s", oauth.GetIssuerUrl())
		}
		// If IssuerURL is present and valid, we allow TokenUrl to be empty (auto-discovery)
	} else if !validation.IsValidURL(oauth.GetTokenUrl()) {
		// Only validate TokenUrl if it is present
		return fmt.Errorf("invalid oauth2 token_url: %s", oauth.GetTokenUrl())
	}

	clientID, err := util.ResolveSecret(ctx, oauth.GetClientId())
	if err != nil {
		return fmt.Errorf("failed to resolve oauth2 client_id: %w", err)
	}
	if clientID == "" {
		return fmt.Errorf("oauth2 client_id is missing or empty")
	}

	clientSecret, err := util.ResolveSecret(ctx, oauth.GetClientSecret())
	if err != nil {
		return fmt.Errorf("failed to resolve oauth2 client_secret: %w", err)
	}
	if clientSecret == "" {
		return fmt.Errorf("oauth2 client_secret is missing or empty")
	}

	return nil
}

func validateOIDCAuth(_ context.Context, oidc *configv1.OIDCAuth) error {
	if oidc.GetIssuer() == "" {
		return fmt.Errorf("oidc issuer is empty")
	}
	if !validation.IsValidURL(oidc.GetIssuer()) {
		return fmt.Errorf("invalid oidc issuer url: %s", oidc.GetIssuer())
	}
	return nil
}

func validateTrustedHeaderAuth(th *configv1.TrustedHeaderAuth) error {
	if th.GetHeaderName() == "" {
		return fmt.Errorf("trusted header name is empty")
	}
	if th.GetHeaderValue() == "" {
		return fmt.Errorf("trusted header value is empty")
	}
	return nil
}

func validateMtlsAuth(mtls *configv1.MTLSAuth) error {
	if mtls.GetClientCertPath() == "" {
		return fmt.Errorf("mtls 'client_cert_path' is empty")
	}
	if mtls.GetClientKeyPath() == "" {
		return fmt.Errorf("mtls 'client_key_path' is empty")
	}
	if err := validation.IsAllowedPath(mtls.GetClientCertPath()); err != nil {
		return fmt.Errorf("mtls 'client_cert_path' is not a secure path: %w", err)
	}
	if err := validation.IsAllowedPath(mtls.GetClientKeyPath()); err != nil {
		return fmt.Errorf("mtls 'client_key_path' is not a secure path: %w", err)
	}
	if mtls.GetCaCertPath() != "" {
		if err := validation.IsAllowedPath(mtls.GetCaCertPath()); err != nil {
			return fmt.Errorf("mtls 'ca_cert_path' is not a secure path: %w", err)
		}
	}

	check := func(path, fieldName string) error {
		if _, err := osStat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("mtls '%s' not found: %w", fieldName, err)
			}
			return fmt.Errorf("mtls '%s' error: %w", fieldName, err)
		}
		return nil
	}

	if err := check(mtls.GetClientCertPath(), "client_cert_path"); err != nil {
		return err
	}
	if err := check(mtls.GetClientKeyPath(), "client_key_path"); err != nil {
		return err
	}
	if mtls.GetCaCertPath() != "" {
		if err := check(mtls.GetCaCertPath(), "ca_cert_path"); err != nil {
			return err
		}
	}
	return nil
}

func validateSchema(s *structpb.Struct) error {
	if s == nil {
		return nil
	}
	fields := s.GetFields()
	if t, ok := fields["type"]; ok {
		if _, ok := t.GetKind().(*structpb.Value_StringValue); !ok {
			return fmt.Errorf("schema 'type' must be a string")
		}
	}

	// Validate against strict JSON Schema using jsonschema library
	// Convert structpb to JSON bytes
	jsonBytes, err := protojson.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to marshal schema to JSON: %w", err)
	}

	// Create a new compiler
	c := jsonschema.NewCompiler()
	// Add the schema as a resource
	// We use "schema.json" as a virtual filename
	if err := c.AddResource("schema.json", bytes.NewReader(jsonBytes)); err != nil {
		return fmt.Errorf("failed to add schema resource: %w", err)
	}

	// Compile validates the schema syntax and structure against the meta-schema
	if _, err := c.Compile("schema.json"); err != nil {
		return &ActionableError{
			Err:        fmt.Errorf("invalid JSON schema: %w", err),
			Suggestion: "Check your 'input_schema' or 'output_schema' definition. Ensure it complies with JSON Schema specification (e.g. types match values).",
		}
	}

	return nil
}

func validateUser(ctx context.Context, user *configv1.User) error {
	if user.GetAuthentication() == nil {
		return nil // No auth config means no authentication required (Open Access)
	}
	if err := validateAuthentication(ctx, user.GetAuthentication()); err != nil {
		return err
	}
	return nil
}

func validateAuditConfig(audit *configv1.AuditConfig) error {
	if audit == nil {
		return nil
	}
	if !audit.GetEnabled() {
		return nil
	}
	switch audit.GetStorageType() {
	case configv1.AuditConfig_STORAGE_TYPE_FILE:
		if audit.GetOutputPath() == "" {
			return fmt.Errorf("output_path is required for file storage")
		}
	case configv1.AuditConfig_STORAGE_TYPE_WEBHOOK:
		if audit.GetWebhookUrl() == "" {
			return fmt.Errorf("webhook_url is required for webhook storage")
		}
		if !validation.IsValidURL(audit.GetWebhookUrl()) {
			return fmt.Errorf("invalid webhook_url: %s", audit.GetWebhookUrl())
		}
	}
	return nil
}

func validateDLPConfig(dlp *configv1.DLPConfig) error {
	if dlp == nil {
		return nil
	}
	for _, pattern := range dlp.GetCustomPatterns() {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid regex pattern %q: %w", pattern, err)
		}
	}
	return nil
}

func validateGCSettings(gc *configv1.GCSettings) error {
	if gc == nil {
		return nil
	}
	if gc.GetInterval() != "" {
		if _, err := time.ParseDuration(gc.GetInterval()); err != nil {
			return fmt.Errorf("invalid interval: %w", err)
		}
	}
	if gc.GetTtl() != "" {
		if _, err := time.ParseDuration(gc.GetTtl()); err != nil {
			return fmt.Errorf("invalid ttl: %w", err)
		}
	}
	if gc.GetEnabled() {
		for _, path := range gc.GetPaths() {
			if path == "" {
				return fmt.Errorf("empty gc path")
			}
			if err := validation.IsAllowedPath(path); err != nil {
				return fmt.Errorf("gc path %q is not secure: %w", path, err)
			}
			// We might also checking if it's absolute?
			if !filepath.IsAbs(path) {
				return fmt.Errorf("gc path %q must be absolute", path)
			}
		}
	}
	return nil
}

func validateProfileDefinition(_ *configv1.ProfileDefinition) error {
	// Add specific profile validation if needed.
	return nil
}

func validateCommandExists(command string, workingDir string) error {
	// If the command is an absolute path, check if it exists and is executable
	if filepath.IsAbs(command) {
		info, err := osStat(command)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("executable not found at %q", command)
			}
			return fmt.Errorf("failed to check executable %q: %w", command, err)
		}
		if info.IsDir() {
			return fmt.Errorf("%q is a directory, not an executable", command)
		}
		// Check for executable permission bit (simplified)
		// Note: os.Access(path, unix.X_OK) is better but unix package adds dependency.
		// For now os.Stat is good enough to check existence.
		// exec.LookPath handles permission checks better.
	}

	// If workingDir is provided and the command contains path separators (relative path),
	// check if it exists relative to workingDir.
	if workingDir != "" && (filepath.Base(command) != command) {
		fullPath := filepath.Join(workingDir, command)
		info, err := osStat(fullPath)
		if err == nil {
			if info.IsDir() {
				return fmt.Errorf("%q is a directory, not an executable (relative to %q)", command, workingDir)
			}
			// Command exists relative to working dir
			return nil
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check executable %q (relative to %q): %w", command, workingDir, err)
		}
		// If not found relative to workingDir, fall through to LookPath (which might find it in PATH or CWD)
	}

	// exec.LookPath searches for the executable in the directories named by the PATH environment variable.
	// If the file contains a slash, it is tried directly and the PATH is not consulted.
	_, err := execLookPath(command)
	if err != nil {
		return &ActionableError{
			Err:        fmt.Errorf("command %q not found in PATH or is not executable: %w", command, err),
			Suggestion: fmt.Sprintf("Ensure %q is installed and listed in your PATH environment variable.", command),
		}
	}
	return nil
}

func validateDirectoryExists(path string) error {
	info, err := osStat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory %q does not exist", path)
		}
		return fmt.Errorf("failed to check directory %q: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", path)
	}
	return nil
}
