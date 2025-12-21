/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  ListServicesResponse,
  GetServiceResponse,
  GetServiceStatusResponse,
  UpstreamServiceConfig,
  RegisterServiceRequest,
  RegisterServiceResponse,
  UpdateServiceRequest,
  UpdateServiceResponse,
  UnregisterServiceRequest,
  UnregisterServiceResponse
} from "./types";

const API_base = process.env.NEXT_PUBLIC_API_URL || "http://localhost:50050";

/**
 * API client for interacting with the MCP Any server.
 */
export const apiClient = {
  /**
   * Lists all registered upstream services.
   *
   * @returns A promise resolving to the list of services response.
   */
  listServices: async (): Promise<ListServicesResponse> => {
    const res = await fetch(`${API_base}/v1/services`);
    if (!res.ok) {
      throw new Error(`Failed to list services: ${res.statusText}`);
    }
    return res.json();
  },

  /**
   * Retrieves a single service configuration by name.
   *
   * @param name - The name of the service to retrieve.
   * @returns A promise resolving to the service details response.
   */
  getService: async (name: string): Promise<GetServiceResponse> => {
    // According to proto: GET /v1/services/{service_name}
    const res = await fetch(`${API_base}/v1/services/${encodeURIComponent(name)}`);
    if (!res.ok) {
      throw new Error(`Service not found or error: ${res.statusText}`);
    }
    return res.json();
  },

  /**
   * Registers a new upstream service.
   *
   * @param config - The configuration of the service to register.
   * @returns A promise resolving to the registration response.
   */
  registerService: async (config: UpstreamServiceConfig): Promise<RegisterServiceResponse> => {
    const payload: RegisterServiceRequest = { config };
    const res = await fetch(`${API_base}/v1/services/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    if (!res.ok) {
       const errorData = await res.json().catch(() => ({}));
       throw new Error(errorData.message || `Failed to register service: ${res.statusText}`);
    }
    return res.json();
  },

  /**
   * Updates an existing upstream service configuration.
   *
   * @param config - The new configuration for the service.
   * @returns A promise resolving to the update response.
   */
  updateService: async (config: UpstreamServiceConfig): Promise<UpdateServiceResponse> => {
    const payload: UpdateServiceRequest = { config };
     const res = await fetch(`${API_base}/v1/services/update`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
    if (!res.ok) {
       const errorData = await res.json().catch(() => ({}));
       throw new Error(errorData.message || `Failed to update service: ${res.statusText}`);
    }
    return res.json();
  },

  /**
   * Unregisters an upstream service.
   *
   * @param serviceName - The name of the service to unregister.
   * @returns A promise resolving to the unregistration response.
   */
  unregisterService: async (serviceName: string): Promise<UnregisterServiceResponse> => {
      const payload: UnregisterServiceRequest = { service_name: serviceName };
      const res = await fetch(`${API_base}/v1/services/unregister`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
      });
      if (!res.ok) {
         const errorData = await res.json().catch(() => ({}));
         throw new Error(errorData.message || `Failed to unregister service: ${res.statusText}`);
      }
      return res.json();
  },

  /**
   * Helper function to toggle the disabled status of a service.
   *
   * @param name - The name of the service.
   * @param disabled - Whether the service should be disabled.
   * @returns A promise resolving to the update response.
   */
  setServiceStatus: async (name: string, disabled: boolean): Promise<UpdateServiceResponse> => {
      const { service } = await apiClient.getService(name);
      // We need to send the full config back.
      const updatedService: UpstreamServiceConfig = { ...service, disable: disabled };
      return apiClient.updateService(updatedService);
  },

  /**
   * Retrieves the runtime status and metrics of a service.
   *
   * @param name - The name of the service.
   * @returns A promise resolving to the service status response.
   */
  getServiceStatus: async (name: string): Promise<GetServiceStatusResponse> => {
      const res = await fetch(`${API_base}/v1/services/${encodeURIComponent(name)}/status`);
       if (!res.ok) {
        console.error(`Failed to get status for ${name}: ${res.statusText}`);
        return { metrics: {} };
      }
      return res.json();
  }
};
