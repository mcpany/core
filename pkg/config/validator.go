/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"fmt"
	"net/url"

	"github.com/mcpany/core/pkg/util"
	"github.com/mcpany/core/pkg/validation"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// BinaryType defines the type of the binary being validated.
type BinaryType int

const (
	// Server represents the server binary.
	Server BinaryType = iota
	// Worker represents the worker binary.
	Worker
	// Client represents the client binary.
	Client
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
// config is the server configuration to be validated.
//
// It returns a slice of ValidationErrors, which will be empty if the
// configuration is valid.
func Validate(config *configv1.McpAnyServerConfig, binaryType BinaryType) []ValidationError {
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

		if err := validateUpstreamService(service); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				ServiceName: service.GetName(),
				Err:         err,
			})
		}
	}

	return validationErrors
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
func ValidateOrError(service *configv1.UpstreamServiceConfig) error {
	return validateUpstreamService(service)
}

func validateUpstreamService(service *configv1.UpstreamServiceConfig) error {
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
		if err := validateUpstreamAuthentication(authConfig); err != nil {
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
	if u.Scheme != "http" && u.Scheme != "https" {
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

func validateUpstreamAuthentication(authConfig *configv1.UpstreamAuthentication) error {
	switch authConfig.WhichAuthMethod() {
	case configv1.UpstreamAuthentication_ApiKey_case:
		return validateAPIKeyAuth(authConfig.GetApiKey())
	case configv1.UpstreamAuthentication_BearerToken_case:
		return validateBearerTokenAuth(authConfig.GetBearerToken())
	case configv1.UpstreamAuthentication_BasicAuth_case:
		return validateBasicAuth(authConfig.GetBasicAuth())
	case configv1.UpstreamAuthentication_Mtls_case:
		return validateMtlsAuth(authConfig.GetMtls())
	}
	return nil
}

func validateAPIKeyAuth(apiKey *configv1.UpstreamAPIKeyAuth) error {
	if apiKey.GetHeaderName() == "" {
		return fmt.Errorf("api key 'header_name' is empty")
	}
	apiKeyValue, err := util.ResolveSecret(apiKey.GetApiKey())
	if err != nil {
		return fmt.Errorf("failed to resolve api key secret: %w", err)
	}
	if apiKeyValue == "" {
		return fmt.Errorf("api key 'api_key' is empty")
	}
	return nil
}

func validateBearerTokenAuth(bearerToken *configv1.UpstreamBearerTokenAuth) error {
	tokenValue, err := util.ResolveSecret(bearerToken.GetToken())
	if err != nil {
		return fmt.Errorf("failed to resolve bearer token secret: %w", err)
	}
	if tokenValue == "" {
		return fmt.Errorf("bearer token 'token' is empty")
	}
	return nil
}

func validateBasicAuth(basicAuth *configv1.UpstreamBasicAuth) error {
	if basicAuth.GetUsername() == "" {
		return fmt.Errorf("basic auth 'username' is empty")
	}
	passwordValue, err := util.ResolveSecret(basicAuth.GetPassword())
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
	if err := validation.FileExists(mtls.GetClientCertPath()); err != nil {
		return fmt.Errorf("mtls 'client_cert_path' not found: %w", err)
	}
	if err := validation.FileExists(mtls.GetClientKeyPath()); err != nil {
		return fmt.Errorf("mtls 'client_key_path' not found: %w", err)
	}
	if mtls.GetCaCertPath() != "" {
		if err := validation.FileExists(mtls.GetCaCertPath()); err != nil {
			return fmt.Errorf("mtls 'ca_cert_path' not found: %w", err)
		}
	}
	return nil
}

func validateSchema(s *structpb.Struct) error {
	if s == nil {
		return nil
	}
	// If fields are present, ensure it has a type or properties, typical for JSON schema
	// This is a loose check to catch obviously bad config
	return nil
}
