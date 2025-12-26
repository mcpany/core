/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Configuration for an upstream service.
 */
export interface UpstreamServiceConfig {
    /** Unique identifier for the service. */
    id?: string;
    /** Display name of the service. */
    name: string;
    /** Version string of the service. */
    version?: string;
    /** Whether the service is disabled. */
    disable?: boolean;
    /** Configuration for HTTP services. */
    http_service?: { address: string };
    /** Configuration for gRPC services. */
    grpc_service?: { address: string };
    /** Configuration for command-line services. */
    command_line_service?: { command: string; args?: string[]; env?: Record<string, string> };
    /** Configuration for MCP services (proxying). */
    mcp_service?: { http_connection?: { http_address: string }; sse_connection?: { sse_address: string }; stdio_connection?: { command: string } };
}

/**
 * Definition of a tool exposed by an MCP server.
 */
export interface ToolDefinition {
    /** Name of the tool. */
    name: string;
    /** Description of what the tool does. */
    description: string;
    /** JSON Schema for the tool's input arguments. */
    schema?: any;
    /** Whether the tool is enabled for use. */
    enabled?: boolean;
    /** Name of the service that provides this tool. */
    serviceName?: string;
}

/**
 * Definition of a resource exposed by an MCP server.
 */
export interface ResourceDefinition {
    /** URI identifying the resource. */
    uri: string;
    /** Name of the resource. */
    name: string;
    /** MIME type of the resource content. */
    mimeType?: string;
    /** Description of the resource. */
    description?: string;
    /** Whether the resource is enabled for access. */
    enabled?: boolean;
    /** Name of the service that provides this resource. */
    serviceName?: string;
}

/**
 * Definition of a prompt template exposed by an MCP server.
 */
export interface PromptDefinition {
    /** Name of the prompt. */
    name: string;
    /** Description of the prompt. */
    description?: string;
    /** List of arguments accepted by the prompt. */
    arguments?: any[];
    /** Whether the prompt is enabled for use. */
    enabled?: boolean;
    /** Name of the service that provides this prompt. */
    serviceName?: string;
}

/**
 * Client for interacting with the MCP Any API.
 */
export const apiClient = {
    // Services

    /**
     * Lists all registered upstream services.
     * @returns A promise resolving to a list of services.
     */
    listServices: async () => {
        const res = await fetch('/api/services');
        if (!res.ok) throw new Error('Failed to fetch services');
        return res.json();
    },

    /**
     * Sets the enabled/disabled status of a service.
     * @param name - The name of the service.
     * @param disable - True to disable, false to enable.
     * @returns A promise resolving to the updated service status.
     */
    setServiceStatus: async (name: string, disable: boolean) => {
        const res = await fetch('/api/services', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ action: 'toggle', name, disable })
        });
        if (!res.ok) throw new Error('Failed to update service status');
        return res.json();
    },

    /**
     * Registers a new upstream service.
     * @param config - The configuration for the new service.
     * @returns A promise resolving to the registered service details.
     */
    registerService: async (config: UpstreamServiceConfig) => {
        const res = await fetch('/api/services', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config)
        });
        if (!res.ok) throw new Error('Failed to register service');
        return res.json();
    },

    /**
     * Updates an existing upstream service.
     * @param config - The updated configuration for the service.
     * @returns A promise resolving to the updated service details.
     */
    updateService: async (config: UpstreamServiceConfig) => {
         const res = await fetch('/api/services', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config)
        });
        if (!res.ok) throw new Error('Failed to update service');
        return res.json();
    },

    // Tools

    /**
     * Lists all available tools.
     * @returns A promise resolving to a list of tools.
     */
    listTools: async () => {
        const res = await fetch('/api/tools');
        if (!res.ok) throw new Error('Failed to fetch tools');
        return res.json();
    },

    /**
     * Sets the enabled/disabled status of a tool.
     * @param name - The name of the tool.
     * @param enabled - True to enable, false to disable.
     * @returns A promise resolving to the updated tool status.
     */
    setToolStatus: async (name: string, enabled: boolean) => {
        const res = await fetch('/api/tools', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, enabled })
        });
        if (!res.ok) throw new Error('Failed to update tool status');
        return res.json();
    },

    // Resources

    /**
     * Lists all available resources.
     * @returns A promise resolving to a list of resources.
     */
    listResources: async () => {
        const res = await fetch('/api/resources');
        if (!res.ok) throw new Error('Failed to fetch resources');
        return res.json();
    },

    /**
     * Sets the enabled/disabled status of a resource.
     * @param uri - The URI of the resource.
     * @param enabled - True to enable, false to disable.
     * @returns A promise resolving to the updated resource status.
     */
    setResourceStatus: async (uri: string, enabled: boolean) => {
         const res = await fetch('/api/resources', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ uri, enabled })
        });
        if (!res.ok) throw new Error('Failed to update resource status');
        return res.json();
    },

    // Prompts

    /**
     * Lists all available prompts.
     * @returns A promise resolving to a list of prompts.
     */
    listPrompts: async () => {
        const res = await fetch('/api/prompts');
        if (!res.ok) throw new Error('Failed to fetch prompts');
        return res.json();
    },

    /**
     * Sets the enabled/disabled status of a prompt.
     * @param name - The name of the prompt.
     * @param enabled - True to enable, false to disable.
     * @returns A promise resolving to the updated prompt status.
     */
    setPromptStatus: async (name: string, enabled: boolean) => {
        const res = await fetch('/api/prompts', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, enabled })
        });
        if (!res.ok) throw new Error('Failed to update prompt status');
        return res.json();
    },
};
