import { Registration } from "@proto/api/v1/registration";
import { UpstreamServiceConfig } from "@proto/config/v1/upstream_service";
import { AnyConfig } from "@proto/config/v1/config";
import { ToolDefinition } from "@proto/config/v1/tool";
import { ResourceDefinition } from "@proto/config/v1/resource";
import { PromptDefinition } from "@proto/config/v1/prompt";
import { AuthConfig } from "@proto/config/v1/auth";

export interface ResourceType {
    id: string;
    name: string;
    description: string;
}

export interface AuthContext {
    user_id: string;
    org_id: string;
    email: string;
    roles: string[];
}

export interface AnalyticsSummary {
    total_requests: number;
    error_rate: number;
    avg_latency_ms: number;
    active_tools: number;
}

export interface MetricPoint {
    timestamp: string;
    value: number;
}

export interface ToolUsage {
    tool_id: string;
    calls: number;
    errors: number;
}

export interface TrafficData {
    time_series: MetricPoint[];
}

export interface ValidationResult {
    valid: boolean;
    errors?: string[];
    warnings?: string[];
}

export interface RegistrationResponse {
    registration: Registration;
}

export interface Secret {
    id: string;
    name: string;
    created_at: string;
}

export interface Collection {
    id: string;
    name: string;
    description: string;
    items: any[];
}

export interface DoctorCheck {
    status: 'ok' | 'warning' | 'error';
    message?: string;
    latency?: string;
    diff?: string;
}

export interface SystemStatus {
    uptime_seconds: number;
    active_connections: number;
    bound_http_port: number;
    bound_grpc_port: number;
    version: string;
    security_warnings?: string[];
    has_tls?: boolean;
    has_auth?: boolean;
}

export interface DiscoveryData {
    servers?: {
        id: string;
        [key: string]: any;
    }[];
}

export interface DoctorReport {
    status: 'healthy' | 'degraded' | 'error';
    timestamp?: string;
    checks: Record<string, DoctorCheck>;
}

export interface SystemStatus {
    uptime_seconds: number;
    active_connections: number;
    bound_http_port: number;
    bound_grpc_port: number;
    version: string;
    security_warnings?: string[];
    has_tls?: boolean;
    has_auth?: boolean;
}

export interface DiscoveryData {
    servers?: {
        id: string;
        [key: string]: any;
    }[];
}

