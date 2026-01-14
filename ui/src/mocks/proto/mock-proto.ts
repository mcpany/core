/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unused-vars */
/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

export class GrpcWebImpl {
    constructor(_host: string, _options: any) {}
}

export class RegistrationServiceClientImpl {
    constructor(_rpc: any) {}
    GetService(_request: any, _metadata: any) { return Promise.resolve({}); }
}

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

export interface ToolDefinition {
    name: string;
    [key: string]: any;
}

export interface ResourceDefinition {
    uri: string;
    name: string;
    mimeType?: string;
    [key: string]: any;
}

export interface PromptDefinition {
    name: string;
    [key: string]: any;
}

export interface Credential {
    id?: string;
    [key: string]: any;
}

export interface Authentication {
    [key: string]: any;
}

export type ListServicesResponse = any;
export type GetServiceResponse = any;
export type GetServiceStatusResponse = any;
