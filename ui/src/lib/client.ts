/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Client for interacting with the MCPAny backend API.
 * Provides methods for managing services, tools, resources, prompts, and secrets.
 */

import { GrpcWebImpl, RegistrationServiceClientImpl } from '../proto/api/v1/registration';
import { UpstreamServiceConfig } from '../proto/config/v1/upstream_service';
import { ToolDefinition } from '../proto/config/v1/tool';
import { ResourceDefinition } from '../proto/config/v1/resource';
import { PromptDefinition } from '../proto/config/v1/prompt';
import { BrowserHeaders } from 'browser-headers';

// Re-export generated types
export type { UpstreamServiceConfig, ToolDefinition, ResourceDefinition, PromptDefinition };

const getBaseUrl = () => {
    if (typeof window !== 'undefined') {
        return window.location.origin;
    }
    return 'http://localhost:8080';
};

const rpc = new GrpcWebImpl(getBaseUrl(), {
  debug: false,
});
const registrationClient = new RegistrationServiceClientImpl(rpc);

const fetchWithAuth = async (input: RequestInfo | URL, init?: RequestInit) => {
    const headers = new Headers(init?.headers);
    const key = process.env.NEXT_PUBLIC_MCPANY_API_KEY;
    if (key) {
        headers.set('X-API-Key', key);
    }
    return fetch(input, { ...init, headers });
};

export interface SecretDefinition {
    id: string;
    name: string;
    key: string;
    value: string;
    provider?: 'openai' | 'anthropic' | 'aws' | 'gcp' | 'custom';
    lastUsed?: string;
    createdAt: string;
}

const getMetadata = () => {
    const key = process.env.NEXT_PUBLIC_MCPANY_API_KEY;
    return key ? new BrowserHeaders({ 'X-API-Key': key }) : undefined;
};

