/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { useState } from "react";

interface ServicePaletteProps {
    onTemplateSelect: (snippet: string) => void;
}

const TEMPLATES = [
    {
        name: "PostgreSQL",
        category: "Database",
        snippet: `  postgres:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: \${POSTGRES_PASSWORD}
    ports:
      - "5432:5432"
`
    },
    {
        name: "Redis",
        category: "Database",
        snippet: `  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
`
    },
    {
        name: "MCP Server (NPM)",
        category: "MCP",
        snippet: `  mcp-server:
    mcp_service:
      stdio_connection:
        command: npx
        args: ["-y", "@modelcontextprotocol/server-memory"]
`
    },
    {
        name: "Python Worker",
        category: "Worker",
        snippet: `  worker:
    image: python:3.9-slim
    command: python app.py
`
    }
];

/**
 * ServicePalette component.
 * Displays a list of service templates that can be dragged or clicked to insert into the stack configuration.
 *
 * @param props - The component props.
 * @param props.onTemplateSelect - Callback when a template is selected.
 * @returns The rendered palette component.
 */
export function ServicePalette({ onTemplateSelect }: ServicePaletteProps) {
    const [search, setSearch] = useState("");

    const filtered = TEMPLATES.filter(t => t.name.toLowerCase().includes(search.toLowerCase()));

    return (
        <div className="flex flex-col h-full bg-muted/5">
            <div className="p-4 border-b">
                <h3 className="font-semibold mb-2 text-sm">Service Templates</h3>
                <div className="relative">
                    <Search className="absolute left-2 top-2 h-3 w-3 text-muted-foreground" />
                    <Input
                        placeholder="Search..."
                        className="pl-7 h-8 text-xs"
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                    />
                </div>
            </div>
            <ScrollArea className="flex-1">
                <div className="p-4 space-y-2">
                    {filtered.map((t) => (
                        <div
                            key={t.name}
                            className="p-3 rounded border bg-card hover:bg-accent cursor-pointer transition-colors group"
                            onClick={() => onTemplateSelect(t.snippet)}
                        >
                            <div className="flex items-center justify-between mb-1">
                                <span className="font-medium text-xs">{t.name}</span>
                                <Badge variant="secondary" className="text-[10px] h-4 px-1">{t.category}</Badge>
                            </div>
                            <div className="text-[10px] text-muted-foreground font-mono truncate opacity-70 group-hover:opacity-100">
                                {t.snippet.split('\n')[0].trim()}...
                            </div>
                        </div>
                    ))}
                </div>
            </ScrollArea>
        </div>
    );
}
