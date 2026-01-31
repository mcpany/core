/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { UpstreamServiceConfig, ToolDefinition } from "@/lib/types";

/**
 * Hook to fetch and format a list of sibling services for navigation.
 *
 * It retrieves all available services excluding the current one,
 * useful for displaying "other services" or navigation menus.
 *
 * @param currentServiceId - The ID of the current service to exclude from the list.
 * @returns An array of objects containing the label and href for each sibling service.
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
 * Hook to fetch and format a list of sibling tools within the same service.
 *
 * It retrieves all tools belonging to the specified service excluding the current one,
 * useful for "other tools in this service" navigation.
 *
 * @param serviceId - The ID of the service containing the tools.
 * @param currentToolName - The name of the current tool to exclude from the list.
 * @returns An array of objects containing the label and href for each sibling tool.
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
