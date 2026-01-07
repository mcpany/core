/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { ListServicesResponse, UpstreamServiceConfig, GetServiceResponse, GetServiceStatusResponse } from "./types";

const mockServices: UpstreamServiceConfig[] = [
  {
    id: "auth-service-prod-123",
    name: "auth-service",
    sanitizedName: "auth-service",
    version: "1.2.1",
    disable: false,
    priority: 0,
    callPolicies: [],
    preCallHooks: [],
    postCallHooks: [],
    profiles: [],
    prompts: [],
    loadBalancingStrategy: 0,
    profileLimits: {},
    autoDiscoverTool: false,
    grpcService: {
      address: "auth.prod.mcpany.io:443",
      useReflection: true,
      tlsConfig: { serverName: "auth.prod.mcpany.io", insecureSkipVerify: false, caCertPath: "", clientCertPath: "", clientKeyPath: "" },
      tools: [
        {
            name: "Login",
            title: "Login",
            description: "Authenticate a user.",
            serviceId: "auth-service-prod-123",
            disable: false,
            isStream: false,
            readOnlyHint: false,
            destructiveHint: false,
            idempotentHint: false,
            openWorldHint: false,
            callId: "",
            profiles: [],
            tags: [],
            mergeStrategy: 0,
        },
        {
            name: "Logout",
            title: "Logout",
            description: "Log out a user.",
            serviceId: "auth-service-prod-123",
            disable: false,
            isStream: false,
            readOnlyHint: false,
            destructiveHint: false,
            idempotentHint: false,
            openWorldHint: false,
            callId: "",
            profiles: [],
            tags: [],
            mergeStrategy: 0,
        },
        {
            name: "CheckStatus",
            title: "Check Status",
            description: "Check authentication status.",
            serviceId: "auth-service-prod-123",
            disable: false,
            isStream: false,
            readOnlyHint: true,
            destructiveHint: false,
            idempotentHint: true,
            openWorldHint: false,
            callId: "",
            profiles: [],
            tags: [],
            mergeStrategy: 0,
        },
      ],
      resources: [
        { name: "User", title: "User", description: "User profile", uri: "user://profile", mimeType: "application/json", size: 0 as any, disable: false, profiles: [] },
        { name: "Session", title: "Session", description: "User session", uri: "user://session", mimeType: "application/json", size: 0 as any, disable: false, profiles: [] },
      ],
      protoDefinitions: [],
      protoCollection: [],
      calls: {},
      prompts: []
    }
  },
  {
    id: "payment-gateway-prod-456",
    name: "payment-gateway",
    sanitizedName: "payment-gateway",
    version: "2.0.0",
    disable: false,
    priority: 0,
    callPolicies: [],
    preCallHooks: [],
    postCallHooks: [],
    profiles: [],
    prompts: [],
    loadBalancingStrategy: 0,
    profileLimits: {},
    autoDiscoverTool: false,
    httpService: {
      address: "https://payments.prod.mcpany.io",
      tlsConfig: { serverName: "payments.prod.mcpany.io", insecureSkipVerify: false, caCertPath: "", clientCertPath: "", clientKeyPath: "" },
      calls: {},
      resources: [],
      prompts: [],
      tools: [
        {
            name: "CreatePayment",
            title: "Create Payment",
            description: "Create a new payment.",
            serviceId: "payment-gateway-prod-456",
            disable: false,
            isStream: false,
            readOnlyHint: false,
            destructiveHint: true,
            idempotentHint: false,
            openWorldHint: true,
            callId: "",
            profiles: [],
            tags: [],
            mergeStrategy: 0,
        },
        {
            name: "GetPaymentStatus",
            title: "Get Payment Status",
            description: "Get the status of a payment.",
            serviceId: "payment-gateway-prod-456",
            disable: false,
            isStream: false,
            readOnlyHint: true,
            destructiveHint: false,
            idempotentHint: true,
            openWorldHint: false,
            callId: "",
            profiles: [],
            tags: [],
            mergeStrategy: 0,
        },
      ],
    }
  },
  {
    id: "user-profiles-staging-789",
    name: "user-profiles",
    sanitizedName: "user-profiles",
    version: "0.5.0-beta",
    disable: true,
    priority: 0,
    callPolicies: [],
    preCallHooks: [],
    postCallHooks: [],
    profiles: [],
    prompts: [],
    loadBalancingStrategy: 0,
    profileLimits: {},
    autoDiscoverTool: false,
    grpcService: {
      address: "users.staging.mcpany.io:8080",
      useReflection: false,
      tools: [],
      resources: [],
      protoDefinitions: [],
      protoCollection: [],
      calls: {},
      prompts: []
    }
  },
  {
    id: "inventory-service-dev-101",
    name: "inventory-service",
    sanitizedName: "inventory-service",
    version: "0.1.0-dev",
    disable: false,
    priority: 0,
    callPolicies: [],
    preCallHooks: [],
    postCallHooks: [],
    profiles: [],
    prompts: [],
    loadBalancingStrategy: 0,
    profileLimits: {},
    autoDiscoverTool: false,
    commandLineService: {
      command: "go run ./cmd/inventory",
      workingDirectory: ".",
      communicationProtocol: 0,
      local: true,
      env: {},
      calls: {},
      resources: [],
      prompts: [],
      tools: [
        {
            name: "CheckStock",
            title: "Check Stock",
            description: "Check stock levels for an item.",
            serviceId: "inventory-service-dev-101",
            disable: false,
            isStream: false,
            readOnlyHint: true,
            destructiveHint: false,
            idempotentHint: true,
            openWorldHint: false,
            callId: "",
            profiles: [],
            tags: [],
            mergeStrategy: 0,
        }
      ]
    }
  },
];

