/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Client for interacting with the MCPAny backend API.
 * Provides methods for managing services, tools, resources, prompts, and secrets.
 */

import { GrpcWebImpl, RegistrationServiceClientImpl } from '@proto/api/v1/registration';
import { UpstreamServiceConfig as BaseUpstreamServiceConfig, HttpUpstreamService } from '@proto/config/v1/upstream_service';
import { ProfileDefinition } from '@proto/config/v1/config';
import { ToolDefinition } from '@proto/config/v1/tool';
import { ResourceDefinition } from '@proto/config/v1/resource';
import { PromptDefinition } from '@proto/config/v1/prompt';
import { Credential, Authentication } from '@proto/config/v1/auth';

import { BrowserHeaders } from 'browser-headers';

/**
 * Extended UpstreamServiceConfig to include runtime error information.
 */
export interface UpstreamServiceConfig extends Omit<BaseUpstreamServiceConfig, 'lastError' | 'toolCount'> {
    /**
     * The last error message encountered by the service, if any.
     */
    lastError?: string;
    /**
     * The number of tools registered for this service.
     */
    toolCount?: number;
    /**
     * Optional description for the service (used in UI templates).
     */
    description?: string;
}

// Re-export generated types
export type { ToolDefinition, ResourceDefinition, PromptDefinition, Credential, Authentication, ProfileDefinition };
export type { ListServicesResponse, GetServiceResponse, GetServiceStatusResponse, ValidateServiceResponse } from '@proto/api/v1/registration';

/**
 * ServiceTemplate defines a template for an upstream service.
 */
export interface ServiceTemplate {
    id: string;
    name: string;
    description: string;
    icon: string;
    tags: string[];
    authType?: string; // Optional helper for UI
    serviceConfig: UpstreamServiceConfig;
}

// Helper to map snake_case config to camelCase UpstreamServiceConfig
const mapUpstreamServiceConfig = (s: any): UpstreamServiceConfig => ({
    ...s,
    connectionPool: s.connection_pool,
    httpService: s.http_service ? HttpUpstreamService.fromJSON(s.http_service) : undefined,
    grpcService: s.grpc_service,
    commandLineService: s.command_line_service,
    mcpService: s.mcp_service,
    upstreamAuth: s.upstream_auth,
    preCallHooks: s.pre_call_hooks,
    postCallHooks: s.post_call_hooks,
    lastError: s.last_error,
    toolCount: s.tool_count,
    toolExportPolicy: s.tool_export_policy,
    promptExportPolicy: s.prompt_export_policy,
    resourceExportPolicy: s.resource_export_policy,
    callPolicies: s.call_policies?.map((p: any) => ({
        defaultAction: p.default_action,
        rules: p.rules?.map((r: any) => ({
            action: r.action,
            nameRegex: r.name_regex,
            argumentRegex: r.argument_regex,
            urlRegex: r.url_regex,
            callIdRegex: r.call_id_regex
        }))
    })),
});

// Initialize gRPC Web Client
const getBaseUrl = () => {
    if (typeof window !== 'undefined') {
        return window.location.origin;
    }
    return process.env.BACKEND_URL || 'http://mcpany:50050'; // Default for SSR in K8s
};

const rpc = new GrpcWebImpl(getBaseUrl(), {
  debug: false,
});
const registrationClient = new RegistrationServiceClientImpl(rpc);

