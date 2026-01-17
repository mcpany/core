/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unused-vars */
/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * GrpcWebImpl is a mock implementation of the gRPC web client.
 */
export class GrpcWebImpl {
    constructor(_host: string, _options: any) {}
}

/**
 * RegistrationServiceClientImpl is a mock implementation of the Registration Service client.
 */
export class RegistrationServiceClientImpl {
    constructor(_rpc: any) {}
    GetService(_request: any, _metadata: any) { return Promise.resolve({}); }
}

/**
 * UpstreamServiceConfig represents the configuration for an upstream service.
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
 * ToolDefinition represents the definition of a tool.
 */
export interface ToolDefinition {
    name: string;
    [key: string]: any;
}

/**
 * ResourceDefinition represents the definition of a resource.
 */
export interface ResourceDefinition {
    uri: string;
    name: string;
    mimeType?: string;
    [key: string]: any;
}

/**
 * PromptDefinition represents the definition of a prompt.
 */
export interface PromptDefinition {
    name: string;
    [key: string]: any;
}

/**
 * Credential represents a credential.
 */
export interface Credential {
    id?: string;
    [key: string]: any;
}

/**
 * Authentication represents authentication settings.
 */
export interface Authentication {
    [key: string]: any;
}

/** ListServicesResponse represents the response for listing services. */
export type ListServicesResponse = any;
/** GetServiceResponse represents the response for getting a service. */
export type GetServiceResponse = any;
/** GetServiceStatusResponse represents the response for getting service status. */
export type GetServiceStatusResponse = any;