const mockMetrics: Record<string, Record<string, number>> = {
    "auth-service-prod-123": {
        "tool_usage:Login": 1024,
        "tool_usage:Logout": 512,
        "tool_usage:CheckStatus": 2048,
        "resource_access:User": 300,
        "resource_access:Session": 1200,
    },
    "payment-gateway-prod-456": {
        "tool_usage:CreatePayment": 78,
        "tool_usage:GetPaymentStatus": 230,
        "prompt_usage:GenerateInvoice": 45,
    },
    "inventory-service-dev-101": {
        "tool_usage:CheckStock": 1500,
    }
}


const updateService = (id: string, update: Partial<UpstreamServiceConfig>): UpstreamServiceConfig | null => {
    const serviceIndex = mockServices.findIndex(s => s.id === id);
    if (serviceIndex !== -1) {
        mockServices[serviceIndex] = { ...mockServices[serviceIndex], ...update };
        return mockServices[serviceIndex];
    }
    return null;
}


export const apiClient = {
  listServices: (): Promise<ListServicesResponse> => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({ services: [...mockServices] });
      }, 500);
    });
  },
  getService: (id: string): Promise<GetServiceResponse> => {
      return new Promise((resolve, reject) => {
          setTimeout(() => {
              const service = mockServices.find(s => s.id === id);
              if (service) {
                  resolve({ service: {...service} });
              } else {
                  reject(new Error("Service not found"));
              }
          }, 300);
      });
  },
  setServiceStatus: (id: string, disabled: boolean): Promise<GetServiceResponse> => {
      return new Promise((resolve, reject) => {
          setTimeout(() => {
              const updatedService = updateService(id, { disable: disabled });
              if (updatedService) {
                resolve({ service: updatedService });
              } else {
                reject(new Error("Service not found to update status"));
              }
          }, 200)
      })
  },
  getServiceStatus: (id: string): Promise<GetServiceStatusResponse> => {
       return new Promise((resolve) => {
          setTimeout(() => {
              resolve({ metrics: (mockMetrics[id] || {}) as any, tools: [] });
          }, 400);
      });
  }
};
