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

	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/util"
	"github.com/mcpany/core/pkg/validation"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// BinaryType defines the type of the binary being validated.
type BinaryType int

const (
	// Server represents the server binary.
	Server BinaryType = iota
	// Worker represents the worker binary.
	Worker
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
			continue // Continue to find other errors
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
	if binaryType == Server {
		if gs.GetBindAddress() != "" {
			if err := validation.IsValidBindAddress(gs.GetBindAddress()); err != nil {
				return fmt.Errorf("invalid bind_address: %w", err)
			}
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
	log := logging.GetLogger().With("component", "configValidator")
	if service.GetMcpService() == nil && service.GetHttpService() == nil && service.GetGrpcService() == nil && service.GetOpenapiService() == nil && service.GetCommandLineService() == nil && service.GetWebsocketService() == nil {
		return fmt.Errorf("service has no service_config")
	}

	if httpService := service.GetHttpService(); httpService != nil {
		if httpService.GetAddress() == "" {
			return fmt.Errorf("http service has empty target_address")
		}
		if !validation.IsValidURL(httpService.GetAddress()) {
			return fmt.Errorf("invalid http target_address: %s", httpService.GetAddress())
		}
	} else if websocketService := service.GetWebsocketService(); websocketService != nil {
		if websocketService.GetAddress() == "" {
			return fmt.Errorf("websocket service has empty target_address")
		}
		if !validation.IsValidURL(websocketService.GetAddress()) {
			return fmt.Errorf("invalid websocket target_address: %s", websocketService.GetAddress())
		}
	} else if grpcService := service.GetGrpcService(); grpcService != nil {
		if grpcService.GetAddress() == "" {
			return fmt.Errorf("gRPC service has empty target_address")
		}
	} else if openapiService := service.GetOpenapiService(); openapiService != nil {
		if openapiService.GetAddress() != "" && !validation.IsValidURL(openapiService.GetAddress()) {
			return fmt.Errorf("invalid openapi target_address: %s", openapiService.GetAddress())
		}
	} else if mcpService := service.GetMcpService(); mcpService != nil {
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
	}

	if service.GetCache() != nil {
		if service.GetCache().GetTtl() != nil && service.GetCache().GetTtl().GetSeconds() < 0 {
			return fmt.Errorf("invalid cache timeout: %v", service.GetCache().GetTtl().AsDuration())
		}
	}

	if authConfig := service.GetUpstreamAuthentication(); authConfig != nil {
		if apiKey := authConfig.GetApiKey(); apiKey != nil {
			if apiKey.GetHeaderName() == "" {
				log.Warn("API key 'header_name' is empty. Authentication may fail.")
			}
			apiKeyValue, err := util.ResolveSecret(apiKey.GetApiKey())
			if err != nil {
				return fmt.Errorf("failed to resolve api key secret: %w", err)
			}
			if apiKeyValue == "" {
				return fmt.Errorf("api key 'api_key' is empty")
			}
		} else if bearerToken := authConfig.GetBearerToken(); bearerToken != nil {
			tokenValue, err := util.ResolveSecret(bearerToken.GetToken())
			if err != nil {
				return fmt.Errorf("failed to resolve bearer token secret: %w", err)
			}
			if tokenValue == "" {
				return fmt.Errorf("bearer token 'token' is empty")
			}
		} else if basicAuth := authConfig.GetBasicAuth(); basicAuth != nil {
			if basicAuth.GetUsername() == "" {
				log.Warn("Basic auth 'username' is empty. Authentication may fail.")
			}
		}
	}
	return nil
}
