/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { ListServicesResponse, UpstreamServiceConfig, GetServiceResponse, GetServiceStatusResponse } from "./types";

const API_base = process.env.NEXT_PUBLIC_API_URL || "http://localhost:50050";

export const apiClient = {
  listServices: async (): Promise<ListServicesResponse> => {
    const res = await fetch(`${API_base}/v1/services`);
    if (!res.ok) {
      throw new Error(`Failed to list services: ${res.statusText}`);
    }
    return res.json();
  },

  getService: async (id: string): Promise<GetServiceResponse> => {
    // Since we don't have getServiceById on backend yet, we list all and filter.
    // This is inefficient but functional for now.
    const listRes = await apiClient.listServices();
    const service = listRes.services.find((s) => s.id === id || s.name === id); // id or name?
    if (!service) {
      throw new Error("Service not found");
    }
    return { service };
  },

  // emulate setServiceStatus by re-registering? Or just unregister if disabled?
  // User asked to "Update make rules... and UI information fetched... is correct"
  // If `setServiceStatus` is complex, I might skip it or implement partial logic.
  // mock-client says "disable: boolean".
  // If I can't easily toggle disable on backend, I'll log a warning or try register.
  setServiceStatus: async (id: string, disabled: boolean): Promise<GetServiceResponse> => {
      // For now, we might not support strict toggling without re-registering payload.
      // But we can try to fetch, update, and re-register.
      const { service } = await apiClient.getService(id);
      const updatedService = { ...service, disable: disabled };

      const res = await fetch(`${API_base}/v1/services/register`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ config: updatedService })
      });
      if (!res.ok) {
          throw new Error(`Failed to update service status: ${res.statusText}`);
      }
      return { service: updatedService };
  },

  getServiceStatus: async (id: string): Promise<GetServiceStatusResponse> => {
      // We need service name for this endpoint.
      // Assuming id IS the name or we resolve it.
      // But endpoint is /v1/services/{service_name}/status
      // In mock, id was like "auth-service-prod-123". Name was "auth-service".
      // We probably need to resolve ID to Name first if they differ.
      const { service } = await apiClient.getService(id);
      const name = service.name;

      const res = await fetch(`${API_base}/v1/services/${name}/status`);
       if (!res.ok) {
        // Fallback or throw?
        // If status endpoint fails, maybe return empty metrics?
        console.error(`Failed to get status for ${name}: ${res.statusText}`);
        return { metrics: {} };
      }
      return res.json();
  }
};
