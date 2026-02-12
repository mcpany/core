/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unused-vars */
/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * GrpcWebImpl is a mock implementation of the gRPC-Web client transport.
 *
 * It is used for testing and development when a real gRPC backend is not available.
 */
export class GrpcWebImpl {
    constructor(_host: string, _options: any) {}
}

/**
 * RegistrationServiceClientImpl simulates the registration service client for testing.
 *
 * It provides mock responses for service discovery and management operations.
 */
export class RegistrationServiceClientImpl {
    constructor(_rpc: any) {}
    /**
     * GetService retrieves a service definition (mock).
     * @param _request - The request object (unused).
     * @param _metadata - Metadata (unused).
     * @returns A promise resolving to an empty object.
     */
    GetService(_request: any, _metadata: any) { return Promise.resolve({}); }
}

/**
 * UpstreamServiceConfig defines the configuration for an upstream service.
 *
 * @property id - Unique identifier for the service.
 * @property name - Display name of the service.
 * @property version - Service version string.
 * @property disable - Whether the service is disabled.
 * @property priority - Execution priority for the service.
 * @property loadBalancingStrategy - Strategy for load balancing requests.
 * @property httpService - Configuration for HTTP-based services.
 * @property grpcService - Configuration for gRPC-based services.
 * @property commandLineService - Configuration for CLI-based services.
 * @property mcpService - Configuration for MCP-based services.
 * @property preCallHooks - Hooks to run before execution.
 * @property postCallHooks - Hooks to run after execution.
 */
export interface UpstreamServiceConfig {
    id?: string;
    name?: string;
    version?: string;
    disable?: boolean;
    priority?: number;
    loadBalancingStrategy?: string;
    httpService?: any;
    grpcService?: any;
    commandLineService?: any;
    mcpService?: any;
    preCallHooks?: any[];
    postCallHooks?: any[];
    [key: string]: any;
}

/**
 * ToolDefinition describes a tool available in the system.
 *
 * @property name - The unique name of the tool.
 */
export interface ToolDefinition {
    name: string;
    [key: string]: any;
}

/**
 * ResourceDefinition describes a resource exposed by the system.
 *
 * @property uri - The URI of the resource.
 * @property name - The display name of the resource.
 * @property mimeType - The MIME type of the resource content.
 */
export interface ResourceDefinition {
    uri: string;
    name: string;
    mimeType?: string;
    [key: string]: any;
}

/**
 * PromptDefinition describes a prompt template.
 *
 * @property name - The unique name of the prompt.
 */
export interface PromptDefinition {
    name: string;
    [key: string]: any;
}

/**
 * Credential represents authentication credentials.
 *
 * @property id - Unique identifier for the credential.
 */
export interface Credential {
    id?: string;
    [key: string]: any;
}

/**
 * Authentication defines authentication settings.
 */
export interface Authentication {
    [key: string]: any;
}

/**
 * ListServicesResponse represents the response for listing services.
 */
export type ListServicesResponse = any;
/**
 * GetServiceResponse represents the response for retrieving a single service.
 */
export type GetServiceResponse = any;
/**
 * GetServiceStatusResponse represents the response for service status checks.
 */
export type GetServiceStatusResponse = any;
