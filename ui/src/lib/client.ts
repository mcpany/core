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
    /**
     * Optional template ID if this config was loaded from a template.
     */
    templateId?: string;
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

/**
 * ServiceStatus represents the possible health states of a service.
 */
export type ServiceStatus = "healthy" | "degraded" | "unhealthy" | "inactive" | "unknown";

/**
 * ServiceHealth describes the current health information of a service.
 */
export interface ServiceHealth {
  /** The unique identifier of the service. */
  id: string;
  /** The display name of the service. */
  name: string;
  /** The current status of the service. */
  status: ServiceStatus;
  /** The latency of the service check. */
  latency: string;
  /** The uptime duration of the service. */
  uptime: string;
  /** An optional message providing more details about the status. */
  message?: string;
}

/**
 * HealthHistoryPoint represents a single data point in the health history of a service.
 */
export interface HealthHistoryPoint {
  /** The timestamp of the health check in milliseconds. */
  timestamp: number;
  /** The status of the service at that time. */
  status: ServiceStatus;
}

/**
 * ServiceHealthResponse represents the response for the health dashboard.
 */
export interface ServiceHealthResponse {
  services: ServiceHealth[];
  history: Record<string, HealthHistoryPoint[]>;
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
     *
     * Summary: Fetches the list of all configured upstream services from the backend.
     *
     * @returns Promise<UpstreamServiceConfig[]>. A promise that resolves to a list of services.
     * @throws Error. If the network request fails or the response is not OK.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/services.
     */
    listServices: async () => {
        // Fallback to REST for E2E reliability until gRPC-Web is stable
        const res = await fetchWithAuth('/api/v1/services');
        if (!res.ok) throw new Error('Failed to fetch services');
        const data = await res.json();
        const list = Array.isArray(data) ? data : (data.services || []);
        // Map snake_case to camelCase for UI compatibility
        return list.map(mapUpstreamServiceConfig);
    },

    /**
     * Lists services from the dynamic catalog.
     *
     * Summary: Retrieves services available in the catalog.
     *
     * @returns Promise<any[]>. A promise that resolves to a list of catalog services.
     * @throws Error. If the network request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/catalog/services.
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
     * @param id - string. The ID of the service to retrieve.
     * @returns Promise<any>. A promise that resolves to the service configuration.
     * @throws Error. If the service is not found or the request fails.
     *
     * Side Effects:
     *   - Makes a gRPC call or GET request to /api/v1/services/:id.
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
     * Summary: Enables or disables a service.
     *
     * @param name - string. The name of the service.
     * @param disable - boolean. True to disable the service, false to enable it.
     * @returns Promise<any>. A promise that resolves to the updated service status.
     * @throws Error. If the request fails.
     *
     * Side Effects:
     *   - Makes a PUT request to /api/v1/services/:name.
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
     * Summary: Retrieves the current status of a service.
     *
     * @param name - string. The name of the service.
     * @returns Promise<any>. A promise that resolves to the service status.
     * @throws Error. If the request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/services/:name/status.
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
     * @param name - string. The name of the service to restart.
     * @returns Promise<any>. A promise that resolves when the service is restarted.
     * @throws Error. If the restart command fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/services/:name/restart.
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
     * @param config - UpstreamServiceConfig. The configuration of the service to register.
     * @returns Promise<any>. A promise that resolves to the registered service configuration.
     * @throws Error. If the registration fails (e.g., validation error, duplicate ID).
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/services.
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
     * Summary: Updates the configuration of an existing service.
     *
     * @param config - UpstreamServiceConfig. The updated configuration of the service.
     * @returns Promise<any>. A promise that resolves to the updated service configuration.
     * @throws Error. If the update fails.
     *
     * Side Effects:
     *   - Makes a PUT request to /api/v1/services/:name.
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
     * Summary: Deletes a registered service.
     *
     * @param id - string. The ID of the service to unregister.
     * @returns Promise<any>. A promise that resolves when the service is unregistered.
     * @throws Error. If the deletion fails.
     *
     * Side Effects:
     *   - Makes a DELETE request to /api/v1/services/:id.
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
     * Summary: checks if a service configuration is valid without registering it.
     *
     * @param config - UpstreamServiceConfig. The service configuration to validate.
     * @returns Promise<any>. A promise that resolves to the validation result.
     * @throws Error. If validation check fails (network error).
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/services/validate.
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

    // Tools (Legacy Fetch - Not yet migrated to Admin/Registration Service completely or keeping as is)
    // admin.proto has ListTools but we are focusing on RegistrationService first.
    // So keep using fetch for Tools/Secrets/etc for now.

    /**
     * Lists all available tools.
     *
     * Summary: Fetches a list of all tools from all registered services.
     *
     * @returns Promise<any>. A promise that resolves to an object containing a list of tools.
     * @throws Error. If the request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/tools.
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
     * Summary: Runs a specific tool with given inputs.
     *
     * @param request - any. The execution request (tool name, arguments, etc.).
     * @param dryRun - boolean. If true, performs a dry run without side effects.
     * @returns Promise<any>. A promise that resolves to the execution result.
     * @throws Error. If tool execution fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/execute.
     *   - May perform actions on upstream services (unless dryRun is true).
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
     * @param name - string. The name of the tool.
     * @param disabled - boolean. True to disable the tool, false to enable it.
     * @returns Promise<any>. A promise that resolves to the updated tool status.
     *
     * Side Effects:
     *   - Makes a PUT request to /api/v1/tools.
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
     * Summary: Fetches all resources available in the system.
     *
     * @returns Promise<any>. A promise that resolves to a list of resources.
     * @throws Error. If the request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/resources.
     */
    listResources: async () => {
        const res = await fetchWithAuth('/api/v1/resources');
        if (!res.ok) throw new Error('Failed to list resources');
        return res.json();
    },

