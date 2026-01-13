// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/validation"
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
//   ctx: The context for the validation (used for secret resolution).
//   config: The server configuration to be validated.
//   binaryType: The type of binary (server, worker) which might affect validation rules.
//
// Returns:
//   []ValidationError: A slice of ValidationErrors, which will be empty if the configuration is valid.
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

	for _, collection := range config.GetUpstreamServiceCollections() {
		if err := validateUpstreamServiceCollection(ctx, collection); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				ServiceName: collection.GetName(),
				Err:         err,
			})
		}
	}

	return validationErrors
}

func validateUpstreamServiceCollection(ctx context.Context, collection *configv1.UpstreamServiceCollection) error {
	if collection.GetName() == "" {
		return fmt.Errorf("collection name is empty")
	}
	if collection.GetHttpUrl() == "" {
		return fmt.Errorf("collection http_url is empty")
	}
	if !validation.IsValidURL(collection.GetHttpUrl()) {
		return fmt.Errorf("invalid collection http_url: %s", collection.GetHttpUrl())
	}
	u, _ := url.Parse(collection.GetHttpUrl())
	if u.Scheme != schemeHTTP && u.Scheme != schemeHTTPS {
		return fmt.Errorf("invalid collection http_url scheme: %s", u.Scheme)
	}

	if authConfig := collection.GetAuthentication(); authConfig != nil {
		if err := validateUpstreamAuthentication(ctx, authConfig); err != nil {
			return err
		}
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
			return fmt.Errorf("environment variable %q is not set", envVar)
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
//   ctx: The context for the validation.
//   service: The upstream service configuration to validate.
//
// Returns:
//   error: An error if validation fails.
func ValidateOrError(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	return validateUpstreamService(ctx, service)
}

func validateUpstreamService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
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
		return fmt.Errorf("http service has empty target_address")
	}
	if !validation.IsValidURL(httpService.GetAddress()) {
		return fmt.Errorf("invalid http target_address: %s", httpService.GetAddress())
	}
	u, _ := url.Parse(httpService.GetAddress())
	if u.Scheme != schemeHTTP && u.Scheme != schemeHTTPS {
		return fmt.Errorf("invalid http target_address scheme: %s", u.Scheme)
	}

	for name, call := range httpService.GetCalls() {
		if err := validateSchema(call.GetInputSchema()); err != nil {
			return fmt.Errorf("http call %q input_schema error: %w", name, err)
		}
		if err := validateSchema(call.GetOutputSchema()); err != nil {
			return fmt.Errorf("http call %q output_schema error: %w", name, err)
		}
	}
	return nil
}

func validateWebSocketService(websocketService *configv1.WebsocketUpstreamService) error {
	if websocketService.GetAddress() == "" {
		return fmt.Errorf("websocket service has empty target_address")
	}
	if !validation.IsValidURL(websocketService.GetAddress()) {
		return fmt.Errorf("invalid websocket target_address: %s", websocketService.GetAddress())
	}
	u, _ := url.Parse(websocketService.GetAddress())
	if u.Scheme != "ws" && u.Scheme != "wss" {
		return fmt.Errorf("invalid websocket target_address scheme: %s", u.Scheme)
	}

	for name, call := range websocketService.GetCalls() {
		if err := validateSchema(call.GetInputSchema()); err != nil {
			return fmt.Errorf("websocket call %q input_schema error: %w", name, err)
		}
		if err := validateSchema(call.GetOutputSchema()); err != nil {
			return fmt.Errorf("websocket call %q output_schema error: %w", name, err)
		}
	}
	return nil
}

func validateGrpcService(grpcService *configv1.GrpcUpstreamService) error {
	if grpcService.GetAddress() == "" {
		return fmt.Errorf("gRPC service has empty target_address")
	}

	for name, call := range grpcService.GetCalls() {
		if err := validateSchema(call.GetInputSchema()); err != nil {
			return fmt.Errorf("grpc call %q input_schema error: %w", name, err)
		}
		if err := validateSchema(call.GetOutputSchema()); err != nil {
			return fmt.Errorf("grpc call %q output_schema error: %w", name, err)
		}
	}
	return nil
}

func validateOpenAPIService(openapiService *configv1.OpenapiUpstreamService) error {
	if openapiService.GetAddress() == "" && openapiService.GetSpecContent() == "" && openapiService.GetSpecUrl() == "" {
		return fmt.Errorf("openapi service must have either an address, spec content or spec url")
	}
	if openapiService.GetAddress() != "" && !validation.IsValidURL(openapiService.GetAddress()) {
		return fmt.Errorf("invalid openapi target_address: %s", openapiService.GetAddress())
	}
	if openapiService.GetSpecUrl() != "" && !validation.IsValidURL(openapiService.GetSpecUrl()) {
		return fmt.Errorf("invalid openapi spec_url: %s", openapiService.GetSpecUrl())
	}
	return nil
}

