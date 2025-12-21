/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

// Types based on mcpany.config.v1 and mcpany.api.v1

/**
 * Represents the definition of a tool provided by a service.
 */
export interface ToolDefinition {
  /** The name of the tool. */
  name: string;
  /** A description of what the tool does. */
  description?: string;
  /** The source of the tool definition (e.g., configured manually or discovered dynamically). */
  source?: 'configured' | 'discovered';
  // Add detailed tool fields if necessary
}

/**
 * Represents the definition of a prompt provided by a service.
 */
export interface PromptDefinition {
  /** The name of the prompt. */
  name: string;
  /** A description of the prompt. */
  description?: string;
}

/**
 * Represents the definition of a resource provided by a service.
 */
export interface ResourceDefinition {
  /** The name of the resource. */
  name: string;
  /** The type of the resource. */
  type?: string;
}

/**
 * Represents a secret value which can be sourced from various locations.
 */
export interface SecretValue {
  /** The raw value of the secret. */
  key?: string;
  /** The name of the environment variable containing the secret. */
  env?: string;
  /** The path to a file containing the secret. */
  file?: string;
  /** A direct string value (use with caution). */
  value?: string;
}

/**
 * Configuration for Transport Layer Security (TLS).
 */
export interface TLSConfig {
  /** The expected server name for verification. */
  server_name?: string;
  /** Path to the CA certificate file. */
  ca_cert_path?: string;
  /** Path to the client certificate file. */
  client_cert_path?: string;
  /** Path to the client key file. */
  client_key_path?: string;
  /** Whether to skip server certificate verification. */
  insecure_skip_verify?: boolean;
}

/**
 * Configuration for connection pooling.
 */
export interface ConnectionPoolConfig {
  /** The maximum number of open connections. */
  max_connections?: number;
  /** The maximum number of idle connections. */
  max_idle_connections?: number;
  /** The maximum amount of time a connection may be reused. */
  idle_timeout?: string;
}

/**
 * Configuration for upstream authentication.
 */
export interface UpstreamAuthentication {
  // Define based on auth.proto if needed
  /** API key for authentication. */
  api_key?: string;
  /** Bearer token for authentication. */
  bearer_token?: string;
  /** Basic authentication credentials. */
  basic_auth?: { username: string; password?: SecretValue };
}

/**
 * Configuration for caching.
 */
export interface CacheConfig {
  /** Whether caching is enabled. */
  enabled?: boolean;
  /** Time-to-live for cached items. */
  ttl?: string;
}

/**
 * Configuration for rate limiting.
 */
export interface RateLimitConfig {
  /** Whether rate limiting is enabled. */
  is_enabled?: boolean;
  /** Maximum number of requests allowed per second. */
  requests_per_second?: number;
  /** Maximum burst size. */
  burst?: number;
}

// --- Service Types ---

/**
 * Configuration for a gRPC upstream service.
 */
export interface GrpcUpstreamService {
  /** The address of the gRPC service. */
  address: string;
  /** Whether to use gRPC reflection to discover services. */
  use_reflection?: boolean;
  /** TLS configuration for the connection. */
  tls_config?: TLSConfig;
  /** List of tools provided by the service. */
  tools?: ToolDefinition[];
  /** List of prompts provided by the service. */
  prompts?: PromptDefinition[];
  /** List of resources provided by the service. */
  resources?: ResourceDefinition[];
  /** Health check configuration. */
  health_check?: any;
}

/**
 * Configuration for an HTTP upstream service.
 */
export interface HttpUpstreamService {
  /** The base URL of the HTTP service. */
  address: string;
  /** TLS configuration for the connection. */
  tls_config?: TLSConfig;
  /** List of tools provided by the service. */
  tools?: ToolDefinition[];
  /** List of prompts provided by the service. */
  prompts?: PromptDefinition[];
  /** List of resources provided by the service. */
  resources?: ResourceDefinition[];
  /** Health check configuration. */
  health_check?: any;
}

/**
 * Configuration for a command-line upstream service.
 */
export interface CommandLineUpstreamService {
  /** The command to execute. */
  command: string;
  /** The working directory for the command. */
  working_directory?: string;
  /** Environment variables for the container/process. */
  container_environment?: any;
  /** Timeout for the command execution. */
  timeout?: string;
  /** The communication protocol to use (e.g., 'JSON'). */
  communication_protocol?: string; // 'JSON' or undefined
  /** Whether to run locally. */
  local?: boolean;
  /** Environment variables map. */
  env?: Record<string, SecretValue>;
  /** List of tools provided. */
  tools?: ToolDefinition[];
  /** List of prompts provided. */
  prompts?: PromptDefinition[];
  /** List of resources provided. */
  resources?: ResourceDefinition[];
}

/**
 * Configuration for an OpenAPI upstream service.
 */
export interface OpenapiUpstreamService {
  /** The base URL of the API. */
  address: string;
  /** The content of the OpenAPI spec. */
  spec_content?: string;
  /** The URL to fetch the OpenAPI spec from. */
  spec_url?: string;
  /** TLS configuration. */
  tls_config?: TLSConfig;
  /** List of tools provided. */
  tools?: ToolDefinition[];
  /** List of prompts provided. */
  prompts?: PromptDefinition[];
  /** List of resources provided. */
  resources?: ResourceDefinition[];
}

/**
 * Configuration for a WebSocket upstream service.
 */
