/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import React, { useEffect } from "react";
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
    Loader2
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { apiClient } from "@/lib/client";

/**
 * ServiceTemplate type definition.
 */
export interface ServiceTemplate {
    id: string;
    name: string;
    description: string;
    icon: React.ElementType;
    category: "Database" | "MCP Server" | "Utility" | "AI" | string;
    yamlSnippet: string;
}

interface ServicePaletteProps {
    onTemplateSelect: (snippet: string) => void;
}

/**
 * ServicePalette.
 *
 * @param { onTemplateSelect - The { onTemplateSelect.
 */
export function ServicePalette({ onTemplateSelect }: ServicePaletteProps) {
    const [search, setSearch] = React.useState("");
    const [filter, setFilter] = React.useState<string | null>(null);
    const [templates, setTemplates] = React.useState<ServiceTemplate[]>([]);
    const [loading, setLoading] = React.useState(true);

    useEffect(() => {
        const fetchTemplates = async () => {
            try {
                const data = await apiClient.listTemplates();
                const mapped: ServiceTemplate[] = data.map((t: any) => ({
                    id: t.id || t.name,
                    name: t.name,
                    description: t.description || "",
                    category: t.category || "Utility",
                    yamlSnippet: t.yamlSnippet || "",
                    icon: getIcon(t.category || "Utility")
                }));
                setTemplates(mapped);
            } catch (error) {
                console.error("Failed to fetch templates:", error);
            } finally {
                setLoading(false);
            }
        };

        fetchTemplates();
    }, []);

    const getIcon = (category: string) => {
        switch (category) {
            case "Database": return Database;
            case "MCP Server": return Server;
            case "Utility": return Globe;
            case "AI": return Cpu;
            default: return Server;
        }
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
                {loading ? (
                    <div className="flex items-center justify-center h-40">
                        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                    </div>
                ) : (
                    <div className="p-3 grid gap-2">
                        {filtered.map(template => (
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
                        {filtered.length === 0 && (
                            <div className="text-center py-8 text-xs text-muted-foreground">
                                No templates found.
                            </div>
                        )}
                    </div>
                )}
            </ScrollArea>
        </div>
    );
}
