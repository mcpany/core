/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unused-vars */
/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Mock implementation of GrpcWebImpl for testing purposes.
 * Simulates a gRPC-Web client connection.
 */
export class GrpcWebImpl {
    constructor(_host: string, _options: any) {}
}

/**
 * Mock implementation of the Registration Service Client.
 * Used for testing interactions with the registration service without a backend.
 */
export class RegistrationServiceClientImpl {
    constructor(_rpc: any) {}
    GetService(_request: any, _metadata: any) { return Promise.resolve({}); }
}

/**
 * Mock configuration for an upstream service.
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
 * Mock definition of a tool.
 */
export interface ToolDefinition {
    name: string;
    [key: string]: any;
}

/**
 * Mock definition of a resource.
 */
export interface ResourceDefinition {
    uri: string;
    name: string;
    mimeType?: string;
    [key: string]: any;
}

/**
 * Mock definition of a prompt.
 */
export interface PromptDefinition {
    name: string;
    [key: string]: any;
}

/**
 * Mock definition of a credential.
 */
export interface Credential {
    id?: string;
    [key: string]: any;
}

/**
 * Mock definition for authentication.
 */
export interface Authentication {
    [key: string]: any;
}

/**
 * Mock response for listing services.
 */
export type ListServicesResponse = any;
/**
 * Mock response for getting a service.
 */
export type GetServiceResponse = any;
/**
 * Mock response for getting service status.
 */
export type GetServiceStatusResponse = any;
