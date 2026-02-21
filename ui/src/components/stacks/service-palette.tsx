/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React from "react";
import {
    Database,
    HardDrive,
    MessageSquare,
    Globe,
    Server,
    Terminal,
    Cpu,
    Search,
    Plus
} from "lucide-react";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { apiClient, ServiceTemplate as BackendServiceTemplate } from "@/lib/client";
import yaml from "js-yaml";

/**
 * ServiceTemplate type definition for UI.
 */
export interface UiServiceTemplate {
    id: string;
    name: string;
    description: string;
    icon: React.ElementType;
    category: string;
    yamlSnippet: string;
}

interface ServicePaletteProps {
    onTemplateSelect: (snippet: string) => void;
}

const ICON_MAP: Record<string, React.ElementType> = {
    "database": Database,
    "hard-drive": HardDrive,
    "message-square": MessageSquare,
    "globe": Globe,
    "terminal": Terminal,
    "cpu": Cpu,
    "server": Server,
    // Add more mappings as needed
    "github": Globe,
    "gitlab": Globe,
    "notion": Globe,
    "linear": Globe,
    "jira": Globe,
    "google-calendar": Globe,
    "slack": MessageSquare
};

const mapTagsToCategory = (tags: string[] = []): string => {
    if (tags.includes("database")) return "Database";
    if (tags.includes("mcp")) return "MCP Server";
    if (tags.includes("utility")) return "Utility";
    if (tags.includes("ai")) return "AI";
    return "Other";
};

// Helper to clean up the config object for cleaner YAML
const cleanConfig = (obj: any): any => {
    if (Array.isArray(obj)) {
        return obj.map(cleanConfig);
    }
    if (typeof obj === 'object' && obj !== null) {
        const newObj: any = {};
        for (const key in obj) {
            const value = obj[key];
            // Skip empty arrays, null, undefined, empty strings (except name/id maybe?)
            if (value === null || value === undefined || (Array.isArray(value) && value.length === 0) || value === "") {
                continue;
            }
            // Handle SecretValue simplification for display if possible
            // But we want valid config.
            // If secret is { plainText: "${VAR}" }, we might want to just output "${VAR}" if the loader supported it,
            // but for now let's keep the structure or try to simplify if we are sure.
            // Actually, for better UX, we might want to manually format some things, but that's risky.
            // Let's just recurse.
            newObj[key] = cleanConfig(value);
        }
        return newObj;
    }
    return obj;
};

const generateYamlSnippet = (template: BackendServiceTemplate): string => {
    // Generate valid YAML for the service config.
    // We wrap it in a list because the editor typically expects a list item.
    // We use cleanConfig to remove empty fields to make the snippet concise.
    const cleaned = cleanConfig(template.serviceConfig);

    // We dump it as a list of one item
    return yaml.dump([cleaned], { lineWidth: -1, noRefs: true });
};

/**
 * ServicePalette.
 *
 * @param { onTemplateSelect - The { onTemplateSelect.
 */
export function ServicePalette({ onTemplateSelect }: ServicePaletteProps) {
    const [search, setSearch] = React.useState("");
    const [filter, setFilter] = React.useState<string | null>(null);
    const [templates, setTemplates] = React.useState<UiServiceTemplate[]>([]);
    const [loading, setLoading] = React.useState(true);

    React.useEffect(() => {
        apiClient.listTemplates()
            .then(backendTemplates => {
                const uiTemplates = backendTemplates.map(t => ({
                    id: t.id,
                    name: t.name,
                    description: t.description || "",
                    icon: ICON_MAP[t.icon] || Server,
                    category: mapTagsToCategory(t.tags),
                    yamlSnippet: generateYamlSnippet(t)
                }));
                setTemplates(uiTemplates);
            })
            .catch(err => {
                console.error("Failed to load templates", err);
            })
            .finally(() => {
                setLoading(false);
            });
    }, []);

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
                    {loading && (
                        <div className="text-center py-8 text-xs text-muted-foreground">
                            Loading templates...
                        </div>
                    )}
                    {!loading && filtered.map(template => (
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
                    ))}
                    {!loading && filtered.length === 0 && (
                        <div className="text-center py-8 text-xs text-muted-foreground">
                            No templates found.
                        </div>
                    )}
                </div>
            </ScrollArea>
        </div>
    );
}
