// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/mcpany/core/pkg/util"
	"github.com/mcpany/core/pkg/validation"
	configv1 "github.com/mcpany/core/proto/config/v1"
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

var osStat = os.Stat

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

	if authConfig := service.GetUpstreamAuthentication(); authConfig != nil {
		if err := validateUpstreamAuthentication(ctx, authConfig); err != nil {
			return err
		}
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
		return validateSqlService(sqlService)
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
			if err := validation.IsRelativePath(dest); err != nil {
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
	case configv1.McpUpstreamService_StdioConnection_case:
		stdioConn := mcpService.GetStdioConnection()
		if len(stdioConn.GetCommand()) == 0 {
			return fmt.Errorf("mcp service with stdio_connection has empty command")
		}
	case configv1.McpUpstreamService_BundleConnection_case:
		bundleConn := mcpService.GetBundleConnection()
		if bundleConn.GetBundlePath() == "" {
			return fmt.Errorf("mcp service with bundle_connection has empty bundle_path")
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

func validateSqlService(sqlService *configv1.SqlUpstreamService) error {
	if sqlService.GetDriver() == "" {
		return fmt.Errorf("sql service has empty driver")
	}
	if sqlService.GetDsn() == "" {
		return fmt.Errorf("sql service has empty dsn")
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

func validateUpstreamAuthentication(ctx context.Context, authConfig *configv1.UpstreamAuthentication) error {
	switch authConfig.WhichAuthMethod() {
	case configv1.UpstreamAuthentication_ApiKey_case:
		return validateAPIKeyAuth(ctx, authConfig.GetApiKey())
	case configv1.UpstreamAuthentication_BearerToken_case:
		return validateBearerTokenAuth(ctx, authConfig.GetBearerToken())
	case configv1.UpstreamAuthentication_BasicAuth_case:
		return validateBasicAuth(ctx, authConfig.GetBasicAuth())
	case configv1.UpstreamAuthentication_Mtls_case:
		return validateMtlsAuth(authConfig.GetMtls())
	}
	return nil
}

func validateAPIKeyAuth(ctx context.Context, apiKey *configv1.UpstreamAPIKeyAuth) error {
	if apiKey.GetHeaderName() == "" {
		return fmt.Errorf("api key 'header_name' is empty")
	}
	apiKeyValue, err := util.ResolveSecret(ctx, apiKey.GetApiKey())
	if err != nil {
		return fmt.Errorf("failed to resolve api key secret: %w", err)
	}
	if apiKeyValue == "" {
		return fmt.Errorf("api key 'api_key' is empty")
	}
	return nil
}

func validateBearerTokenAuth(ctx context.Context, bearerToken *configv1.UpstreamBearerTokenAuth) error {
	tokenValue, err := util.ResolveSecret(ctx, bearerToken.GetToken())
	if err != nil {
		return fmt.Errorf("failed to resolve bearer token secret: %w", err)
	}
	if tokenValue == "" {
		return fmt.Errorf("bearer token 'token' is empty")
	}
	return nil
}

func validateBasicAuth(ctx context.Context, basicAuth *configv1.UpstreamBasicAuth) error {
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

func validateMtlsAuth(mtls *configv1.UpstreamMTLSAuth) error {
	if mtls.GetClientCertPath() == "" {
		return fmt.Errorf("mtls 'client_cert_path' is empty")
	}
	if mtls.GetClientKeyPath() == "" {
		return fmt.Errorf("mtls 'client_key_path' is empty")
	}
	if err := validation.IsSecurePath(mtls.GetClientCertPath()); err != nil {
		return fmt.Errorf("mtls 'client_cert_path' is not a secure path: %w", err)
	}
	if err := validation.IsSecurePath(mtls.GetClientKeyPath()); err != nil {
		return fmt.Errorf("mtls 'client_key_path' is not a secure path: %w", err)
	}
	if mtls.GetCaCertPath() != "" {
		if err := validation.IsSecurePath(mtls.GetCaCertPath()); err != nil {
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
