/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unused-vars */
/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * GrpcWebImpl is a helper class for gRPC-Web clients.
 *
 * Summary: helper class for gRPC-Web clients.
 */
export class GrpcWebImpl {
    /**
     * Constructs a new GrpcWebImpl.
     *
     * Summary: Initializes the gRPC-Web implementation.
     *
     * Parameters:
     * - _host: string. The host address.
     * - _options: any. Configuration options.
     */
    constructor(_host: string, _options: any) {}
}

/**
 * The RegistrationServiceClientImpl class.
 *
 * Summary: Client implementation for the Registration Service.
 */
export class RegistrationServiceClientImpl {
    /**
     * Constructs a new RegistrationServiceClientImpl.
     *
     * Summary: Initializes the client.
     *
     * Parameters:
     * - _rpc: any. The RPC implementation to use.
     */
    constructor(_rpc: any) {}

    /**
     * GetService retrieves a service definition.
     *
     * Summary: Fetches service details.
     *
     * Parameters:
     * - _request: any. The request object.
     * - _metadata: any. metadata.
     *
     * Returns:
     * - Promise<any>: The service definition.
     */
    GetService(_request: any, _metadata: any) { return Promise.resolve({}); }
}

/**
 * UpstreamServiceConfig type definition.
 *
 * Summary: Configuration for an upstream service.
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
 *
 * Summary: Defines an MCP tool.
 */
export interface ToolDefinition {
    name: string;
    description?: string;
    [key: string]: any;
}

/**
 * ResourceDefinition type definition.
 *
 * Summary: Defines an MCP resource.
 */
export interface ResourceDefinition {
    uri: string;
    name: string;
    mimeType?: string;
    [key: string]: any;
}

/**
 * PromptDefinition type definition.
 *
 * Summary: Defines an MCP prompt.
 */
export interface PromptDefinition {
    name: string;
    [key: string]: any;
}

/**
 * Credential type definition.
 *
 * Summary: Defines credentials for authentication.
 */
export interface Credential {
    id?: string;
    [key: string]: any;
}

/**
 * Authentication type definition.
 *
 * Summary: Defines authentication configuration.
 */
export interface Authentication {
    [key: string]: any;
}

/**
 * ListServicesResponse type definition.
 *
 * Summary: Response for listing services.
 */
export type ListServicesResponse = any;
/**
 * GetServiceResponse type definition.
 *
 * Summary: Response for getting a service.
 */
export type GetServiceResponse = any;
/**
 * GetServiceStatusResponse type definition.
 *
 * Summary: Response for getting service status.
 */
export type GetServiceStatusResponse = any;

// Mock types for HttpCallDefinition

/**
 * ParameterType defines the supported data types.
 *
 * Summary: Enumerates supported parameter types.
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
 * HttpCallDefinition_HttpMethod defines supported HTTP methods.
 *
 * Summary: Enumerates standard HTTP methods.
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
 * OutputTransformer_OutputFormat defines output formats.
 *
 * Summary: Enumerates supported output formats.
 */
export enum OutputTransformer_OutputFormat {
    JSON = 0,
    XML = 1,
    TEXT = 2,
    RAW_BYTES = 3,
    JQ = 4,
}

/**
 * InputTransformer defines how to transform input.
 *
 * Summary: Configures input transformation.
 */
export interface InputTransformer {
    template?: string;
    webhook?: any;
}

/**
 * OutputTransformer defines how to transform output.
 *
 * Summary: Configures output transformation.
 */
export interface OutputTransformer {
    format: OutputTransformer_OutputFormat;
    extractionRules?: { [key: string]: string };
    template?: string;
    jqQuery?: string;
}

/**
 * HttpParameterMapping defines parameter mapping.
 *
 * Summary: Maps input parameters to HTTP request parameters.
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
 * HttpCallDefinition defines an HTTP call.
 *
 * Summary: detailed definition of an HTTP API call.
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
