/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unused-vars */
/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Mock implementation of the gRPC-Web client transport.
 * Used for testing environments where a real gRPC connection is not available.
 */
export class GrpcWebImpl {
    /**
     * Creates an instance of GrpcWebImpl.
     * @param _host - The host URL of the gRPC server.
     * @param _options - Configuration options for the client.
     */
    constructor(_host: string, _options: any) {}
}

/**
 * Mock implementation of the RegistrationServiceClient.
 * Simulates the client-side stub for the Registration Service.
 */
export class RegistrationServiceClientImpl {
    /**
     * Creates an instance of RegistrationServiceClientImpl.
     * @param _rpc - The RPC transport implementation.
     */
    constructor(_rpc: any) {}

    /**
     * Simulates fetching service details.
     * @param _request - The request object.
     * @param _metadata - Optional metadata.
     * @returns A promise resolving to an empty object.
     */
    GetService(_request: any, _metadata: any) { return Promise.resolve({}); }
}

/**
 * Configuration definition for an upstream service.
 * Matches the structure of the `UpstreamServiceConfig` protobuf message.
 */
export interface UpstreamServiceConfig {
    /** Unique identifier for the service. */
    id?: string;
    /** Display name of the service. */
    name?: string;
    /** Version string of the service. */
    version?: string;
    /** Whether the service is currently disabled. */
    disable?: boolean;
    /** Priority order for service selection. */
    priority?: number;
    /** Strategy used for load balancing requests. */
    loadBalancingStrategy?: string;
    /** Configuration for HTTP-based services. */
    httpService?: any;
    /** Configuration for gRPC-based services. */
    grpcService?: any;
    /** Configuration for local command-line tool services. */
    commandLineService?: any;
    /** Configuration for MCP-compliant services. */
    mcpService?: any;
    /** Hooks executed before calling the service. */
    preCallHooks?: any[];
    /** Hooks executed after calling the service. */
    postCallHooks?: any[];
    /** Allow for additional properties. */
    [key: string]: any;
}

/**
 * Definition of an MCP Tool.
 * Matches the structure of the `ToolDefinition` protobuf message.
 */
export interface ToolDefinition {
    /** The unique name of the tool. */
    name: string;
    /** Allow for additional properties. */
    [key: string]: any;
}

/**
 * Definition of an MCP Resource.
 * Matches the structure of the `ResourceDefinition` protobuf message.
 */
export interface ResourceDefinition {
    /** The URI of the resource. */
    uri: string;
    /** The human-readable name of the resource. */
    name: string;
    /** The MIME type of the resource content. */
    mimeType?: string;
    /** Allow for additional properties. */
    [key: string]: any;
}

/**
 * Definition of an MCP Prompt.
 * Matches the structure of the `PromptDefinition` protobuf message.
 */
export interface PromptDefinition {
    /** The unique name of the prompt. */
    name: string;
    /** Allow for additional properties. */
    [key: string]: any;
}

/**
 * Definition of a Credential.
 * Represents a stored secret or credential entity.
 */
export interface Credential {
    /** Unique identifier for the credential. */
    id?: string;
    /** Allow for additional properties. */
    [key: string]: any;
}

/**
 * Authentication configuration.
 * Defines how to authenticate with a service.
 */
export interface Authentication {
    /** Allow for additional properties. */
    [key: string]: any;
}

/**
 * Response type for listing services.
 */
export type ListServicesResponse = any;
/**
 * Response type for getting service details.
 */
export type GetServiceResponse = any;
/**
 * Response type for getting service status.
 */
export type GetServiceStatusResponse = any;