const fetchWithAuth = async (input: RequestInfo | URL, init?: RequestInit) => {
    const headers = new Headers(init?.headers);
    // Inject Authorization header from localStorage if available
    if (typeof window !== 'undefined') {
        const token = localStorage.getItem('mcp_auth_token');
        if (token) {
            headers.set('Authorization', `Basic ${token}`);
        }
    } else {
        // Server-side: Inject API Key from env
        const apiKey = process.env.MCPANY_API_KEY;
        if (apiKey) {
            headers.set('X-API-Key', apiKey);
        }
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

export interface ResourceContent {
    uri: string;
    mimeType: string;
    text?: string;
    blob?: string;
}

export interface ReadResourceResponse {
    contents: ResourceContent[];
}

export interface CheckResult {
    status: string;
    message?: string;
    latency?: string;
    diff?: string;
}

export interface DoctorReport {
    status: string;
    timestamp: string;
    checks: Record<string, CheckResult>;
}

export interface ToolFailureStats {
    name: string;
    serviceId: string;
    failureRate: number;
    totalCalls: number;
}

export interface ToolAnalytics {
    name: string;
    serviceId: string;
    totalCalls: number;
    successRate: number;
}

export interface Metric {
    label: string;
    value: string;
    change?: string;
    trend?: "up" | "down" | "neutral";
    icon: string;
    subLabel?: string;
}

export interface SystemStatus {
    uptime_seconds: number;
    active_connections: number;
    bound_http_port: number;
    bound_grpc_port: number;
    version: string;
    security_warnings: string[];
}

export type ServiceStatus = "healthy" | "degraded" | "unhealthy" | "inactive" | "unknown";

export interface ServiceHealth {
  id: string;
  name: string;
  status: ServiceStatus;
  latency: string;
  uptime: string;
  message?: string;
}

export interface HealthHistoryPoint {
  timestamp: number;
  status: ServiceStatus;
}

export interface ServiceHealthResponse {
  services: ServiceHealth[];
  history: Record<string, HealthHistoryPoint[]>;
}

const getMetadata = () => {
    // Metadata for gRPC calls.
    // If gRPC endpoint is proxied via Next.js rewrites, middleware injects auth.
    return undefined;
};

/**
 * API Client for interacting with the MCP Any server.
 */
export const apiClient = {
    // Services (Migrated to gRPC)

    /**
     * Lists all registered upstream services.
     */
    listServices: async () => {
        // We use REST endpoint which is stable and proxied correctly by Next.js
        const res = await fetchWithAuth('/api/v1/services');
        if (!res.ok) throw new Error('Failed to fetch services');
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.services || []);
        // Map snake_case to camelCase for UI compatibility
        return list.map(mapUpstreamServiceConfig);
    },

    /**
     * Lists services from the dynamic catalog.
     */
    listCatalog: async () => {
        const res = await fetchWithAuth('/api/v1/catalog/services');
        if (!res.ok) throw new Error('Failed to fetch catalog');
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.services || []);

        return list.map((s: any) => ({
            ...s,
            connectionPool: s.connection_pool,
            httpService: s.http_service ? HttpUpstreamService.fromJSON(s.http_service) : undefined,
            grpcService: s.grpc_service,
            commandLineService: s.command_line_service,
            mcpService: s.mcp_service,
            upstreamAuth: s.upstream_auth,
            preCallHooks: s.pre_call_hooks,
            postCallHooks: s.post_call_hooks,
            lastError: s.last_error,
            toolCount: s.tool_count,
            toolExportPolicy: s.tool_export_policy,
            promptExportPolicy: s.prompt_export_policy,
            resourceExportPolicy: s.resource_export_policy,
        }));
    },

    /**
     * Gets a single service by its ID.
     */
    getService: async (id: string) => {
         try {
             // Try gRPC-Web first
             const response = await registrationClient.GetService({ serviceName: id }, getMetadata());
             return response;
         } catch (e) {
             // Fallback to REST if gRPC fails (e.g. environment issues or proxy misconfiguration)
             const res = await fetchWithAuth(`/api/v1/services/${id}`);
             if (res.ok) {
                 const data = await res.json();
                 // REST returns { service: ... }, gRPC returns { service: ... }
                 // Map snake_case to camelCase for ServiceDetail
                 if (data.service) {
                     const s = data.service;
                     data.service = {
                         ...s,
                         connectionPool: s.connection_pool,
                         httpService: s.http_service ? HttpUpstreamService.fromJSON(s.http_service) : undefined,
                         grpcService: s.grpc_service,
                         commandLineService: s.command_line_service,
                         mcpService: s.mcp_service,
                         upstreamAuth: s.upstream_auth,
                         preCallHooks: s.pre_call_hooks,
                         postCallHooks: s.post_call_hooks,
                         toolExportPolicy: s.tool_export_policy,
                         promptExportPolicy: s.prompt_export_policy,
                         resourceExportPolicy: s.resource_export_policy,
                         callPolicies: s.call_policies?.map((p: any) => ({
                            defaultAction: p.default_action,
                            rules: p.rules?.map((r: any) => ({
                                action: r.action,
                                nameRegex: r.name_regex,
                                argumentRegex: r.argument_regex,
                                urlRegex: r.url_regex,
                                callIdRegex: r.call_id_regex
                            }))
                        })),
                     };
                 }
                 return data;
             }
             throw e;
         }
    },

    /**
     * Sets the status (enabled/disabled) of a service.
     */
    setServiceStatus: async (name: string, disable: boolean) => {
        const response = await fetchWithAuth(`/api/v1/services/${name}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ disable })
        });
        if (!response.ok) throw new Error('Failed to update service status');
        return response.json();
    },

    /**
     * Gets the status of a service.
     */
    getServiceStatus: async (name: string) => {
        const res = await fetchWithAuth(`/api/v1/services/${name}/status`);
        if (!res.ok) throw new Error('Failed to fetch service status');
        return res.json();
    },

    /**
     * Restarts a service.
     */
    restartService: async (name: string) => {
        const response = await fetchWithAuth(`/api/v1/services/${name}/restart`, {
            method: 'POST'
        });
        if (!response.ok) throw new Error('Failed to restart service');
        return {};
    },

    /**
     * Registers a new upstream service.
     */
    registerService: async (config: UpstreamServiceConfig) => {
        // Map camelCase (UI) to snake_case (Server REST)
        const payload: any = {
            id: config.id,
            name: config.name,
            version: config.version,
            disable: config.disable,
            priority: config.priority,
            load_balancing_strategy: config.loadBalancingStrategy,
            tags: config.tags,
        };

        if (config.httpService) {
            payload.http_service = HttpUpstreamService.toJSON(config.httpService);
        }
        if (config.grpcService) {
            payload.grpc_service = { address: config.grpcService.address };
        }
        if (config.commandLineService) {
            payload.command_line_service = {
                command: config.commandLineService.command,
                working_directory: config.commandLineService.workingDirectory,
                environment: config.commandLineService.env,
                env: config.commandLineService.env
            };
        }
        if (config.mcpService) {
            payload.mcp_service = { ...config.mcpService };
        }
        if (config.openapiService) {
            payload.openapi_service = {
                address: config.openapiService.address,
                spec_url: config.openapiService.specUrl,
                spec_content: config.openapiService.specContent,
                tools: config.openapiService.tools,
                resources: config.openapiService.resources,
                prompts: config.openapiService.prompts,
                calls: config.openapiService.calls,
                health_check: config.openapiService.healthCheck,
                tls_config: config.openapiService.tlsConfig
            };
        }
        if (config.preCallHooks) {
            payload.pre_call_hooks = config.preCallHooks;
        }
        if (config.postCallHooks) {
            payload.post_call_hooks = config.postCallHooks;
        }
        if (config.callPolicies) {
            payload.call_policies = config.callPolicies.map((p: any) => ({
                default_action: p.defaultAction,
                rules: p.rules?.map((r: any) => ({
                    action: r.action,
                    name_regex: r.nameRegex,
                    argument_regex: r.argumentRegex,
                    url_regex: r.urlRegex,
                    call_id_regex: r.callIdRegex
                }))
            }));
        }
        if (config.toolExportPolicy) {
            payload.tool_export_policy = config.toolExportPolicy;
        }
        if (config.promptExportPolicy) {
            payload.prompt_export_policy = config.promptExportPolicy;
        }
        if (config.resourceExportPolicy) {
            payload.resource_export_policy = config.resourceExportPolicy;
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

    /**
     * Updates an existing upstream service.
     */
    updateService: async (config: UpstreamServiceConfig) => {
        const payload: any = {
             id: config.id,
            name: config.name,
            version: config.version,
            disable: config.disable,
            priority: config.priority,
            load_balancing_strategy: config.loadBalancingStrategy,
            tags: config.tags,
        };
         if (config.httpService) {
            payload.http_service = HttpUpstreamService.toJSON(config.httpService);
        }
        if (config.grpcService) {
            payload.grpc_service = { address: config.grpcService.address };
        }
        if (config.commandLineService) {
            payload.command_line_service = {
                command: config.commandLineService.command,
                working_directory: config.commandLineService.workingDirectory,
            };
        }
        if (config.mcpService) {
            payload.mcp_service = { ...config.mcpService };
        }
        if (config.openapiService) {
            payload.openapi_service = {
                address: config.openapiService.address,
                spec_url: config.openapiService.specUrl,
                spec_content: config.openapiService.specContent,
                tools: config.openapiService.tools,
                resources: config.openapiService.resources,
                prompts: config.openapiService.prompts,
                calls: config.openapiService.calls,
                health_check: config.openapiService.healthCheck,
                tls_config: config.openapiService.tlsConfig
            };
        }
        if (config.preCallHooks) {
            payload.pre_call_hooks = config.preCallHooks;
        }
        if (config.postCallHooks) {
            payload.post_call_hooks = config.postCallHooks;
        }
        if (config.callPolicies) {
            payload.call_policies = config.callPolicies.map((p: any) => ({
                default_action: p.defaultAction,
                rules: p.rules?.map((r: any) => ({
                    action: r.action,
                    name_regex: r.nameRegex,
                    argument_regex: r.argumentRegex,
                    url_regex: r.urlRegex,
                    call_id_regex: r.callIdRegex
                }))
            }));
        }
        if (config.toolExportPolicy) {
            payload.tool_export_policy = config.toolExportPolicy;
        }
        if (config.promptExportPolicy) {
            payload.prompt_export_policy = config.promptExportPolicy;
        }
        if (config.resourceExportPolicy) {
            payload.resource_export_policy = config.resourceExportPolicy;
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

    /**
     * Unregisters (deletes) an upstream service.
     */
    unregisterService: async (id: string) => {
         const response = await fetchWithAuth(`/api/v1/services/${id}`, {
            method: 'DELETE'
         });
         if (!response.ok) throw new Error('Failed to unregister service');
         return {};
    },

    /**
     * Validates a service configuration.
     */
    validateService: async (config: UpstreamServiceConfig) => {
        const payload: any = {
            id: config.id,
            name: config.name,
            version: config.version,
            disable: config.disable,
            priority: config.priority,
            load_balancing_strategy: config.loadBalancingStrategy,
            tags: config.tags,
        };

        if (config.httpService) {
            payload.http_service = HttpUpstreamService.toJSON(config.httpService);
        }
        if (config.grpcService) {
            payload.grpc_service = { address: config.grpcService.address };
        }
        if (config.commandLineService) {
            payload.command_line_service = {
                command: config.commandLineService.command,
                working_directory: config.commandLineService.workingDirectory,
                env: config.commandLineService.env,
                container_environment: config.commandLineService.containerEnvironment,
            };
        }
        if (config.mcpService) {
            payload.mcp_service = { ...config.mcpService };
        }
        if (config.openapiService) {
            payload.openapi_service = {
                address: config.openapiService.address,
                spec_url: config.openapiService.specUrl,
                spec_content: config.openapiService.specContent,
                tools: config.openapiService.tools,
                resources: config.openapiService.resources,
                prompts: config.openapiService.prompts,
                calls: config.openapiService.calls,
                health_check: config.openapiService.healthCheck,
                tls_config: config.openapiService.tlsConfig
            };
        }
        if (config.preCallHooks) {
            payload.pre_call_hooks = config.preCallHooks;
        }
        if (config.postCallHooks) {
            payload.post_call_hooks = config.postCallHooks;
        }
        if (config.callPolicies) {
            payload.call_policies = config.callPolicies.map((p: any) => ({
                default_action: p.defaultAction,
                rules: p.rules?.map((r: any) => ({
                    action: r.action,
                    name_regex: r.nameRegex,
                    argument_regex: r.argumentRegex,
                    url_regex: r.urlRegex,
                    call_id_regex: r.callIdRegex
                }))
            }));
        }
        if (config.toolExportPolicy) {
            payload.tool_export_policy = config.toolExportPolicy;
        }
        if (config.promptExportPolicy) {
            payload.prompt_export_policy = config.promptExportPolicy;
        }
        if (config.resourceExportPolicy) {
            payload.resource_export_policy = config.resourceExportPolicy;
        }

        const response = await fetchWithAuth('/api/v1/services/validate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });

        const text = await response.text();
        let data: any;
        try {
            data = JSON.parse(text);
        } catch (e) {
            // Not JSON
        }

        if (!response.ok) {
            if (data && data.error) {
                 return data; // Return the error object (valid: false, error: ...)
            }
            throw new Error(`Failed to validate service: ${response.status} ${text}`);
        }
        return data;
    },

    /**
     * Lists all available tools.
     */
    listTools: async () => {
        const res = await fetchWithAuth('/api/v1/tools');
        if (!res.ok) throw new Error('Failed to fetch tools');
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.tools || []);
        return {
            tools: list.map((t: any) => ({
                ...t,
                serviceId: t.serviceId || t.service_id,
                inputSchema: t.inputSchema || t.input_schema,
                outputSchema: t.outputSchema || t.output_schema,
            }))
        };
    },

    /**
     * Executes a tool with the provided arguments.
     */
    executeTool: async (request: any, dryRun?: boolean) => {
        try {
            const payload = { ...request };
            if (dryRun) {
                payload.dryRun = true;
            }
            const res = await fetchWithAuth('/api/v1/execute', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            if (!res.ok) {
                const text = await res.text();
                let errorMsg = null;
                try {
                    const json = JSON.parse(text);
                    if (json.error) errorMsg = json.error;
                } catch (e) {
                    // ignore
                }
                if (errorMsg) throw new Error(errorMsg);
                throw new Error(`Failed to execute tool: ${text || res.statusText}`);
            }
            return res.json();
        } catch (e) {
            console.warn("DEBUG: fetch failed:", e);
            throw e;
        }
    },

    /**
     * Sets the status (enabled/disabled) of a tool.
     */
    setToolStatus: async (name: string, disabled: boolean) => {
        const res = await fetchWithAuth('/api/v1/tools', {
            method: 'PUT',
            body: JSON.stringify({ name, disable: disabled })
        });
        return res.json();
    },

    // Resources

    /**
     * Lists all available resources.
     */
    listResources: async () => {
        const res = await fetchWithAuth('/api/v1/resources');
        if (!res.ok) throw new Error('Failed to list resources');
        return res.json();
    },

    /**
     * Reads the content of a resource.
     */
    readResource: async (uri: string): Promise<ReadResourceResponse> => {
        const res = await fetchWithAuth(`/api/v1/resources/read?uri=${encodeURIComponent(uri)}`);
        if (!res.ok) throw new Error('Failed to read resource');
        return res.json();
    },

    /**
     * Sets the status (enabled/disabled) of a resource.
     */
    setResourceStatus: async (uri: string, disabled: boolean) => {
        const res = await fetchWithAuth('/api/v1/resources', {
            method: 'PUT',
            body: JSON.stringify({ uri, disable: disabled })
        });
        return res.json();
    },

    // Prompts

    /**
     * Lists all available prompts.
     */
    listPrompts: async () => {
        const res = await fetchWithAuth('/api/v1/prompts');
        if (!res.ok) throw new Error(`Failed to fetch prompts: ${res.status}`);
        return res.json();
    },

    /**
     * Sets the status (enabled/disabled) of a prompt.
     */
    setPromptStatus: async (name: string, enabled: boolean) => {
        const res = await fetchWithAuth('/api/v1/prompts', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, enabled })
        });
        return res.json();
    },

    /**
     * Executes a prompt with the given arguments.
     */
    executePrompt: async (name: string, args: Record<string, string>) => {
        const res = await fetchWithAuth('/api/v1/prompts/execute', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, arguments: args })
        });
        if (!res.ok) throw new Error('Failed to execute prompt');
        return res.json();
    },

    // Wizard Helpers

    /**
     * Returns a list of available service templates for the wizard.
     */
    getServiceTemplates: async () => {
        const res = await fetchWithAuth('/api/v1/templates');
        if (!res.ok) throw new Error('Failed to fetch templates');
        const data = await res.json();
        const list = Array.isArray(data) ? data : [];

        return list.map((t: any) => {
            const sc = t.service_config || {};
            const auth = sc.upstream_auth;
            let authType = 'none';
            if (auth?.oauth2) authType = 'oauth2';
            else if (auth?.api_key) authType = 'apiKey';
            else if (auth?.bearer_token) authType = 'token';
            else if (auth?.basic_auth) authType = 'basic';

            return {
                id: t.id,
                name: t.name,
                description: t.description,
                icon: t.icon,
                tags: t.tags || [],
                authType: authType,
                serviceConfig: {
                    ...sc,
                    connectionPool: sc.connection_pool,
                    httpService: sc.http_service ? HttpUpstreamService.fromJSON(sc.http_service) : undefined,
                    grpcService: sc.grpc_service,
                    commandLineService: sc.command_line_service,
                    mcpService: sc.mcp_service,
                    upstreamAuth: sc.upstream_auth,
                    preCallHooks: sc.pre_call_hooks,
                    postCallHooks: sc.post_call_hooks,
                    toolExportPolicy: sc.tool_export_policy,
                    promptExportPolicy: sc.prompt_export_policy,
                    resourceExportPolicy: sc.resource_export_policy,
                    callPolicies: sc.call_policies?.map((p: any) => ({
                        defaultAction: p.default_action,
                        rules: p.rules?.map((r: any) => ({
                            action: r.action,
                            nameRegex: r.name_regex,
                            argumentRegex: r.argument_regex,
                            urlRegex: r.url_regex,
                            callIdRegex: r.call_id_regex
                        }))
                    })),
                    openapiService: sc.openapi_service ? {
                        address: sc.openapi_service.address,
                        specUrl: sc.openapi_service.spec_url,
                        specContent: sc.openapi_service.spec_content,
                        tools: sc.openapi_service.tools,
                        resources: sc.openapi_service.resources,
                        prompts: sc.openapi_service.prompts,
                        calls: sc.openapi_service.calls,
                        healthCheck: sc.openapi_service.health_check,
                        tlsConfig: sc.openapi_service.tls_config
                    } : undefined
                }
            };
        });
    },

    /**
     * Initiates an OAuth flow for a specific service.
     */
    initiateOAuth: async (serviceId: string, credentialId: string, redirectUrl: string) => {
        const res = await fetchWithAuth('/api/v1/auth/oauth/initiate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                service_id: serviceId,
                credential_id: credentialId,
                redirect_url: redirectUrl
            })
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to initiate OAuth: ${txt}`);
        }
        return res.json();
    },

    // Profiles

    /**
     * Creates a new profile.
     */
    createProfile: async (profileData: any) => {
        const res = await fetchWithAuth('/api/v1/profiles', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(profileData)
        });
        if (!res.ok) throw new Error('Failed to create profile');
        return res.json();
    },

    /**
     * Updates an existing profile.
     */
    updateProfile: async (profileData: any) => {
        const res = await fetchWithAuth(`/api/v1/profiles/${profileData.name}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(profileData)
        });
        if (!res.ok) throw new Error('Failed to update profile');
        return res.json();
    },

    /**
    * Deletes a profile.
    */
    deleteProfile: async (name: string) => {
        const res = await fetchWithAuth(`/api/v1/profiles/${name}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete profile');
        return {};
    },

    /**
     * Lists all profiles.
     */
    listProfiles: async () => {
        const res = await fetchWithAuth('/api/v1/profiles');
        if (!res.ok) throw new Error('Failed to list profiles');
        const data = await res.json();
        return data.profiles || [];
    },

    // Secrets

    /**
     * Lists all stored secrets.
     */
    listSecrets: async () => {
        const res = await fetchWithAuth('/api/v1/secrets');
        if (!res.ok) throw new Error('Failed to fetch secrets');
        const data = await res.json();
        return Array.isArray(data) ? data : (data.secrets || []);
    },

    /**
     * Reveals a secret value.
     */
    revealSecret: async (id: string): Promise<{ value: string }> => {
        const res = await fetchWithAuth(`/api/v1/secrets/${id}/reveal`, {
            method: 'POST'
        });
        if (!res.ok) throw new Error('Failed to reveal secret');
        return res.json();
    },

    /**
     * Saves a secret.
     */
    saveSecret: async (secret: SecretDefinition) => {
        const res = await fetchWithAuth('/api/v1/secrets', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(secret)
        });
        if (!res.ok) throw new Error('Failed to save secret');
        return res.json();
    },

    /**
     * Deletes a secret.
     */
    deleteSecret: async (id: string) => {
        const res = await fetchWithAuth(`/api/v1/secrets/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete secret');
        return {};
    },

    // Global Settings

    /**
     * Gets the global server settings.
     */
    getGlobalSettings: async () => {
        const res = await fetchWithAuth('/api/v1/settings');
        if (!res.ok) throw new Error('Failed to fetch global settings');
        return res.json();
    },

    /**
     * Saves the global server settings.
     */
    saveGlobalSettings: async (settings: any) => {
        const res = await fetchWithAuth('/api/v1/settings', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(settings)
        });
        if (!res.ok) throw new Error('Failed to save global settings');
    },

    /**
     * Gets the dashboard traffic history.
     */
    getDashboardTraffic: async (serviceId?: string, timeRange?: string) => {
        let url = '/api/v1/dashboard/traffic';
        const params = new URLSearchParams();
        if (serviceId) params.append('serviceId', serviceId);
        if (timeRange) params.append('timeRange', timeRange);

        if (params.toString()) url += `?${params.toString()}`;

        const res = await fetchWithAuth(url);
        if (!res.ok) throw new Error('Failed to fetch dashboard traffic');
        return res.json();
    },

    /**
     * Gets the top used tools.
     */
    getTopTools: async (serviceId?: string) => {
        let url = '/api/v1/dashboard/top-tools';
        if (serviceId) url += `?serviceId=${encodeURIComponent(serviceId)}`;
        const res = await fetchWithAuth(url);
        if (!res.ok) return [];
        return res.json();
    },

    // Alerts

    /**
     * Lists all alerts.
     */
    listAlerts: async () => {
        const res = await fetchWithAuth('/api/v1/alerts');
        if (!res.ok) throw new Error('Failed to fetch alerts');
        return res.json();
    },

    /**
     * Lists all alert rules.
     */
    listAlertRules: async () => {
        const res = await fetchWithAuth('/api/v1/alerts/rules');
        if (!res.ok) throw new Error('Failed to fetch alert rules');
        return res.json();
    },

    /**
     * Creates a new alert rule.
     */
    createAlertRule: async (rule: any) => {
        const res = await fetchWithAuth('/api/v1/alerts/rules', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(rule)
        });
        if (!res.ok) throw new Error('Failed to create alert rule');
        return res.json();
    },

    /**
     * Gets an alert rule by ID.
     */
    getAlertRule: async (id: string) => {
        const res = await fetchWithAuth(`/api/v1/alerts/rules/${id}`);
        if (!res.ok) throw new Error('Failed to get alert rule');
        return res.json();
    },

    /**
     * Updates an alert rule.
     */
    updateAlertRule: async (rule: any) => {
        const res = await fetchWithAuth(`/api/v1/alerts/rules/${rule.id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(rule)
        });
        if (!res.ok) throw new Error('Failed to update alert rule');
        return res.json();
    },

    /**
     * Deletes an alert rule.
     */
    deleteAlertRule: async (id: string) => {
        const res = await fetchWithAuth(`/api/v1/alerts/rules/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete alert rule');
        return {};
    },

    /**
     * Gets the tools with highest failure rates.
     */
    getToolFailures: async (serviceId?: string): Promise<ToolFailureStats[]> => {
        let url = '/api/v1/dashboard/tool-failures';
        if (serviceId) url += `?serviceId=${encodeURIComponent(serviceId)}`;
        const res = await fetchWithAuth(url);
        if (!res.ok) return [];
        return res.json();
    },

    /**
     * Gets the tool usage analytics.
     */
    getToolUsage: async (serviceId?: string): Promise<ToolAnalytics[]> => {
        let url = '/api/v1/dashboard/tool-usage';
        if (serviceId) url += `?serviceId=${encodeURIComponent(serviceId)}`;
        const res = await fetchWithAuth(url);
        if (!res.ok) return [];
        return res.json();
    },


    /**
     * Gets the system status.
     */
    getSystemStatus: async (): Promise<SystemStatus> => {
        const res = await fetchWithAuth('/api/v1/system/status');
        if (!res.ok) throw new Error('Failed to fetch system status');
        return res.json();
    },

    /**
     * Gets the dashboard health status and history.
     */
    getDashboardHealth: async (): Promise<ServiceHealthResponse> => {
        const res = await fetchWithAuth('/api/v1/dashboard/health');
        if (!res.ok) throw new Error('Failed to fetch dashboard health');
        return res.json();
    },

    /**
     * Gets the dashboard metrics.
     */
    getDashboardMetrics: async (serviceId?: string): Promise<Metric[]> => {
        let url = '/api/v1/dashboard/metrics';
        if (serviceId) url += `?serviceId=${encodeURIComponent(serviceId)}`;
        const res = await fetchWithAuth(url);
        if (!res.ok) throw new Error('Failed to fetch dashboard metrics. Is the server running and authenticated?');
        return res.json();
    },

    /**
     * Gets the latest execution traces.
     */
    getTraces: async (options?: { limit?: number }): Promise<any[]> => {
        let url = '/api/v1/traces';
        if (options?.limit) {
            url += `?limit=${options.limit}`;
        }
        const res = await fetchWithAuth(url);
        if (!res.ok) throw new Error('Failed to fetch traces');
        return res.json();
    },

    /**
     * Seeds the dashboard traffic history (Debug/Test only).
     */
    seedTrafficData: async (points: any[]) => {
        const res = await fetchWithAuth('/api/v1/debug/seed_traffic', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(points)
        });
        if (!res.ok) throw new Error('Failed to seed traffic data');
    },

    /**
     * Updates an alert status.
     */
    updateAlertStatus: async (id: string, status: string) => {
        const res = await fetchWithAuth(`/api/v1/alerts/${id}`, {
            method: 'PATCH',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status })
        });
        if (!res.ok) throw new Error('Failed to update alert status');
        return res.json();
    },

    /**
     * Gets the configured global webhook URL for alerts.
     */
    getWebhookURL: async (): Promise<{ url: string }> => {
        const res = await fetchWithAuth('/api/v1/alerts/webhook');
        if (!res.ok) throw new Error('Failed to fetch webhook URL');
        return res.json();
    },

    /**
     * Saves the configured global webhook URL for alerts.
     */
    saveWebhookURL: async (url: string) => {
        const res = await fetchWithAuth('/api/v1/alerts/webhook', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ url })
        });
        if (!res.ok) throw new Error('Failed to save webhook URL');
        return res.json();
    },

    // Stack Management (Collections)

    /**
     * Lists all service collections (stacks).
     */
    listCollections: async () => {
        const res = await fetchWithAuth('/api/v1/collections');
        if (!res.ok) throw new Error('Failed to list collections');
        return res.json();
    },

    /**
     * Gets a single service collection (stack) by its name.
     */
    getCollection: async (name: string) => {
        const res = await fetchWithAuth(`/api/v1/collections/${name}`);
        if (!res.ok) throw new Error('Failed to get collection');
        return res.json();
    },

    /**
     * Saves a service collection (stack).
     */
    saveCollection: async (collection: any) => {
        const res = await fetchWithAuth(`/api/v1/collections/${collection.name}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(collection)
        });
        if (!res.ok) {
             throw new Error('Failed to save collection');
        }
        return res.json();
    },

    /**
     * Deletes a service collection (stack).
     */
    deleteCollection: async (name: string) => {
        const res = await fetchWithAuth(`/api/v1/collections/${name}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete collection');
        return {};
    },

    /**
     * Gets the configuration for a stack (Compatibility wrapper).
     */
    getStackConfig: async (stackId: string) => {
        // Map to getCollection
        return apiClient.getCollection(stackId);
    },

    /**
     * Saves the configuration for a stack (Compatibility wrapper).
     */
    saveStackConfig: async (stackId: string, config: any) => {
        // Map to saveCollection. Ensure name is set.
        const collection = typeof config === 'string' ? JSON.parse(config) : config;
        if (!collection.name) collection.name = stackId;
        return apiClient.saveCollection(collection);
    },

    /**
     * Gets the stack configuration as YAML.
     */
    getStackYaml: async (stackId: string) => {
        const res = await fetchWithAuth(`/api/v1/stacks/${stackId}/config`);
        if (!res.ok) throw new Error('Failed to get stack config');
        return res.text();
    },

    /**
     * Saves the stack configuration from YAML.
     */
    saveStackYaml: async (stackId: string, yamlContent: string) => {
        const res = await fetchWithAuth(`/api/v1/stacks/${stackId}/config`, {
            method: 'POST',
            headers: { 'Content-Type': 'text/yaml' }, // or text/plain depending on server handler
            body: yamlContent
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to save stack config: ${txt}`);
        }
        return res.json();
    },

    // Audit Logs

    /**
     * Lists audit logs.
     */
    listAuditLogs: async (filters: {
        start_time?: string;
        end_time?: string;
        tool_name?: string;
        user_id?: string;
        profile_id?: string;
        limit?: number;
        offset?: number;
    }) => {
        const query = new URLSearchParams();
        if (filters.start_time) query.set('start_time', filters.start_time);
        if (filters.end_time) query.set('end_time', filters.end_time);
        if (filters.tool_name) query.set('tool_name', filters.tool_name);
        if (filters.user_id) query.set('user_id', filters.user_id);
        if (filters.profile_id) query.set('profile_id', filters.profile_id);
        if (filters.limit) query.set('limit', filters.limit.toString());
        if (filters.offset) query.set('offset', filters.offset.toString());

        const res = await fetchWithAuth(`/api/v1/audit/logs?${query.toString()}`);
        if (!res.ok) throw new Error('Failed to fetch audit logs');
        return res.json();
    },

    // Credentials

    /**
     * Lists all credentials.
     */
    listCredentials: async () => {
        const res = await fetchWithAuth('/credentials');
        if (!res.ok) throw new Error('Failed to list credentials');
        return res.json();
    },

    /**
     * Creates a new credential.
     */
    createCredential: async (credential: any) => {
        const res = await fetchWithAuth('/credentials', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(credential)
        });
        if (!res.ok) throw new Error('Failed to create credential');
        return res.json();
    },

    /**
     * Gets a single credential by ID.
     */
    getCredential: async (id: string) => {
        const res = await fetchWithAuth(`/credentials/${id}`);
        if (!res.ok) throw new Error('Failed to get credential');
        return res.json();
    },

    /**
     * Updates a credential.
     */
    updateCredential: async (id: string, credential: any) => {
        const res = await fetchWithAuth(`/credentials/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(credential)
        });
        if (!res.ok) throw new Error('Failed to update credential');
        return res.json();
    },

    /**
     * Deletes a credential.
     */
    deleteCredential: async (id: string) => {
        const res = await fetchWithAuth(`/credentials/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete credential');
        return {};
    },

    /**
     * Runs the doctor check.
     */
    getDoctorStatus: async () => {
        const res = await fetchWithAuth('/doctor');
        if (!res.ok) throw new Error('Failed to run doctor check');
        return res.json();
    }
};
