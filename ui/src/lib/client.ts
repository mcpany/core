
/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Client for interacting with the MCPAny backend API.
 * Provides methods for managing services, tools, resources, prompts, and secrets.
 */

// NOTE: Adjusted to point to local Next.js API routes for this UI overhaul task
// In a real deployment, these might be /api/v1/... proxied to backend

export interface UpstreamServiceConfig {
    id?: string;
    name: string;
    version?: string;
    disable?: boolean;
    priority?: number;
    http_service?: { address: string; tls_config?: unknown; tools?: unknown[]; prompts?: unknown[] };
    grpc_service?: { address: string; use_reflection?: boolean; tls_config?: unknown; tools?: unknown[]; resources?: unknown[] };
    command_line_service?: { command: string; args?: string[]; env?: Record<string, string>; tools?: unknown[] };
    mcp_service?: {
        http_connection?: { http_address: string; tls_config?: unknown };
        sse_connection?: { sse_address: string };
        stdio_connection?: { command: string };
        bundle_connection?: { bundle_path: string };
        tools?: unknown[];
    };
    openapi_service?: { address: string; spec_url?: string; spec_content?: string; tools?: unknown[]; tls_config?: unknown };
    websocket_service?: { address: string; tls_config?: unknown };
    webrtc_service?: { address: string; tls_config?: unknown };
    graphql_service?: { address: string };
}

export interface ToolDefinition {
    name: string;
    description: string;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    schema?: any; // Keeping any for schema structure as it's dynamic JSON
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    input_schema?: any; // keeping both for compatibility
    enabled?: boolean;
    serviceName?: string;
    source?: string;
}

export interface ResourceDefinition {
    uri: string;
    name: string;
    mimeType?: string;
    description?: string;
    enabled?: boolean;
    serviceName?: string;
    type?: string;
}

export interface PromptDefinition {
    name: string;
    description?: string;
    arguments?: unknown[];
    enabled?: boolean;
    serviceName?: string;
    type?: string;
}

export interface SecretDefinition {
    id: string;
    name: string;
    key: string;
    value: string;
    provider?: 'openai' | 'anthropic' | 'aws' | 'gcp' | 'custom';
    lastUsed?: string;
    createdAt: string;
}

export const apiClient = {
    // Services
    listServices: async () => {
        const res = await fetch('/api/services');
        if (!res.ok) throw new Error('Failed to fetch services');
        return res.json();
    },
    getService: async (id: string) => {
         const res = await fetch(`/api/services/${id}`);
         if (!res.ok) throw new Error('Failed to fetch service');
         return res.json();
    },
    setServiceStatus: async (name: string, disable: boolean) => {
        const res = await fetch('/api/services', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ action: 'toggle', name, disable })
        });
        if (!res.ok) throw new Error('Failed to update service status');
        return res.json();
    },
    getServiceStatus: async (_name: string) => {
        // Mock implementation for now
        return {
            enabled: true,
            metrics: { uptime: 99.9, latency: 45 }
        };
    },
    registerService: async (config: UpstreamServiceConfig) => {
        const res = await fetch('/api/services', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config)
        });
        if (!res.ok) throw new Error('Failed to register service');
        return res.json();
    },
    updateService: async (config: UpstreamServiceConfig) => {
         const res = await fetch('/api/services', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config)
        });
        if (!res.ok) throw new Error('Failed to update service');
        return res.json();
    },
    unregisterService: async (_id: string) => {
         // Mock
        return {};
    },

    // Tools
    listTools: async () => {
        const res = await fetch('/api/tools');
        if (!res.ok) throw new Error('Failed to fetch tools');
        return res.json();
    },
    executeTool: async (_request: unknown) => {
        // Mock execution
        return { output: "Mock execution result", success: true };
    },
    setToolStatus: async (name: string, enabled: boolean) => {
        const res = await fetch('/api/tools', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, enabled })
        });
        return res.json();
    },

    // Resources
    listResources: async () => {
        const res = await fetch('/api/resources');
        if (!res.ok) throw new Error('Failed to fetch resources');
        return res.json();
    },
    setResourceStatus: async (uri: string, enabled: boolean) => {
         const res = await fetch('/api/resources', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ uri, enabled })
        });
        return res.json();
    },

    // Prompts
    listPrompts: async () => {
        const res = await fetch('/api/prompts');
        if (!res.ok) throw new Error('Failed to fetch prompts');
        return res.json();
    },
    setPromptStatus: async (name: string, enabled: boolean) => {
        const res = await fetch('/api/prompts', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, enabled })
        });
        return res.json();
    },

    // Secrets
    listSecrets: async () => {
        const res = await fetch('/api/secrets');
        if (!res.ok) throw new Error('Failed to fetch secrets');
        return res.json();
    },
    saveSecret: async (secret: SecretDefinition) => {
        const res = await fetch('/api/secrets', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(secret)
        });
        if (!res.ok) throw new Error('Failed to save secret');
        return res.json();
    },
    deleteSecret: async (id: string) => {
        const res = await fetch(`/api/secrets/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete secret');
        return {};
    },

    // Global Settings
    getGlobalSettings: async () => {
        const res = await fetch('/api/settings');
        if (!res.ok) throw new Error('Failed to fetch global settings');
        return res.json();
    },
    saveGlobalSettings: async (settings: unknown) => {
        const res = await fetch('/api/settings', {
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
