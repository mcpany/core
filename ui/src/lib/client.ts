/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

export interface UpstreamServiceConfig {
    id?: string;
    name: string;
    version?: string;
    disable?: boolean;
    http_service?: { address: string };
    grpc_service?: { address: string };
    command_line_service?: { command: string; args?: string[]; env?: Record<string, string> };
    mcp_service?: { http_connection?: { http_address: string }; sse_connection?: { sse_address: string }; stdio_connection?: { command: string } };
}

export interface ToolDefinition {
    name: string;
    description: string;
    schema?: any;
    enabled?: boolean;
    serviceName?: string;
}

export interface ResourceDefinition {
    uri: string;
    name: string;
    mimeType?: string;
    description?: string;
    enabled?: boolean;
    serviceName?: string;
}

export interface PromptDefinition {
    name: string;
    description?: string;
    arguments?: any[];
    enabled?: boolean;
    serviceName?: string;
}

export const apiClient = {
    // Services
    listServices: async () => {
        const res = await fetch('/api/services');
        if (!res.ok) throw new Error('Failed to fetch services');
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

    // Secrets (Mock)
    listSecrets: async () => {
        // Mock delay
        await new Promise(resolve => setTimeout(resolve, 500));
        const stored = localStorage.getItem('mcp-secrets');
        return stored ? JSON.parse(stored) : [];
    },
    saveSecret: async (secret: SecretDefinition) => {
        await new Promise(resolve => setTimeout(resolve, 500));
        const stored = localStorage.getItem('mcp-secrets');
        const secrets: SecretDefinition[] = stored ? JSON.parse(stored) : [];
        const index = secrets.findIndex(s => s.id === secret.id);
        if (index >= 0) {
            secrets[index] = secret;
        } else {
            secrets.push(secret);
        }
        localStorage.setItem('mcp-secrets', JSON.stringify(secrets));
        return secret;
    },
    deleteSecret: async (id: string) => {
        await new Promise(resolve => setTimeout(resolve, 500));
        const stored = localStorage.getItem('mcp-secrets');
        if (!stored) return;
        const secrets: SecretDefinition[] = JSON.parse(stored);
        const filtered = secrets.filter(s => s.id !== id);
        localStorage.setItem('mcp-secrets', JSON.stringify(filtered));
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
