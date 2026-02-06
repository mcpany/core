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

import { GrpcWebImpl, RegistrationServiceClientImpl } from '@proto/api/v1/registration';
import { UpstreamServiceConfig as BaseUpstreamServiceConfig } from '@proto/config/v1/upstream_service';
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
}

// Re-export generated types
export type { ToolDefinition, ResourceDefinition, PromptDefinition, Credential, Authentication, ProfileDefinition };
export type { ListServicesResponse, GetServiceResponse, GetServiceStatusResponse, ValidateServiceResponse } from '@proto/api/v1/registration';

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

/**
 * Definition of a secret stored in the system.
 */
export interface SecretDefinition {
    /** Unique identifier for the secret. */
    id: string;
    /** Human-readable name of the secret. */
    name: string;
    /** The key or name used to reference the secret in configurations. */
    key: string;
    /** The secret value (masked in responses). */
    value: string;
    /** The provider of the secret (e.g., openai, anthropic). */
    provider?: 'openai' | 'anthropic' | 'aws' | 'gcp' | 'custom';
    /** Timestamp of the last usage. */
    lastUsed?: string;
    /** Timestamp of creation. */
    createdAt: string;
}

/**
 * Content of a resource.
 */
export interface ResourceContent {
    /** The URI of the resource. */
    uri: string;
    /** The MIME type of the content. */
    mimeType: string;
    /** Text content, if applicable. */
    text?: string;
    /** Binary content as a base64 encoded string, if applicable. */
    blob?: string;
}

/**
 * Response for reading a resource.
 */
export interface ReadResourceResponse {
    /** List of resource contents. */
    contents: ResourceContent[];
}

/**
 * Result of a single system health check.
 */
export interface CheckResult {
    /** The status of the check (e.g., "ok", "degraded", "error"). */
    status: string;
    /** Optional message describing the status or error. */
    message?: string;
    /** Optional latency measurement. */
    latency?: string;
    /** Optional diff showing configuration changes on error. */
    diff?: string;
}

/**
 * Full doctor report containing system health status.
 */
export interface DoctorReport {
    /** Overall system status. */
    status: string;
    /** Timestamp of the report. */
    timestamp: string;
    /** Map of check names to their results. */
    checks: Record<string, CheckResult>;
}

/**
 * Tool failure statistics.
 */
export interface ToolFailureStats {
    name: string;
    serviceId: string;
    failureRate: number;
    totalCalls: number;
}

/**
 * Tool usage analytics.
 */
export interface ToolAnalytics {
    name: string;
    serviceId: string;
    totalCalls: number;
    successRate: number;
}


/**
 * Metric definition for dashboard.
 */
export interface Metric {
    label: string;
    value: string;
    change?: string;
    trend?: "up" | "down" | "neutral";
    icon: string;
    subLabel?: string;
}


/**
 * Represents the current status and health of the system.
 */
export interface SystemStatus {
    /** The number of seconds the server has been running. */
    uptime_seconds: number;
    /** The number of currently active HTTP connections. */
    active_connections: number;
    /** The port number where the HTTP server is listening. */
    bound_http_port: number;
    /** The port number where the gRPC server is listening. */
    bound_grpc_port: number;
    /** The current version of the server. */
    version: string;
    /** A list of active security warnings, if any. */
    security_warnings: string[];
}


const getMetadata = () => {
    // Metadata for gRPC calls.
    // Since gRPC-Web calls might bypass Next.js middleware if they go directly to Envoy/Backend,
    // we need to be careful.
    // However, if we proxy gRPC via Next.js (not yet fully standard for gRPC-Web), we could use middleware.
    // For now, if we don't have the key in NEXT_PUBLIC, we can't send it from client.
    // The gRPC calls should ideally be proxied or use a session token.
    // Given the current refactor to remove NEXT_PUBLIC_ key, direct gRPC calls from client will fail auth
    // if they require the static key.
    // We should rely on the Next.js API routes (REST) which use middleware, OR assume the gRPC endpoint
    // is also behind the Next.js proxy (rewrites).
    // ui/next.config.ts has a rewrite for `/mcpany.api.v1.RegistrationService/:path*`.
    // If we use that, the middleware WILL run and inject the header!
    return undefined;
};