export const apiClient = {
    // Services
    listServices: async () => {
        // Fallback to REST until gRPC-Web setup is fully verified end-to-end
        const res = await fetchWithAuth('/api/v1/services');
        if (!res.ok) throw new Error('Failed to fetch services');
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.services || []);
        return list.map((s: any) => ({
            ...s,
            // normalize snake_case from backend to camelCase if needed,
            // though generated types usually handle this if using gRPC.
            // For REST fallback, we manually map if necessary.
            // Assuming the UI expects the generated type structure.
            connectionPool: s.connection_pool,
            httpService: s.http_service,
            grpcService: s.grpc_service,
            commandLineService: s.command_line_service,
            mcpService: s.mcp_service
        }));
    },
    getService: async (id: string) => {
         const response = await registrationClient.GetService({ serviceName: id }, getMetadata());
         return response.service;
    },
    setServiceStatus: async (name: string, disable: boolean) => {
        const response = await fetchWithAuth(`/api/v1/services/${name}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ disable })
        });
        if (!response.ok) throw new Error('Failed to update service status');
        return response.json();
    },
    registerService: async (config: UpstreamServiceConfig) => {
        // Map camelCase (UI) to snake_case (Server REST expectation)
        const payload: any = {
            id: config.id,
            name: config.name,
            version: config.version,
            disable: config.disable,
            priority: config.priority,
            load_balancing_strategy: config.loadBalancingStrategy,
        };

        if (config.httpService) {
            payload.http_service = { address: config.httpService.address };
        }
        if (config.grpcService) {
            payload.grpc_service = { address: config.grpcService.address };
        }
        if (config.commandLineService) {
            payload.command_line_service = {
                command: config.commandLineService.command,
                working_directory: config.commandLineService.workingDirectory,
                env: config.commandLineService.env
            };
        }
        if (config.mcpService) {
            payload.mcp_service = { ...config.mcpService };
        }

        const response = await fetchWithAuth('/api/v1/services', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
             const txt = await response.text();
             throw new Error(`Failed to register service: ${response.status} ${txt}`);
        }
        return response.json();
    },
    updateService: async (config: UpstreamServiceConfig) => {
        const payload: any = {
            id: config.id,
            name: config.name,
            version: config.version,
            disable: config.disable,
            priority: config.priority,
            load_balancing_strategy: config.loadBalancingStrategy,
        };
        if (config.httpService) {
            payload.http_service = { address: config.httpService.address };
        }
        if (config.grpcService) {
            payload.grpc_service = { address: config.grpcService.address };
        }
        if (config.commandLineService) {
            payload.command_line_service = {
                command: config.commandLineService.command,
                working_directory: config.commandLineService.workingDirectory,
                env: config.commandLineService.env
            };
        }
        if (config.mcpService) {
            payload.mcp_service = { ...config.mcpService };
        }

        const response = await fetchWithAuth(`/api/v1/services/${config.name}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

         if (!response.ok) {
             const txt = await response.text();
             throw new Error(`Failed to update service: ${response.status} ${txt}`);
        }
        return response.json();
    },
    unregisterService: async (id: string) => {
         const response = await fetchWithAuth(`/api/v1/services/${id}`, {
            method: 'DELETE'
         });
         if (!response.ok) throw new Error('Failed to unregister service');
         return {};
    },

    // Tools
    listTools: async () => {
        const res = await fetchWithAuth('/api/v1/tools');
        if (!res.ok) throw new Error('Failed to fetch tools');
        return res.json();
    },
    executeTool: async (request: any) => {
        const res = await fetchWithAuth('/api/v1/execute', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(request)
        });
        if (!res.ok) throw new Error('Failed to execute tool');
        return res.json();
    },
    setToolStatus: async (name: string, enabled: boolean) => {
        const res = await fetchWithAuth('/api/v1/tools', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, enabled })
        });
        return res.json();
    },

    // Resources
    listResources: async () => {
        const res = await fetchWithAuth('/api/v1/resources');
        if (!res.ok) throw new Error('Failed to fetch resources');
        return res.json();
    },
    setResourceStatus: async (uri: string, enabled: boolean) => {
         const res = await fetchWithAuth('/api/v1/resources', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ uri, enabled })
        });
        return res.json();
    },

    // Prompts
    listPrompts: async () => {
        const res = await fetchWithAuth('/api/v1/prompts');
        if (!res.ok) throw new Error('Failed to fetch prompts');
        return res.json();
    },
    setPromptStatus: async (name: string, enabled: boolean) => {
        const res = await fetchWithAuth('/api/v1/prompts', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, enabled })
        });
        return res.json();
    },
    executePrompt: async (name: string, args: Record<string, string>) => {
        try {
            const res = await fetch(`/api/v1/prompts/${name}/execute`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(args)
            });
            if (res.ok) return res.json();
        } catch (e) {
            console.warn("Backend execution failed, falling back to simulation", e);
        }

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
                                text: `Mock response for prompt '${name}'.`
                            }
                        }
                    ]
                });
            }, 800);
        });
    },

    // Secrets
    listSecrets: async () => {
        const res = await fetchWithAuth('/api/v1/secrets');
        if (!res.ok) throw new Error('Failed to fetch secrets');
        return res.json();
    },
    saveSecret: async (secret: SecretDefinition) => {
        const res = await fetchWithAuth('/api/v1/secrets', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(secret)
        });
        if (!res.ok) throw new Error('Failed to save secret');
        return res.json();
    },
    deleteSecret: async (id: string) => {
        const res = await fetchWithAuth(`/api/v1/secrets/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete secret');
        return {};
    },

    // Global Settings
    getGlobalSettings: async () => {
        const res = await fetchWithAuth('/api/v1/settings');
        if (!res.ok) throw new Error('Failed to fetch global settings');
        return res.json();
    },
    saveGlobalSettings: async (settings: any) => {
        const res = await fetchWithAuth('/api/v1/settings', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(settings)
        });
        if (!res.ok) throw new Error('Failed to save global settings');
    },

    // Profiles
    listProfiles: async () => {
         // Mock profiles for now or fetch from backend if available
         // The config.proto has ProfileDefinition
         const res = await fetchWithAuth('/api/v1/profiles');
         if (!res.ok) {
             // Return mock data for development if backend missing
             return [
                 { id: 'dev', name: 'Development', active: true, roles: ['admin'] },
                 { id: 'prod', name: 'Production', active: false, roles: ['viewer'] },
                 { id: 'debug', name: 'Debug Mode', active: false, roles: ['tester'] },
             ];
         }
         return res.json();
    },
    toggleProfile: async (id: string, enabled: boolean) => {
         // Implement toggle logic
         return { success: true };
    },

    // Middleware
    listMiddleware: async () => {
        const res = await fetchWithAuth('/api/v1/middleware');
        if (!res.ok) {
            // Mock
             return [
                 { id: 'auth', name: 'Authentication', priority: 1, enabled: true },
                 { id: 'logging', name: 'Logging', priority: 2, enabled: true },
                 { id: 'ratelimit', name: 'Rate Limiting', priority: 3, enabled: false },
             ];
        }
        return res.json();
    },
    saveMiddleware: async (middleware: any[]) => {
        // save
    },

    // Webhooks
    listWebhooks: async () => {
        const res = await fetchWithAuth('/api/v1/webhooks');
        if (!res.ok) {
             // Mock
             return [
                 { id: 'wh_1', url: 'https://webhook.site/xyz', events: ['service.down'], enabled: true },
             ];
        }
        return res.json();
    },
    saveWebhook: async (webhook: any) => {
        // save
    },
    testWebhook: async (id: string) => {
        // trigger test
        return { success: true };
    }
};
