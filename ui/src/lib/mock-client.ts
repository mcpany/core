/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { UpstreamServiceConfig } from "@/lib/client";

// Mock types no longer needed as we match client.ts signatures
// import { ListServicesResponse, UpstreamServiceConfig, GetServiceResponse, GetServiceStatusResponse } from "./types";

const mockServices: UpstreamServiceConfig[] = [
  {
    id: "auth-service-prod-123",
    name: "auth-service",
    version: "1.2.1",
    disable: false,
    priority: 0,
    grpcService: {
      address: "auth.prod.mcpany.io:443",
      useReflection: true,
      tlsConfig: { serverName: "auth.prod.mcpany.io", insecureSkipVerify: false } as any,
      tools: [
        { name: "Login", description: "Authenticate a user." },
        { name: "Logout", description: "Log out a user." },
        { name: "CheckStatus", description: "Check authentication status." },
      ] as any,
      resources: [
        { name: "User", uri: "user://profile" },
        { name: "Session", uri: "user://session" },
      ] as any,
      prompts: [],
       calls: {},
       protoDefinitions: [],
        protoCollection: [],
       healthCheck: undefined
    }
  },
  {
    id: "payment-gateway-prod-456",
    name: "payment-gateway",
    version: "2.0.0",
    disable: false,
    priority: 0,
    httpService: {
      address: "https://payments.prod.mcpany.io",
      tlsConfig: { serverName: "payments.prod.mcpany.io", insecureSkipVerify: false } as any,
      tools: [
        { name: "CreatePayment", description: "Create a new payment." },
        { name: "GetPaymentStatus", description: "Get the status of a payment." },
      ] as any,
      prompts: [
        { name: "GenerateInvoice", description: "Generate a PDF invoice for a payment." }
      ] as any,
      resources: [],
      calls: {},
      healthCheck: undefined
    }
  },
  {
    id: "user-profiles-staging-789",
    name: "user-profiles",
    version: "0.5.0-beta",
    disable: true,
    priority: 0,
    grpcService: {
      address: "users.staging.mcpany.io:8080",
      useReflection: false,
       tools: [],
       resources: [],
       prompts: [],
       calls: {},
       protoDefinitions: [],
       protoCollection: [],
       healthCheck: undefined,
       tlsConfig: undefined
    }
  },
  {
    id: "inventory-service-dev-101",
    name: "inventory-service",
    version: "0.1.0-dev",
    disable: false,
    priority: 0,
    commandLineService: {
      command: "go run ./cmd/inventory",
      workingDirectory: ".",
      tools: [
        { name: "CheckStock", description: "Check stock levels for an item." }
      ] as any,
      resources: [],
      prompts: [],
      calls: {},
      healthCheck: undefined,
      cache: undefined,
      containerEnvironment: undefined,
      timeout: undefined,
      local: true,
      communicationProtocol: 0,
      env: {}
    }
  },
] as any;

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
  listServices: (): Promise<UpstreamServiceConfig[]> => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve([...mockServices]);
      }, 500);
    });
  },
  getService: (id: string): Promise<UpstreamServiceConfig> => {
      return new Promise((resolve, reject) => {
          setTimeout(() => {
              const service = mockServices.find(s => s.id === id);
              if (service) {
                  resolve({...service});
              } else {
                  reject(new Error("Service not found"));
              }
          }, 300);
      });
  },
  setServiceStatus: (id: string, disabled: boolean): Promise<any> => {
      return new Promise((resolve, reject) => {
          setTimeout(() => {
              const updatedService = updateService(id, { disable: disabled });
              if (updatedService) {
                resolve({ service: updatedService }); // Keep as object for now if UI expects it, or unwrap?
                // Checking client.ts: setServiceStatus returns res.json().
                // And in client.ts usage... let's check.
                // client.ts: return res.json();
                // usage: does it use the return value?
              } else {
                reject(new Error("Service not found to update status"));
              }
          }, 200)
      })
  },
  getServiceStatus: (id: string): Promise<any> => {
       return new Promise((resolve) => {
          setTimeout(() => {
              resolve({ metrics: mockMetrics[id] || {} });
          }, 400);
      });
  },
    registerService: async (config: UpstreamServiceConfig) => {
        return config;
    },
    updateService: async (config: UpstreamServiceConfig) => {
        return config;
    },
    unregisterService: async (id: string) => {
        return {};
    },
    listTools: async () => Promise.resolve([]),
    executeTool: async () => Promise.resolve({}),
    setToolStatus: async () => Promise.resolve({}),
    listResources: async () => Promise.resolve([]),
    setResourceStatus: async () => Promise.resolve({}),
    listPrompts: async () => Promise.resolve([]),
    setPromptStatus: async () => Promise.resolve({}),
    listSecrets: async () => Promise.resolve([]),
    saveSecret: async () => Promise.resolve({} as any),
    deleteSecret: async () => Promise.resolve({}),
    getGlobalSettings: async () => Promise.resolve({} as any),
    saveGlobalSettings: async () => Promise.resolve(),
    getStackConfig: async () => Promise.resolve(""),
    saveStackConfig: async () => Promise.resolve({})
};