    /**
     * Reads the content of a resource.
     *
     * Summary: Fetches the content of a specific resource.
     *
     * @param uri - string. The URI of the resource to read.
     * @returns Promise<ReadResourceResponse>. A promise that resolves to the resource content.
     * @throws Error. If the resource cannot be read.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/resources/read.
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
     * @param uri - string. The URI of the resource.
     * @param disabled - boolean. True to disable the resource, false to enable it.
     * @returns Promise<any>. A promise that resolves to the updated resource status.
     *
     * Side Effects:
     *   - Makes a PUT request to /api/v1/resources.
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
     * Summary: Fetches all prompts available in the system.
     *
     * @returns Promise<any>. A promise that resolves to a list of prompts.
     * @throws Error. If the request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/prompts.
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
     * @param name - string. The name of the prompt.
     * @param enabled - boolean. True to enable the prompt, false to disable it.
     * @returns Promise<any>. A promise that resolves to the updated prompt status.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/prompts.
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
     * Summary: Runs a specific prompt.
     *
     * @param name - string. The name of the prompt.
     * @param args - Record<string, string>. The arguments for the prompt.
     * @returns Promise<any>. A promise that resolves to the prompt execution result.
     * @throws Error. If prompt execution fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/prompts/execute.
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
     * Summary: Fetches available service templates from the marketplace.
     *
     * @returns Promise<any[]>. A list of templates.
     * @throws Error. If the request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/templates.
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
                    // Map snake_case to camelCase for UI consumption
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
                    // Specific mapping for openapi_service
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
     *
     * Summary: Starts the OAuth authorization process.
     *
     * @param serviceId - string. The ID of the service (e.g. "google_calendar").
     * @param credentialId - string. The ID of the credential to bind.
     * @param redirectUrl - string. The URL to redirect back to after auth.
     * @returns Promise<any>. The authorization URL and state.
     * @throws Error. If initiation fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/auth/oauth/initiate.
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

    /**
     * Lists all authentication credentials.
     *
     * Summary: Fetches all stored credentials.
     *
     * @returns Promise<Credential[]>. A list of credentials.
     * @throws Error. If the request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/credentials.
     */
    listCredentials: async (): Promise<Credential[]> => {
        const res = await fetchWithAuth('/api/v1/credentials');
        if (!res.ok) throw new Error('Failed to fetch credentials');
        const data = await res.json();
        return Array.isArray(data) ? data : (data.credentials || []);
    },

