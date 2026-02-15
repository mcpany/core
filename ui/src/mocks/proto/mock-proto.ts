/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unused-vars */
/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Mock implementation of the gRPC web client for testing purposes.
 */
export class GrpcWebImpl {
    /**
     * Creates a new instance of the mock gRPC web client.
     */
    constructor(_host: string, _options: any) {}
}

/**
 * Mock implementation of the Registration Service Client.
 */
export class RegistrationServiceClientImpl {
    /**
     * Creates a new instance of the mock Registration Service Client.
     */
    constructor(_rpc: any) {}

    /**
     * Mock implementation of the GetService method.
     */
    GetService(_request: any, _metadata: any) { return Promise.resolve({}); }
}

/**
 * Configuration for an upstream service, defining how MCP Any connects to backend systems.
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
 * Defines a tool exposed by an upstream service or the core server.
 */
export interface ToolDefinition {
    name: string;
    [key: string]: any;
}

/**
 * Defines a resource exposed by an upstream service.
 */
export interface ResourceDefinition {
    uri: string;
    name: string;
    mimeType?: string;
    [key: string]: any;
}

/**
 * Defines a prompt template exposed by an upstream service.
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
 * Defines authentication settings for a service.
 */
export interface Authentication {
    [key: string]: any;
}

/**
 * Response payload for listing available services.
 */
export type ListServicesResponse = any;
/**
 * Response payload for retrieving a specific service configuration.
 */
export type GetServiceResponse = any;
/**
 * Response payload for retrieving the status of a service.
 */
export type GetServiceStatusResponse = any;
