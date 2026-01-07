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

import { UpstreamServiceConfig } from '../proto/config/v1/upstream_service';
import { ToolDefinition } from '../proto/config/v1/tool';
import { ResourceDefinition } from '../proto/config/v1/resource';
import { PromptDefinition } from '../proto/config/v1/prompt';

// Re-export generated types
export type { UpstreamServiceConfig, ToolDefinition, ResourceDefinition, PromptDefinition };

export interface SecretDefinition {
    id: string;
    name: string;
    key: string;
    value: string;
    provider?: 'openai' | 'anthropic' | 'aws' | 'gcp' | 'custom';
    lastUsed?: string;
    createdAt: string;
}

// Initialize REST Client Helpers
const getBaseUrl = () => {
    if (typeof window !== 'undefined') {
        return window.location.origin;
    }
    return 'http://localhost:8080'; // Default for SSR
};

// Helper to identify API Key
const getApiKey = () => {
    if (typeof process !== 'undefined' && process.env) {
        return process.env.NEXT_PUBLIC_MCPANY_API_KEY;
    }
    return undefined;
};

// Helper to handle REST requests with Protobuf Types or JSON
async function restRequest<Req, Res>(
    path: string,
    method: 'GET' | 'POST' | 'DELETE' | 'PUT',
    reqBody: Req | undefined,
    resType: { fromJSON: (o: any) => Res }
): Promise<Res> {
    const url = `${getBaseUrl()}${path}`;
    const headers: HeadersInit = {
        'Content-Type': 'application/json',
    };

    // Inject API Key if present
    const key = getApiKey();
    if (key) {
        (headers as any)['X-API-Key'] = key;
    }

    const options: RequestInit = {
        method,
        headers,
    };

    if (reqBody) {
        options.body = JSON.stringify(reqBody);
    }

    const response = await fetch(url, options);
    if (!response.ok) {
        const txt = await response.text().catch(() => '');
        throw new Error(`API request failed: ${response.status} ${response.statusText} ${txt}`);
    }

    // Handle empty response for DELETE or void returns
    const text = await response.text();
    if (!text) {
        // If we expect a response type, this might be an error if it's not optional.
        // But for {} return types, fromJSON({}) works.
        return resType.fromJSON({});
    }

    try {
        const json = JSON.parse(text);
        return resType.fromJSON(json);
    } catch (e) {
        // Fallback for plain text response if expected
        // Check if resType handles string? Unlikely for fromJSON.
        // Assuming JSON always for this helper unless specific case.
        throw new Error(`Invalid JSON response: ${text.substring(0, 100)}...`);
    }
}

// Pass-through for plain JSON objects
const identity = { fromJSON: (json: any) => json };

