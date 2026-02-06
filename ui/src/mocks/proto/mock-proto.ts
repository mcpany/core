/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unused-vars */
/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Mock implementation of GrpcWebImpl.
 */
export class GrpcWebImpl {
    constructor(_host: string, _options: any) {}
}

/**
 * The RegistrationServiceClientImpl class.
 */
export class RegistrationServiceClientImpl {
    constructor(_rpc: any) {}
    GetService(_request: any, _metadata: any) { return Promise.resolve({}); }
}

/**
 * UpstreamServiceConfig type definition.
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
 * ToolDefinition type definition.
 */
export interface ToolDefinition {
    name: string;
    [key: string]: any;
}

/**
 * ResourceDefinition type definition.
 */
export interface ResourceDefinition {
    uri: string;
    name: string;
    mimeType?: string;
    [key: string]: any;
}

/**
 * PromptDefinition type definition.
 */
export interface PromptDefinition {
    name: string;
    [key: string]: any;
}

/**
 * Credential type definition.
 */
export interface Credential {
    id?: string;
    [key: string]: any;
}

/**
 * Authentication type definition.
 */
export interface Authentication {
    [key: string]: any;
}

/**
 * ListServicesResponse type definition.
 */
export type ListServicesResponse = any;
/**
 * GetServiceResponse type definition.
 */
export type GetServiceResponse = any;
/**
 * GetServiceStatusResponse type definition.
 */
export type GetServiceStatusResponse = any;
