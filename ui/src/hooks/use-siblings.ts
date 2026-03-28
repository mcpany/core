/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { UpstreamServiceConfig, ToolDefinition } from "@/lib/types";

/**
 * Intent: Document useServiceSiblings
 *
 * Params:
 *   - Documented below.
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * useServiceSiblings.
 *
 * @param currentServiceId - The currentServiceId.
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
 * Intent: Document useToolSiblings
 *
 * Params:
 *   - Documented below.
 *
 * Returns:
 *   - None
 *
 * Errors:
 *   - None
 *
 * Side Effects:
 *   - None
 *
 * useToolSiblings.
 *
 * @param serviceId - The serviceId.
 * @param currentToolName - The currentToolName.
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
