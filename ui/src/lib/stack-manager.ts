/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

/* eslint-disable @typescript-eslint/no-explicit-any */

import { apiClient, UpstreamServiceConfig } from "@/lib/client";

export interface Stack {
  name: string;
  status: "Active" | "Partially Active" | "Inactive" | "Error";
  services: UpstreamServiceConfig[];
  serviceCount: number;
}

// Partial definition to handle loose inputs without strict import of ServiceCollection
interface ServiceCollectionLike {
    services?: UpstreamServiceConfig[] | Record<string, UpstreamServiceConfig>;
}

export const stackManager = {
  /**
   * List all stacks.
   * Group services by `stack:<name>` tag.
   */
  listStacks: async (): Promise<Stack[]> => {
    const services = await apiClient.listServices();
    const stacksMap: Record<string, UpstreamServiceConfig[]> = {};

    services.forEach((svc) => {
      const stackTag = svc.tags?.find((t) => t.startsWith("stack:"));
      if (stackTag) {
        const stackName = stackTag.split(":")[1];
        if (!stacksMap[stackName]) stacksMap[stackName] = [];
        stacksMap[stackName].push(svc);
      }
    });

    return Object.entries(stacksMap).map(([name, svcs]) => {
      // Determine status
      let running = 0;
      let error = 0;
      svcs.forEach((s) => {
        if (s.lastError) error++;
        else if (!s.disable) running++;
      });

      let status: Stack["status"] = "Inactive";
      if (error > 0) status = "Error";
      else if (running === svcs.length) status = "Active";
      else if (running > 0) status = "Partially Active";

      return {
        name,
        status,
        services: svcs,
        serviceCount: svcs.length,
      };
    });
  },

  /**
   * Get a single stack by name.
   */
  getStack: async (name: string): Promise<Stack | null> => {
    const services = await apiClient.listServices();
    const stackServices = services.filter((s) =>
      s.tags?.includes(`stack:${name}`)
    );

    if (stackServices.length === 0) return null;

    // Determine status (same logic)
    let running = 0;
    let error = 0;
    stackServices.forEach((s) => {
        if (s.lastError) error++;
        else if (!s.disable) running++;
    });
    let status: Stack["status"] = "Inactive";
    if (error > 0) status = "Error";
    else if (running === stackServices.length) status = "Active";
    else if (running > 0) status = "Partially Active";

    return {
      name,
      status,
      services: stackServices,
      serviceCount: stackServices.length,
    };
  },

  /**
   * Save a stack (Create or Update).
   * @param name Stack Name
   * @param config The stack configuration (list of services or ServiceCollection)
   */
  saveStack: async (name: string, config: UpstreamServiceConfig[] | ServiceCollectionLike | any) => {
    let newServices: UpstreamServiceConfig[] = [];
    if (Array.isArray(config)) {
      newServices = config;
    } else if (config && typeof config === 'object' && 'services' in config) {
        if (Array.isArray(config.services)) {
            newServices = config.services;
        } else {
             // Map format?
             newServices = Object.values(config.services || {});
        }
    }

    // fetch current stack to know what to update vs delete?
    const currentStack = await stackManager.getStack(name);
    const existingServices = currentStack ? currentStack.services : [];
    const existingNames = new Set(existingServices.map(s => s.name));
    const newNames = new Set(newServices.map(s => s.name));

    // 1. Register/Update new services
    for (const svc of newServices) {
      // Inject tag
      const tags = new Set(svc.tags || []);
      tags.add(`stack:${name}`);
      svc.tags = Array.from(tags);
      // Ensure ID logic if missing?
      if (!svc.id) svc.id = svc.name;

      // Ensure defaults if missing
      if (svc.disable === undefined) svc.disable = false;
      if (svc.version === undefined) svc.version = "1.0.0";

      if (existingNames.has(svc.name)) {
        await apiClient.updateService(svc);
      } else {
        await apiClient.registerService(svc);
      }
    }

    // 2. Delete removed services
    for (const oldSvc of existingServices) {
      if (!newNames.has(oldSvc.name)) {
        await apiClient.unregisterService(oldSvc.name);
      }
    }
  },

  /**
   * Delete a stack.
   */
  deleteStack: async (name: string) => {
    const stack = await stackManager.getStack(name);
    if (!stack) return;

    for (const svc of stack.services) {
      await apiClient.unregisterService(svc.name);
    }
  },
};