// REST Client Implementation
export const apiClient = {
    // Services
    listServices: async () => {
        const response = await restRequest(
            '/api/v1/services',
            'GET',
            undefined,
            identity
        );
        if (Array.isArray(response)) {
            return response as UpstreamServiceConfig[];
        }
        return (response as any).services as UpstreamServiceConfig[];
    },
    getService: async (id: string) => {
        const response = await restRequest(
            `/api/v1/services/${id}`,
            'GET',
            undefined,
            identity
        );
        return (response as any).service as UpstreamServiceConfig;
    },
    setServiceStatus: async (name: string, disable: boolean) => {
         const service = await apiClient.getService(name);
         if (!service) throw new Error('Service not found');

         service.disable = disable;

         await restRequest(
             `/api/v1/services/${name}`,
             'PUT',
             service,
             identity
         );
         return service;
    },
    getServiceStatus: async (name: string) => {
        return restRequest(
            `/api/v1/services/${name}/status`,
            'GET',
            undefined,
            identity
        );
    },
    registerService: async (config: UpstreamServiceConfig) => {
        return restRequest(
            '/api/v1/services',
            'POST',
            config,
            identity
        );
    },
    updateService: async (config: UpstreamServiceConfig) => {
        return restRequest(
            `/api/v1/services/${config.name}`,
            'PUT',
            config,
            identity
        );
    },
    unregisterService: async (id: string) => {
         await restRequest(
             `/api/v1/services/${id}`,
             'DELETE',
             undefined,
             identity
         );
         return {};
    },

    // Tools
    listTools: async () => {
        return restRequest('/api/v1/tools', 'GET', undefined, identity);
    },
    executeTool: async (request: any) => {
        return restRequest('/api/v1/execute', 'POST', request, identity);
    },
    setToolStatus: async (name: string, enabled: boolean) => {
        return restRequest('/api/v1/tools', 'POST', { name, enabled }, identity);
    },

    // Resources
    listResources: async () => {
        return restRequest('/api/v1/resources', 'GET', undefined, identity);
    },
    setResourceStatus: async (uri: string, enabled: boolean) => {
        return restRequest('/api/v1/resources', 'POST', { uri, enabled }, identity);
    },

    // Prompts
    listPrompts: async () => {
        return restRequest('/api/v1/prompts', 'GET', undefined, identity);
    },
    setPromptStatus: async (name: string, enabled: boolean) => {
        return restRequest('/api/v1/prompts', 'POST', { name, enabled }, identity);
    },
    executePrompt: async (name: string, args: Record<string, string>) => {
        // Attempt to call backend
        try {
            const res = await fetch(`/api/v1/prompts/${name}/execute`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(args)
            });
            if (res.ok) return res.json();
        } catch (e) {
            console.warn("Backend execution failed, falling back to simulation for UI demo", e);
        }

        // Mock simulation if backend fails (for demo purposes)
        return new Promise((resolve) => {
            setTimeout(() => {
                resolve({
                    messages: [
                        {
                            role: "user",
                            content: {
                                type: "text",
                                text: `Execute prompt '${name}' with args: ${JSON.stringify(args)}`
                            }
                        },
                        {
                            role: "assistant",
                            content: {
                                type: "text",
                                text: `This is a simulated response for the prompt template '${name}'.\n\nArguments used:\n${Object.entries(args).map(([k, v]) => `- ${k}: ${v}`).join('\n')}\n\nThe backend endpoint /api/v1/prompts/${name}/execute is not yet implemented, so this mock response is provided for UI verification.`
                            }
                        }
                    ]
                });
            }, 800);
        });
    },

    // Secrets
    listSecrets: async () => {
        return restRequest('/api/v1/secrets', 'GET', undefined, identity);
    },
    saveSecret: async (secret: SecretDefinition) => {
        return restRequest('/api/v1/secrets', 'POST', secret, identity);
    },
    deleteSecret: async (id: string) => {
        return restRequest(`/api/v1/secrets/${id}`, 'DELETE', undefined, identity);
    },

    // Global Settings
    getGlobalSettings: async () => {
        return restRequest('/api/v1/settings', 'GET', undefined, identity);
    },
    saveGlobalSettings: async (settings: any) => {
        return restRequest('/api/v1/settings', 'POST', settings, identity);
    },

    // Stack Management
    getStackConfig: async (stackId: string) => {
        // This endpoint returns text/plain usually?
        // restRequest expects JSON response.
        // We might need a specialized fetch for text.
        const url = `${getBaseUrl()}/api/v1/stacks/${stackId}/config`;
        const headers: HeadersInit = {};
        const key = getApiKey();
        if (key) {
            (headers as any)['X-API-Key'] = key;
        }
        const res = await fetch(url, { headers });
        if (!res.ok) throw new Error('Failed to fetch stack config');
        return res.text();
    },
    saveStackConfig: async (stackId: string, config: string) => {
        const url = `${getBaseUrl()}/api/v1/stacks/${stackId}/config`;
        const headers: HeadersInit = { 'Content-Type': 'text/plain' };
        const key = getApiKey();
        if (key) {
            (headers as any)['X-API-Key'] = key;
        }
        const res = await fetch(url, {
            method: 'POST',
            headers,
            body: config
        });
        if (!res.ok) throw new Error('Failed to save stack config');
        return res.json();
    }
};