func validateCommandLineService(commandLineService *configv1.CommandLineUpstreamService) error {
	if commandLineService.GetCommand() == "" {
		return fmt.Errorf("command_line_service has empty command")
	}

	// Only validate command existence if not running in a container
	if commandLineService.GetContainerEnvironment().GetImage() == "" {
		if err := validateCommandExists(commandLineService.GetCommand()); err != nil {
			return fmt.Errorf("command_line_service command validation failed: %w", err)
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
			return fmt.Errorf("mcp service with http_connection has empty http_address")
		}
		if !validation.IsValidURL(httpConn.GetHttpAddress()) {
			return fmt.Errorf("mcp service with http_connection has invalid http_address: %s", httpConn.GetHttpAddress())
		}
		u, _ := url.Parse(httpConn.GetHttpAddress())
		if u.Scheme != schemeHTTP && u.Scheme != schemeHTTPS {
			return fmt.Errorf("mcp service with http_connection has invalid http_address scheme: %s", u.Scheme)
		}
	case configv1.McpUpstreamService_StdioConnection_case:
		stdioConn := mcpService.GetStdioConnection()
		if len(stdioConn.GetCommand()) == 0 {
			return fmt.Errorf("mcp service with stdio_connection has empty command")
		}

		// If running in Docker (container_image is set), we don't enforce host path/command restrictions
		if stdioConn.GetContainerImage() == "" {
			if err := validateCommandExists(stdioConn.GetCommand()); err != nil {
				return fmt.Errorf("mcp service with stdio_connection command validation failed: %w", err)
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
			return fmt.Errorf("mcp call %q input_schema error: %w", name, err)
		}
		if err := validateSchema(call.GetOutputSchema()); err != nil {
			return fmt.Errorf("mcp call %q output_schema error: %w", name, err)
		}
	}
	return nil
}

func validateSQLService(sqlService *configv1.SqlUpstreamService) error {
	if sqlService.GetDriver() == "" {
		return fmt.Errorf("sql service has empty driver")
	}
	if sqlService.GetDsn() == "" {
		return fmt.Errorf("sql service has empty dsn")
	}

	for name, call := range sqlService.GetCalls() {
		if call.GetQuery() == "" {
			return fmt.Errorf("sql call %q query is empty", name)
		}
		if err := validateSchema(call.GetInputSchema()); err != nil {
			return fmt.Errorf("sql call %q input_schema error: %w", name, err)
		}
		if err := validateSchema(call.GetOutputSchema()); err != nil {
			return fmt.Errorf("sql call %q output_schema error: %w", name, err)
		}
	}
	return nil
}

func validateGraphQLService(graphqlService *configv1.GraphQLUpstreamService) error {
	if graphqlService.GetAddress() == "" {
		return fmt.Errorf("graphql service has empty address")
	}
	if !validation.IsValidURL(graphqlService.GetAddress()) {
		return fmt.Errorf("invalid graphql target_address: %s", graphqlService.GetAddress())
	}
	u, _ := url.Parse(graphqlService.GetAddress())
	if u.Scheme != schemeHTTP && u.Scheme != schemeHTTPS {
		return fmt.Errorf("invalid graphql target_address scheme: %s", u.Scheme)
	}
	return nil
}

func validateWebrtcService(webrtcService *configv1.WebrtcUpstreamService) error {
	if webrtcService.GetAddress() == "" {
		return fmt.Errorf("webrtc service has empty address")
	}
	if !validation.IsValidURL(webrtcService.GetAddress()) {
		return fmt.Errorf("invalid webrtc target_address: %s", webrtcService.GetAddress())
	}
	u, _ := url.Parse(webrtcService.GetAddress())
	if u.Scheme != schemeHTTP && u.Scheme != schemeHTTPS {
		return fmt.Errorf("invalid webrtc target_address scheme: %s", u.Scheme)
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
		return fmt.Errorf("api key 'param_name' is empty")
	}
	// Value might be nil but usually we want to validate it resolves if present,
	// or ensure it IS present if required. For upstream, it's usually required.
	if apiKey.GetValue() == nil && apiKey.GetVerificationValue() == "" {
		// If both are missing, it's invalid?
		// VerificationValue is for incoming, Value (SecretValue) is for outgoing.
		// Detailed validation might depend on context (incoming vs outgoing),
		// but here we are validating "UpstreamServiceConfig" which implies outgoing.
		return fmt.Errorf("api key 'value' is missing (required for outgoing auth)")
	}
	if apiKey.GetValue() != nil {
		apiKeyValue, err := util.ResolveSecret(ctx, apiKey.GetValue())
		if err != nil {
			return fmt.Errorf("failed to resolve api key secret: %w", err)
		}
		if apiKeyValue == "" {
			return fmt.Errorf("resolved api key value is empty")
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
		return fmt.Errorf("bearer token 'token' is empty")
	}
	return nil
}

func validateBasicAuth(ctx context.Context, basicAuth *configv1.BasicAuth) error {
	if basicAuth.GetUsername() == "" {
		return fmt.Errorf("basic auth 'username' is empty")
	}
	passwordValue, err := util.ResolveSecret(ctx, basicAuth.GetPassword())
	if err != nil {
		return fmt.Errorf("failed to resolve basic auth password secret: %w", err)
	}
	if passwordValue == "" {
		return fmt.Errorf("basic auth 'password' is empty")
	}
	return nil
}

func validateOAuth2Auth(_ context.Context, oauth *configv1.OAuth2Auth) error {
	if oauth.GetTokenUrl() == "" {
		return fmt.Errorf("oauth2 token_url is empty")
	}
	if !validation.IsValidURL(oauth.GetTokenUrl()) {
		return fmt.Errorf("invalid oauth2 token_url: %s", oauth.GetTokenUrl())
	}
	u, _ := url.Parse(oauth.GetTokenUrl())
	if u.Scheme != schemeHTTP && u.Scheme != schemeHTTPS {
		return fmt.Errorf("invalid oauth2 token_url scheme: %s", u.Scheme)
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
		u, _ := url.Parse(audit.GetWebhookUrl())
		if u.Scheme != schemeHTTP && u.Scheme != schemeHTTPS {
			return fmt.Errorf("invalid webhook_url scheme: %s", u.Scheme)
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

func validateCommandExists(command string) error {
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

	// exec.LookPath searches for the executable in the directories named by the PATH environment variable.
	// If the file contains a slash, it is tried directly and the PATH is not consulted.
	_, err := execLookPath(command)
	if err != nil {
		return fmt.Errorf("command %q not found in PATH or is not executable: %w", command, err)
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
