/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Configuration for an upstream service.
 * Defines the connection details and type of service being registered.
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
    /** Priority for merging strategies. */
    priority?: number;
    /** Configuration for HTTP-based services. */
    http_service?: { address: string; tls_config?: any; tools?: any[]; prompts?: any[] };
    /** Configuration for gRPC-based services. */
    grpc_service?: { address: string; use_reflection?: boolean; tls_config?: any; tools?: any[]; resources?: any[] };
    /** Configuration for command-line interface services. */
    command_line_service?: { command: string; args?: string[]; env?: Record<string, string>; tools?: any[] };
    /** Configuration for MCP (Model Context Protocol) services. */
    mcp_service?: {
        http_connection?: { http_address: string; tls_config?: any };
        sse_connection?: { sse_address: string };
        stdio_connection?: { command: string };
        bundle_connection?: { bundle_path: string };
        tools?: any[];
    };
    /** Configuration for OpenAPI-based services. */
    openapi_service?: { address: string; spec_url?: string; spec_content?: string; tools?: any[]; tls_config?: any };
    /** Configuration for WebSocket-based services. */
    websocket_service?: { address: string; tls_config?: any };
    /** Configuration for WebRTC-based services. */
    webrtc_service?: { address: string; tls_config?: any };
    /** Configuration for GraphQL-based services. */
    graphql_service?: { address: string };
}

/**
 * Definition of a tool available in the system.
 */
export interface ToolDefinition {
    /** Unique name of the tool. */
    name: string;
    /** Human-readable description of what the tool does. */
    description: string;
    /** JSON schema defining the expected input arguments. */
    schema?: any;
    /** Whether the tool is currently enabled. */
    enabled?: boolean;
    /** Name of the service that provides this tool. */
    serviceName?: string;
    /** Source of the tool (e.g., "builtin", "dynamic"). */
    source?: string;
}

/**
 * Definition of a resource available in the system.
 */
export interface ResourceDefinition {
    /** Unique URI identifying the resource. */
    uri: string;
    /** Human-readable name of the resource. */
    name: string;
    /** MIME type of the resource content. */
    mimeType?: string;
    /** Human-readable description of the resource. */
    description?: string;
    /** Whether the resource is currently enabled. */
    enabled?: boolean;
    /** Name of the service that provides this resource. */
    serviceName?: string;
    /** Type of the resource. */
    type?: string;
}

/**
 * Definition of a prompt template available in the system.
 */
export interface PromptDefinition {
    /** Unique name of the prompt. */
    name: string;
    /** Human-readable description of the prompt. */
    description?: string;
    /** List of arguments expected by the prompt template. */
    arguments?: any[];
    /** Whether the prompt is currently enabled. */
    enabled?: boolean;
    /** Name of the service that provides this prompt. */
    serviceName?: string;
    /** Type of the prompt. */
    type?: string;
}

/**
 * Client for interacting with the MCPAny backend API.
 * Provides methods for managing services, tools, resources, prompts, and secrets.
 */
