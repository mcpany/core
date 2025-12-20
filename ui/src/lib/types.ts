/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// Types based on mcpany.config.v1 and mcpany.api.v1

export interface ToolDefinition {
  name: string;
  description?: string;
  source?: 'configured' | 'discovered';
  // Add detailed tool fields if necessary
}

export interface PromptDefinition {
  name: string;
  description?: string;
}

export interface ResourceDefinition {
  name: string;
  type?: string;
}

export interface SecretValue {
  key?: string;
  env?: string;
  file?: string;
  value?: string;
}

export interface TLSConfig {
  server_name?: string;
  ca_cert_path?: string;
  client_cert_path?: string;
  client_key_path?: string;
  insecure_skip_verify?: boolean;
}

export interface ConnectionPoolConfig {
  max_connections?: number;
  max_idle_connections?: number;
  idle_timeout?: string;
}

export interface UpstreamAuthentication {
  // Define based on auth.proto if needed
  api_key?: string;
  bearer_token?: string;
  basic_auth?: { username: string; password?: SecretValue };
}

export interface CacheConfig {
  enabled?: boolean;
  ttl?: string;
}

export interface RateLimitConfig {
  is_enabled?: boolean;
  requests_per_second?: number;
  burst?: number;
}

// --- Service Types ---

export interface GrpcUpstreamService {
  address: string;
  use_reflection?: boolean;
  tls_config?: TLSConfig;
  tools?: ToolDefinition[];
  prompts?: PromptDefinition[];
  resources?: ResourceDefinition[];
  health_check?: any;
}

export interface HttpUpstreamService {
  address: string;
  tls_config?: TLSConfig;
  tools?: ToolDefinition[];
  prompts?: PromptDefinition[];
  resources?: ResourceDefinition[];
  health_check?: any;
}

export interface CommandLineUpstreamService {
  command: string;
  working_directory?: string;
  container_environment?: any;
  timeout?: string;
  communication_protocol?: string; // 'JSON' or undefined
  local?: boolean;
  env?: Record<string, SecretValue>;
  tools?: ToolDefinition[];
  prompts?: PromptDefinition[];
  resources?: ResourceDefinition[];
}

export interface OpenapiUpstreamService {
  address: string;
  spec_content?: string;
  spec_url?: string;
  tls_config?: TLSConfig;
  tools?: ToolDefinition[];
  prompts?: PromptDefinition[];
  resources?: ResourceDefinition[];
}

export interface WebsocketUpstreamService {
  address: string;
  tls_config?: TLSConfig;
  tools?: ToolDefinition[];
  prompts?: PromptDefinition[];
  resources?: ResourceDefinition[];
}

export interface WebrtcUpstreamService {
  address: string;
  tls_config?: TLSConfig;
  tools?: ToolDefinition[];
  prompts?: PromptDefinition[];
  resources?: ResourceDefinition[];
}

export interface GraphQLUpstreamService {
  address: string;
  calls?: Record<string, any>;
}

export interface McpUpstreamService {
  http_connection?: {
    http_address: string;
    tls_config?: TLSConfig;
  };
  stdio_connection?: {
    command: string;
    args?: string[];
    working_directory?: string;
    env?: Record<string, SecretValue>;
  };
  bundle_connection?: {
    bundle_path: string;
    env?: Record<string, SecretValue>;
  };
  tool_auto_discovery?: boolean;
  tools?: ToolDefinition[];
  prompts?: PromptDefinition[];
  resources?: ResourceDefinition[];
}

// --- Main Configuration ---

export interface UpstreamServiceConfig {
  id?: string; // Generated on server usually
  name: string;
  sanitized_name?: string;
  version?: string;
  disable?: boolean;
  priority?: number;

  connection_pool?: ConnectionPoolConfig;
  upstream_authentication?: UpstreamAuthentication;
  cache?: CacheConfig;
  rate_limit?: RateLimitConfig;

  // OneOf service_config
  grpc_service?: GrpcUpstreamService;
  http_service?: HttpUpstreamService;
  command_line_service?: CommandLineUpstreamService;
  openapi_service?: OpenapiUpstreamService;
  websocket_service?: WebsocketUpstreamService;
  webrtc_service?: WebrtcUpstreamService;
  graphql_service?: GraphQLUpstreamService;
  mcp_service?: McpUpstreamService;
}

// --- API Request/Response Types ---

export interface ListServicesResponse {
  services: UpstreamServiceConfig[];
}

export interface GetServiceResponse {
  service: UpstreamServiceConfig;
}

export interface GetServiceStatusResponse {
  tools?: ToolDefinition[];
  metrics: Record<string, number>;
}

export interface RegisterServiceRequest {
  config: UpstreamServiceConfig;
}

export interface RegisterServiceResponse {
  message: string;
  service_key: string;
}

export interface UpdateServiceRequest {
  config: UpstreamServiceConfig;
}

export interface UpdateServiceResponse {
  config: UpstreamServiceConfig;
}

export interface UnregisterServiceRequest {
  service_name: string;
  namespace?: string;
}

export interface UnregisterServiceResponse {
  message: string;
}
