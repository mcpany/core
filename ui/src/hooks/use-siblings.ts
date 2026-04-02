/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { UpstreamServiceConfig, ToolDefinition } from "@/lib/types";

/**
 * Summary: Computes previous and next service siblings to enable sequential navigation.
 *
 * Params:
 *   - currentServiceId (string): The ID of the currently active service.
 *
 * Returns:
 *   - Object: { prev: ServiceRegistryItem | null, next: ServiceRegistryItem | null }
 *
 * Errors:
 *   - N/A: Safe computation; returns null if no valid sibling exists.
 *
 * Side Effects:
 *   - None: Computes sequentially based on the static `SERVICE_REGISTRY`.
 */
export function useServiceSiblings(currentServiceId: string) {
    const [siblings, setSiblings] = useState<{ label: string; href: string }[]>([]);

    useEffect(() => {
        apiClient.listServices().then((services: UpstreamServiceConfig[]) => {
            const list = Array.isArray(services) ? services : [];
            setSiblings(list
                .filter((s) => s.id !== currentServiceId)
                .map((s) => ({ label: s.name, href: `/service/${s.id}` }))
            );
        });
    }, [currentServiceId]);

    return siblings;
}

/**
 * Summary: Computes previous and next tool siblings to enable sequential tool navigation within a service.
 *
 * Params:
 *   - serviceId (string): The ID of the parent service.
 *   - currentToolName (string): The name of the active tool.
 *
 * Returns:
 *   - Object: { prevTool: string | null, nextTool: string | null }
 *
 * Errors:
 *   - N/A: Returns null if no valid sibling exists or service is missing.
 *
 * Side Effects:
 *   - None: Computed purely.
 */
export function useToolSiblings(serviceId: string, currentToolName: string) {
    const [siblings, setSiblings] = useState<{ label: string; href: string }[]>([]);

    useEffect(() => {
        apiClient.listTools().then((res: { tools?: ToolDefinition[] }) => {
            const tools = res.tools || [];
            const decodedName = decodeURIComponent(currentToolName);
            setSiblings(tools
                .filter((t) => t.serviceId === serviceId && t.name !== decodedName)
                .map((t) => ({ label: t.name, href: `/service/${serviceId}/tool/${t.name}` }))
            );
        });
    }, [serviceId, currentToolName]);

    return siblings;
}
