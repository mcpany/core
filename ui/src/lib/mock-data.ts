
import { UpstreamServiceConfig, ToolDefinition, ResourceDefinition, PromptDefinition } from "./client";

// Shared Mock Database

export interface ProfileDefinition {
    id: string;
    name: string;
    description: string;
    env: Record<string, string>;
    active: boolean;
}

export interface MockDB {
    services: UpstreamServiceConfig[];
    tools: ToolDefinition[];
    resources: ResourceDefinition[];
    prompts: PromptDefinition[];
    profiles: ProfileDefinition[];
}

export const db: MockDB = {
    services: [
        {
            id: "svc-1",
            name: "Payment Gateway",
            version: "v1.2.0",
            disable: false,
            priority: 1,
            http_service: {
                address: "https://api.payments.com",
            },
        },
        {
            id: "svc-2",
            name: "User Service",
            version: "v2.0.1",
            disable: false,
            priority: 2,
            grpc_service: {
                address: "localhost:50051",
            },
        },
        {
            id: "svc-3",
            name: "Legacy Inventory",
            version: "v0.9.5",
            disable: true,
            priority: 5,
            http_service: {
                address: "http://inventory-internal",
            },
        }
    ],
    tools: [
        {
            name: "calculate_tax",
            description: "Calculates sales tax for a given location and amount.",
            enabled: true,
            serviceName: "Payment Gateway",
            schema: {
                type: "object",
                properties: {
                    amount: { type: "number" },
                    zip_code: { type: "string" }
                }
            }
        },
        {
            name: "get_user_profile",
            description: "Retrieves user profile by ID.",
            enabled: true,
            serviceName: "User Service",
            schema: {
                type: "object",
                properties: {
                    user_id: { type: "string" }
                }
            }
        },
        {
            name: "check_inventory",
            description: "Checks stock level for an item.",
            enabled: false,
            serviceName: "Legacy Inventory",
            schema: {
                type: "object",
                properties: {
                    sku: { type: "string" }
                }
            }
        }
    ],
    resources: [
        {
            uri: "file:///logs/access.log",
            name: "Access Logs",
            description: "Web server access logs.",
            enabled: true,
            serviceName: "System",
            mimeType: "text/plain"
        },
        {
            uri: "s3://bucket/reports/daily.csv",
            name: "Daily Report",
            description: "Daily transaction report.",
            enabled: true,
            serviceName: "Reporting",
            mimeType: "text/csv"
        }
    ],
    prompts: [
        {
            name: "summarize_email",
            description: "Summarizes an email thread.",
            enabled: true,
            serviceName: "Email Service",
            arguments: [
                { name: "email_content", type: "string" }
            ]
        },
        {
            name: "generate_sql",
            description: "Generates SQL query from natural language.",
            enabled: true,
            serviceName: "Data Service",
            arguments: [
                { name: "question", type: "string" },
                { name: "schema", type: "string" }
            ]
        }
    ],
    profiles: [
        {
            id: "prof-dev",
            name: "Development",
            description: "Profile for local development with verbose logging.",
            env: { "LOG_LEVEL": "debug", "ENV": "dev" },
            active: true
        },
        {
            id: "prof-prod",
            name: "Production",
            description: "Production environment settings.",
            env: { "LOG_LEVEL": "warn", "ENV": "prod" },
            active: false
        }
    ]
};
