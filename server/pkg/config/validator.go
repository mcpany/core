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

// AuthValidationContext defines the context for authentication validation.
type AuthValidationContext int

const (
	// AuthValidationContextIncoming represents incoming authentication (e.g., Users).
	AuthValidationContextIncoming AuthValidationContext = iota
	// AuthValidationContextOutgoing represents outgoing authentication (e.g., Upstream Services).
	AuthValidationContextOutgoing
)

type contextKey string

const (
	// SkipSecretValidationKey is the context key to skip secret validation (e.g. for config check API).
	// Value should be a boolean.
	SkipSecretValidationKey contextKey = "skip_secret_validation"

	// SkipFilesystemCheckKey is the context key to skip filesystem existence checks (e.g. for config check API).
	// Value should be a boolean.
	SkipFilesystemCheckKey contextKey = "skip_filesystem_check"
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
		if err := validateGlobalSettings(ctx, gs, binaryType); err != nil {
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
		if err := validateUpstreamAuthentication(ctx, authConfig, AuthValidationContextOutgoing); err != nil {
			return err
		}
	}
	return nil
}

func validateStdioArgs(ctx context.Context, command string, args []string, workingDir string) error {
	// Only check for common interpreters to avoid false positives on arbitrary flags
	baseCmd := filepath.Base(command)
	isInterpreter := false
	// List of common interpreters that take a script file as an argument
	interpreters := []string{"python", "python3", "node", "deno", "bun", "ruby", "perl", "bash", "sh", "zsh", "go"}
	for _, i := range interpreters {
		if baseCmd == i {
			isInterpreter = true
			break
		}
		// Special handling for python versions (e.g. python3.9, python3.11)
		// We only check prefix if it is python/python3
		if strings.HasPrefix(i, "python") && strings.HasPrefix(baseCmd, i) {
			suffix := baseCmd[len(i):]
			if len(suffix) > 0 {
				// Check if the entire suffix is composed of digits and dots
				isValidVersion := true
				for _, r := range suffix {
					if r != '.' && (r < '0' || r > '9') {
						isValidVersion = false
						break
					}
				}
				if isValidVersion {
					isInterpreter = true
					break
				}
			}
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
			// Special case for -c (command) or -e/--eval (eval)
			// If we see these flags, the next argument is code, not a file.
			if arg == "-c" || arg == "-e" || arg == "--eval" || arg == "-p" || arg == "--print" {
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
		if err := validateFileExists(ctx, arg, workingDir); err != nil {
			return WrapActionableError(fmt.Sprintf("argument %q looks like a script file but does not exist", arg), err)
		}

		// We only check the FIRST script argument for interpreters.
		// Subsequent args are likely arguments to the script itself.
		return nil
	}
	return nil
}

func validateFileExists(ctx context.Context, path string, workingDir string) error {
	// Security: If context requests skipping filesystem checks, return success immediately.
	if skip, ok := ctx.Value(SkipFilesystemCheckKey).(bool); ok && skip {
		return nil
	}

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

func validateSecretMap(ctx context.Context, secrets map[string]*configv1.SecretValue) error {
	for key, secret := range secrets {
		if err := validateSecretValue(ctx, secret); err != nil {
			return fmt.Errorf("%q: %w", key, err)
		}
	}
	return nil
}

func validateSecretValue(ctx context.Context, secret *configv1.SecretValue) error {
	if secret == nil {
		return nil
	}
	// Security: If context requests skipping secret validation, we skip strict existence checks
	// to prevent information leakage (oracle attacks) where a user can probe environment variables or files.
	skipExistenceCheck, _ := ctx.Value(SkipSecretValidationKey).(bool)
	skipFilesystemCheck, _ := ctx.Value(SkipFilesystemCheckKey).(bool)
	if skipFilesystemCheck {
		skipExistenceCheck = true
	}

	switch secret.WhichValue() {
	case configv1.SecretValue_EnvironmentVariable_case:
		envVar := secret.GetEnvironmentVariable()
		if _, exists := os.LookupEnv(envVar); !exists {
			if skipExistenceCheck {
				return nil
			}
			suggestion := fmt.Sprintf("Set the environment variable %q in your shell or .env file before starting the server.", envVar)
			if similar := findSimilarEnvVar(envVar); similar != "" {
				suggestion = fmt.Sprintf("Did you mean %q? %s", similar, suggestion)
			}
			return &ActionableError{
				Err:        fmt.Errorf("environment variable %q is not set", envVar),
				Suggestion: suggestion,
			}
		}
	case configv1.SecretValue_FilePath_case:
		if skipFilesystemCheck {
			if err := validation.IsPathTraversalSafe(secret.GetFilePath()); err != nil {
				return fmt.Errorf("invalid secret file path %q: %w", secret.GetFilePath(), err)
			}
		} else {
			if err := validation.IsAllowedPath(secret.GetFilePath()); err != nil {
				return fmt.Errorf("invalid secret file path %q: %w", secret.GetFilePath(), err)
			}
		}
		// Validate that the file actually exists to fail fast
		// We skip this check if we are in validation mode to prevent file existence enumeration.
		if !skipExistenceCheck {
			if err := validation.FileExists(secret.GetFilePath()); err != nil {
				return &ActionableError{
					Err:        fmt.Errorf("secret file %q does not exist: %w", secret.GetFilePath(), err),
					Suggestion: fmt.Sprintf("Ensure the file exists at %q and the server process has read permissions.", secret.GetFilePath()),
				}
			}
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

	if secret.GetValidationRegex() != "" {
		re, err := regexp.Compile(secret.GetValidationRegex())
		if err != nil {
			return fmt.Errorf("invalid validation regex %q: %w", secret.GetValidationRegex(), err)
		}

		// Security: If context requests skipping secret validation, stop here.
		// This prevents information leakage (oracle attacks) where a user can probe secret values via regex.
		if skip, ok := ctx.Value(SkipSecretValidationKey).(bool); ok && skip {
			return nil
		}

		var valueToValidate string
		var shouldValidate bool

		switch secret.WhichValue() {
		case configv1.SecretValue_PlainText_case:
			valueToValidate = strings.TrimSpace(secret.GetPlainText())
			shouldValidate = true
		case configv1.SecretValue_EnvironmentVariable_case:
			valueToValidate = strings.TrimSpace(os.Getenv(secret.GetEnvironmentVariable()))
			shouldValidate = true
		case configv1.SecretValue_FilePath_case:
			// We do not read the file content for validation regex to prevent
			// Blind File Read vulnerabilities via the validation API.
			return fmt.Errorf("validation regex is not supported for secret file paths")
		}

		if shouldValidate && !re.MatchString(valueToValidate) {
			return fmt.Errorf("secret value does not match validation regex %q", secret.GetValidationRegex())
		}
	}

	return nil
}

func validateGlobalSettings(ctx context.Context, gs *configv1.GlobalSettings, binaryType BinaryType) error {
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

	if err := validateGCSettings(ctx, gs.GetGcSettings()); err != nil {
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

	if err := validateServiceConfig(ctx, service); err != nil {
		return err
	}

	if service.GetCache() != nil {
		if service.GetCache().GetTtl() != nil && service.GetCache().GetTtl().GetSeconds() < 0 {
			return fmt.Errorf("invalid cache timeout: %v", service.GetCache().GetTtl().AsDuration())
		}
	}

	if authConfig := service.GetUpstreamAuth(); authConfig != nil {
		if err := validateAuthentication(ctx, authConfig, AuthValidationContextOutgoing); err != nil {
			return err
		}
	}
	return nil
}

// ... (other functions)

func validateAuthentication(ctx context.Context, authConfig *configv1.Authentication, authCtx AuthValidationContext) error {
	switch authConfig.WhichAuthMethod() {
	case configv1.Authentication_ApiKey_case:
		return validateAPIKeyAuth(ctx, authConfig.GetApiKey(), authCtx)
	case configv1.Authentication_BearerToken_case:
		return validateBearerTokenAuth(ctx, authConfig.GetBearerToken())
	case configv1.Authentication_BasicAuth_case:
		return validateBasicAuth(ctx, authConfig.GetBasicAuth(), authCtx)
	case configv1.Authentication_Mtls_case:
		return validateMtlsAuth(ctx, authConfig.GetMtls())
	case configv1.Authentication_Oauth2_case:
		return validateOAuth2Auth(ctx, authConfig.GetOauth2())
	case configv1.Authentication_Oidc_case:
		return validateOIDCAuth(ctx, authConfig.GetOidc())
	case configv1.Authentication_TrustedHeader_case:
		return validateTrustedHeaderAuth(authConfig.GetTrustedHeader())
	}
	return nil
}

func validateBasicAuth(ctx context.Context, basicAuth *configv1.BasicAuth, authCtx AuthValidationContext) error {
	if basicAuth.GetUsername() == "" {
		return &ActionableError{
			Err:        fmt.Errorf("basic auth 'username' is empty"),
			Suggestion: "Set the 'username' field.",
		}
	}

	// For incoming auth (Users), we accept either a plain text password (for seeding) OR a hash.
	if authCtx == AuthValidationContextIncoming && basicAuth.GetPasswordHash() != "" {
		return nil
	}

	if basicAuth.GetPassword() == nil {
		return &ActionableError{
			Err:        fmt.Errorf("basic auth 'password' is missing"),
			Suggestion: "Set the 'password' field.",
		}
	}

	if err := validateSecretValue(ctx, basicAuth.GetPassword()); err != nil {
		return WrapActionableError("basic auth password validation failed", err)
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

func validateServiceConfig(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	if httpService := service.GetHttpService(); httpService != nil {
		return validateHTTPService(httpService)
	} else if websocketService := service.GetWebsocketService(); websocketService != nil {
		return validateWebSocketService(websocketService)
	} else if grpcService := service.GetGrpcService(); grpcService != nil {
		return validateGrpcService(grpcService)
	} else if openapiService := service.GetOpenapiService(); openapiService != nil {
		return validateOpenAPIService(openapiService)
	} else if commandLineService := service.GetCommandLineService(); commandLineService != nil {
		return validateCommandLineService(ctx, commandLineService)
	} else if mcpService := service.GetMcpService(); mcpService != nil {
		return validateMcpService(ctx, mcpService)
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
	if strings.TrimSpace(httpService.GetAddress()) != httpService.GetAddress() {
		return &ActionableError{
			Err:        fmt.Errorf("http address contains hidden whitespace"),
			Suggestion: "The address contains hidden whitespace (spaces or tabs). Fix: Check your configuration or environment variables and remove any trailing spaces.",
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
	if strings.TrimSpace(websocketService.GetAddress()) != websocketService.GetAddress() {
		return &ActionableError{
			Err:        fmt.Errorf("websocket address contains hidden whitespace"),
			Suggestion: "The address contains hidden whitespace (spaces or tabs). Fix: Check your configuration or environment variables and remove any trailing spaces.",
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
	if openapiService.GetAddress() != "" {
		if strings.TrimSpace(openapiService.GetAddress()) != openapiService.GetAddress() {
			return &ActionableError{
				Err:        fmt.Errorf("openapi address contains hidden whitespace"),
				Suggestion: "The address contains hidden whitespace (spaces or tabs). Fix: Check your configuration or environment variables and remove any trailing spaces.",
			}
		}
		if !validation.IsValidURL(openapiService.GetAddress()) {
			return &ActionableError{
				Err:        fmt.Errorf("invalid openapi address: %s", openapiService.GetAddress()),
				Suggestion: "Ensure the address is a valid URL.",
			}
		}
	}
	if openapiService.GetSpecUrl() != "" {
		if strings.TrimSpace(openapiService.GetSpecUrl()) != openapiService.GetSpecUrl() {
			return &ActionableError{
				Err:        fmt.Errorf("openapi spec_url contains hidden whitespace"),
				Suggestion: "The URL contains hidden whitespace (spaces or tabs). Fix: Check your configuration or environment variables and remove any trailing spaces.",
			}
		}
		if !validation.IsValidURL(openapiService.GetSpecUrl()) {
			return &ActionableError{
				Err:        fmt.Errorf("invalid openapi spec_url: %s", openapiService.GetSpecUrl()),
				Suggestion: "Ensure the spec_url is a valid URL.",
			}
		}
	}
	return nil
}

func validateCommandLineService(ctx context.Context, commandLineService *configv1.CommandLineUpstreamService) error {
	if commandLineService.GetCommand() == "" {
		return &ActionableError{
			Err:        fmt.Errorf("command_line_service has empty command"),
			Suggestion: "Set the 'command' field (e.g., './my-script.sh' or 'python my_script.py').",
		}
	}

	// Only validate command existence if not running in a container
	if commandLineService.GetContainerEnvironment().GetImage() == "" {
		if err := validateCommandExists(ctx, commandLineService.GetCommand(), commandLineService.GetWorkingDirectory()); err != nil {
			return WrapActionableError("command_line_service command validation failed", err)
		}
	}

	if err := validateContainerEnvironment(ctx, commandLineService.GetContainerEnvironment()); err != nil {
		return err
	}
	return nil
}

func validateContainerEnvironment(ctx context.Context, env *configv1.ContainerEnvironment) error {
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
			if skip, ok := ctx.Value(SkipFilesystemCheckKey).(bool); ok && skip {
				if err := validation.IsPathTraversalSafe(dest); err != nil {
					return fmt.Errorf("container environment volume host path %q is not a secure path: %w", dest, err)
				}
			} else {
				if err := validation.IsAllowedPath(dest); err != nil {
					return fmt.Errorf("container environment volume host path %q is not a secure path: %w", dest, err)
				}
			}
		}
	}

	return nil
}

func validateMcpService(ctx context.Context, mcpService *configv1.McpUpstreamService) error {
	switch mcpService.WhichConnectionType() {
	case configv1.McpUpstreamService_HttpConnection_case:
		httpConn := mcpService.GetHttpConnection()
		if httpConn.GetHttpAddress() == "" {
			return &ActionableError{
				Err:        fmt.Errorf("mcp service with http_connection has empty http_address"),
				Suggestion: "Set the 'http_address' field (e.g., 'http://localhost:8080').",
			}
		}
		if strings.TrimSpace(httpConn.GetHttpAddress()) != httpConn.GetHttpAddress() {
			return &ActionableError{
				Err:        fmt.Errorf("mcp http_address contains hidden whitespace"),
				Suggestion: "The address contains hidden whitespace (spaces or tabs). Fix: Check your configuration or environment variables and remove any trailing spaces.",
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
			if err := validateCommandExists(ctx, stdioConn.GetCommand(), stdioConn.GetWorkingDirectory()); err != nil {
				return WrapActionableError("mcp service with stdio_connection command validation failed", err)
			}

			if err := validateStdioArgs(ctx, stdioConn.GetCommand(), stdioConn.GetArgs(), stdioConn.GetWorkingDirectory()); err != nil {
				return WrapActionableError("mcp service with stdio_connection argument validation failed", err)
			}

			if stdioConn.GetWorkingDirectory() != "" {
				if err := validation.IsAllowedPath(stdioConn.GetWorkingDirectory()); err != nil {
					return fmt.Errorf("mcp service with stdio_connection has insecure working_directory %q: %w", stdioConn.GetWorkingDirectory(), err)
				}
				if err := validateDirectoryExists(ctx, stdioConn.GetWorkingDirectory()); err != nil {
					return fmt.Errorf("mcp service with stdio_connection working_directory validation failed: %w", err)
				}
			}
		}

		if err := validateSecretMap(ctx, stdioConn.GetEnv()); err != nil {
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
		if err := validateSecretMap(ctx, bundleConn.GetEnv()); err != nil {
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
	if strings.TrimSpace(graphqlService.GetAddress()) != graphqlService.GetAddress() {
		return &ActionableError{
			Err:        fmt.Errorf("graphql address contains hidden whitespace"),
			Suggestion: "The address contains hidden whitespace (spaces or tabs). Fix: Check your configuration or environment variables and remove any trailing spaces.",
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
	if strings.TrimSpace(webrtcService.GetAddress()) != webrtcService.GetAddress() {
		return &ActionableError{
			Err:        fmt.Errorf("webrtc address contains hidden whitespace"),
			Suggestion: "The address contains hidden whitespace (spaces or tabs). Fix: Check your configuration or environment variables and remove any trailing spaces.",
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

func validateUpstreamAuthentication(ctx context.Context, authConfig *configv1.Authentication, authCtx AuthValidationContext) error {
	switch authConfig.WhichAuthMethod() {
	case configv1.Authentication_ApiKey_case:
		return validateAPIKeyAuth(ctx, authConfig.GetApiKey(), authCtx)
	case configv1.Authentication_BearerToken_case:
		return validateBearerTokenAuth(ctx, authConfig.GetBearerToken())
	case configv1.Authentication_BasicAuth_case:
		return validateBasicAuth(ctx, authConfig.GetBasicAuth(), authCtx)
	case configv1.Authentication_Mtls_case:
		return validateMtlsAuth(ctx, authConfig.GetMtls())
	case configv1.Authentication_Oauth2_case:
		return validateOAuth2Auth(ctx, authConfig.GetOauth2())
	case configv1.Authentication_Oidc_case:
		return validateOIDCAuth(ctx, authConfig.GetOidc())
	case configv1.Authentication_TrustedHeader_case:
		return validateTrustedHeaderAuth(authConfig.GetTrustedHeader())
	}
	return nil
}

func validateAPIKeyAuth(ctx context.Context, apiKey *configv1.APIKeyAuth, authCtx AuthValidationContext) error {
	if apiKey.GetParamName() == "" {
		return &ActionableError{
			Err:        fmt.Errorf("api key 'param_name' is empty"),
			Suggestion: "Set the 'param_name' (e.g., 'X-API-Key' or 'api_key').",
		}
	}

	// Validation depends on context (Incoming vs Outgoing)

	if authCtx == AuthValidationContextOutgoing {
		// Outgoing: We need the 'Value' (the secret to send)
		if apiKey.GetValue() == nil {
			return &ActionableError{
				Err:        fmt.Errorf("api key 'value' is missing (required for outgoing auth)"),
				Suggestion: "Set the 'value' field using a secret (environment variable or file) for the API key.",
			}
		}
	} else {
		// Incoming: We need either 'VerificationValue' (hardcoded) or 'Value' (secret reference) to verify against
		if apiKey.GetValue() == nil && apiKey.GetVerificationValue() == "" {
			return &ActionableError{
				Err:        fmt.Errorf("api key configuration is empty"),
				Suggestion: "Set either 'verification_value' (for static keys) or 'value' (for secret-based keys).",
			}
		}
	}

	if apiKey.GetValue() != nil {
		if err := validateSecretValue(ctx, apiKey.GetValue()); err != nil {
			return WrapActionableError("api key secret validation failed", err)
		}

		// If we are skipping secret validation, we should also skip attempting to resolve it for "not empty" check
		// because ResolveSecret will read the value.
		if skip, ok := ctx.Value(SkipSecretValidationKey).(bool); ok && skip {
			return nil
		}

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
	if err := validateSecretValue(ctx, bearerToken.GetToken()); err != nil {
		return WrapActionableError("bearer token validation failed", err)
	}

	if skip, ok := ctx.Value(SkipSecretValidationKey).(bool); ok && skip {
		return nil
	}

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

func validateOAuth2Auth(ctx context.Context, oauth *configv1.OAuth2Auth) error {
	if oauth.GetTokenUrl() == "" {
		if oauth.GetIssuerUrl() == "" {
			return &ActionableError{
				Err:        fmt.Errorf("oauth2 token_url is empty and no issuer_url provided"),
				Suggestion: "Provide 'token_url' OR 'issuer_url' to enable auto-discovery.",
			}
		}
		if strings.TrimSpace(oauth.GetIssuerUrl()) != oauth.GetIssuerUrl() {
			return &ActionableError{
				Err:        fmt.Errorf("oauth2 issuer_url contains hidden whitespace"),
				Suggestion: "The URL contains hidden whitespace (spaces or tabs). Fix: Check your configuration or environment variables and remove any trailing spaces.",
			}
		}
		if !validation.IsValidURL(oauth.GetIssuerUrl()) {
			return fmt.Errorf("invalid oauth2 issuer_url: %s", oauth.GetIssuerUrl())
		}
		// If IssuerURL is present and valid, we allow TokenUrl to be empty (auto-discovery)
	} else {
		if strings.TrimSpace(oauth.GetTokenUrl()) != oauth.GetTokenUrl() {
			return &ActionableError{
				Err:        fmt.Errorf("oauth2 token_url contains hidden whitespace"),
				Suggestion: "The URL contains hidden whitespace (spaces or tabs). Fix: Check your configuration or environment variables and remove any trailing spaces.",
			}
		}
		if !validation.IsValidURL(oauth.GetTokenUrl()) {
			// Only validate TokenUrl if it is present
			return fmt.Errorf("invalid oauth2 token_url: %s", oauth.GetTokenUrl())
		}
	}

	if err := validateSecretValue(ctx, oauth.GetClientId()); err != nil {
		return WrapActionableError("oauth2 client_id validation failed", err)
	}

	if skip, ok := ctx.Value(SkipSecretValidationKey).(bool); !ok || !skip {
		clientID, err := util.ResolveSecret(ctx, oauth.GetClientId())
		if err != nil {
			return fmt.Errorf("failed to resolve oauth2 client_id: %w", err)
		}
		if clientID == "" {
			return fmt.Errorf("oauth2 client_id is missing or empty")
		}
	}

	if err := validateSecretValue(ctx, oauth.GetClientSecret()); err != nil {
		return WrapActionableError("oauth2 client_secret validation failed", err)
	}

	if skip, ok := ctx.Value(SkipSecretValidationKey).(bool); !ok || !skip {
		clientSecret, err := util.ResolveSecret(ctx, oauth.GetClientSecret())
		if err != nil {
			return fmt.Errorf("failed to resolve oauth2 client_secret: %w", err)
		}
		if clientSecret == "" {
			return fmt.Errorf("oauth2 client_secret is missing or empty")
		}
	}

	return nil
}

func validateOIDCAuth(_ context.Context, oidc *configv1.OIDCAuth) error {
	if oidc.GetIssuer() == "" {
		return fmt.Errorf("oidc issuer is empty")
	}
	if strings.TrimSpace(oidc.GetIssuer()) != oidc.GetIssuer() {
		return &ActionableError{
			Err:        fmt.Errorf("oidc issuer url contains hidden whitespace"),
			Suggestion: "The URL contains hidden whitespace (spaces or tabs). Fix: Check your configuration or environment variables and remove any trailing spaces.",
		}
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

func validateMtlsAuth(ctx context.Context, mtls *configv1.MTLSAuth) error {
	if mtls.GetClientCertPath() == "" {
		return fmt.Errorf("mtls 'client_cert_path' is empty")
	}
	if mtls.GetClientKeyPath() == "" {
		return fmt.Errorf("mtls 'client_key_path' is empty")
	}
	skip, ok := ctx.Value(SkipFilesystemCheckKey).(bool)
	checkPath := func(path, name string) error {
		if ok && skip {
			if err := validation.IsPathTraversalSafe(path); err != nil {
				return fmt.Errorf("mtls '%s' is not a secure path: %w", name, err)
			}
		} else {
			if err := validation.IsAllowedPath(path); err != nil {
				return fmt.Errorf("mtls '%s' is not a secure path: %w", name, err)
			}
		}
		return nil
	}

	if err := checkPath(mtls.GetClientCertPath(), "client_cert_path"); err != nil {
		return err
	}
	if err := checkPath(mtls.GetClientKeyPath(), "client_key_path"); err != nil {
		return err
	}
	if mtls.GetCaCertPath() != "" {
		if err := checkPath(mtls.GetCaCertPath(), "ca_cert_path"); err != nil {
			return err
		}
	}

	check := func(path, fieldName string) error {
		// Security: If context requests skipping filesystem checks, return success immediately.
		if skip, ok := ctx.Value(SkipFilesystemCheckKey).(bool); ok && skip {
			return nil
		}
		if _, err := osStat(path); err != nil {
			if os.IsNotExist(err) {
				return &ActionableError{
					Err:        fmt.Errorf("mtls '%s' not found at %q: %w", fieldName, path, err),
					Suggestion: fmt.Sprintf("Ensure the %s file exists and is accessible.", fieldName),
				}
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
	if err := validateAuthentication(ctx, user.GetAuthentication(), AuthValidationContextIncoming); err != nil {
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
		if strings.TrimSpace(audit.GetWebhookUrl()) != audit.GetWebhookUrl() {
			return &ActionableError{
				Err:        fmt.Errorf("webhook_url contains hidden whitespace"),
				Suggestion: "The URL contains hidden whitespace (spaces or tabs). Fix: Check your configuration or environment variables and remove any trailing spaces.",
			}
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

func validateGCSettings(ctx context.Context, gc *configv1.GCSettings) error {
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
			if skip, ok := ctx.Value(SkipFilesystemCheckKey).(bool); ok && skip {
				// Just check string safety to prevent traversal, but skip disk checks (IsAllowedPath calls EvalSymlinks)
				if err := validation.IsPathTraversalSafe(path); err != nil {
					return fmt.Errorf("gc path %q is not secure: %w", path, err)
				}
			} else {
				if err := validation.IsAllowedPath(path); err != nil {
					return fmt.Errorf("gc path %q is not secure: %w", path, err)
				}
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

func validateCommandExists(ctx context.Context, command string, workingDir string) error {
	// Security: If context requests skipping filesystem checks, return success immediately.
	if skip, ok := ctx.Value(SkipFilesystemCheckKey).(bool); ok && skip {
		return nil
	}

	// If the command is an absolute path, check if it exists and is executable
	if filepath.IsAbs(command) {
		info, err := osStat(command)
		if err != nil {
			if os.IsNotExist(err) {
				return &ActionableError{
					Err:        fmt.Errorf("executable not found at %q", command),
					Suggestion: fmt.Sprintf("Check if the file exists at %q.", command),
				}
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

func validateDirectoryExists(ctx context.Context, path string) error {
	// Security: If context requests skipping filesystem checks, return success immediately.
	if skip, ok := ctx.Value(SkipFilesystemCheckKey).(bool); ok && skip {
		return nil
	}

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

func findSimilarEnvVar(target string) string {
	environ := os.Environ()
	bestMatch := ""
	minDist := 1000

	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		key := parts[0]

		// Skip exact match (shouldn't happen if we are here)
		if key == target {
			continue
		}

		dist := util.LevenshteinDistance(target, key)

		// Threshold: Allow edits up to 1/3 of the string length, but at least 1 and max 5.
		// For very short strings (len < 3), require exact match (so threshold 0, which means no fuzzy match).
		if len(target) < 3 {
			continue
		}

		threshold := len(target) / 3
		if threshold > 5 {
			threshold = 5
		}

		if dist <= threshold && dist < minDist {
			minDist = dist
			bestMatch = key
		}
	}

	return bestMatch
}
