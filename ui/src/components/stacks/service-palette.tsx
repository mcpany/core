/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useEffect, useState } from "react";
import {
    Database,
    HardDrive,
    MessageSquare,
    Globe,
    Server,
    Terminal,
    Cpu,
    Search,
    Plus,
    Activity,
    Box
} from "lucide-react";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { apiClient, ServiceTemplate as BackendServiceTemplate } from "@/lib/client";
import yaml from "js-yaml";

/**
 * ServiceTemplate type definition for UI.
 */
export interface ServiceTemplate {
    id: string;
    name: string;
    description: string;
    icon: React.ElementType;
    category: string;
    yamlSnippet: string;
}

const ICON_MAP: Record<string, React.ElementType> = {
    "database": Database,
    "hard-drive": HardDrive,
    "message-square": MessageSquare,
    "slack": MessageSquare,
    "globe": Globe,
    "server": Server,
    "terminal": Terminal,
    "cpu": Cpu,
    "notion": Globe, // Placeholder
    "linear": Activity,
    "jira": Activity,
    "github": Globe,
    "gitlab": Globe,
    "google-calendar": Globe,
};

interface ServicePaletteProps {
    onTemplateSelect: (snippet: string) => void;
}

/**
 * ServicePalette.
 *
 * @param { onTemplateSelect - The { onTemplateSelect.
 */
export function ServicePalette({ onTemplateSelect }: ServicePaletteProps) {
    const [templates, setTemplates] = useState<ServiceTemplate[]>([]);
    const [search, setSearch] = React.useState("");
    const [filter, setFilter] = React.useState<string | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchTemplates = async () => {
            try {
                // Use getServiceTemplates which handles mapping from snake_case to camelCase
                const backendTemplates = await apiClient.getServiceTemplates();
                const uiTemplates = backendTemplates.map(t => mapBackendTemplateToUI(t));
                setTemplates(uiTemplates);
            } catch (err) {
                console.error("Failed to fetch templates:", err);
            } finally {
                setLoading(false);
            }
        };
        fetchTemplates();
    }, []);

    const mapBackendTemplateToUI = (t: BackendServiceTemplate): ServiceTemplate => {
        // Handle case where serviceConfig might be missing or raw
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const rawConfig = t.serviceConfig || (t as any).service_config || {};

        // Strip internal fields
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        const { id, sanitizedName, configError, lastError, toolCount, ...config } = rawConfig as any;

        // Wrap in list item
        const serviceItem = {
            name: config.name || t.name,
            ...config
        };

        // Generate the snippet
        // We use a list to match the "upstream_services" list structure expected by the editor
        // The snippet is appended to the existing YAML, so it should look like a list item.
        const snippetObj = [serviceItem];
        const yamlSnippet = yaml.dump(snippetObj, { indent: 2, lineWidth: -1 });

        // Map category from tags
        let category = "Other";
        if (t.tags && t.tags.length > 0) {
            // Capitalize first tag
            category = t.tags[0].charAt(0).toUpperCase() + t.tags[0].slice(1);
        }

        return {
            id: t.id,
            name: t.name,
            description: t.description,
            icon: ICON_MAP[t.icon] || Box,
            category: category,
            yamlSnippet: yamlSnippet
        };
    };

    const filtered = templates.filter(t => {
        const matchesSearch = t.name.toLowerCase().includes(search.toLowerCase()) ||
                              t.description.toLowerCase().includes(search.toLowerCase());
        const matchesFilter = filter ? t.category === filter : true;
        return matchesSearch && matchesFilter;
    });

    const categories = Array.from(new Set(templates.map(t => t.category)));

    return (
        <div className="flex flex-col h-full bg-muted/10 border-r w-[280px]">
            <div className="p-4 space-y-4 border-b">
                <div className="flex items-center gap-2 font-semibold">
                    <Server className="h-5 w-5 text-primary" />
                    <span>Service Palette</span>
                </div>
                <div className="relative">
                    <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                        placeholder="Search templates..."
                        value={search}
                        onChange={e => setSearch(e.target.value)}
                        className="pl-8 h-9 text-xs"
                    />
                </div>
                <div className="flex flex-wrap gap-1">
                    <Badge
                        variant={filter === null ? "secondary" : "outline"}
                        className="cursor-pointer text-[10px] hover:bg-muted"
                        onClick={() => setFilter(null)}
                    >
                        All
                    </Badge>
                    {categories.map(cat => (
                        <Badge
                            key={cat}
                            variant={filter === cat ? "secondary" : "outline"}
                            className="cursor-pointer text-[10px] hover:bg-muted"
                            onClick={() => setFilter(cat === filter ? null : cat)}
                        >
                            {cat}
                        </Badge>
                    ))}
                </div>
            </div>

            <ScrollArea className="flex-1">
                <div className="p-3 grid gap-2">
                    {loading ? (
                        <div className="text-center py-8 text-xs text-muted-foreground">
                            Loading templates...
                        </div>
                    ) : filtered.length === 0 ? (
                        <div className="text-center py-8 text-xs text-muted-foreground">
                            No templates found.
                        </div>
                    ) : (
                        filtered.map(template => (
                            <Card
                                key={template.id}
                                className="cursor-pointer transition-all hover:bg-accent hover:border-primary/50 group"
                                onClick={() => onTemplateSelect(template.yamlSnippet)}
                            >
                                <CardContent className="p-3 flex items-start gap-3">
                                    <div className="mt-1 p-2 bg-muted rounded-md group-hover:bg-background transition-colors">
                                        <template.icon className="h-4 w-4 text-muted-foreground group-hover:text-primary" />
                                    </div>
                                    <div className="space-y-1">
                                        <div className="flex items-center justify-between">
                                            <h4 className="font-medium text-xs">{template.name}</h4>
                                            <Plus className="h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity text-primary" />
                                        </div>
                                        <p className="text-[10px] text-muted-foreground leading-tight">
                                            {template.description}
                                        </p>
                                    </div>
                                </CardContent>
                            </Card>
                        ))
                    )}
                </div>
            </ScrollArea>
        </div>
    );
}
