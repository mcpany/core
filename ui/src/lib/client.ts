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
    /** Human-readable name of the service. */
    name: string;
    /** Version string of the service. */
    version?: string;
    /** Whether the service is disabled. */
    disable?: boolean;
    /** Priority of the service (higher value means higher priority). */
    priority?: number;
    /** Configuration for an HTTP upstream service. */
    http_service?: { address: string; tls_config?: any; tools?: any[]; prompts?: any[] };
    /** Configuration for a gRPC upstream service. */
    grpc_service?: { address: string; use_reflection?: boolean; tls_config?: any; tools?: any[]; resources?: any[] };
    /** Configuration for a command-line upstream service. */
    command_line_service?: { command: string; args?: string[]; env?: Record<string, string>; tools?: any[] };
    /** Configuration for an MCP upstream service. */
    mcp_service?: {
        http_connection?: { http_address: string; tls_config?: any };
        sse_connection?: { sse_address: string };
        stdio_connection?: { command: string };
        bundle_connection?: { bundle_path: string };
        tools?: any[];
    };
    /** Configuration for an OpenAPI upstream service. */
    openapi_service?: { address: string; spec_url?: string; spec_content?: string; tools?: any[]; tls_config?: any };
    /** Configuration for a WebSocket upstream service. */
    websocket_service?: { address: string; tls_config?: any };
    /** Configuration for a WebRTC upstream service. */
    webrtc_service?: { address: string; tls_config?: any };
    /** Configuration for a GraphQL upstream service. */
    graphql_service?: { address: string };
}

/**
 * Definition of a tool exposed by a service.
 */
export interface ToolDefinition {
    /** Unique name of the tool. */
    name: string;
    /** Description of what the tool does. */
    description: string;
    /** JSON schema defining the tool's input arguments. */
    schema?: any;
    /** Whether the tool is currently enabled. */
    enabled?: boolean;
    /** Name of the service that provides this tool. */
    serviceName?: string;
    /** Source of the tool definition (e.g., "config", "discovery"). */
    source?: string;
}

/**
 * Definition of a resource exposed by a service.
 */
export interface ResourceDefinition {
    /** URI identifying the resource. */
    uri: string;
    /** Human-readable name of the resource. */
    name: string;
    /** MIME type of the resource content. */
    mimeType?: string;
    /** Description of the resource. */
    description?: string;
    /** Whether the resource is currently enabled. */
    enabled?: boolean;
    /** Name of the service that provides this resource. */
    serviceName?: string;
    /** Type of the resource. */
    type?: string;
}

/**
 * Definition of a prompt exposed by a service.
 */
export interface PromptDefinition {
    /** Unique name of the prompt. */
    name: string;
    /** Description of what the prompt does. */
    description?: string;
    /** List of arguments accepted by the prompt. */
    arguments?: any[];
    /** Whether the prompt is currently enabled. */
    enabled?: boolean;
    /** Name of the service that provides this prompt. */
    serviceName?: string;
    /** Type of the prompt. */
    type?: string;
}

/**
 * Client for interacting with the backend API.
 */
export const apiClient = {
    // Services
    /**
     * Lists all registered services.
     * @returns A promise resolving to a list of services.
     */
    listServices: async () => {
        const res = await fetch('/api/services');
        if (!res.ok) throw new Error('Failed to fetch services');
        return res.json();
    },
    /**
     * Retrieves details for a specific service.
     * @param id - The ID of the service to retrieve.
     * @returns A promise resolving to the service details.
     */
    getService: async (id: string) => {
         const res = await fetch(`/api/services?id=${id}`); // Mock
         if (!res.ok) throw new Error('Failed to fetch service');
         return res.json();
    },
    /**
     * Updates the status (enabled/disabled) of a service.
     * @param name - The name of the service.
     * @param disable - True to disable, false to enable.
     * @returns A promise resolving to the updated status.
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
     * Gets the current status and metrics of a service.
     * @param name - The name of the service.
     * @returns A promise resolving to the status and metrics.
     */
    getServiceStatus: async (name: string) => {
        const res = await fetch(`/api/services?name=${name}`);
        if (!res.ok) return { enabled: false, metrics: { uptime: 0, latency: 0 } };
        const data = await res.json();
        // Assuming the API returns the service object or list
        // This is a best-effort mock
        return {
            enabled: !data.disable,
            metrics: { uptime: 99.9, latency: 45 } // Mock metrics as numbers
        };
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
     * @param config - The updated configuration.
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
    /**
     * Unregisters (deletes) a service.
     * @param _id - The ID of the service to unregister.
     * @returns A promise resolving to an empty object on success.
     */
    unregisterService: async (_id: string) => {
        // Mock
        return {};
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
     * Updates the status (enabled/disabled) of a tool.
     * @param name - The name of the tool.
     * @param enabled - True to enable, false to disable.
     * @returns A promise resolving to the updated status.
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
     * Updates the status (enabled/disabled) of a resource.
     * @param uri - The URI of the resource.
     * @param enabled - True to enable, false to disable.
     * @returns A promise resolving to the updated status.
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
     * Updates the status (enabled/disabled) of a prompt.
     * @param name - The name of the prompt.
     * @param enabled - True to enable, false to disable.
     * @returns A promise resolving to the updated status.
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

    // Secrets
    /**
     * Lists all stored secrets.
     * @returns A promise resolving to a list of secrets.
     */
    listSecrets: async () => {
        const res = await fetch('/api/secrets');
        if (!res.ok) throw new Error('Failed to fetch secrets');
        return res.json();
    },
    /**
     * Saves a new secret or updates an existing one.
     * @param secret - The secret definition to save.
     * @returns A promise resolving to the saved secret.
     */
    saveSecret: async (secret: SecretDefinition) => {
        const res = await fetch('/api/secrets', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(secret)
        });
        if (!res.ok) throw new Error('Failed to save secret');
        return res.json();
    },
    /**
     * Deletes a secret by its ID.
     * @param id - The ID of the secret to delete.
     * @returns A promise resolving to the deletion confirmation.
     */
    deleteSecret: async (id: string) => {
        const res = await fetch(`/api/secrets/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete secret');
        return res.json();
    }
};

/**
 * Definition of a secret.
 */
export interface SecretDefinition {
    /** Unique identifier for the secret. */
    id: string;
    /** Human-readable name of the secret. */
    name: string;
    /** The environment variable key or usage key. */
    key: string;
    /** The actual secret value. */
    value: string;
    /** The provider associated with the secret (optional). */
    provider?: 'openai' | 'anthropic' | 'aws' | 'gcp' | 'custom';
    /** Timestamp when the secret was last used (optional). */
    lastUsed?: string;
    /** Timestamp when the secret was created. */
    createdAt: string;
}
