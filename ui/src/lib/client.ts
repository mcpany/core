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
        const res = await fetch('/api/v1/services');
        if (!res.ok) throw new Error('Failed to fetch services');
        return res.json();
    },
    getService: async (id: string) => {
         const res = await fetch(`/api/v1/services/${id}`);
         if (!res.ok) throw new Error('Failed to fetch service');
         return res.json();
    },
    setServiceStatus: async (name: string, disable: boolean) => {
        // First get the service config to get the full object (including ID/Address/etc)
        const getRes = await fetch(`/api/v1/services/${name}`);
        if (!getRes.ok) throw new Error('Failed to fetch service for status update');
        const service = await getRes.json();

        // Update disable status
        service.disable = disable;

        // Save back
        const res = await fetch(`/api/v1/services/${name}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(service)
        });
        if (!res.ok) throw new Error('Failed to update service status');
        return res.json();
    },
    getServiceStatus: async (name: string) => {
        const res = await fetch(`/api/v1/services/${name}/status`);
        if (!res.ok) return { enabled: false, metrics: {} };
        return res.json();
    },
    registerService: async (config: UpstreamServiceConfig) => {
        const res = await fetch('/api/v1/services', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config)
        });
        if (!res.ok) throw new Error('Failed to register service');
        return res.json();
    },
    updateService: async (config: UpstreamServiceConfig) => {
         const res = await fetch('/api/v1/services', {
            method: 'PUT', // handleServiceDetail uses PUT for updates? No, handleServices POST handles create/update?
            // handleServices POST calls SaveService.
            // handleServiceDetail PUT calls SaveService with name forced.
            // Let's use handleServiceDetail PUT if updating specific service.
            // But config has ID/Name.
            // If we use POST /api/v1/services, it saves it.
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
         if (!res.ok) throw new Error('Failed to unregister service');
         return {};
    },

    // Tools
    listTools: async () => {
        const res = await fetch('/api/v1/tools');
        if (!res.ok) throw new Error('Failed to fetch tools');
        return res.json();
    },
    executeTool: async (request: any) => {
        console.error("DEBUG: Calling executeTool with:", JSON.stringify(request));
        try {
            const res = await fetch('/api/v1/execute', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(request)
            });
            console.error("DEBUG: fetch returned status:", res.status);
            if (!res.ok) throw new Error('Failed to execute tool');
            return res.json();
        } catch (e) {
            console.error("DEBUG: fetch failed:", e);
            throw e;
        }
    },
    setToolStatus: async (name: string, enabled: boolean) => {
        // Not implemented in backend yet? handleTools only GET
        // So keeping as mock or throwing?
        // Let's keep as fetch to /api/v1/tools to fail properly or if I add it later.
        const res = await fetch('/api/v1/tools', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, enabled })
        });
        return res.json();
    },

    // Resources
    listResources: async () => {
        const res = await fetch('/api/v1/resources');
        if (!res.ok) throw new Error('Failed to fetch resources');
        return res.json();
    },
    setResourceStatus: async (uri: string, enabled: boolean) => {
         const res = await fetch('/api/v1/resources', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ uri, enabled })
        });
        return res.json();
    },

    // Prompts
    listPrompts: async () => {
        const res = await fetch('/api/v1/prompts');
        if (!res.ok) throw new Error('Failed to fetch prompts');
        return res.json();
    },
    setPromptStatus: async (name: string, enabled: boolean) => {
        const res = await fetch('/api/v1/prompts', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, enabled })
        });
        return res.json();
    },

    // Secrets
    listSecrets: async () => {
        const res = await fetch('/api/v1/secrets');
        if (!res.ok) throw new Error('Failed to fetch secrets');
        return res.json();
    },
    saveSecret: async (secret: SecretDefinition) => {
        const res = await fetch('/api/v1/secrets', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(secret)
        });
        if (!res.ok) throw new Error('Failed to save secret');
        return res.json();
    },
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
