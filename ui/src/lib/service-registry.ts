/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "./types";

export interface ServiceTemplate {
    id: string;
    name: string;
    description: string;
    icon: string;
    config: Partial<UpstreamServiceConfig>;
    configurationSchema?: Record<string, any>;
}

export const SERVICE_REGISTRY: ServiceTemplate[] = [
    {
        id: "stripe",
        name: "Stripe",
        description: "Payment processing platform",
        icon: "credit-card",
        config: {
            httpService: {
                address: "https://api.stripe.com",
            }
        },
        configurationSchema: {
            type: "object",
            properties: {
                STRIPE_SECRET_KEY: {
                    type: "string",
                    title: "Secret Key",
                    description: "Stripe API Secret Key (starts with sk_live_...)"
                }
            },
            required: ["STRIPE_SECRET_KEY"]
        }
    },
    {
        id: "slack",
        name: "Slack",
        description: "Messaging app for business",
        icon: "message-square",
        config: {
            httpService: {
                address: "https://slack.com/api",
            }
        },
        configurationSchema: {
            type: "object",
            properties: {
                SLACK_BOT_TOKEN: {
                    type: "string",
                    title: "Bot User OAuth Token",
                    description: "Token starting with xoxb-..."
                }
            },
            required: ["SLACK_BOT_TOKEN"]
        }
    },
    {
        id: "github",
        name: "GitHub",
        description: "Code hosting platform",
        icon: "github",
        config: {
            httpService: {
                address: "https://api.github.com",
            }
        },
        configurationSchema: {
            type: "object",
            properties: {
                GITHUB_TOKEN: {
                    type: "string",
                    title: "Personal Access Token",
                    description: "GitHub PAT with repo scopes"
                }
            },
            required: ["GITHUB_TOKEN"]
        }
    },
    {
        id: "openai",
        name: "OpenAI",
        description: "AI research and deployment company",
        icon: "cpu",
        config: {
            httpService: {
                address: "https://api.openai.com/v1",
            }
        },
        configurationSchema: {
            type: "object",
            properties: {
                OPENAI_API_KEY: {
                    type: "string",
                    title: "API Key",
                    description: "OpenAI API Key"
                }
            },
            required: ["OPENAI_API_KEY"]
        }
    }
];
