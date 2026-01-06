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

import { GrpcWebImpl, RegistrationServiceClientImpl } from '../proto/api/v1/registration';
import { UpstreamServiceConfig } from '../proto/config/v1/upstream_service';
import { ToolDefinition } from '../proto/config/v1/tool';
import { ResourceDefinition } from '../proto/config/v1/resource';
import { PromptDefinition } from '../proto/config/v1/prompt';

// Re-export generated types
export type { UpstreamServiceConfig, ToolDefinition, ResourceDefinition, PromptDefinition };

// Initialize gRPC Web Client
// Note: In development, we use localhost:8081 (envoy) or the Go server port if configured for gRPC-Web?
// server.go wraps gRPC-Web on the SAME port as HTTP (8080 usually).
// So we can point to window.location.origin or relative?
// GrpcWebImpl needs a full URL host usually.
// If running in browser, we can use empty string or relative?
// GrpcWebImpl implementation uses `this.host`. If empty?
// Let's assume we point to the current origin.
const getBaseUrl = () => {
    if (typeof window !== 'undefined') {
        return window.location.origin;
    }
    return 'http://localhost:8080'; // Default for SSR
};

const rpc = new GrpcWebImpl(getBaseUrl(), {
  debug: false,
});
const registrationClient = new RegistrationServiceClientImpl(rpc);

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
    // Services (Migrated to gRPC)
    listServices: async () => {
        const response = await registrationClient.ListServices({});
        return response.services;
    },
    getService: async (id: string) => {
         const response = await registrationClient.GetService({ serviceName: id });
         return response.service;
    },
    setServiceStatus: async (name: string, disable: boolean) => {
        const response = await registrationClient.GetService({ serviceName: name });
        const service = response.service;
        if (!service) throw new Error('Service not found');

        service.disable = disable;
        // RegisterService expects RegisterServiceRequest which has 'config' field
        await registrationClient.RegisterService({ config: service });
        return service;
    },
    getServiceStatus: async (name: string) => {
        const response = await registrationClient.GetServiceStatus({ serviceName: name, namespace: '' });
        return response;
    },
    registerService: async (config: UpstreamServiceConfig) => {
        const response = await registrationClient.RegisterService({ config });
        return response;
    },
    updateService: async (config: UpstreamServiceConfig) => {
        // gRPC uses RegisterService for both create and update
        const response = await registrationClient.RegisterService({ config });
        return response;
    },
    unregisterService: async (id: string) => {
         await registrationClient.UnregisterService({ serviceName: id, namespace: '' });
         return {};
    },

    // Tools (Legacy Fetch - Not yet migrated to Admin/Registration Service completely or keeping as is)
    // admin.proto has ListTools but we are focusing on RegistrationService first.
    // So keep using fetch for Tools/Secrets/etc for now.
    listTools: async () => {
        const res = await fetch('/api/v1/tools');
        if (!res.ok) throw new Error('Failed to fetch tools');
        return res.json();
    },
    executeTool: async (request: any) => {
        try {
            const res = await fetch('/api/v1/execute', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(request)
            });
            if (!res.ok) throw new Error('Failed to execute tool');
            return res.json();
        } catch (e) {
            console.error("DEBUG: fetch failed:", e);
            throw e;
        }
    },
    setToolStatus: async (name: string, enabled: boolean) => {
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
