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
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export interface ServiceTemplate {
    id: string;
    name: string;
    description: string;
    icon: React.ElementType;
    category: "Database" | "MCP Server" | "Utility" | "AI";
    yamlSnippet: string;
}

const TEMPLATES: ServiceTemplate[] = [
    {
        id: "postgres",
        name: "PostgreSQL",
        description: "Standard SQL Database",
        icon: Database,
        category: "Database",
        yamlSnippet: `  postgres-db:
    image: postgres:15
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: ${"${POSTGRES_PASSWORD}"}
      POSTGRES_DB: mydb
    ports:
      - "5432:5432"
`
    },
    {
        id: "redis",
        name: "Redis",
        description: "In-memory key-value store",
        icon: Database,
        category: "Database",
        yamlSnippet: `  redis-cache:
    image: redis:alpine
    ports:
      - "6379:6379"
`
    },
    {
        id: "filesystem",
        name: "Filesystem MCP",
        description: "Local file access",
        icon: HardDrive,
        category: "MCP Server",
        yamlSnippet: `  filesystem-mcp:
    command: npx -y @modelcontextprotocol/server-filesystem /path/to/allowed/dir
    environment:
      NODE_ENV: production
`
    },
    {
        id: "slack",
        name: "Slack MCP",
        description: "Slack integration",
        icon: MessageSquare,
        category: "MCP Server",
        yamlSnippet: `  slack-mcp:
    command: npx -y @modelcontextprotocol/server-slack
    environment:
      SLACK_BOT_TOKEN: \${SLACK_BOT_TOKEN}
      SLACK_SIGNING_SECRET: \${SLACK_SIGNING_SECRET}
`
    },
    {
        id: "memory",
        name: "Memory MCP",
        description: "Graph-based memory",
        icon: Cpu,
        category: "MCP Server",
        yamlSnippet: `  memory-mcp:
    command: npx -y @modelcontextprotocol/server-memory
`
    },
    {
        id: "generic-http",
        name: "HTTP Service",
        description: "Generic HTTP API",
        icon: Globe,
        category: "Utility",
        yamlSnippet: `  my-api-service:
    image: my-repo/api:latest
    environment:
      PORT: 8080
`
    },
    {
        id: "generic-cmd",
        name: "Command Line",
        description: "Local script execution",
        icon: Terminal,
        category: "Utility",
        yamlSnippet: `  local-script:
    command: python3 ./scripts/worker.py
    working_dir: ./
`
    }
];

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

    const filtered = TEMPLATES.filter(t => {
        const matchesSearch = t.name.toLowerCase().includes(search.toLowerCase()) ||
                              t.description.toLowerCase().includes(search.toLowerCase());
        const matchesFilter = filter ? t.category === filter : true;
        return matchesSearch && matchesFilter;
    });

    const categories = Array.from(new Set(TEMPLATES.map(t => t.category)));

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
            </ScrollArea>
        </div>
    );
}