export interface WebsocketUpstreamService {
  /** The WebSocket URL. */
  address: string;
  /** TLS configuration. */
  tls_config?: TLSConfig;
  /** List of tools provided. */
  tools?: ToolDefinition[];
  /** List of prompts provided. */
  prompts?: PromptDefinition[];
  /** List of resources provided. */
  resources?: ResourceDefinition[];
}

/**
 * Configuration for a WebRTC upstream service.
 */
export interface WebrtcUpstreamService {
  /** The WebRTC signaling address. */
  address: string;
  /** TLS configuration. */
  tls_config?: TLSConfig;
  /** List of tools provided. */
  tools?: ToolDefinition[];
  /** List of prompts provided. */
  prompts?: PromptDefinition[];
  /** List of resources provided. */
  resources?: ResourceDefinition[];
}

/**
 * Configuration for a GraphQL upstream service.
 */
export interface GraphQLUpstreamService {
  /** The GraphQL endpoint URL. */
  address: string;
  /** Map of query/mutation names to their definitions. */
  calls?: Record<string, any>;
}

/**
 * Configuration for an MCP upstream service (proxying another MCP server).
 */
export interface McpUpstreamService {
  /** Configuration for connecting via HTTP (SSE). */
  http_connection?: {
    http_address: string;
    tls_config?: TLSConfig;
  };
  /** Configuration for connecting via standard I/O (stdio). */
  stdio_connection?: {
    command: string;
    args?: string[];
    working_directory?: string;
    env?: Record<string, SecretValue>;
  };
  /** Configuration for connecting via a bundle. */
  bundle_connection?: {
    bundle_path: string;
    env?: Record<string, SecretValue>;
  };
  /** Whether to automatically discover tools from the upstream MCP server. */
  tool_auto_discovery?: boolean;
  /** List of tools explicitly defined. */
  tools?: ToolDefinition[];
  /** List of prompts explicitly defined. */
  prompts?: PromptDefinition[];
  /** List of resources explicitly defined. */
  resources?: ResourceDefinition[];
}

// --- Main Configuration ---

/**
 * Configuration for an upstream service.
 */
export interface UpstreamServiceConfig {
  /** The unique identifier of the service (generated by server). */
  id?: string;
  /** The display name of the service. */
  name: string;
  /** A sanitized version of the name suitable for use in identifiers. */
  sanitized_name?: string;
  /** The version of the service configuration. */
  version?: string;
  /** Whether the service is disabled. */
  disable?: boolean;
  /** The priority of the service (higher values take precedence). */
  priority?: number;

  /** Connection pool settings. */
  connection_pool?: ConnectionPoolConfig;
  /** Upstream authentication settings. */
  upstream_authentication?: UpstreamAuthentication;
  /** Caching settings. */
  cache?: CacheConfig;
  /** Rate limiting settings. */
  rate_limit?: RateLimitConfig;

  // OneOf service_config
  /** gRPC service configuration. */
  grpc_service?: GrpcUpstreamService;
  /** HTTP service configuration. */
  http_service?: HttpUpstreamService;
  /** Command-line service configuration. */
  command_line_service?: CommandLineUpstreamService;
  /** OpenAPI service configuration. */
  openapi_service?: OpenapiUpstreamService;
  /** WebSocket service configuration. */
  websocket_service?: WebsocketUpstreamService;
  /** WebRTC service configuration. */
  webrtc_service?: WebrtcUpstreamService;
  /** GraphQL service configuration. */
  graphql_service?: GraphQLUpstreamService;
  /** MCP service configuration. */
  mcp_service?: McpUpstreamService;
}

// --- API Request/Response Types ---

/**
 * Response object for listing services.
 */
export interface ListServicesResponse {
  /** The list of upstream service configurations. */
  services: UpstreamServiceConfig[];
}

/**
 * Response object for getting a single service.
 */
export interface GetServiceResponse {
  /** The requested service configuration. */
  service: UpstreamServiceConfig;
}

/**
 * Response object for getting service status.
 */
export interface GetServiceStatusResponse {
  /** The list of tools currently available from the service. */
  tools?: ToolDefinition[];
  /** Key-value map of metrics for the service. */
  metrics: Record<string, number>;
}

/**
 * Request object for registering a service.
 */
export interface RegisterServiceRequest {
  /** The configuration of the service to register. */
  config: UpstreamServiceConfig;
}

/**
 * Response object for registering a service.
 */
export interface RegisterServiceResponse {
  /** A message indicating the result of the registration. */
  message: string;
  /** The unique key assigned to the registered service. */
  service_key: string;
}

/**
 * Request object for updating a service.
 */
export interface UpdateServiceRequest {
  /** The new configuration for the service. */
  config: UpstreamServiceConfig;
}

/**
 * Response object for updating a service.
 */
export interface UpdateServiceResponse {
  /** The updated service configuration. */
  config: UpstreamServiceConfig;
}

/**
 * Request object for unregistering a service.
 */
export interface UnregisterServiceRequest {
  /** The name of the service to unregister. */
  service_name: string;
  /** The namespace of the service (optional). */
  namespace?: string;
}

/**
 * Response object for unregistering a service.
 */
export interface UnregisterServiceResponse {
  /** A message indicating the result of the unregistration. */
  message: string;
}
export type Tool = ToolDefinition;
export type Resource = ResourceDefinition;

export interface ListToolsResponse {
    tools: Tool[];
}

export interface ListResourcesResponse {
    resources: Resource[];
}
