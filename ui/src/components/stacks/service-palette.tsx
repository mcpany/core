/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

import { useState, useEffect } from "react";
import { ServiceTemplate } from "@/lib/client";
import { apiClient } from "@/lib/client";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Search, Plus, Server, Code, Terminal, Webhook } from "lucide-react";
import YAML from "yaml";

interface ServicePaletteProps {
    onTemplateSelect: (yamlSnippet: string) => void;
}

const getIcon = (template: ServiceTemplate) => {
    if (template.tags.includes('database')) return Server;
    if (template.tags.includes('api')) return Webhook;
    if (template.tags.includes('local')) return Terminal;
    return Code;
};

/**
 * ServicePalette component.
 * @param props - The component props.
 * @param props.onTemplateSelect - The callback when a template is selected.
 * @returns The rendered component.
 */
export function ServicePalette({ onTemplateSelect }: ServicePaletteProps) {
    const [search, setSearch] = useState("");
    const [templates, setTemplates] = useState<ServiceTemplate[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const fetchTemplates = async () => {
            try {
                setLoading(true);
                const data = await apiClient.listTemplates();
                // Using \`YAML.stringify\` to properly construct a YAML snippet.

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
        const configToYaml: any = {
            name: t.serviceConfig.name || t.name.toLowerCase().replace(/\s+/g, '-')
        };

        if (t.serviceConfig.commandLineService) {
            configToYaml.command = t.serviceConfig.commandLineService.command;
            if (t.serviceConfig.commandLineService.workingDirectory) {
                configToYaml.working_dir = t.serviceConfig.commandLineService.workingDirectory;
            }
            if (t.serviceConfig.commandLineService.env && Object.keys(t.serviceConfig.commandLineService.env).length > 0) {
                configToYaml.environment = t.serviceConfig.commandLineService.env;
            }
        } else if (t.serviceConfig.mcpService) {
            configToYaml.mcp = { url: t.serviceConfig.mcpService.url };
        }

        const yamlString = YAML.stringify([configToYaml]);
        return yamlString.split('\n').map(line => `  ${line}`).join('\n');
    };

    const filtered = templates.filter(t =>
        t.name.toLowerCase().includes(search.toLowerCase()) ||
        t.description.toLowerCase().includes(search.toLowerCase()) ||
        t.tags.some(tag => tag.toLowerCase().includes(search.toLowerCase()))
    );

    return (
        <div className="flex flex-col h-full bg-background border-l">
            <div className="p-4 border-b">
                <h3 className="font-semibold mb-3 flex items-center gap-2">
                    <Server className="h-4 w-4" />
                    Service Catalog
                </h3>
                <div className="relative">
                    <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                        placeholder="Search services..."
                        className="pl-8 bg-muted/50"
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                    />
                </div>
            </div>

            <ScrollArea className="flex-1 p-4">
                <div className="space-y-4">
                    {loading && (
                        <div className="flex items-center justify-center h-24 text-muted-foreground text-sm">
                            Loading catalog...
                        </div>
                    )}
                    {error && (
                        <div className="text-destructive text-sm text-center p-4">
                            {error}
                        </div>
                    )}
                    {!loading && !error && (
                        <div className="grid gap-3">
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
                                No services match your search.
                            </div>
                        )}
                        </div>
                    )}
                </div>
            </ScrollArea>
        </div>
    );
}