export const apiClient = {
    // ---- Registration Management ----
    async registerService(config: UpstreamServiceConfig): Promise<RegistrationResponse> {
        const res = await fetch('/api/v1/registration', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config)
        });
        if (!res.ok) throw new Error('Failed to register service');
        return res.json();
    },

    async getRegistration(id: string): Promise<RegistrationResponse> {
        const res = await fetch(`/api/v1/registration/${id}`);
        if (!res.ok) throw new Error('Failed to get registration');
        return res.json();
    },

    async listRegistrations(): Promise<{ registrations: Registration[] }> {
        const res = await fetch('/api/v1/registrations');
        if (!res.ok) throw new Error('Failed to list registrations');
        return res.json();
    },

    async deleteRegistration(id: string): Promise<void> {
        const res = await fetch(`/api/v1/registration/${id}`, { method: 'DELETE' });
        if (!res.ok) throw new Error('Failed to delete registration');
    },

    // ---- Service Management ----
    async listServices(): Promise<{ services: UpstreamServiceConfig[] }> {
        const res = await fetch('/api/v1/services');
        if (!res.ok) throw new Error('Failed to list services');
        return res.json();
    },

    async getService(id: string): Promise<UpstreamServiceConfig> {
        const res = await fetch(`/api/v1/services/${id}`);
        if (!res.ok) throw new Error('Failed to get service');
        return res.json();
    },

    async validateService(id: string): Promise<ValidationResult> {
        const res = await fetch(`/api/v1/services/${id}/validate`);
        if (!res.ok) throw new Error('Failed to validate service');
        return res.json();
    },

    async getServiceStatus(id: string): Promise<DoctorReport> {
        const res = await fetch(`/api/v1/services/${id}/status`);
        if (!res.ok) throw new Error('Failed to fetch service status');
        return res.json();
    },

    // ---- Diagnostics ----
    async getDoctorStatus(): Promise<DoctorReport> {
        const res = await fetch('/api/v1/doctor');
        if (!res.ok) throw new Error('Failed to fetch doctor status');
        return res.json();
    },

    async getSystemStatus(): Promise<SystemStatus> {
        const res = await fetch('/api/v1/system');
        if (!res.ok) throw new Error('Failed to fetch system status');
        return res.json();
    },

    async getDiscoveryStatus(): Promise<DiscoveryData> {
        const res = await fetch('/api/v1/discovery');
        if (!res.ok) throw new Error('Failed to fetch discovery status');
        return res.json();
    },

    // ---- Auth Management ----
    async getAuthConfig(): Promise<AuthConfig> {
        const res = await fetch('/api/v1/auth/config');
        if (!res.ok) throw new Error('Failed to fetch auth config');
        return res.json();
    },

    async getAuthContext(): Promise<AuthContext> {
        const res = await fetch('/api/v1/auth/context');
        if (!res.ok) throw new Error('Failed to fetch auth context');
        return res.json();
    },

    // ---- Tool Management ----
    async listTools(serviceId?: string): Promise<{ tools: ToolDefinition[] }> {
        const url = serviceId ? `/api/v1/tools?service_id=${serviceId}` : '/api/v1/tools';
        const res = await fetch(url);
        if (!res.ok) throw new Error('Failed to list tools');
        return res.json();
    },

    async executeTool(id: string, args: any): Promise<any> {
        const res = await fetch(`/api/v1/tools/${id}/execute`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(args)
        });
        if (!res.ok) throw new Error('Failed to execute tool');
        return res.json();
    },

    async getTopTools(): Promise<ToolUsage[]> {
        const res = await fetch('/api/v1/analytics/tools/top');
        if (!res.ok) throw new Error('Failed to fetch top tools');
        return res.json();
    },

    async getToolUsage(toolId: string): Promise<ToolUsage> {
        const res = await fetch(`/api/v1/analytics/tools/${toolId}/usage`);
        if (!res.ok) throw new Error('Failed to fetch tool usage');
        return res.json();
    },

    async setToolStatus(toolId: string, status: string): Promise<void> {
        const res = await fetch(`/api/v1/tools/${toolId}/status`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ status })
        });
        if (!res.ok) throw new Error('Failed to set tool status');
    },

    // ---- Prompt Management ----
    async listPrompts(): Promise<{ prompts: PromptDefinition[] }> {
        const res = await fetch('/api/v1/prompts');
        if (!res.ok) throw new Error('Failed to list prompts');
        return res.json();
    },

    async executePrompt(id: string, inputs: any): Promise<any> {
        const res = await fetch(`/api/v1/prompts/${id}/execute`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(inputs)
        });
        if (!res.ok) throw new Error('Failed to execute prompt');
        return res.json();
    },

    // ---- Resource Management ----
    async listResources(): Promise<{ resources: ResourceDefinition[] }> {
        const res = await fetch('/api/v1/resources');
        if (!res.ok) throw new Error('Failed to list resources');
        return res.json();
    },

    async getResourceTypes(): Promise<ResourceType[]> {
        const res = await fetch('/api/v1/resources/types');
        if (!res.ok) throw new Error('Failed to fetch resource types');
        return res.json();
    },

    // ---- System Config ----
    async getSystemConfig(): Promise<AnyConfig> {
        const res = await fetch('/api/v1/config');
        if (!res.ok) throw new Error('Failed to get system config');
        return res.json();
    },

    async validateConfig(config: AnyConfig): Promise<ValidationResult> {
        const res = await fetch('/api/v1/config/validate', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(config)
        });
        if (!res.ok) throw new Error('Failed to validate config');
        return res.json();
    },

    // ---- Analytics ----
    async getAnalyticsSummary(): Promise<AnalyticsSummary> {
        const res = await fetch('/api/v1/analytics/summary');
        if (!res.ok) throw new Error('Failed to fetch analytics summary');
        return res.json();
    },

    async getDashboardTraffic(): Promise<TrafficData> {
        const res = await fetch('/api/v1/analytics/traffic');
        if (!res.ok) throw new Error('Failed to fetch traffic data');
        return res.json();
    },

    // ---- Secrets Management ----
    async listSecrets(): Promise<{ secrets: Secret[] }> {
        const res = await fetch('/api/v1/secrets');
        if (!res.ok) throw new Error('Failed to list secrets');
        return res.json();
    },

    async saveSecret(secret: Partial<Secret>): Promise<Secret> {
        const res = await fetch('/api/v1/secrets', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(secret)
        });
        if (!res.ok) throw new Error('Failed to save secret');
        return res.json();
    },

    async deleteSecret(id: string): Promise<void> {
        const res = await fetch(`/api/v1/secrets/${id}`, { method: 'DELETE' });
        if (!res.ok) throw new Error('Failed to delete secret');
    },

    // ---- Collections / Stacks ----
    async getCollection(id: string): Promise<Collection> {
        const res = await fetch(`/api/v1/collections/${id}`);
        if (!res.ok) throw new Error('Failed to get collection');
        return res.json();
    },

    async saveCollection(collection: Partial<Collection>): Promise<Collection> {
        const res = await fetch('/api/v1/collections', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(collection)
        });
        if (!res.ok) throw new Error('Failed to save collection');
        return res.json();
    }
};
