/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/**
 * Mock GrpcWebImpl for tests to avoid actual grpc-web client dependencies
 */
export class GrpcWebImpl {
  constructor(url: string, options: any) {
    // mock implementation
  }
}

/**
 * Mock RegistrationServiceClientImpl for tests
 */
export class RegistrationServiceClientImpl {
  constructor(rpc: any) {
    // mock implementation
  }

  RegisterUpstreamService(request: any): Promise<any> {
    return Promise.resolve({ service: {} });
  }

  GetUpstreamService(request: any): Promise<any> {
    return Promise.resolve({ service: {} });
  }

  UpdateUpstreamService(request: any): Promise<any> {
    return Promise.resolve({ service: {} });
  }

  UnregisterUpstreamService(request: any): Promise<any> {
    return Promise.resolve({});
  }

  ListUpstreamServices(request: any): Promise<any> {
    return Promise.resolve({ services: [] });
  }
}

/**
 * Mock enum values for tests
 */
export enum CallPolicyRule_Action {
  ALLOW = 0,
  DENY = 1,
}

/** Mock enum */
export enum ExportRule_Action {
  EXPORT = 0,
  UNEXPORT = 1,
}

/**
 * Mock type placeholders for policy-related proto messages.
 */
export const CallPolicy = {};
/** Mock type placeholder for CallPolicyRule */
export const CallPolicyRule = {};
/** Mock type placeholder for ExportPolicy */
export const ExportPolicy = {};
/** Mock type placeholder for ExportRule */
export const ExportRule = {};

/**
 * Mock constant for ProfileDefinition used in tests.
 */
export const ProfileDefinition = {};

/**
 * Mock ToolDefinition used in tests
 */
export const ToolDefinition = {};

/**
 * Mock UpstreamServiceConfig used in tests
 */
export const UpstreamServiceConfig = {};

/**
 * Mock ServiceTemplate used in tests
 */
export const ServiceTemplate = {};

/**
 * Mock ResourceDefinition used in tests
 */
export const ResourceDefinition = {};

/**
 * Mock PromptDefinition used in tests
 */
export const PromptDefinition = {};

/**
 * Mock AdminServiceClientImpl for tests
 */
export class AdminServiceClientImpl {
  constructor(rpc: any) {
    // mock implementation
  }

  GetConfig(request: any): Promise<any> {
    return Promise.resolve({ config: {} });
  }

  UpdateConfig(request: any): Promise<any> {
    return Promise.resolve({ config: {} });
  }

  GetStatus(request: any): Promise<any> {
    return Promise.resolve({
      version: "mock-version",
      status: "running"
    });
  }
}
