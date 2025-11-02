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

	"github.com/mcpany/core/pkg/validation"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// Validate inspects the given McpxServerConfig for correctness and consistency.
// It iterates through the list of upstream services, checking for valid
// service definitions, addresses, cache settings, and authentication
// configurations.
//
// Invalid services are logged and removed from the configuration. The function
// modifies the input config in place by setting its upstream services list to
// only the valid ones. It returns the modified configuration.
//
// config is the server configuration to be validated.
//
// It returns the validated configuration, which may have invalid services
// removed, and an error if a fundamental validation issue occurs.
func Validate(config *configv1.McpxServerConfig) (*configv1.McpxServerConfig, error) {
	if err := validateGlobalSettings(config.GetGlobalSettings()); err != nil {
		return nil, fmt.Errorf("invalid global settings: %w", err)
	}

	validServices := []*configv1.UpstreamServiceConfig{}
	serviceNames := make(map[string]bool)

	for _, service := range config.GetUpstreamServices() {
		if _, exists := serviceNames[service.GetName()]; exists {
			return nil, fmt.Errorf("duplicate service name found: %s", service.GetName())
		}
		serviceNames[service.GetName()] = true

		if err := validateUpstreamServiceConfig(service); err != nil {
			return nil, fmt.Errorf("invalid upstream service config for service %s: %w", service.GetName(), err)
		}
		validServices = append(validServices, service)
	}
	config.SetUpstreamServices(validServices)
	return config, nil
}

func validateUpstreamServiceConfig(service *configv1.UpstreamServiceConfig) error {
	if service.GetName() == "" {
		return fmt.Errorf("service name is empty")
	}

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
			return fmt.Errorf("grpc service has empty target_address")
		}
	} else if openapiService := service.GetOpenapiService(); openapiService != nil {
		if openapiService.GetAddress() != "" && !validation.IsValidURL(openapiService.GetAddress()) {
			return fmt.Errorf("invalid target_address for openapi service: %s", openapiService.GetAddress())
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
				return fmt.Errorf("api key 'header_name' is empty")
			}
			if apiKey.GetApiKey() == "" {
				return fmt.Errorf("api key 'api_key' is empty")
			}
		} else if bearerToken := authConfig.GetBearerToken(); bearerToken != nil {
			if bearerToken.GetToken() == "" {
				return fmt.Errorf("bearer token 'token' is empty")
			}
		} else if basicAuth := authConfig.GetBasicAuth(); basicAuth != nil {
			if basicAuth.GetUsername() == "" {
				return fmt.Errorf("basic auth 'username' is empty")
			}
		}
	}

	return nil
}

func validateGlobalSettings(settings *configv1.GlobalSettings) error {
	if settings == nil {
		return nil
	}
	if settings.GetBindAddress() == "" {
		return fmt.Errorf("bind_address is empty")
	}
	if _, ok := configv1.GlobalSettings_LogLevel_name[int32(settings.GetLogLevel())]; !ok {
		return fmt.Errorf("invalid log_level: %v", settings.GetLogLevel())
	}
	return nil
}
