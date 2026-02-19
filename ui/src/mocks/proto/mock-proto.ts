/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unused-vars */
/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
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
    description?: string;
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

// Mock types for HttpCallDefinition
/**
 * ParameterType enum.
 */
export enum ParameterType {
    STRING = 0,
    NUMBER = 1,
    INTEGER = 2,
    BOOLEAN = 3,
    ARRAY = 4,
    OBJECT = 5,
}

/**
 * HttpMethod enum for HTTP call definition.
 */
export enum HttpCallDefinition_HttpMethod {
    HTTP_METHOD_UNSPECIFIED = 0,
    HTTP_METHOD_GET = 1,
    HTTP_METHOD_POST = 2,
    HTTP_METHOD_PUT = 3,
    HTTP_METHOD_DELETE = 4,
    HTTP_METHOD_PATCH = 5,
}

/**
 * Output format enum for output transformer.
 */
export enum OutputTransformer_OutputFormat {
    JSON = 0,
    XML = 1,
    TEXT = 2,
    RAW_BYTES = 3,
    JQ = 4,
}

/**
 * InputTransformer interface.
 */
export interface InputTransformer {
    template?: string;
    webhook?: any;
}

/**
 * OutputTransformer interface.
 */
export interface OutputTransformer {
    format: OutputTransformer_OutputFormat;
    extractionRules?: { [key: string]: string };
    template?: string;
    jqQuery?: string;
}

/**
 * HttpParameterMapping interface.
 */
export interface HttpParameterMapping {
    schema?: {
        name: string;
        description?: string;
        type: ParameterType;
        isRequired?: boolean;
        defaultValue?: any;
    };
    secret?: any;
    disableEscape?: boolean;
}

/**
 * HttpCallDefinition interface.
 */
export interface HttpCallDefinition {
    id?: string;
    method: HttpCallDefinition_HttpMethod;
    endpointPath: string;
    parameters: HttpParameterMapping[];
    inputTransformer?: InputTransformer;
    outputTransformer?: OutputTransformer;
    [key: string]: any;
}