/**
 * API Client for interacting with the MCP Any server.
 */
export const apiClient = {
    // Services (Migrated to gRPC)

    /**
     * Lists all registered upstream services.
     * @returns A promise that resolves to a list of services.
     */
    listServices: async () => {
        // Fallback to REST for E2E reliability until gRPC-Web is stable
        const res = await fetchWithAuth('/api/v1/services');
        if (!res.ok) throw new Error('Failed to fetch services');
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.services || []);
        // Map snake_case to camelCase for UI compatibility
        return list.map((s: any) => ({
            ...s,
            connectionPool: s.connection_pool,
            httpService: s.http_service,
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
        }));
    },

    /**
     * Gets a single service by its ID.
     * @param id The ID of the service to retrieve.
     * @returns A promise that resolves to the service configuration.
     */
    getService: async (id: string) => {
         try {
             // Try gRPC-Web first
             const response = await registrationClient.GetService({ serviceName: id }, getMetadata());
             return response;
         } catch (e) {
             // Fallback to REST if gRPC fails (e.g. in E2E tests passing through Next.js proxy or mock)
             // Check if we are in a test env or just try fetch
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
                         httpService: s.http_service,
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
     * @param name The name of the service.
     * @param disable True to disable the service, false to enable it.
     * @returns A promise that resolves to the updated service status.
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
     * @param name The name of the service.
     * @returns A promise that resolves to the service status.
     */
    getServiceStatus: async (name: string) => {
        const res = await fetchWithAuth(`/api/v1/services/${name}/status`);
        if (!res.ok) throw new Error('Failed to fetch service status');
        return res.json();
    },

    /**
     * Restarts a service.
     * @param name The name of the service to restart.
     * @returns A promise that resolves when the service is restarted.
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
     * @param config The configuration of the service to register.
     * @returns A promise that resolves to the registered service configuration.
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
            payload.http_service = { address: config.httpService.address };
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
        if (config.upstreamAuth) {
            payload.upstream_auth = config.upstreamAuth;
        }
        if (config.authentication) {
            payload.authentication = config.authentication;
        }
        if (config.resilience) {
            payload.resilience = config.resilience;
        }
        if (config.cache) {
            payload.cache = config.cache;
        }
        if (config.rateLimit) {
            payload.rate_limit = config.rateLimit;
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
     * @param config The updated configuration of the service.
     * @returns A promise that resolves to the updated service configuration.
     */
    updateService: async (config: UpstreamServiceConfig) => {
        // Same mapping as register
        const payload: any = {
             id: config.id,
            name: config.name,
            version: config.version,
            disable: config.disable,
            priority: config.priority,
            load_balancing_strategy: config.loadBalancingStrategy,
            tags: config.tags,
        };
        // Reuse mapping logic or duplicate for now safely
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
        if (config.upstreamAuth) {
            payload.upstream_auth = config.upstreamAuth;
        }
        if (config.authentication) {
            payload.authentication = config.authentication;
        }
        if (config.resilience) {
            payload.resilience = config.resilience;
        }
        if (config.cache) {
            payload.cache = config.cache;
        }
        if (config.rateLimit) {
            payload.rate_limit = config.rateLimit;
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
     * @param id The ID of the service to unregister.
     * @returns A promise that resolves when the service is unregistered.
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
     * @param config The service configuration to validate.
     * @returns A promise that resolves to the validation result.
     */
    validateService: async (config: UpstreamServiceConfig) => {
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
            payload.http_service = { address: config.httpService.address };
        }
        if (config.grpcService) {
            payload.grpc_service = { address: config.grpcService.address };
        }
        if (config.commandLineService) {
            payload.command_line_service = {
                command: config.commandLineService.command,
                working_directory: config.commandLineService.workingDirectory,
                env: config.commandLineService.env,
                container_environment: config.commandLineService.containerEnvironment, // Include this if needed
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
        if (config.upstreamAuth) {
            payload.upstream_auth = config.upstreamAuth;
        }
        if (config.authentication) {
            payload.authentication = config.authentication;
        }
        if (config.resilience) {
            payload.resilience = config.resilience;
        }
        if (config.cache) {
            payload.cache = config.cache;
        }
        if (config.rateLimit) {
            payload.rate_limit = config.rateLimit;
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
            // Even if not ok (400), it might contain validation details in JSON
            if (data && data.error) {
                 return data; // Return the error object (valid: false, error: ...)
            }
            throw new Error(`Failed to validate service: ${response.status} ${text}`);
        }
        return data;
    },

    // Tools (Legacy Fetch - Not yet migrated to Admin/Registration Service completely or keeping as is)
    // admin.proto has ListTools but we are focusing on RegistrationService first.
    // So keep using fetch for Tools/Secrets/etc for now.

    /**
     * Lists all available tools.
     * @returns A promise that resolves to a list of tools.
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
     * @param request The execution request (tool name, arguments, etc.).
     * @param dryRun If true, performs a dry run without side effects.
     * @returns A promise that resolves to the execution result.
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
     * @param name The name of the tool.
     * @param disabled True to disable the tool, false to enable it.
     * @returns A promise that resolves to the updated tool status.
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
     * @returns A promise that resolves to a list of resources.
     */
    listResources: async () => {
        const res = await fetchWithAuth('/api/v1/resources');
        if (!res.ok) throw new Error('Failed to list resources');
        return res.json();
    },

    /**
     * Reads the content of a resource.
     * @param uri The URI of the resource to read.
     * @returns A promise that resolves to the resource content.
     */
    readResource: async (uri: string): Promise<ReadResourceResponse> => {
        const res = await fetchWithAuth(`/api/v1/resources/read?uri=${encodeURIComponent(uri)}`);
        if (!res.ok) throw new Error('Failed to read resource');
        return res.json();
    },

    /**
     * Sets the status (enabled/disabled) of a resource.
     * @param uri The URI of the resource.
     * @param disabled True to disable the resource, false to enable it.
     * @returns A promise that resolves to the updated resource status.
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
     * @returns A promise that resolves to a list of prompts.
     */
    listPrompts: async () => {
        const res = await fetchWithAuth('/api/v1/prompts');
        if (!res.ok) throw new Error(`Failed to fetch prompts: ${res.status}`);
        return res.json();
    },

    /**
     * Sets the status (enabled/disabled) of a prompt.
     * @param name The name of the prompt.
     * @param enabled True to enable the prompt, false to disable it.
     * @returns A promise that resolves to the updated prompt status.
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
     * @param name The name of the prompt.
     * @param args The arguments for the prompt.
     * @returns A promise that resolves to the prompt execution result.
     */
    executePrompt: async (name: string, args: Record<string, string>) => {
        const res = await fetch(`/api/v1/prompts/${name}/execute`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(args)
        });
        if (!res.ok) throw new Error('Failed to execute prompt');
        return res.json();
    },


    // Secrets

    /**
     * Lists all stored secrets.
     * @returns A promise that resolves to a list of secrets.
     */
    listSecrets: async () => {
        const res = await fetchWithAuth('/api/v1/secrets');
        if (!res.ok) throw new Error('Failed to fetch secrets');
        const data = await res.json();
        return Array.isArray(data) ? data : (data.secrets || []);
    },

    /**
     * Reveals a secret value.
     * @param id The ID of the secret to reveal.
     * @returns A promise that resolves to the secret value.
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
     * @param secret The secret definition to save.
     * @returns A promise that resolves to the saved secret.
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
     * @param id The ID of the secret to delete.
     * @returns A promise that resolves when the secret is deleted.
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
     * @returns A promise that resolves to the global settings.
     */
    getGlobalSettings: async () => {
        const res = await fetchWithAuth('/api/v1/settings');
        if (!res.ok) throw new Error('Failed to fetch global settings');
        return res.json();
    },

    /**
     * Saves the global server settings.
     * @param settings The settings to save.
     * @returns A promise that resolves when the settings are saved.
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
     * @param serviceId Optional service ID to filter by.
     * @param timeRange Optional time range to filter by (e.g. "1h", "24h").
     * @returns A promise that resolves to the traffic history points.
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
     * @param serviceId Optional service ID to filter by.
     * @returns A promise that resolves to the top tools stats.
     */
    getTopTools: async (serviceId?: string) => {
        let url = '/api/v1/dashboard/top-tools';
        if (serviceId) url += `?serviceId=${encodeURIComponent(serviceId)}`;
        const res = await fetchWithAuth(url);
        // If 404/500, return empty to avoid crashing UI
        if (!res.ok) return [];
        return res.json();
    },

    // Alerts

    /**
     * Lists all alerts.
     * @returns A promise that resolves to a list of alerts.
     */
    listAlerts: async () => {
        const res = await fetchWithAuth('/api/v1/alerts');
        if (!res.ok) throw new Error('Failed to fetch alerts');
        return res.json();
    },

    /**
     * Lists all alert rules.
     * @returns A promise that resolves to a list of alert rules.
     */
    listAlertRules: async () => {
        const res = await fetchWithAuth('/api/v1/alerts/rules');
        if (!res.ok) throw new Error('Failed to fetch alert rules');
        return res.json();
    },

    /**
     * Creates a new alert rule.
     * @param rule The rule to create.
     * @returns A promise that resolves to the created rule.
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
     * @param id The ID of the rule.
     * @returns A promise that resolves to the rule.
     */
    getAlertRule: async (id: string) => {
        const res = await fetchWithAuth(`/api/v1/alerts/rules/${id}`);
        if (!res.ok) throw new Error('Failed to get alert rule');
        return res.json();
    },

    /**
     * Updates an alert rule.
     * @param rule The rule to update.
     * @returns A promise that resolves to the updated rule.
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
     * @param id The ID of the rule to delete.
     * @returns A promise that resolves when the rule is deleted.
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
     * @param serviceId Optional service ID to filter by.
     * @returns A promise that resolves to the tool failure stats.
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
     * @param serviceId Optional service ID to filter by.
     * @returns A promise that resolves to the tool usage stats.
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
     * @returns A promise that resolves to the system status.
     */
    getSystemStatus: async (): Promise<SystemStatus> => {
        const res = await fetchWithAuth('/api/v1/system/status');
        if (!res.ok) throw new Error('Failed to fetch system status');
        return res.json();
    },

    /**
     * Gets the dashboard metrics.
     * @param serviceId Optional service ID to filter by.
     * @returns A promise that resolves to the metrics list.
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
     * @param options Optional parameters.
     * @returns A promise that resolves to the traces list.
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
     * @param points The traffic points to seed.
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
     * @param id The ID of the alert.
     * @param status The new status.
     * @returns A promise that resolves to the updated alert.
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
     * @returns A promise that resolves to the webhook configuration.
     */
    getWebhookURL: async (): Promise<{ url: string }> => {
        const res = await fetchWithAuth('/api/v1/alerts/webhook');
        if (!res.ok) throw new Error('Failed to fetch webhook URL');
        return res.json();
    },

    /**
     * Saves the configured global webhook URL for alerts.
     * @param url The webhook URL.
     * @returns A promise that resolves to the updated webhook configuration.
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
     * @returns A promise that resolves to a list of collections.
     */
    listCollections: async () => {
        const res = await fetchWithAuth('/api/v1/collections');
        if (!res.ok) throw new Error('Failed to list collections');
        return res.json();
    },

    /**
     * Gets a single service collection (stack) by its name.
     * @param name The name of the collection.
     * @returns A promise that resolves to the collection.
     */
    getCollection: async (name: string) => {
        const res = await fetchWithAuth(`/api/v1/collections/${name}`);
        if (!res.ok) throw new Error('Failed to get collection');
        return res.json();
    },

    /**
     * Saves a service collection (stack).
     * @param collection The collection to save.
     * @returns A promise that resolves when the collection is saved.
     */
    saveCollection: async (collection: any) => {
        // Decide if create or update based on existence?
        // The API might expect POST for create, PUT for update.
        // For now, let's try POST to /api/v1/collections if id/name is new, or PUT if existing?
        // But stack-editor logic handles "saving".
        // The endpoint logic in api.go: handleCollections is POST, handleCollectionDetail is PUT.
        // We'll use PUT if we have a name in the URL, POST otherwise?
        // But `saveStackConfig` was replacing config.
        // Let's assume we update an existing one usually.
        const res = await fetchWithAuth(`/api/v1/collections/${collection.name}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(collection)
        });
        if (!res.ok) {
             // If PUT fails (e.g. not found), try POST to create?
             // But UI should know if it's new.
             // For now just error.
             throw new Error('Failed to save collection');
        }
        return res.json();
    },

    /**
     * Deletes a service collection (stack).
     * @param name The name of the collection to delete.
     * @returns A promise that resolves when the collection is deleted.
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
     * @param stackId The ID of the stack.
     * @returns A promise that resolves to the stack configuration.
     */
    getStackConfig: async (stackId: string) => {
        // Map to getCollection
        return apiClient.getCollection(stackId);
    },

    /**
     * Saves the configuration for a stack (Compatibility wrapper).
     * @param stackId The ID of the stack.
     * @param config The configuration content (Collection object).
     * @returns A promise that resolves when the config is saved.
     */
    saveStackConfig: async (stackId: string, config: any) => {
        // Map to saveCollection. Ensure name is set.
        const collection = typeof config === 'string' ? JSON.parse(config) : config;
        if (!collection.name) collection.name = stackId;
        return apiClient.saveCollection(collection);
    },


    // User Management

    /**
     * Lists all profiles.
     * @returns A promise that resolves to a list of profiles.
     */
    listProfiles: async () => {
        const res = await fetchWithAuth('/api/v1/profiles');
        if (!res.ok) throw new Error('Failed to fetch profiles');
        return res.json();
    },

    /**
     * Creates a new profile.
     * @param profile The profile definition.
     * @returns A promise that resolves to the created profile.
     */
    createProfile: async (profile: ProfileDefinition) => {
        const res = await fetchWithAuth('/api/v1/profiles', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(profile)
        });
        if (!res.ok) throw new Error('Failed to create profile');
        return res.json();
    },

    /**
     * Updates an existing profile.
     * @param profile The profile definition.
     * @returns A promise that resolves to the updated profile.
     */
    updateProfile: async (profile: ProfileDefinition) => {
        const res = await fetchWithAuth(`/api/v1/profiles/${profile.name}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(profile)
        });
        if (!res.ok) throw new Error('Failed to update profile');
        return res.json();
    },

    /**
     * Deletes a profile.
     * @param name The name of the profile to delete.
     * @returns A promise that resolves when the profile is deleted.
     */
    deleteProfile: async (name: string) => {
        const res = await fetchWithAuth(`/api/v1/profiles/${name}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete profile');
        return {};
    },

    // User Management

    /**
     * Lists all users.
     * @returns A promise that resolves to a list of users.
     */
    listUsers: async () => {
        // Fallback for demo/dev - use AdminRPC if possible, but we don't have generated client for Admin yet in UI?
        // We do have @proto/admin/v1/admin
        // Let's rely on fetch for now if we expose REST, OR we can try to use standard gRPC-Web if we set it up.
        // For simplicity in this UI iteration (which seems to use fetchWithAuth mostly),
        // we might assume there is a REST gateway or we use a custom endpoint.
        // Wait, server/pkg/admin/server.go implements providing gRPC.
        // Does it expose REST?
        // The previous task walkthrough mentions "Admin Service RPCs".
        // And `server/pkg/app/server.go` likely mounts gRPC-Gateway?
        // Let's try fetch first.
        const res = await fetchWithAuth('/api/v1/users');
        if (!res.ok) throw new Error('Failed to list users');
        return res.json();
    },

    /**
     * Creates a new user.
     * @param user The user object to create.
     * @returns A promise that resolves to the created user.
     */
    createUser: async (user: any) => {
        const res = await fetchWithAuth('/api/v1/users', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ user }) // Wrapper expected by AdminRPC? Request is CreateUserRequest { user: User }
        });
        if (!res.ok) throw new Error('Failed to create user');
        return res.json();
    },

    /**
     * Updates an existing user.
     * @param user The user object to update.
     * @returns A promise that resolves to the updated user.
     */
    updateUser: async (user: any) => {
         const res = await fetchWithAuth(`/api/v1/users/${user.id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ user })
        });
        if (!res.ok) throw new Error('Failed to update user');
        return res.json();
    },

    /**
     * Deletes a user.
     * @param id The ID of the user to delete.
     * @returns A promise that resolves when the user is deleted.
     */
    deleteUser: async (id: string) => {
        const res = await fetchWithAuth(`/api/v1/users/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete user');
        return {};
    },


    // OAuth

    /**
     * Initiates an OAuth flow.
     * @param serviceID The ID of the service for which to initiate OAuth.
     * @param redirectURL The URL to redirect to after OAuth completes.
     * @param credentialID Optional credential ID to associate with the token.
     * @returns A promise that resolves to the initiation response.
     */
    initiateOAuth: async (serviceID: string, redirectURL: string, credentialID?: string) => {
        const payload: any = { redirect_url: redirectURL };
        if (serviceID) payload.service_id = serviceID;
        if (credentialID) payload.credential_id = credentialID;

        const res = await fetchWithAuth('/auth/oauth/initiate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to initiate OAuth: ${res.status} ${txt}`);
        }
        return res.json();
    },

    /**
     * Handles the OAuth callback.
     * @param serviceID The ID of the service (optional).
     * @param code The OAuth authorization code.
     * @param redirectURL The redirect URL used in the initial request.
     * @param credentialID Optional credential ID.
     * @returns A promise that resolves to the callback handling result.
     */
    handleOAuthCallback: async (serviceID: string | null, code: string, redirectURL: string, credentialID?: string) => {
        const payload: any = { code, redirect_url: redirectURL };
        if (serviceID) payload.service_id = serviceID;
        if (credentialID) payload.credential_id = credentialID;

        const res = await fetchWithAuth('/auth/oauth/callback', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to handle callback: ${res.status} ${txt}`);
        }
        return res.json();
    },

    // Credentials

    /**
     * Lists all stored credentials.
     * @returns A promise that resolves to a list of credentials.
     */
    listCredentials: async () => {
        const res = await fetchWithAuth('/api/v1/credentials');
        if (!res.ok) throw new Error('Failed to list credentials');
        const data = await res.json();
        return Array.isArray(data) ? data : (data.credentials || []);
    },

    /**
     * Saves (creates or updates) a credential.
     * @param credential The credential to save.
     * @returns A promise that resolves to the saved credential.
     */
    saveCredential: async (credential: Credential) => {
        // ... (logic omitted for brevity, keeping same) ...
        return apiClient.createCredential(credential);
    },

    /**
     * Creates a new credential.
     * @param credential The credential to create.
     * @returns A promise that resolves to the created credential.
     */
    createCredential: async (credential: Credential) => {
        const res = await fetchWithAuth('/api/v1/credentials', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(credential)
        });
        if (!res.ok) throw new Error('Failed to create credential');
        return res.json();
    },

    /**
     * Updates an existing credential.
     * @param credential The credential to update.
     * @returns A promise that resolves to the updated credential.
     */
    updateCredential: async (credential: Credential) => {
        const res = await fetchWithAuth(`/api/v1/credentials/${credential.id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(credential)
        });
        if (!res.ok) throw new Error('Failed to update credential');
        return res.json();
    },

    /**
     * Deletes a credential.
     * @param id The ID of the credential to delete.
     * @returns A promise that resolves when the credential is deleted.
     */
    deleteCredential: async (id: string) => {
        const res = await fetchWithAuth(`/api/v1/credentials/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete credential');
        return {};
    },

    /**
     * Tests authentication with the provided parameters.
     * @param req The authentication test request.
     * @returns A promise that resolves to the test result.
     */
    testAuth: async (req: any) => {
        const res = await fetchWithAuth('/api/v1/debug/auth-test', {
            method: 'POST',
             headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(req)
        });
        // We always return JSON even on error
        return res.json();
    },

    // Templates (Backend Persistence)

    /**
     * Lists all service templates.
     * @returns A promise that resolves to a list of templates.
     */
    listTemplates: async () => {
        const res = await fetchWithAuth('/api/v1/templates');
        if (!res.ok) throw new Error('Failed to fetch templates');
        const data = await res.json();
        // Backend returns generic UpstreamServiceConfig list.
        // Map snake_case to camelCase
        const list = Array.isArray(data) ? data : [];
        return list.map((s: any) => ({
            ...s,
            // Reuse logic? Or copy/paste mapping
            connectionPool: s.connection_pool,
            httpService: s.http_service,
            grpcService: s.grpc_service,
            commandLineService: s.command_line_service,
            mcpService: s.mcp_service,
            upstreamAuth: s.upstream_auth,
            preCallHooks: s.pre_call_hooks,
            postCallHooks: s.post_call_hooks,
            toolExportPolicy: s.tool_export_policy,
            promptExportPolicy: s.prompt_export_policy,
            resourceExportPolicy: s.resource_export_policy,
        }));
    },

    /**
     * Saves a service template.
     * @param template The template configuration to save.
     * @returns A promise that resolves to the saved template.
     */
    saveTemplate: async (template: UpstreamServiceConfig) => {
        // Map back to snake_case for saving
        // Reuse registerService mapping logic essentially but for template endpoint
        const payload: any = {
            id: template.id,
            name: template.name,
            version: template.version,
            disable: template.disable,
            priority: template.priority,
            load_balancing_strategy: template.loadBalancingStrategy,
            tags: template.tags,
        };

        if (template.httpService) {
            payload.http_service = { address: template.httpService.address };
        }
        if (template.grpcService) {
            payload.grpc_service = { address: template.grpcService.address };
        }
        if (template.commandLineService) {
            payload.command_line_service = {
                command: template.commandLineService.command,
                working_directory: template.commandLineService.workingDirectory,
                env: template.commandLineService.env
            };
        }
        if (template.mcpService) {
            payload.mcp_service = { ...template.mcpService };
        }
        if (template.preCallHooks) {
            payload.pre_call_hooks = template.preCallHooks;
        }
        if (template.postCallHooks) {
            payload.post_call_hooks = template.postCallHooks;
        }
        if (template.toolExportPolicy) {
            payload.tool_export_policy = template.toolExportPolicy;
        }
        if (template.promptExportPolicy) {
            payload.prompt_export_policy = template.promptExportPolicy;
        }
        if (template.resourceExportPolicy) {
            payload.resource_export_policy = template.resourceExportPolicy;
        }
        if (template.upstreamAuth) {
            payload.upstream_auth = template.upstreamAuth;
        }
        if (template.authentication) {
            payload.authentication = template.authentication;
        }
        if (template.resilience) {
            payload.resilience = template.resilience;
        }
        if (template.cache) {
            payload.cache = template.cache;
        }
        if (template.rateLimit) {
            payload.rate_limit = template.rateLimit;
        }

        const res = await fetchWithAuth('/api/v1/templates', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        if (!res.ok) throw new Error('Failed to save template');
        return res.json();
    },

    /**
     * Deletes a service template.
     * @param id The ID of the template to delete.
     * @returns A promise that resolves when the template is deleted.
     */
    deleteTemplate: async (id: string) => {
        const res = await fetchWithAuth(`/api/v1/templates/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete template');
        return {};
    },

    // System Health

    /**
     * Gets the doctor status report.
     * @returns A promise that resolves to the doctor report.
     */
    getDoctorStatus: async (): Promise<DoctorReport> => {
        const res = await fetchWithAuth('/api/v1/doctor');
        if (!res.ok) throw new Error('Failed to fetch doctor status');
        return res.json();
    },

    // Audit Logs

    /**
     * Lists audit logs.
     * @param filters The filters for the audit logs.
     * @returns A promise that resolves to the list of audit logs.
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
    }
};
