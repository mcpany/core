/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */



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
    Loader2
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { apiClient, ServiceTemplate } from "@/lib/client";
import yaml from "js-yaml";

// Map icons by name or category if dynamic
const iconMap: Record<string, React.ElementType> = {
    "postgres": Database,
    "redis": Database,
    "filesystem": HardDrive,
    "slack": MessageSquare,
    "memory": Cpu,
    "generic-http": Globe,
    "generic-cmd": Terminal,
    "database": Database,
    "mcp server": Server,
    "utility": Terminal,
    "ai": Cpu
};

// Fallback icon
const DefaultIcon = Server;

interface ServicePaletteProps {
    onTemplateSelect: (snippet: string) => void;
}

/**
 * Intent: Document ServicePalette
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
 * ServicePalette component.
 * Fetches service templates from the API and displays them for selection.
 *
 * @param {ServicePaletteProps} props - Component props.
 */
export function ServicePalette({ onTemplateSelect }: ServicePaletteProps) {
    const [search, setSearch] = useState("");
    const [filter, setFilter] = useState<string | null>(null);
    const [templates, setTemplates] = useState<ServiceTemplate[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchTemplates = async () => {
            try {
                setLoading(true);
                const data = await apiClient.listTemplates();
                // Map API response to UI model if needed, but ServiceTemplate matches closely.
                // We might need to generate yamlSnippet if not present?
                // The API ServiceTemplate has `serviceConfig`. We need to convert it to YAML snippet.
                // For now, let's assume the API returns what we need or we construct it.
                // Wait, the API `ServiceTemplate` definition in client.ts doesn't have `yamlSnippet`.
                // It has `serviceConfig: UpstreamServiceConfig`.
                // We need to serialize `serviceConfig` to YAML.
                // Ideally `apiClient` or backend handles this, or we do it here.
                // Construct a proper YAML snippet from the template configuration using js-yaml.
                setTemplates(data);
            } catch (err) {
                console.error("Failed to fetch templates", err);
                setError("Failed to load templates");
            } finally {
                setLoading(false);
            }
        };
        fetchTemplates();
    }, []);

    const generateYamlSnippet = (t: ServiceTemplate): string => {
        const serviceDef: any = {
            name: t.serviceConfig.name || t.name.toLowerCase().replace(/\s+/g, '-')
        };

        if (t.serviceConfig.commandLineService) {
            serviceDef.command = t.serviceConfig.commandLineService.command;

            if (t.serviceConfig.commandLineService.workingDirectory) {
                serviceDef.working_dir = t.serviceConfig.commandLineService.workingDirectory;
            }

            if (t.serviceConfig.commandLineService.env && Object.keys(t.serviceConfig.commandLineService.env).length > 0) {
                serviceDef.environment = {};
                for (const [k, v] of Object.entries(t.serviceConfig.commandLineService.env)) {
                     // The backend might return an object like { plainText: "..." } or a string.
                     if (typeof v === 'string') {
                         serviceDef.environment[k] = v;
                     } else if (v && typeof v === 'object' && ('plainText' in v || 'plain_text' in v)) {
                         serviceDef.environment[k] = (v as any).plainText || (v as any).plain_text;
                     } else {
                         serviceDef.environment[k] = v;
                     }
                }
            }
        } else if (t.serviceConfig.httpService) {
             serviceDef.url = t.serviceConfig.httpService.address;
        }

        // We use an array so it outputs the correct list format: `- name: ...`
        const yamlStr = yaml.dump([serviceDef], {
            indent: 2,
            lineWidth: -1,
            noRefs: true
        });

        // yaml.dump with array puts it at root level. Usually the stack expects it indented under `upstream_services:`.
        // The previous code returned `  - name: ...`. Let's prepend spaces to match indentation requirement.
        return yamlStr.split('\n').filter(line => line.length > 0).map(line => `  ${line}`).join('\n') + '\n';
    };

    const getIcon = (t: ServiceTemplate) => {
        // Try id, then name, then category
        if (iconMap[t.id]) return iconMap[t.id];
        const cat = t.tags && t.tags.length > 0 ? t.tags[0].toLowerCase() : "utility";
        if (iconMap[cat]) return iconMap[cat];
        return DefaultIcon;
    };

    const filtered = templates.filter(t => {
        const matchesSearch = t.name.toLowerCase().includes(search.toLowerCase()) ||
                              t.description.toLowerCase().includes(search.toLowerCase());
        // Category is not strict in API template, use tags?
        // Let's assume tags[0] is category for now or map it.
        const category = t.tags && t.tags.length > 0 ? t.tags[0] : "Other";
        const matchesFilter = filter ? category === filter : true;
        return matchesSearch && matchesFilter;
    });

    const categories = Array.from(new Set(templates.map(t => t.tags && t.tags.length > 0 ? t.tags[0] : "Other")));

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
                    <div className="flex items-center justify-center py-8">
                        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                    </div>
                ) : error ? (
                    <div className="text-center py-8 text-xs text-destructive">
                        {error}
                    </div>
                ) : (
                    <div className="p-3 grid gap-2">
                        {filtered.map(template => {
                            const Icon = getIcon(template);
                            return (
                                <Card
                                    key={template.id}
                                    className="cursor-pointer transition-all hover:bg-accent hover:border-primary/50 group"
                                    onClick={() => onTemplateSelect(generateYamlSnippet(template))}
                                >
                                    <CardContent className="p-3 flex items-start gap-3">
                                        <div className="mt-1 p-2 bg-muted rounded-md group-hover:bg-background transition-colors">
                                            <Icon className="h-4 w-4 text-muted-foreground group-hover:text-primary" />
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
                            );
                        })}
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
