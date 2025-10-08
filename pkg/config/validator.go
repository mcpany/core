/*
 * Copyright 2025 Author(s) of MCP-XY
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
	"github.com/mcpxy/core/pkg/logging"
	"github.com/mcpxy/core/pkg/validation"
	configv1 "github.com/mcpxy/core/proto/config/v1"
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
func Validate(config *configv1.McpxServerConfig) (*configv1.McpxServerConfig, error) {
	log := logging.GetLogger().With("component", "configValidator")
	validServices := []*configv1.UpstreamServiceConfig{}

	for _, service := range config.GetUpstreamServices() {
		serviceLog := log.With("serviceName", service.GetName())
		isValidService := true

		if service.GetMcpService() == nil && service.GetHttpService() == nil && service.GetGrpcService() == nil && service.GetOpenapiService() == nil && service.GetCommandLineService() == nil {
			serviceLog.Warn("Service has no service_config. Skipping service.")
			continue
		}

		if httpService := service.GetHttpService(); httpService != nil {
			if httpService.GetAddress() == "" {
				serviceLog.Warn("HTTP service has empty target_address. Skipping service.")
				isValidService = false
			} else if !validation.IsValidURL(httpService.GetAddress()) {
				serviceLog.Warn("Invalid HTTP target_address. Skipping service.", "address", httpService.GetAddress())
				isValidService = false
			}
		} else if grpcService := service.GetGrpcService(); grpcService != nil {
			if grpcService.GetAddress() == "" {
				serviceLog.Warn("gRPC service has empty target_address. Skipping service.")
				isValidService = false
			}
		} else if openapiService := service.GetOpenapiService(); openapiService != nil {
			if openapiService.GetAddress() != "" && !validation.IsValidURL(openapiService.GetAddress()) {
				serviceLog.Warn("Invalid TargetAddress for OpenAPI service. This default target will be ignored if spec contains servers.", "address", openapiService.GetAddress())
			}
		} else if mcpService := service.GetMcpService(); mcpService != nil {
			switch mcpService.WhichConnectionType() {
			case configv1.McpUpstreamService_HttpConnection_case:
				httpConn := mcpService.GetHttpConnection()
				if httpConn.GetHttpAddress() == "" {
					serviceLog.Warn("MCP service with http_connection has empty http_address. Skipping service.")
					isValidService = false
				} else if !validation.IsValidURL(httpConn.GetHttpAddress()) {
					serviceLog.Warn("MCP service with http_connection has invalid http_address. Skipping service.", "address", httpConn.GetHttpAddress())
					isValidService = false
				}
			case configv1.McpUpstreamService_StdioConnection_case:
				stdioConn := mcpService.GetStdioConnection()
				if len(stdioConn.GetCommand()) == 0 {
					serviceLog.Warn("MCP service with stdio_connection has empty command. Skipping service.")
					isValidService = false
				}
			default:
				serviceLog.Warn("MCP service has no connection_type. Skipping service.")
				isValidService = false
			}
		} else if cmdService := service.GetCommandLineService(); cmdService != nil {
			if cmdService.GetCommand() == "" {
				serviceLog.Warn("Command line service has empty command. Skipping service.")
				isValidService = false
			}
		} else {
			serviceLog.Warn("Unknown service type. Skipping service.")
			isValidService = false
		}

		if !isValidService {
			continue
		}

		if service.GetCache() != nil {
			cacheLog := serviceLog.With("cache", "enabled")
			if service.GetCache().GetTtl() != nil && service.GetCache().GetTtl().GetSeconds() < 0 {
				cacheLog.Warn("Invalid cache timeout. Cache will be disabled.", "timeout", service.GetCache().GetTtl().AsDuration())
				val := false
				service.GetCache().SetIsEnabled(val)
			}
		}

		if authConfig := service.GetUpstreamAuthentication(); authConfig != nil {
			authLog := serviceLog.With("component", "authValidator")
			if apiKey := authConfig.GetApiKey(); apiKey != nil {
				if apiKey.GetHeaderName() == "" {
					authLog.Warn("API key 'header_name' is empty. Authentication may fail.")
				}
				if apiKey.GetApiKey() == "" {
					authLog.Warn("API key 'api_key' is empty. Authentication will fail.")
				}
			} else if bearerToken := authConfig.GetBearerToken(); bearerToken != nil {
				if bearerToken.GetToken() == "" {
					authLog.Warn("Bearer token 'token' is empty. Authentication will fail.")
				}
			} else if basicAuth := authConfig.GetBasicAuth(); basicAuth != nil {
				if basicAuth.GetUsername() == "" {
					authLog.Warn("Basic auth 'username' is empty. Authentication may fail.")
				}
			}
		}
		validServices = append(validServices, service)
	}
	config.SetUpstreamServices(validServices)
	return config, nil
}
