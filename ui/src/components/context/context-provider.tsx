/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { createContext, useContext, useState, useEffect, useMemo } from "react";
import { apiClient, ToolDefinition, UpstreamServiceConfig } from "@/lib/client";
import { estimateTokens } from "@/lib/tokens";

interface ContextState {
    tools: ToolDefinition[];
    services: UpstreamServiceConfig[];
    loading: boolean;

    // Simulation State
    disabledToolIds: Set<string>; // ID = serviceId + "." + toolName
    toggleTool: (serviceId: string, toolName: string) => void;
    enableAll: () => void;

    // Stats
    totalTokens: number;
    projectedTokens: number;

    // Helper to get cost of a specific tool
    getToolCost: (tool: ToolDefinition) => number;
}

const ContextContext = createContext<ContextState | undefined>(undefined);

export function ContextProvider({ children }: { children: React.ReactNode }) {
    const [tools, setTools] = useState<ToolDefinition[]>([]);
    const [services, setServices] = useState<UpstreamServiceConfig[]>([]);
    const [loading, setLoading] = useState(true);
    const [disabledToolIds, setDisabledToolIds] = useState<Set<string>>(new Set());

    // Load data
    useEffect(() => {
        const load = async () => {
            setLoading(true);
            try {
                const [toolsRes, servicesRes] = await Promise.all([
                    apiClient.listTools(),
                    apiClient.listServices()
                ]);

                const toolList = toolsRes.tools || [];
                setTools(toolList);

                const serviceList = Array.isArray(servicesRes) ? servicesRes : [];
                setServices(serviceList);

                // Initialize disabled state based on backend 'disable' flag?
                // For now, we start with what the backend says is enabled.
                // But the simulator allows us to toggle *without* writing to backend immediately?
                // "Simulator" implies simulated state.
                // Let's assume we start with everything enabled (or respecting tool.disable)
                const initialDisabled = new Set<string>();
                toolList.forEach(t => {
                    if (t.disable) {
                        initialDisabled.add(`${t.serviceId}.${t.name}`);
                    }
                });
                setDisabledToolIds(initialDisabled);

            } catch (e) {
                console.error("Failed to load context data", e);
            } finally {
                setLoading(false);
            }
        };
        load();
    }, []);

    const getToolCost = (tool: ToolDefinition) => {
        // Memoize if needed, but estimateTokens is fast enough for now
        return estimateTokens(JSON.stringify(tool));
    };

    const toggleTool = (serviceId: string, toolName: string) => {
        const id = `${serviceId}.${toolName}`;
        setDisabledToolIds(prev => {
            const next = new Set(prev);
            if (next.has(id)) {
                next.delete(id);
            } else {
                next.add(id);
            }
            return next;
        });
    };

    const enableAll = () => {
        setDisabledToolIds(new Set());
    };

    const totalTokens = useMemo(() => {
        return tools.reduce((acc, tool) => acc + getToolCost(tool), 0);
    }, [tools]);

    const projectedTokens = useMemo(() => {
        return tools.reduce((acc, tool) => {
            const id = `${tool.serviceId}.${tool.name}`;
            if (disabledToolIds.has(id)) return acc;
            return acc + getToolCost(tool);
        }, 0);
    }, [tools, disabledToolIds]);

    const value = {
        tools,
        services,
        loading,
        disabledToolIds,
        toggleTool,
        enableAll,
        totalTokens,
        projectedTokens,
        getToolCost
    };

    return (
        <ContextContext.Provider value={value}>
            {children}
        </ContextContext.Provider>
    );
}

export function useRecursiveContext() {
    const context = useContext(ContextContext);
    if (!context) {
        throw new Error("useRecursiveContext must be used within a ContextProvider");
    }
    return context;
}