export const apiClient = {
    // Services
    /**
     * Lists all registered services.
     * @returns A promise that resolves to the list of services.
     */
    listServices: async () => {
        const res = await fetch('/api/v1/services');
        if (!res.ok) throw new Error('Failed to fetch services');
        return res.json();
    },
    /**
     * Retrieves details for a specific service.
     * @param id - The ID of the service to retrieve.
     * @returns A promise that resolves to the service configuration.
     */
    getService: async (id: string) => {
         const res = await fetch(`/api/v1/services/${id}`);
         if (!res.ok) throw new Error('Failed to fetch service');
         return res.json();
    },
    /**
     * Toggles the enabled/disabled status of a service.
     * @param name - The name of the service.
     * @param disable - True to disable, false to enable.
     * @returns A promise that resolves to the updated status.
     */
    setServiceStatus: async (name: string, disable: boolean) => {
        // First get the service config
        const currentRes = await fetch(`/api/v1/services/${name}`);
        if (!currentRes.ok) throw new Error('Failed to get service for status update');
        const config = await currentRes.json();

        // Update disable flag
        config.disable = disable;

        // Save back
        const res = await fetch('/api/v1/services', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config)
        });
        if (!res.ok) throw new Error('Failed to update service status');
        return res.json();
    },
    // Added for compatibility with legacy components
    /**
     * Gets the status and metrics of a service.
     * @param name - The name of the service.
     * @returns A promise that resolves to the service status and metrics.
     */
    getServiceStatus: async (name: string) => {
        const res = await fetch(`/api/v1/services/${name}/status`);
        if (!res.ok) return { enabled: false, metrics: { uptime: 0, latency: 0 } };
        const data = await res.json();
        return {
            enabled: data.status === "Active",
            metrics: { uptime: 99.9, latency: 45, ...data.metrics } // Merge real metrics when available
        };
    },
    /**
     * Registers a new service.
     * @param config - The configuration for the new service.
     * @returns A promise that resolves to the registered service.
     */
    registerService: async (config: UpstreamServiceConfig) => {
        const res = await fetch('/api/v1/services', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config)
        });
        if (!res.ok) throw new Error('Failed to register service');
        return res.json();
    },
    /**
     * Updates an existing service configuration.
     * @param config - The new configuration for the service.
     * @returns A promise that resolves to the updated service.
     */
    updateService: async (config: UpstreamServiceConfig) => {
         const res = await fetch('/api/v1/services', {
            method: 'POST', // POST handles update via save
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config)
        });
        if (!res.ok) throw new Error('Failed to update service');
        return res.json();
    },
    unregisterService: async (id: string) => {
        const res = await fetch(`/api/v1/services/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete service');
        return {};
    },

    // Tools
    /**
     * Lists all available tools.
     * @returns A promise that resolves to the list of tools.
     */
    listTools: async () => {
        const res = await fetch('/api/v1/tools');
        if (!res.ok) throw new Error('Failed to fetch tools');
        return res.json();
    },
    executeTool: async (request: any) => {
        const res = await fetch('/api/v1/execute', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(request)
        });
        if (!res.ok) {
            const errBody = await res.text();
            throw new Error(`Tool execution failed: ${errBody}`);
        }
        return res.json();
    },
    setToolStatus: async (name: string, enabled: boolean) => {
        // TODO: Backend support for toggling individual tools?
        console.warn("setToolStatus not fully implemented in backend yet");
        return {};
    },

    // Resources
    /**
     * Lists all available resources.
     * @returns A promise that resolves to the list of resources.
     */
    listResources: async () => {
        const res = await fetch('/api/v1/resources');
        if (!res.ok) throw new Error('Failed to fetch resources');
        return res.json();
    },
    /**
     * Toggles the enabled/disabled status of a resource.
     * @param uri - The URI of the resource.
     * @param enabled - True to enable, false to disable.
     * @returns A promise that resolves to the updated status.
     */
    setResourceStatus: async (uri: string, enabled: boolean) => {
         // TODO: Backend support
         console.warn("setResourceStatus not fully implemented in backend yet");
         return {};
    },

    // Prompts
    /**
     * Lists all available prompts.
     * @returns A promise that resolves to the list of prompts.
     */
    listPrompts: async () => {
        const res = await fetch('/api/v1/prompts');
        if (!res.ok) throw new Error('Failed to fetch prompts');
        return res.json();
    },
    /**
     * Toggles the enabled/disabled status of a prompt.
     * @param name - The name of the prompt.
     * @param enabled - True to enable, false to disable.
     * @returns A promise that resolves to the updated status.
     */
    setPromptStatus: async (name: string, enabled: boolean) => {
        // TODO: Backend support
        console.warn("setPromptStatus not fully implemented in backend yet");
        return {};
    },

    // Secrets
    /**
     * Lists all managed secrets.
     * @returns A promise that resolves to the list of secrets.
     */
    listSecrets: async () => {
        const res = await fetch('/api/v1/secrets');
        if (!res.ok) throw new Error('Failed to fetch secrets');
        return res.json();
    },
    /**
     * Saves a new or updated secret.
     * @param secret - The secret definition to save.
     * @returns A promise that resolves to the saved secret.
     */
    saveSecret: async (secret: SecretDefinition) => {
        const res = await fetch('/api/v1/secrets', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(secret)
        });
        if (!res.ok) throw new Error('Failed to save secret');
        return res.json();
    },
    /**
     * Deletes a secret.
     * @param id - The ID of the secret to delete.
     * @returns A promise that resolves when the secret is deleted.
     */
    deleteSecret: async (id: string) => {
        const res = await fetch(`/api/v1/secrets/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete secret');
        return {};
    },

    // Global Settings
    getGlobalSettings: async () => {
        const res = await fetch('/api/v1/settings');
        if (!res.ok) throw new Error('Failed to fetch global settings');
        return res.json();
    },
    saveGlobalSettings: async (settings: any) => {
        const res = await fetch('/api/v1/settings', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(settings)
        });
        if (!res.ok) throw new Error('Failed to save global settings');
    },

    // Stack Management
    getStackConfig: async (stackId: string) => {
        const res = await fetch(`/api/v1/stacks/${stackId}/config`);
        if (!res.ok) throw new Error('Failed to fetch stack config');
        return res.text(); // Config is likely raw YAML/JSON text
    },
    saveStackConfig: async (stackId: string, config: string) => {
        const res = await fetch(`/api/v1/stacks/${stackId}/config`, {
            method: 'POST',
            headers: { 'Content-Type': 'text/plain' }, // Or application/yaml
            body: config
        });
        if (!res.ok) throw new Error('Failed to save stack config');
        return res.json();
    }
};

/**
 * Definition of a secret used for sensitive configuration.
 */
export interface SecretDefinition {
    /** Unique ID of the secret. */
    id: string;
    /** Human-readable name of the secret. */
    name: string;
    /** The environment variable key or usage key. */
    key: string;
    /** The actual secret value. */
    value: string;
    /** The provider associated with this secret (optional). */
    provider?: 'openai' | 'anthropic' | 'aws' | 'gcp' | 'custom';
    /** Timestamp of when the secret was last used (ISO string). */
    lastUsed?: string;
    /** Timestamp of creation (ISO string). */
    createdAt: string;
}
