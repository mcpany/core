/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unused-vars */
/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Mock implementation of GrpcWebImpl for testing purposes.
 */
export class GrpcWebImpl {
    /**
     * Creates a new instance of GrpcWebImpl.
     * @param _host - The host address.
     * @param _options - Configuration options.
     */
    constructor(_host: string, _options: any) {}
}

/**
 * Mock implementation of RegistrationServiceClient.
 */
export class RegistrationServiceClientImpl {
    /**
     * Creates a new instance of RegistrationServiceClientImpl.
     * @param _rpc - The RPC implementation.
     */
    constructor(_rpc: any) {}

    /**
     * Mocks the GetService method.
     * @param _request - The request object.
     * @param _metadata - Metadata.
     * @returns A promise resolving to an empty object.
     */
    GetService(_request: any, _metadata: any) { return Promise.resolve({}); }
}

/**
 * Defines the configuration for an upstream service.
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
 * Defines a tool and its properties.
 */
export interface ToolDefinition {
    name: string;
    [key: string]: any;
}

/**
 * Defines a resource available in the system.
 */
export interface ResourceDefinition {
    uri: string;
    name: string;
    mimeType?: string;
    [key: string]: any;
}

/**
 * Defines a prompt available in the system.
 */
export interface PromptDefinition {
    name: string;
    [key: string]: any;
}

/**
 * Represents a credential used for authentication.
 */
export interface Credential {
    id?: string;
    [key: string]: any;
}

/**
 * Defines authentication settings.
 */
export interface Authentication {
    [key: string]: any;
}

/**
 * Response type for listing services.
 */
export type ListServicesResponse = any;
/**
 * Response type for getting a specific service.
 */
export type GetServiceResponse = any;
/**
 * Response type for getting service status.
 */
export type GetServiceStatusResponse = any;
