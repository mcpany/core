/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

export interface UpstreamServiceConfig {
    id?: string;
    name: string;
    version?: string;
    disable?: boolean;
    priority?: number;
    http_service?: { address: string; tls_config?: any; tools?: any[]; prompts?: any[] };
    grpc_service?: { address: string; use_reflection?: boolean; tls_config?: any; tools?: any[]; resources?: any[] };
    command_line_service?: { command: string; args?: string[]; env?: Record<string, string>; tools?: any[] };
    mcp_service?: {
        http_connection?: { http_address: string; tls_config?: any };
        sse_connection?: { sse_address: string };
        stdio_connection?: { command: string };
        bundle_connection?: { bundle_path: string };
        tools?: any[];
    };
    openapi_service?: { address: string; spec_url?: string; spec_content?: string; tools?: any[]; tls_config?: any };
    websocket_service?: { address: string; tls_config?: any };
    webrtc_service?: { address: string; tls_config?: any };
    graphql_service?: { address: string };
}

export interface ToolDefinition {
    name: string;
    description: string;
    schema?: any;
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
    arguments?: any[];
    enabled?: boolean;
    serviceName?: string;
    type?: string;
}

export const apiClient = {
    // Services
    listServices: async () => {
        const res = await fetch('/api/services');
        if (!res.ok) throw new Error('Failed to fetch services');
        return res.json();
    },
    getService: async (id: string) => {
         const res = await fetch(`/api/services?id=${id}`); // Mock
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
    // Added for compatibility with legacy components
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
        if (!res.ok) throw new Error('Failed to update resource status');
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
        if (!res.ok) throw new Error('Failed to update prompt status');
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
        return res.json();
    }
};

export interface SecretDefinition {
    id: string;
    name: string;
    key: string; // The environment variable key or usage key
    value: string; // The actual secret
    provider?: 'openai' | 'anthropic' | 'aws' | 'gcp' | 'custom';
    lastUsed?: string;
    createdAt: string;
}
