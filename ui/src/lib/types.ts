/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// Mock types based on the protobuf definition for mcpany.config.v1

export interface ToolDefinition {
  name: string;
  description: string;
  // This flag is a client-side concept to differentiate between
  // statically configured tools and those discovered at runtime.
  source: 'configured' | 'discovered';
}

export interface PromptDefinition {
    name: string;
    description: string;
}

export interface ResourceDefinition {
    name: string;
    type: string;
}

export interface UpstreamServiceConfig {
  id: string;
  name: string;
  version?: string;
  grpc_service?: GrpcUpstreamService;
  http_service?: HttpUpstreamService;
  command_line_service?: CommandLineUpstreamService;
  disable?: boolean;
}

export interface GrpcUpstreamService {
  address: string;
  use_reflection: boolean;
  tls_config?: TLSConfig;
  tools?: ToolDefinition[];
  prompts?: PromptDefinition[];
  resources?: ResourceDefinition[];
}

export interface HttpUpstreamService {
  address: string;
  tls_config?: TLSConfig;
  tools?: ToolDefinition[];
  prompts?: PromptDefinition[];
  resources?: ResourceDefinition[];
}

export interface CommandLineUpstreamService {
  command: string;
  tools?: ToolDefinition[];
  prompts?: PromptDefinition[];
  resources?: ResourceDefinition[];
}

export interface TLSConfig {
  server_name: string;
  insecure_skip_verify: boolean;
}

// Mock type for the response from mcpany.api.v1.ListServices
export interface ListServicesResponse {
  services: UpstreamServiceConfig[];
}

// Mock type for a single service response
export interface GetServiceResponse {
    service: UpstreamServiceConfig;
}

// Mock types for GetServiceStatus
export interface GetServiceStatusRequest {
    service_id: string;
}

export interface GetServiceStatusResponse {
    metrics: Record<string, number>;
}
