/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { UpstreamServiceConfig, ToolDefinition } from "@/lib/types";

/**
 * useServiceSiblings.
 *
 * @param currentServiceId - The currentServiceId.
 */
export function useServiceSiblings(currentServiceId: string) {
    const [siblings, setSiblings] = useState<{ label: string; href: string }[]>([]);

    useEffect(() => {
        apiClient.listServices().then((services: UpstreamServiceConfig[]) => {
            const list = Array.isArray(services) ? services : [];
            // ⚡ BOLT: Optimized O(2N) filter/map chain to O(N) single-pass reduce to avoid intermediate array allocation
            // Randomized Selection from Top 5 High-Impact Targets (Algorithmic)
            setSiblings(
                list.reduce<{ label: string; href: string }[]>((acc, s) => {
                    if (s.id !== currentServiceId) {
                        acc.push({ label: s.name, href: `/service/${s.id}` });
                    }
                    return acc;
                }, [])
            );
        });
    }, [currentServiceId]);

    return siblings;
}

/**
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
            // ⚡ BOLT: Optimized O(2N) filter/map chain to O(N) single-pass reduce to avoid intermediate array allocation
            // Randomized Selection from Top 5 High-Impact Targets (Algorithmic)
            setSiblings(
                tools.reduce<{ label: string; href: string }[]>((acc, t) => {
                    if (t.serviceId === serviceId && t.name !== decodedName) {
                        acc.push({ label: t.name, href: `/service/${serviceId}/tool/${t.name}` });
                    }
                    return acc;
                }, [])
            );
        });
    }, [serviceId, currentToolName]);

    return siblings;
}
