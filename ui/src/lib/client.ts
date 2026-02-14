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
    return undefined;
};

/**
 * API Client for interacting with the MCP Any server.
 */
export const apiClient = {
    // Services (Migrated to gRPC)

    /**
     * Lists all registered upstream services.
     *
     * Summary: Fetches the list of all configured upstream services from the backend.
     *
     * @returns {Promise<any[]>} A promise that resolves to a list of services.
     * @throws {Error} If the network request fails or the response is not OK.
     *
     * Side Effects: Makes a GET request to /api/v1/services.
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
        }));
    },

    /**
     * Lists services from the dynamic catalog.
     *
     * Summary: Fetches the list of available services from the dynamic catalog.
     *
     * @returns {Promise<any[]>} A promise that resolves to a list of catalog services.
     * @throws {Error} If the network request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/catalog/services.
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
     *
     * Summary: Retrieves the configuration details for a specific upstream service.
     *
     * @param {string} id - The ID of the service to retrieve.
     * @returns {Promise<any>} A promise that resolves to the service configuration.
     * @throws {Error} If the service is not found or the request fails.
     *
     * Side Effects: Makes a gRPC call or GET request to /api/v1/services/:id.
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
     *
     * Summary: Enables or disables a specific service.
     *
     * @param {string} name - The name of the service.
     * @param {boolean} disable - True to disable the service, false to enable it.
     * @returns {Promise<any>} A promise that resolves to the updated service status.
     * @throws {Error} If the update fails.
     *
     * Side Effects: Makes a PUT request to /api/v1/services/:name.
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
     *
     * Summary: Retrieves the current running status of a service.
     *
     * @param {string} name - The name of the service.
     * @returns {Promise<any>} A promise that resolves to the service status.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/services/:name/status.
     */
    getServiceStatus: async (name: string) => {
        const res = await fetchWithAuth(`/api/v1/services/${name}/status`);
        if (!res.ok) throw new Error('Failed to fetch service status');
        return res.json();
    },

    /**
     * Restarts a service.
     *
     * Summary: Triggers a restart of the specified service.
     *
     * @param {string} name - The name of the service to restart.
     * @returns {Promise<any>} A promise that resolves when the service is restarted.
     * @throws {Error} If the restart fails.
     *
     * Side Effects: Makes a POST request to /api/v1/services/:name/restart.
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
     *
     * Summary: Registers a new upstream service with the provided configuration.
     *
     * @param {UpstreamServiceConfig} config - The configuration of the service to register.
     * @returns {Promise<any>} A promise that resolves to the registered service configuration.
     * @throws {Error} If the registration fails (e.g., validation error, duplicate ID).
     *
     * Side Effects: Makes a POST request to /api/v1/services.
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
     *
     * Summary: Updates the configuration of an existing upstream service.
     *
     * @param {UpstreamServiceConfig} config - The updated configuration of the service.
     * @returns {Promise<any>} A promise that resolves to the updated service configuration.
     * @throws {Error} If the update fails.
     *
     * Side Effects: Makes a PUT request to /api/v1/services/:name.
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
     *
     * Summary: Deletes the specified upstream service configuration.
     *
     * @param {string} id - The ID of the service to unregister.
     * @returns {Promise<any>} A promise that resolves when the service is unregistered.
     * @throws {Error} If the deletion fails.
     *
     * Side Effects: Makes a DELETE request to /api/v1/services/:id.
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
     *
     * Summary: Checks if the provided service configuration is valid without persisting it.
     *
     * @param {UpstreamServiceConfig} config - The service configuration to validate.
     * @returns {Promise<any>} A promise that resolves to the validation result.
     * @throws {Error} If the validation request fails.
     *
     * Side Effects: Makes a POST request to /api/v1/services/validate.
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

    // Tools

    /**
     * Lists all available tools.
     *
     * Summary: Fetches a list of all tools exposed by registered services.
     *
     * @returns {Promise<any>} A promise that resolves to a list of tools.
     * @throws {Error} If the network request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/tools.
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
     *
     * Summary: Triggers the execution of a specific tool.
     *
     * @param {any} request - The execution request object (containing tool name, arguments, etc.).
     * @param {boolean} [dryRun] - If true, performs a dry run without side effects.
     * @returns {Promise<any>} A promise that resolves to the execution result.
     * @throws {Error} If the tool execution fails.
     *
     * Side Effects: Makes a POST request to /api/v1/execute.
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
     *
     * Summary: Enables or disables a specific tool.
     *
     * @param {string} name - The name of the tool.
     * @param {boolean} disabled - True to disable the tool, false to enable it.
     * @returns {Promise<any>} A promise that resolves to the updated tool status.
     * @throws {Error} If the update fails.
     *
     * Side Effects: Makes a PUT request to /api/v1/tools.
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
     *
     * Summary: Fetches a list of all resources exposed by registered services.
     *
     * @returns {Promise<any>} A promise that resolves to a list of resources.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/resources.
     */
    listResources: async () => {
        const res = await fetchWithAuth('/api/v1/resources');
        if (!res.ok) throw new Error('Failed to list resources');
        return res.json();
    },

    /**
     * Reads the content of a resource.
     *
     * Summary: Retrieves the content of a specific resource.
     *
     * @param {string} uri - The URI of the resource to read.
     * @returns {Promise<ReadResourceResponse>} A promise that resolves to the resource content.
     * @throws {Error} If the read operation fails.
     *
     * Side Effects: Makes a GET request to /api/v1/resources/read.
     */
    readResource: async (uri: string): Promise<ReadResourceResponse> => {
        const res = await fetchWithAuth(`/api/v1/resources/read?uri=${encodeURIComponent(uri)}`);
        if (!res.ok) throw new Error('Failed to read resource');
        return res.json();
    },

    /**
     * Sets the status (enabled/disabled) of a resource.
     *
     * Summary: Enables or disables a specific resource.
     *
     * @param {string} uri - The URI of the resource.
     * @param {boolean} disabled - True to disable the resource, false to enable it.
     * @returns {Promise<any>} A promise that resolves to the updated resource status.
     * @throws {Error} If the update fails.
     *
     * Side Effects: Makes a PUT request to /api/v1/resources.
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
     *
     * Summary: Fetches a list of all prompts exposed by registered services.
     *
     * @returns {Promise<any>} A promise that resolves to a list of prompts.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/prompts.
     */
    listPrompts: async () => {
        const res = await fetchWithAuth('/api/v1/prompts');
        if (!res.ok) throw new Error(`Failed to fetch prompts: ${res.status}`);
        return res.json();
    },

    /**
     * Sets the status (enabled/disabled) of a prompt.
     *
     * Summary: Enables or disables a specific prompt.
     *
     * @param {string} name - The name of the prompt.
     * @param {boolean} enabled - True to enable the prompt, false to disable it.
     * @returns {Promise<any>} A promise that resolves to the updated prompt status.
     * @throws {Error} If the update fails.
     *
     * Side Effects: Makes a POST request to /api/v1/prompts.
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
     *
     * Summary: Runs a specific prompt template with arguments.
     *
     * @param {string} name - The name of the prompt.
     * @param {Record<string, string>} args - The arguments for the prompt.
     * @returns {Promise<any>} A promise that resolves to the prompt execution result.
     * @throws {Error} If the execution fails.
     *
     * Side Effects: Makes a POST request to /api/v1/prompts/execute.
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
     *
     * Summary: Provides pre-defined service templates for the setup wizard.
     *
     * @returns {Promise<any[]>} A promise that resolves to a list of service templates.
     *
     * Side Effects: None (currently mocks data, future implementations may fetch from backend).
     */
    getServiceTemplates: async () => {
        // Mock data mirroring server/examples/popular_services
        // In a real implementation, this should come from an endpoint like /api/v1/templates/services
        return [
            {
                id: "google-calendar",
                name: "Google Calendar",
                description: "Manage events and calendars.",
                icon: "calendar",
                tags: ["google", "productivity"],
                authType: "oauth2",
                serviceConfig: {
                    name: "google_calendar",
                    upstreamAuth: {
                        oauth2: {
                            tokenUrl: "https://oauth2.googleapis.com/token",
                            clientId: { plainText: "" }, // User must provide or we use env vars if set?
                            clientSecret: { plainText: "" },
                            scopes: "https://www.googleapis.com/auth/calendar"
                        }
                    },
                    openapiService: {
                        specUrl: "https://api.apis.guru/v2/specs/googleapis.com/calendar/v3/openapi.yaml"
                    }
                }
            },
            {
                id: "github",
                name: "GitHub",
                description: "Interact with repositories, issues, and PRs.",
                icon: "github",
                tags: ["dev", "git"],
                authType: "token", // Can also be oauth2 but token is easier for wizard?
                serviceConfig: {
                    name: "github",
                    upstreamAuth: {
                        bearerToken: { token: { plainText: "" } }
                    },
                    openapiService: {
                        address: "https://api.github.com",
                        specUrl: "https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.yaml"
                    }
                }
            },
            {
                id: "linear",
                name: "Linear",
                description: "Issue tracking and project management.",
                icon: "linear",
                tags: ["dev", "pm"],
                authType: "oauth2", // or api key
                serviceConfig: {
                    name: "linear",
                    upstreamAuth: {
                        apiKey: { plainText: "" } // Usually API key for simplicity
                    },
                    openapiService: {
                        // Placeholder
                        specUrl: "https://raw.githubusercontent.com/linear/linear/master/api/openapi.yaml"
                    }
                }
            }
        ];
    },

    /**
     * Initiates an OAuth flow for a specific service.
     *
     * Summary: Starts the OAuth authorization process.
     *
     * @param {string} serviceId - The ID of the service (e.g. "google_calendar").
     * @param {string} credentialId - The ID of the credential to bind (usually same as service name).
     * @param {string} redirectUrl - The URL to redirect back to after auth.
     * @returns {Promise<any>} A promise that resolves to the authorization URL info.
     * @throws {Error} If the initiation fails.
     *
     * Side Effects: Makes a POST request to /api/v1/auth/oauth/initiate.
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
        return res.json(); // { authorization_url: "...", state: "..." }
    },

    // Profiles

    /**
     * Creates a new profile.
     *
     * Summary: Adds a new profile configuration to the system.
     *
     * @param {any} profileData - The profile configuration data.
     * @returns {Promise<any>} A promise that resolves to the created profile.
     * @throws {Error} If the creation fails.
     *
     * Side Effects: Makes a POST request to /api/v1/profiles.
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
     *
     * Summary: Modifies the configuration of an existing profile.
     *
     * @param {any} profileData - The updated profile configuration data.
     * @returns {Promise<any>} A promise that resolves to the updated profile.
     * @throws {Error} If the update fails.
     *
     * Side Effects: Makes a PUT request to /api/v1/profiles/:name.
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
     *
     * Summary: Removes a profile from the system.
     *
     * @param {string} name - The name of the profile to delete.
     * @returns {Promise<any>} A promise that resolves when the profile is deleted.
     * @throws {Error} If the deletion fails.
     *
     * Side Effects: Makes a DELETE request to /api/v1/profiles/:name.
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
     *
     * Summary: Fetches a list of all defined profiles.
     *
     * @returns {Promise<any[]>} A promise that resolves to a list of profiles.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/profiles.
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
     *
     * Summary: Fetches a list of all secrets stored in the system.
     *
     * @returns {Promise<any[]>} A promise that resolves to a list of secrets.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/secrets.
     */
    listSecrets: async () => {
        const res = await fetchWithAuth('/api/v1/secrets');
        if (!res.ok) throw new Error('Failed to fetch secrets');
        const data = await res.json();
        return Array.isArray(data) ? data : (data.secrets || []);
    },

    /**
     * Reveals a secret value.
     *
     * Summary: Retrieves the unmasked value of a secret.
     *
     * @param {string} id - The ID of the secret to reveal.
     * @returns {Promise<{value: string}>} A promise that resolves to an object containing the secret value.
     * @throws {Error} If the reveal operation fails.
     *
     * Side Effects: Makes a POST request to /api/v1/secrets/:id/reveal.
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
     *
     * Summary: Creates or updates a secret in the system.
     *
     * @param {SecretDefinition} secret - The secret definition to save.
     * @returns {Promise<any>} A promise that resolves to the saved secret.
     * @throws {Error} If the save operation fails.
     *
     * Side Effects: Makes a POST request to /api/v1/secrets.
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
     *
     * Summary: Removes a secret from the system.
     *
     * @param {string} id - The ID of the secret to delete.
     * @returns {Promise<any>} A promise that resolves when the secret is deleted.
     * @throws {Error} If the deletion fails.
     *
     * Side Effects: Makes a DELETE request to /api/v1/secrets/:id.
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
     *
     * Summary: Retrieves the global configuration settings for the server.
     *
     * @returns {Promise<any>} A promise that resolves to the global settings.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/settings.
     */
    getGlobalSettings: async () => {
        const res = await fetchWithAuth('/api/v1/settings');
        if (!res.ok) throw new Error('Failed to fetch global settings');
        return res.json();
    },

    /**
     * Saves the global server settings.
     *
     * Summary: Updates the global configuration settings for the server.
     *
     * @param {any} settings - The settings object to save.
     * @returns {Promise<void>} A promise that resolves when the settings are saved.
     * @throws {Error} If the update fails.
     *
     * Side Effects: Makes a POST request to /api/v1/settings.
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
     *
     * Summary: Retrieves traffic metrics for the dashboard.
     *
     * @param {string} [serviceId] - Optional service ID to filter by.
     * @param {string} [timeRange] - Optional time range to filter by (e.g. "1h", "24h").
     * @returns {Promise<any>} A promise that resolves to the traffic history points.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/dashboard/traffic.
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
     *
     * Summary: Retrieves statistics for the most frequently used tools.
     *
     * @param {string} [serviceId] - Optional service ID to filter by.
     * @returns {Promise<any[]>} A promise that resolves to the top tools stats.
     *
     * Side Effects: Makes a GET request to /api/v1/dashboard/top-tools.
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
     *
     * Summary: Fetches a list of all active alerts.
     *
     * @returns {Promise<any[]>} A promise that resolves to a list of alerts.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/alerts.
     */
    listAlerts: async () => {
        const res = await fetchWithAuth('/api/v1/alerts');
        if (!res.ok) throw new Error('Failed to fetch alerts');
        return res.json();
    },

    /**
     * Lists all alert rules.
     *
     * Summary: Fetches a list of all defined alert rules.
     *
     * @returns {Promise<any[]>} A promise that resolves to a list of alert rules.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/alerts/rules.
     */
    listAlertRules: async () => {
        const res = await fetchWithAuth('/api/v1/alerts/rules');
        if (!res.ok) throw new Error('Failed to fetch alert rules');
        return res.json();
    },

    /**
     * Creates a new alert rule.
     *
     * Summary: Defines a new rule for triggering alerts.
     *
     * @param {any} rule - The rule configuration.
     * @returns {Promise<any>} A promise that resolves to the created rule.
     * @throws {Error} If the creation fails.
     *
     * Side Effects: Makes a POST request to /api/v1/alerts/rules.
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
     *
     * Summary: Retrieves the configuration of a specific alert rule.
     *
     * @param {string} id - The ID of the rule.
     * @returns {Promise<any>} A promise that resolves to the rule configuration.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/alerts/rules/:id.
     */
    getAlertRule: async (id: string) => {
        const res = await fetchWithAuth(`/api/v1/alerts/rules/${id}`);
        if (!res.ok) throw new Error('Failed to get alert rule');
        return res.json();
    },

    /**
     * Updates an alert rule.
     *
     * Summary: Modifies an existing alert rule.
     *
     * @param {any} rule - The updated rule configuration (must include id).
     * @returns {Promise<any>} A promise that resolves to the updated rule.
     * @throws {Error} If the update fails.
     *
     * Side Effects: Makes a PUT request to /api/v1/alerts/rules/:id.
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
     *
     * Summary: Removes an alert rule from the system.
     *
     * @param {string} id - The ID of the rule to delete.
     * @returns {Promise<any>} A promise that resolves when the rule is deleted.
     * @throws {Error} If the deletion fails.
     *
     * Side Effects: Makes a DELETE request to /api/v1/alerts/rules/:id.
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
     *
     * Summary: Retrieves statistics for tools causing errors.
     *
     * @param {string} [serviceId] - Optional service ID to filter by.
     * @returns {Promise<ToolFailureStats[]>} A promise that resolves to the tool failure stats.
     *
     * Side Effects: Makes a GET request to /api/v1/dashboard/tool-failures.
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
     *
     * Summary: Retrieves usage statistics for tools.
     *
     * @param {string} [serviceId] - Optional service ID to filter by.
     * @returns {Promise<ToolAnalytics[]>} A promise that resolves to the tool usage stats.
     *
     * Side Effects: Makes a GET request to /api/v1/dashboard/tool-usage.
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
     *
     * Summary: Retrieves the current health and version of the system.
     *
     * @returns {Promise<SystemStatus>} A promise that resolves to the system status.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/system/status.
     */
    getSystemStatus: async (): Promise<SystemStatus> => {
        const res = await fetchWithAuth('/api/v1/system/status');
        if (!res.ok) throw new Error('Failed to fetch system status');
        return res.json();
    },

    /**
     * Gets the dashboard metrics.
     *
     * Summary: Retrieves high-level metrics for the dashboard overview.
     *
     * @param {string} [serviceId] - Optional service ID to filter by.
     * @returns {Promise<Metric[]>} A promise that resolves to the metrics list.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/dashboard/metrics.
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
     *
     * Summary: Retrieves a list of recent tool execution traces.
     *
     * @param {Object} [options] - Optional parameters.
     * @param {number} [options.limit] - Limit the number of traces.
     * @returns {Promise<any[]>} A promise that resolves to the traces list.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/traces.
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
     *
     * Summary: Injects fake traffic data for testing purposes.
     *
     * @param {any[]} points - The traffic points to seed.
     * @returns {Promise<void>} A promise that resolves when seeding is complete.
     * @throws {Error} If the seeding fails.
     *
     * Side Effects: Makes a POST request to /api/v1/debug/seed_traffic.
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
     *
     * Summary: Changes the status of an alert (e.g. acknowledge, resolve).
     *
     * @param {string} id - The ID of the alert.
     * @param {string} status - The new status.
     * @returns {Promise<any>} A promise that resolves to the updated alert.
     * @throws {Error} If the update fails.
     *
     * Side Effects: Makes a PATCH request to /api/v1/alerts/:id.
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
     *
     * Summary: Retrieves the global webhook URL.
     *
     * @returns {Promise<{url: string}>} A promise that resolves to the webhook configuration.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/alerts/webhook.
     */
    getWebhookURL: async (): Promise<{ url: string }> => {
        const res = await fetchWithAuth('/api/v1/alerts/webhook');
        if (!res.ok) throw new Error('Failed to fetch webhook URL');
        return res.json();
    },

    /**
     * Saves the configured global webhook URL for alerts.
     *
     * Summary: Updates the global webhook URL.
     *
     * @param {string} url - The webhook URL.
     * @returns {Promise<any>} A promise that resolves to the updated webhook configuration.
     * @throws {Error} If the update fails.
     *
     * Side Effects: Makes a POST request to /api/v1/alerts/webhook.
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
     *
     * Summary: Fetches a list of all defined stacks.
     *
     * @returns {Promise<any>} A promise that resolves to a list of collections.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/collections.
     */
    listCollections: async () => {
        const res = await fetchWithAuth('/api/v1/collections');
        if (!res.ok) throw new Error('Failed to list collections');
        return res.json();
    },

    /**
     * Gets a single service collection (stack) by its name.
     *
     * Summary: Retrieves the configuration of a specific stack.
     *
     * @param {string} name - The name of the collection.
     * @returns {Promise<any>} A promise that resolves to the collection.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/collections/:name.
     */
    getCollection: async (name: string) => {
        const res = await fetchWithAuth(`/api/v1/collections/${name}`);
        if (!res.ok) throw new Error('Failed to get collection');
        return res.json();
    },

    /**
     * Saves a service collection (stack).
     *
     * Summary: Creates or updates a service stack.
     *
     * @param {any} collection - The collection configuration to save.
     * @returns {Promise<any>} A promise that resolves when the collection is saved.
     * @throws {Error} If the save operation fails.
     *
     * Side Effects: Makes a PUT request to /api/v1/collections/:name.
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
     *
     * Summary: Removes a service stack from the system.
     *
     * @param {string} name - The name of the collection to delete.
     * @returns {Promise<any>} A promise that resolves when the collection is deleted.
     * @throws {Error} If the deletion fails.
     *
     * Side Effects: Makes a DELETE request to /api/v1/collections/:name.
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
     *
     * Summary: Alias for getCollection.
     *
     * @param {string} stackId - The ID of the stack.
     * @returns {Promise<any>} A promise that resolves to the stack configuration.
     *
     * Side Effects: Delegates to getCollection.
     */
    getStackConfig: async (stackId: string) => {
        // Map to getCollection
        return apiClient.getCollection(stackId);
    },

    /**
     * Saves the configuration for a stack (Compatibility wrapper).
     *
     * Summary: Alias for saveCollection, ensuring name is set.
     *
     * @param {string} stackId - The ID of the stack.
     * @param {any} config - The configuration content (Collection object).
     * @returns {Promise<any>} A promise that resolves when the config is saved.
     *
     * Side Effects: Delegates to saveCollection.
     */
    saveStackConfig: async (stackId: string, config: any) => {
        // Map to saveCollection. Ensure name is set.
        const collection = typeof config === 'string' ? JSON.parse(config) : config;
        if (!collection.name) collection.name = stackId;
        return apiClient.saveCollection(collection);
    },

    /**
     * Gets the stack configuration as YAML.
     *
     * Summary: Retrieves the stack configuration in YAML format.
     *
     * @param {string} stackId - The ID of the stack.
     * @returns {Promise<string>} A promise that resolves to the YAML string.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/stacks/:stackId/config.
     */
    getStackYaml: async (stackId: string) => {
        const res = await fetchWithAuth(`/api/v1/stacks/${stackId}/config`);
        if (!res.ok) throw new Error('Failed to get stack config');
        return res.text();
    },

    /**
     * Saves the stack configuration from YAML.
     *
     * Summary: Updates the stack configuration from a YAML string.
     *
     * @param {string} stackId - The ID of the stack.
     * @param {string} yamlContent - The YAML configuration content.
     * @returns {Promise<any>} A promise that resolves when the config is saved.
     * @throws {Error} If the save operation fails.
     *
     * Side Effects: Makes a POST request to /api/v1/stacks/:stackId/config.
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



    // User Management

    /**
     * Lists all users.
     *
     * Summary: Fetches a list of all users in the system.
     *
     * @returns {Promise<any[]>} A promise that resolves to a list of users.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/users.
     */
    listUsers: async () => {
        const res = await fetchWithAuth('/api/v1/users');
        if (!res.ok) throw new Error('Failed to list users');
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.users || []);

        // Map snake_case to camelCase
        return list.map((u: any) => ({
            id: u.id,
            roles: u.roles || [],
            profileIds: u.profile_ids || [],
            authentication: u.authentication ? {
                apiKey: u.authentication.api_key ? {
                    paramName: u.authentication.api_key.param_name,
                    in: u.authentication.api_key.in,
                    verificationValue: u.authentication.api_key.verification_value
                } : undefined,
                basicAuth: u.authentication.basic_auth ? {
                    username: u.authentication.basic_auth.username,
                    passwordHash: u.authentication.basic_auth.password_hash
                } : undefined
            } : undefined
        }));
    },

    /**
     * Creates a new user.
     *
     * Summary: Registers a new user with the system.
     *
     * @param {any} user - The user object to create.
     * @returns {Promise<any>} A promise that resolves to the created user.
     * @throws {Error} If the creation fails.
     *
     * Side Effects: Makes a POST request to /api/v1/users.
     */
    createUser: async (user: any) => {
        // Map camelCase to snake_case
        const payload: any = {
            id: user.id,
            roles: user.roles,
            profile_ids: user.profileIds
        };

        if (user.authentication) {
            payload.authentication = {};
            if (user.authentication.apiKey) {
                payload.authentication.api_key = {
                    param_name: user.authentication.apiKey.paramName,
                    in: user.authentication.apiKey.in,
                    verification_value: user.authentication.apiKey.verificationValue
                };
            }
            if (user.authentication.basicAuth) {
                payload.authentication.basic_auth = {
                    username: user.authentication.basicAuth.username,
                    password_hash: user.authentication.basicAuth.passwordHash
                };
            }
        }

        const res = await fetchWithAuth('/api/v1/users', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ user: payload })
        });
        if (!res.ok) throw new Error('Failed to create user');
        return res.json();
    },

    /**
     * Updates an existing user.
     *
     * Summary: Modifies the details of an existing user.
     *
     * @param {any} user - The user object to update.
     * @returns {Promise<any>} A promise that resolves to the updated user.
     * @throws {Error} If the update fails.
     *
     * Side Effects: Makes a PUT request to /api/v1/users/:id.
     */
    updateUser: async (user: any) => {
        // Map camelCase to snake_case
        const payload: any = {
            id: user.id,
            roles: user.roles,
            profile_ids: user.profileIds
        };

        if (user.authentication) {
            payload.authentication = {};
            if (user.authentication.apiKey) {
                payload.authentication.api_key = {
                    param_name: user.authentication.apiKey.paramName,
                    in: user.authentication.apiKey.in,
                    verification_value: user.authentication.apiKey.verificationValue
                };
            }
            if (user.authentication.basicAuth) {
                payload.authentication.basic_auth = {
                    username: user.authentication.basicAuth.username,
                    password_hash: user.authentication.basicAuth.passwordHash
                };
            }
        }

         const res = await fetchWithAuth(`/api/v1/users/${user.id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ user: payload })
        });
        if (!res.ok) throw new Error('Failed to update user');
        return res.json();
    },

    /**
     * Deletes a user.
     *
     * Summary: Removes a user from the system.
     *
     * @param {string} id - The ID of the user to delete.
     * @returns {Promise<any>} A promise that resolves when the user is deleted.
     * @throws {Error} If the deletion fails.
     *
     * Side Effects: Makes a DELETE request to /api/v1/users/:id.
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
     *
     * Summary: Starts the OAuth authorization process for a service.
     *
     * @param {string} serviceID - The ID of the service for which to initiate OAuth.
     * @param {string} redirectURL - The URL to redirect to after OAuth completes.
     * @param {string} [credentialID] - Optional credential ID to associate with the token.
     * @returns {Promise<any>} A promise that resolves to the initiation response.
     * @throws {Error} If the initiation fails.
     *
     * Side Effects: Makes a POST request to /auth/oauth/initiate.
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
     *
     * Summary: Processes the OAuth callback code.
     *
     * @param {string | null} serviceID - The ID of the service (optional).
     * @param {string} code - The OAuth authorization code.
     * @param {string} redirectURL - The redirect URL used in the initial request.
     * @param {string} [credentialID] - Optional credential ID.
     * @returns {Promise<any>} A promise that resolves to the callback handling result.
     * @throws {Error} If the callback processing fails.
     *
     * Side Effects: Makes a POST request to /auth/oauth/callback.
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
     *
     * Summary: Fetches a list of all credentials stored in the system.
     *
     * @returns {Promise<any[]>} A promise that resolves to a list of credentials.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/credentials.
     */
    listCredentials: async () => {
        const res = await fetchWithAuth('/api/v1/credentials');
        if (!res.ok) throw new Error('Failed to list credentials');
        const data = await res.json();
        return Array.isArray(data) ? data : (data.credentials || []);
    },

    /**
     * Saves (creates or updates) a credential.
     *
     * Summary: Wrapper for creating a credential (legacy/convenience).
     *
     * @param {Credential} credential - The credential to save.
     * @returns {Promise<any>} A promise that resolves to the saved credential.
     *
     * Side Effects: Delegates to createCredential.
     */
    saveCredential: async (credential: Credential) => {
        // ... (logic omitted for brevity, keeping same) ...
        return apiClient.createCredential(credential);
    },

    /**
     * Creates a new credential.
     *
     * Summary: Adds a new credential to the system.
     *
     * @param {Credential} credential - The credential to create.
     * @returns {Promise<any>} A promise that resolves to the created credential.
     * @throws {Error} If the creation fails.
     *
     * Side Effects: Makes a POST request to /api/v1/credentials.
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
     *
     * Summary: Modifies an existing credential.
     *
     * @param {Credential} credential - The credential to update.
     * @returns {Promise<any>} A promise that resolves to the updated credential.
     * @throws {Error} If the update fails.
     *
     * Side Effects: Makes a PUT request to /api/v1/credentials/:id.
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
     *
     * Summary: Removes a credential from the system.
     *
     * @param {string} id - The ID of the credential to delete.
     * @returns {Promise<any>} A promise that resolves when the credential is deleted.
     * @throws {Error} If the deletion fails.
     *
     * Side Effects: Makes a DELETE request to /api/v1/credentials/:id.
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
     *
     * Summary: Verifies authentication credentials against a service.
     *
     * @param {any} req - The authentication test request.
     * @returns {Promise<any>} A promise that resolves to the test result.
     *
     * Side Effects: Makes a POST request to /api/v1/debug/auth-test.
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
     *
     * Summary: Fetches a list of all stored service templates.
     *
     * @returns {Promise<any[]>} A promise that resolves to a list of templates.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/templates.
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
     *
     * Summary: Creates or updates a service template.
     *
     * @param {UpstreamServiceConfig} template - The template configuration to save.
     * @returns {Promise<any>} A promise that resolves to the saved template.
     * @throws {Error} If the save operation fails.
     *
     * Side Effects: Makes a POST request to /api/v1/templates.
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
     *
     * Summary: Removes a service template from the system.
     *
     * @param {string} id - The ID of the template to delete.
     * @returns {Promise<any>} A promise that resolves when the template is deleted.
     * @throws {Error} If the deletion fails.
     *
     * Side Effects: Makes a DELETE request to /api/v1/templates/:id.
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
     *
     * Summary: Retrieves a comprehensive system health report.
     *
     * @returns {Promise<DoctorReport>} A promise that resolves to the doctor report.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/doctor.
     */
    getDoctorStatus: async (): Promise<DoctorReport> => {
        const res = await fetchWithAuth('/api/v1/doctor');
        if (!res.ok) throw new Error('Failed to fetch doctor status');
        return res.json();
    },

    // Audit Logs

    /**
     * Lists audit logs.
     *
     * Summary: Fetches audit logs based on provided filters.
     *
     * @param {Object} filters - The filters for the audit logs.
     * @param {string} [filters.start_time] - Start time for logs.
     * @param {string} [filters.end_time] - End time for logs.
     * @param {string} [filters.tool_name] - Filter by tool name.
     * @param {string} [filters.user_id] - Filter by user ID.
     * @param {string} [filters.profile_id] - Filter by profile ID.
     * @param {number} [filters.limit] - Limit results.
     * @param {number} [filters.offset] - Offset results.
     * @returns {Promise<any>} A promise that resolves to the list of audit logs.
     * @throws {Error} If the request fails.
     *
     * Side Effects: Makes a GET request to /api/v1/audit/logs.
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
