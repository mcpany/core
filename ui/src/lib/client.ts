/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from '@proto/config/v1/upstream_service';
import { ToolDefinition } from '@proto/config/v1/tool';
import { ResourceDefinition } from '@proto/config/v1/resource';
import { PromptDefinition } from '@proto/config/v1/prompt';
import { Secret as SecretDefinition } from '@proto/config/v1/config';
import { GlobalSettings } from '@proto/config/v1/config';

/**
 * Client for interacting with the MCPAny backend API.
 * Provides methods for managing services, tools, resources, prompts, and secrets.
 */

// NOTE: Adjusted to point to local Next.js API routes for this UI overhaul task
// In a real deployment, these might be /api/v1/... proxied to backend

// Export generated types for consumption
export type { UpstreamServiceConfig, ToolDefinition, ResourceDefinition, PromptDefinition, SecretDefinition, GlobalSettings };

export const apiClient = {
    // Services
    listServices: async (): Promise<UpstreamServiceConfig[]> => {
        const res = await fetch('/api/services');
        if (!res.ok) throw new Error('Failed to fetch services');
        const json = await res.json();
        return json.map((s: any) => UpstreamServiceConfig.fromJSON(s));
    },
    getService: async (id: string): Promise<UpstreamServiceConfig> => {
         const res = await fetch(`/api/services/${id}`);
         if (!res.ok) throw new Error('Failed to fetch service');
         return UpstreamServiceConfig.fromJSON(await res.json());
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
    getServiceStatus: async (name: string) => {
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
            body: JSON.stringify(UpstreamServiceConfig.toJSON(config))
        });
        if (!res.ok) throw new Error('Failed to register service');
        return UpstreamServiceConfig.fromJSON(await res.json());
    },
    updateService: async (config: UpstreamServiceConfig) => {
         const res = await fetch('/api/services', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
             body: JSON.stringify(UpstreamServiceConfig.toJSON(config))
        });
        if (!res.ok) throw new Error('Failed to update service');
        return UpstreamServiceConfig.fromJSON(await res.json());
    },
    unregisterService: async (id: string) => {
         // Mock
        return {};
    },

    // Tools
    listTools: async (): Promise<ToolDefinition[]> => {
        const res = await fetch('/api/tools');
        if (!res.ok) throw new Error('Failed to fetch tools');
        const json = await res.json();
        return json.map((t: any) => ToolDefinition.fromJSON(t));
    },
    executeTool: async (request: any) => {
        // Mock execution
        return { output: "Mock execution result", success: true };
    },
    setToolStatus: async (name: string, enabled: boolean) => {
        const res = await fetch('/api/tools', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            // Note: Tool definition uses 'disable', but API might expect enabled/disabled toggle.
            // Keeping payload as is for now assuming backend handles 'enabled' for toggle action.
            body: JSON.stringify({ name, enabled })
        });
        return res.json();
    },

    // Resources
    listResources: async (): Promise<ResourceDefinition[]> => {
        const res = await fetch('/api/resources');
        if (!res.ok) throw new Error('Failed to fetch resources');
         const json = await res.json();
        return json.map((r: any) => ResourceDefinition.fromJSON(r));
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
    listPrompts: async (): Promise<PromptDefinition[]> => {
        const res = await fetch('/api/prompts');
        if (!res.ok) throw new Error('Failed to fetch prompts');
        const json = await res.json();
        return json.map((p: any) => PromptDefinition.fromJSON(p));
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
    listSecrets: async (): Promise<SecretDefinition[]> => {
        const res = await fetch('/api/secrets');
        if (!res.ok) throw new Error('Failed to fetch secrets');
        const json = await res.json();
        // SecretList wrapper? Or list of secrets?
        // Checking config.proto: message SecretList { repeated Secret secrets = 1; }
        // Attempt to parse as list or SecretList
        if (json.secrets) {
             return json.secrets.map((s: any) => SecretDefinition.fromJSON(s));
        }
        return Array.isArray(json) ? json.map((s: any) => SecretDefinition.fromJSON(s)) : [];
    },
    saveSecret: async (secret: SecretDefinition) => {
        const res = await fetch('/api/secrets', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(SecretDefinition.toJSON(secret))
        });
        if (!res.ok) throw new Error('Failed to save secret');
        return SecretDefinition.fromJSON(await res.json());
    },
    deleteSecret: async (id: string) => {
        const res = await fetch(`/api/secrets/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete secret');
        return {};
    },

    // Global Settings
    getGlobalSettings: async (): Promise<GlobalSettings> => {
        const res = await fetch('/api/settings');
        if (!res.ok) throw new Error('Failed to fetch global settings');
        return GlobalSettings.fromJSON(await res.json());
    },
    saveGlobalSettings: async (settings: GlobalSettings) => {
        const res = await fetch('/api/settings', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(GlobalSettings.toJSON(settings))
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
