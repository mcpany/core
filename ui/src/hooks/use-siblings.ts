/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { UpstreamServiceConfig, ToolDefinition } from "@/lib/types";

/**
 * useServiceSiblings serves as a public interface for interacting with useServiceSiblings.
 *
 * Summary: Use the service siblings appropriately based on current system conditions.
 *
 * Parameters:
 *   - Refer to the function signature for strongly-typed input arguments.
 *
 * Returns:
 *   - Returns the expected domain model or execution state.
 *
 * Throws/Errors:
 *   - Propagates exceptions from underlying validation layers.
 *
 * Side Effects:
 *   - May mutate state or perform network I/O depending on implementation.
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
 * useToolSiblings serves as a public interface for interacting with useToolSiblings.
 *
 * Summary: Use the tool siblings appropriately based on current system conditions.
 *
 * Parameters:
 *   - Refer to the function signature for strongly-typed input arguments.
 *
 * Returns:
 *   - Returns the expected domain model or execution state.
 *
 * Throws/Errors:
 *   - Propagates exceptions from underlying validation layers.
 *
 * Side Effects:
 *   - May mutate state or perform network I/O depending on implementation.
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
