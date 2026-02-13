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
    /**
     * Creates an instance of GrpcWebImpl.
     *
     * @param _host - The host address (unused).
     * @param _options - Configuration options (unused).
     */
    constructor(_host: string, _options: any) {}
}

/**
 * RegistrationServiceClientImpl mocks the service client registration.
 */
export class RegistrationServiceClientImpl {
    /**
     * Creates an instance of RegistrationServiceClientImpl.
     *
     * @param _rpc - The RPC implementation (unused).
     */
    constructor(_rpc: any) {}

    /**
     * GetService returns a mock service response.
     *
     * @param _request - The request object (unused).
     * @param _metadata - Metadata (unused).
     * @returns A promise resolving to an empty object.
     */
    GetService(_request: any, _metadata: any) { return Promise.resolve({}); }
}

/**
 * UpstreamServiceConfig defines the configuration for an upstream service.
 *
 * @property id - Unique identifier.
 * @property name - Service name.
 * @property version - Service version.
 * @property disable - Whether the service is disabled.
 * @property priority - Execution priority.
 * @property loadBalancingStrategy - Strategy for load balancing.
 * @property httpService - HTTP specific configuration.
 * @property grpcService - gRPC specific configuration.
 * @property commandLineService - CLI specific configuration.
 * @property mcpService - MCP specific configuration.
 * @property preCallHooks - Hooks to run before calls.
 * @property postCallHooks - Hooks to run after calls.
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
 * ToolDefinition defines a tool exposed by a service.
 *
 * @property name - The name of the tool.
 */
export interface ToolDefinition {
    name: string;
    [key: string]: any;
}

/**
 * ResourceDefinition defines a resource exposed by a service.
 *
 * @property uri - The URI of the resource.
 * @property name - The name of the resource.
 * @property mimeType - The MIME type of the resource content.
 */
export interface ResourceDefinition {
    uri: string;
    name: string;
    mimeType?: string;
    [key: string]: any;
}

/**
 * PromptDefinition defines a prompt template exposed by a service.
 *
 * @property name - The name of the prompt.
 */
export interface PromptDefinition {
    name: string;
    [key: string]: any;
}

/**
 * Credential defines security credentials.
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
 * GetServiceResponse represents the response for getting a single service.
 */
export type GetServiceResponse = any;

/**
 * GetServiceStatusResponse represents the status of a service.
 */
export type GetServiceStatusResponse = any;