    /**
     * Creates a new authentication credential.
     *
     * Summary: Stores a new credential.
     *
     * @param credential - any. The credential data.
     * @returns Promise<Credential>. The created credential.
     * @throws Error. If creation fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/credentials.
     */
    createCredential: async (credential: any): Promise<Credential> => {
        const res = await fetchWithAuth('/api/v1/credentials', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(credential)
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to create credential: ${txt}`);
        }
        return res.json();
    },

    /**
     * Updates an existing authentication credential.
     *
     * Summary: Updates a stored credential.
     *
     * @param credential - any. The updated credential data.
     * @returns Promise<Credential>. The updated credential.
     * @throws Error. If update fails.
     *
     * Side Effects:
     *   - Makes a PUT request to /api/v1/credentials/:id.
     */
    updateCredential: async (credential: any): Promise<Credential> => {
        const res = await fetchWithAuth(`/api/v1/credentials/${credential.id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(credential)
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to update credential: ${txt}`);
        }
        return res.json();
    },

    /**
     * Deletes an authentication credential.
     *
     * Summary: Removes a stored credential.
     *
     * @param id - string. The ID of the credential to delete.
     * @returns Promise<void>.
     * @throws Error. If deletion fails.
     *
     * Side Effects:
     *   - Makes a DELETE request to /api/v1/credentials/:id.
     */
    deleteCredential: async (id: string): Promise<void> => {
        const res = await fetchWithAuth(`/api/v1/credentials/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete credential');
    },

    /**
     * Tests an authentication configuration.
     *
     * Summary: Verifies if a credential works against a target.
     *
     * @param request - any. The test parameters.
     * @returns Promise<any>. The test result.
     * @throws Error. If test fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/debug/auth-test.
     */
    testAuth: async (request: any): Promise<any> => {
        const res = await fetchWithAuth('/api/v1/debug/auth-test', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(request)
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to test auth: ${txt}`);
        }
        return res.json();
    },

    /**
     * Exchanges an OAuth code for a token.
     *
     * Summary: Completes the OAuth flow by exchanging code for token.
     *
     * @param code - string. The OAuth code.
     * @param state - string. The OAuth state.
     * @param redirectUri - string. The redirect URI.
     * @returns Promise<any>. The token response.
     * @throws Error. If exchange fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/oauth/exchange.
     */
    exchangeOAuthCode: async (code: string, state: string, redirectUri: string): Promise<any> => {
        const res = await fetchWithAuth('/api/v1/oauth/exchange', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ code, state, redirect_uri: redirectUri })
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to exchange OAuth code: ${txt}`);
        }
        return res.json();
    },

    /**
     * Handles the OAuth callback by exchanging the code for a token and associating it.
     *
     * Summary: Handles OAuth callback and links credential.
     *
     * @param serviceId - string | null. The service ID being authenticated.
     * @param code - string. The authorization code from the provider.
     * @param redirectUrl - string. The redirect URL used in the flow.
     * @param credentialId - string | undefined. Optional specific credential ID to update.
     * @returns Promise<any>. The result of the callback handler.
     * @throws Error. If callback handling fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/auth/oauth/callback.
     */
    handleOAuthCallback: async (serviceId: string | null, code: string, redirectUrl: string, credentialId?: string) => {
        const res = await fetchWithAuth('/api/v1/auth/oauth/callback', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                service_id: serviceId,
                code: code,
                redirect_url: redirectUrl,
                credential_id: credentialId
            })
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to handle OAuth callback: ${txt}`);
        }
        return res.json();
    },

    /**
     * Lists all users.
     *
     * Summary: Fetches all users.
     *
     * @returns Promise<any>. A list of users.
     * @throws Error. If the request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/users.
     */
    listUsers: async (): Promise<any> => {
        const res = await fetchWithAuth('/api/v1/users');
        if (!res.ok) throw new Error('Failed to fetch users');
        return res.json();
    },

    /**
     * Creates a new user.
     *
     * Summary: Adds a new user.
     *
     * @param user - any. The user data.
     * @returns Promise<any>. The created user.
     * @throws Error. If creation fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/users.
     */
    createUser: async (user: any): Promise<any> => {
        const res = await fetchWithAuth('/api/v1/users', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(user)
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to create user: ${txt}`);
        }
        return res.json();
    },

    /**
     * Updates an existing user.
     *
     * Summary: Modifies a user.
     *
     * @param user - any. The updated user data.
     * @returns Promise<any>. The updated user.
     * @throws Error. If update fails.
     *
     * Side Effects:
     *   - Makes a PUT request to /api/v1/users/:id.
     */
    updateUser: async (user: any): Promise<any> => {
        const res = await fetchWithAuth(`/api/v1/users/${user.id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(user)
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to update user: ${txt}`);
        }
        return res.json();
    },

    /**
     * Deletes a user.
     *
     * Summary: Removes a user.
     *
     * @param id - string. The ID of the user to delete.
     * @returns Promise<void>.
     * @throws Error. If deletion fails.
     *
     * Side Effects:
     *   - Makes a DELETE request to /api/v1/users/:id.
     */
    deleteUser: async (id: string): Promise<void> => {
        const res = await fetchWithAuth(`/api/v1/users/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete user');
    },

    /**
     * Lists all skills.
     *
     * Summary: Fetches all skills.
     *
     * @returns Promise<any[]>. A list of skills.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/skills.
     */
    listSkills: async (): Promise<any[]> => {
        const res = await fetchWithAuth('/api/v1/skills');
        if (!res.ok) throw new Error('Failed to fetch skills');
        const data = await res.json();
        return data.skills || [];
    },

    /**
     * Gets a skill by name.
     *
     * Summary: Fetches details of a specific skill.
     *
     * @param name - string. The name of the skill.
     * @returns Promise<any>. The skill details.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/skills/:name.
     */
    getSkill: async (name: string): Promise<any> => {
        const res = await fetchWithAuth(`/api/v1/skills/${name}`);
        if (!res.ok) throw new Error('Failed to fetch skill');
        const data = await res.json();
        return data.skill;
    },

    /**
     * Creates a new skill.
     *
     * Summary: Adds a new skill.
     *
     * @param skill - any. The skill data.
     * @returns Promise<any>. The created skill.
     * @throws Error. If creation fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/skills.
     */
    createSkill: async (skill: any): Promise<any> => {
        const res = await fetchWithAuth('/api/v1/skills', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(skill)
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to create skill: ${txt}`);
        }
        const data = await res.json();
        return data.skill;
    },

    /**
     * Updates an existing skill.
     *
     * Summary: Modifies a skill.
     *
     * @param originalName - string. The original name of the skill.
     * @param skill - any. The updated skill data.
     * @returns Promise<any>. The updated skill.
     * @throws Error. If update fails.
     *
     * Side Effects:
     *   - Makes a PUT request to /api/v1/skills/:originalName.
     */
    updateSkill: async (originalName: string, skill: any): Promise<any> => {
        const res = await fetchWithAuth(`/api/v1/skills/${originalName}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(skill)
        });
        if (!res.ok) {
            const txt = await res.text();
            throw new Error(`Failed to update skill: ${txt}`);
        }
        const data = await res.json();
        return data.skill;
    },

    /**
     * Deletes a skill.
     *
     * Summary: Removes a skill.
     *
     * @param name - string. The name of the skill to delete.
     * @returns Promise<void>.
     * @throws Error. If deletion fails.
     *
     * Side Effects:
     *   - Makes a DELETE request to /api/v1/skills/:name.
     */
    deleteSkill: async (name: string): Promise<void> => {
        const res = await fetchWithAuth(`/api/v1/skills/${name}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete skill');
    },

    // Profiles

    /**
     * Creates a new profile.
     *
     * Summary: Adds a new profile.
     *
     * @param profileData - any. The profile configuration.
     * @returns Promise<any>. The created profile.
     * @throws Error. If creation fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/profiles.
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
     * Summary: Modifies a profile.
     *
     * @param profileData - any. The profile configuration.
     * @returns Promise<any>. The updated profile.
     * @throws Error. If update fails.
     *
     * Side Effects:
     *   - Makes a PUT request to /api/v1/profiles/:name.
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
    * Summary: Removes a profile.
    *
    * @param name - string. The name of the profile to delete.
    * @returns Promise<any>.
    * @throws Error. If deletion fails.
    *
    * Side Effects:
    *   - Makes a DELETE request to /api/v1/profiles/:name.
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
     * Summary: Fetches all profiles.
     *
     * @returns Promise<any[]>. A list of profiles.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/profiles.
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
     * Summary: Fetches all secrets.
     *
     * @returns Promise<any[]>. A list of secrets.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/secrets.
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
     * Summary: Fetches the plaintext value of a secret.
     *
     * @param id - string. The ID of the secret to reveal.
     * @returns Promise<{ value: string }>. The secret value.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/secrets/:id/reveal.
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
     * Summary: Stores a new secret.
     *
     * @param secret - SecretDefinition. The secret to save.
     * @returns Promise<any>. The saved secret.
     * @throws Error. If save fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/secrets.
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
     * Summary: Removes a secret.
     *
     * @param id - string. The ID of the secret to delete.
     * @returns Promise<any>.
     * @throws Error. If deletion fails.
     *
     * Side Effects:
     *   - Makes a DELETE request to /api/v1/secrets/:id.
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
     * @returns Promise<any>. A promise that resolves to the global settings.
     * @throws Error. If the request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/settings.
     */
    getGlobalSettings: async () => {
        const res = await fetchWithAuth('/api/v1/settings');
        if (!res.ok) throw new Error('Failed to fetch global settings');
        return res.json();
    },

    /**
     * Saves the global server settings.
     *
     * Summary: Updates the global configuration.
     *
     * @param settings - any. The settings to save.
     * @returns Promise<void>. A promise that resolves when the settings are saved.
     * @throws Error. If save fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/settings.
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
     * Summary: Fetches traffic metrics for the dashboard.
     *
     * @param serviceId - string | undefined. Optional service ID to filter by.
     * @param timeRange - string | undefined. Optional time range to filter by (e.g. "1h", "24h").
     * @returns Promise<any>. A promise that resolves to the traffic history points.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/dashboard/traffic.
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
     * Summary: Fetches statistics for top used tools.
     *
     * @param serviceId - string | undefined. Optional service ID to filter by.
     * @returns Promise<any[]>. A promise that resolves to the top tools stats.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/dashboard/top-tools.
     */
    getTopTools: async (serviceId?: string) => {
        let url = '/api/v1/dashboard/top-tools';
        if (serviceId) url += `?serviceId=${encodeURIComponent(serviceId)}`;
        const res = await fetchWithAuth(url);
        // If 404/500, return empty to avoid crashing UI
        if (!res.ok) return [];
        const data = await res.json();
        return data || [];
    },

    // Alerts

    /**
     * Lists all alerts.
     *
     * Summary: Fetches all system alerts.
     *
     * @returns Promise<any>. A promise that resolves to a list of alerts.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/alerts.
     */
    listAlerts: async () => {
        const res = await fetchWithAuth('/api/v1/alerts');
        if (!res.ok) throw new Error('Failed to fetch alerts');
        return res.json();
    },

    /**
     * Lists all alert rules.
     *
     * Summary: Fetches configured alert rules.
     *
     * @returns Promise<any>. A promise that resolves to a list of alert rules.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/alerts/rules.
     */
    listAlertRules: async () => {
        const res = await fetchWithAuth('/api/v1/alerts/rules');
        if (!res.ok) throw new Error('Failed to fetch alert rules');
        return res.json();
    },

    /**
     * Creates a new alert rule.
     *
     * Summary: Adds a new alert rule.
     *
     * @param rule - any. The rule to create.
     * @returns Promise<any>. The created rule.
     * @throws Error. If creation fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/alerts/rules.
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
     * Summary: Fetches details of a specific alert rule.
     *
     * @param id - string. The ID of the rule.
     * @returns Promise<any>. The rule details.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/alerts/rules/:id.
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
     * @param rule - any. The rule to update (must include id).
     * @returns Promise<any>. The updated rule.
     * @throws Error. If update fails.
     *
     * Side Effects:
     *   - Makes a PUT request to /api/v1/alerts/rules/:id.
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
     * Summary: Removes an alert rule.
     *
     * @param id - string. The ID of the rule to delete.
     * @returns Promise<any>.
     * @throws Error. If deletion fails.
     *
     * Side Effects:
     *   - Makes a DELETE request to /api/v1/alerts/rules/:id.
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
     * Summary: Fetches failure statistics for tools.
     *
     * @param serviceId - string | undefined. Optional service ID to filter by.
     * @returns Promise<ToolFailureStats[]>. A promise that resolves to the tool failure stats.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/dashboard/tool-failures.
     */
    getToolFailures: async (serviceId?: string): Promise<ToolFailureStats[]> => {
        let url = '/api/v1/dashboard/tool-failures';
        if (serviceId) url += `?serviceId=${encodeURIComponent(serviceId)}`;
        const res = await fetchWithAuth(url);
        if (!res.ok) return [];
        const data = await res.json();
        return data || [];
    },

    /**
     * Gets the tool usage analytics.
     *
     * Summary: Fetches usage statistics for tools.
     *
     * @param serviceId - string | undefined. Optional service ID to filter by.
     * @returns Promise<ToolAnalytics[]>. A promise that resolves to the tool usage stats.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/dashboard/tool-usage.
     */
    getToolUsage: async (serviceId?: string): Promise<ToolAnalytics[]> => {
        let url = '/api/v1/dashboard/tool-usage';
        if (serviceId) url += `?serviceId=${encodeURIComponent(serviceId)}`;
        const res = await fetchWithAuth(url);
        if (!res.ok) return [];
        const data = await res.json();
        return data || [];
    },


    /**
     * Gets the system status.
     *
     * Summary: Fetches current system status information.
     *
     * @returns Promise<SystemStatus>. A promise that resolves to the system status.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/system/status.
     */
    getSystemStatus: async (): Promise<SystemStatus> => {
        const res = await fetchWithAuth('/api/v1/system/status');
        if (!res.ok) throw new Error('Failed to fetch system status');
        return res.json();
    },

    /**
     * Gets the doctor health report.
     *
     * Summary: Runs system health checks and returns a report.
     *
     * @returns Promise<DoctorReport>. A promise that resolves to the doctor report.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/doctor.
     */
    getDoctorStatus: async (): Promise<DoctorReport> => {
        const res = await fetchWithAuth('/api/v1/doctor');
        if (!res.ok) throw new Error('Failed to fetch doctor status');
        return res.json();
    },

    /**
     * Gets the dashboard health status and history.
     *
     * Summary: Fetches overall health status for the dashboard.
     *
     * @returns Promise<ServiceHealthResponse>. A promise that resolves to the health response.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/dashboard/health.
     */
    getDashboardHealth: async (): Promise<ServiceHealthResponse> => {
        const res = await fetchWithAuth('/api/v1/dashboard/health');
        if (!res.ok) throw new Error('Failed to fetch dashboard health');
        return res.json();
    },

    /**
     * Gets the dashboard metrics.
     *
     * Summary: Fetches key performance metrics.
     *
     * @param serviceId - string | undefined. Optional service ID to filter by.
     * @returns Promise<Metric[]>. A promise that resolves to the metrics list.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/dashboard/metrics.
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
     * Summary: Fetches recent tool execution traces.
     *
     * @param options - { limit?: number }. Optional parameters.
     * @returns Promise<any[]>. A promise that resolves to the traces list.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/traces.
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
     * Gets the network topology graph.
     *
     * Summary: Fetches the current system topology.
     *
     * @returns Promise<any>. A promise that resolves to the graph data.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/topology.
     */
    getTopology: async () => {
        const res = await fetchWithAuth('/api/v1/topology');
        if (!res.ok) throw new Error('Failed to fetch topology');
        return res.json();
    },

    /**
     * Seeds the dashboard traffic history (Debug/Test only).
     *
     * Summary: Injects fake traffic data for testing.
     *
     * @param points - any[]. The traffic points to seed.
     * @returns Promise<void>.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/debug/seed_traffic.
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
     * Summary: Changes the status of an alert (e.g. acknowledge).
     *
     * @param id - string. The ID of the alert.
     * @param status - string. The new status.
     * @returns Promise<any>. The updated alert.
     * @throws Error. If update fails.
     *
     * Side Effects:
     *   - Makes a PATCH request to /api/v1/alerts/:id.
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
     * Summary: Fetches the alert webhook configuration.
     *
     * @returns Promise<{ url: string }>. The webhook URL.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/alerts/webhook.
     */
    getWebhookURL: async (): Promise<{ url: string }> => {
        const res = await fetchWithAuth('/api/v1/alerts/webhook');
        if (!res.ok) throw new Error('Failed to fetch webhook URL');
        return res.json();
    },

    /**
     * Saves the configured global webhook URL for alerts.
     *
     * Summary: Updates the alert webhook configuration.
     *
     * @param url - string. The webhook URL.
     * @returns Promise<any>. The updated webhook configuration.
     * @throws Error. If save fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/alerts/webhook.
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
     * Summary: Fetches all collections.
     *
     * @returns Promise<any>. A list of collections.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/collections.
     */
    listCollections: async () => {
        const res = await fetchWithAuth('/api/v1/collections');
        if (!res.ok) throw new Error('Failed to list collections');
        return res.json();
    },

    /**
     * Gets a single service collection (stack) by its name.
     *
     * Summary: Fetches details of a specific collection.
     *
     * @param name - string. The name of the collection.
     * @returns Promise<any>. The collection details.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/collections/:name.
     */
    getCollection: async (name: string) => {
        const res = await fetchWithAuth(`/api/v1/collections/${name}`);
        if (!res.ok) throw new Error('Failed to get collection');
        return res.json();
    },

    /**
     * Saves a service collection (stack).
     *
     * Summary: Creates or updates a collection.
     *
     * @param collection - any. The collection data.
     * @returns Promise<any>. The saved collection.
     * @throws Error. If save fails.
     *
     * Side Effects:
     *   - Makes a PUT request to /api/v1/collections/:name.
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
     * Summary: Removes a collection.
     *
     * @param name - string. The name of the collection to delete.
     * @returns Promise<any>.
     * @throws Error. If deletion fails.
     *
     * Side Effects:
     *   - Makes a DELETE request to /api/v1/collections/:name.
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
     * Summary: Wrapper for getCollection.
     *
     * @param stackId - string. The ID of the stack.
     * @returns Promise<any>. The stack configuration.
     *
     * Side Effects:
     *   - Calls getCollection.
     */
    getStackConfig: async (stackId: string) => {
        // Map to getCollection
        return apiClient.getCollection(stackId);
    },

    /**
     * Saves the configuration for a stack (Compatibility wrapper).
     *
     * Summary: Wrapper for saveCollection.
     *
     * @param stackId - string. The ID of the stack.
     * @param config - any. The configuration content (Collection object).
     * @returns Promise<any>.
     *
     * Side Effects:
     *   - Calls saveCollection.
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
     * Summary: Fetches stack config in YAML format.
     *
     * @param stackId - string. The ID of the stack.
     * @returns Promise<string>. The YAML string.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/stacks/:stackId/config.
     */
    getStackYaml: async (stackId: string) => {
        const res = await fetchWithAuth(`/api/v1/stacks/${stackId}/config`);
        if (!res.ok) throw new Error('Failed to get stack config');
        return res.text();
    },

    /**
     * Lists all service templates from the marketplace.
     *
     * Summary: Fetches marketplace templates.
     *
     * @returns Promise<ServiceTemplate[]>. A list of service templates.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/templates.
     */
    listTemplates: async (): Promise<ServiceTemplate[]> => {
        const res = await fetchWithAuth('/api/v1/templates');
        if (!res.ok) throw new Error('Failed to fetch templates');
        return res.json();
    },

    /**
     * Saves a service template to the marketplace.
     *
     * Summary: Creates a new template in the marketplace.
     *
     * @param template - ServiceTemplate. The template to save.
     * @returns Promise<any>. The saved template.
     * @throws Error. If save fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/templates.
     */
    saveTemplate: async (template: ServiceTemplate) => {
        const res = await fetchWithAuth('/api/v1/templates', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(template)
        });
        if (!res.ok) throw new Error('Failed to save template');
        return res.json();
    },

    /**
     * Deletes a service template from the marketplace.
     *
     * Summary: Removes a template from the marketplace.
     *
     * @param id - string. The ID of the template to delete.
     * @returns Promise<any>.
     * @throws Error. If deletion fails.
     *
     * Side Effects:
     *   - Makes a DELETE request to /api/v1/templates/:id.
     */
    deleteTemplate: async (id: string) => {
        const res = await fetchWithAuth(`/api/v1/templates/${id}`, {
            method: 'DELETE'
        });
        if (!res.ok) throw new Error('Failed to delete template');
        return {};
    },

    /**
     * Saves the stack configuration from YAML.
     *
     * Summary: Updates a stack using YAML config.
     *
     * @param stackId - string. The ID of the stack.
     * @param yamlContent - string. The YAML configuration content.
     * @returns Promise<any>. The updated stack.
     * @throws Error. If save fails.
     *
     * Side Effects:
     *   - Makes a POST request to /api/v1/stacks/:stackId/config.
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
     *
     * Summary: Fetches audit logs with filtering.
     *
     * @param filters - Object. The filters for the audit logs.
     * @returns Promise<any>. A list of audit logs.
     * @throws Error. If request fails.
     *
     * Side Effects:
     *   - Makes a GET request to /api/v1/audit/logs.
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
