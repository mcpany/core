// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"

	cpb "github.com/mcpany/core/proto/config/v1"
)

// Validate checks the given MCP Any server configuration for correctness.
// It returns a slice of strings, where each string is an error message.
// An empty or nil slice indicates that the configuration is valid.
func Validate(config *cpb.McpAnyServerConfig) []string {
	var errs []string

	if config == nil {
		errs = append(errs, "top-level configuration is nil")
		return errs
	}

	errs = append(errs, validateGlobalSettings(config.GetGlobalSettings())...)
	errs = append(errs, validateUpstreamServices(config.GetUpstreamServices())...)
	errs = append(errs, validateUpstreamServiceCollections(config.GetUpstreamServiceCollections())...)

	return errs
}

func validateGlobalSettings(settings *cpb.GlobalSettings) []string {
	var errs []string
	if settings == nil {
		// This is a valid case, so we don't return an error.
		return errs
	}
	if settings.GetMcpListenAddress() == "" {
		errs = append(errs, "global_settings.mcp_listen_address is not set")
	}
	return errs
}

func validateUpstreamServices(services []*cpb.UpstreamServiceConfig) []string {
	var errs []string
	for i, service := range services {
		errs = append(errs, validateUpstreamService(service, fmt.Sprintf("upstream_services[%d]", i))...)
	}
	return errs
}

func validateUpstreamService(service *cpb.UpstreamServiceConfig, path string) []string {
	var errs []string
	if service.GetName() == "" {
		errs = append(errs, fmt.Sprintf("%s.name is not set", path))
	}

	if service.GetServiceConfig() == nil {
		errs = append(errs, fmt.Sprintf("%s.service_config is not set", path))
		return errs
	}

	switch c := service.GetServiceConfig().(type) {
	case *cpb.UpstreamServiceConfig_HttpService:
		if c.HttpService.GetAddress() == "" {
			errs = append(errs, fmt.Sprintf("%s.http_service.address is not set", path))
		}
	case *cpb.UpstreamServiceConfig_GrpcService:
		if c.GrpcService.GetAddress() == "" {
			errs = append(errs, fmt.Sprintf("%s.grpc_service.address is not set", path))
		}
	case *cpb.UpstreamServiceConfig_OpenapiService:
		if c.OpenapiService.GetAddress() == "" && c.OpenapiService.GetOpenapiSpec() == "" {
			errs = append(errs, fmt.Sprintf("%s.openapi_service requires either an address or an openapi_spec", path))
		}
	case *cpb.UpstreamServiceConfig_CommandLineService:
		if c.CommandLineService.GetCommand() == "" {
			errs = append(errs, fmt.Sprintf("%s.command_line_service.command is not set", path))
		}
	}
	return errs
}

func validateUpstreamServiceCollections(collections []*cpb.UpstreamServiceCollection) []string {
	var errs []string
	for i, collection := range collections {
		path := fmt.Sprintf("upstream_service_collections[%d]", i)
		if collection.GetName() == "" {
			errs = append(errs, fmt.Sprintf("%s.name is not set", path))
		}
		if collection.GetHttpUrl() == "" {
			errs = append(errs, fmt.Sprintf("%s.http_url is not set", path))
		}
	}
	return errs
}
