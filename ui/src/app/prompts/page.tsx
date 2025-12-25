/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Input } from "@/components/ui/input";
import { Search, MessageSquare } from "lucide-react";

const initialPrompts = [
  { name: "summarize_text", description: "Summarizes the given text", arguments: ["text", "length"], enabled: true },
  { name: "code_review", description: "Reviews code for bugs", arguments: ["code", "language"], enabled: true },
  { name: "creative_story", description: "Generates a creative story", arguments: ["topic", "genre"], enabled: false },
];

export default function PromptsPage() {
  const [prompts, setPrompts] = useState(initialPrompts);
  const [search, setSearch] = useState("");

  const filteredPrompts = prompts.filter(p =>
    p.name.toLowerCase().includes(search.toLowerCase()) ||
    p.description.toLowerCase().includes(search.toLowerCase())
  );

  const togglePrompt = (name: string) => {
    setPrompts(prompts.map(p => p.name === name ? { ...p, enabled: !p.enabled } : p));
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Prompts</h2>
         <div className="relative w-64">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                    placeholder="Search prompts..."
                    className="pl-8"
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                />
        </div>
      </div>
      <div className="rounded-md border bg-background">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Arguments</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredPrompts.map((prompt) => (
              <TableRow key={prompt.name}>
                <TableCell className="font-medium flex items-center gap-2">
                    <MessageSquare className="h-4 w-4 text-muted-foreground" />
                    {prompt.name}
                </TableCell>
                <TableCell>{prompt.description}</TableCell>
                <TableCell>
                    <div className="flex gap-1">
                        {prompt.arguments.map(arg => (
                            <Badge key={arg} variant="secondary" className="text-xs font-normal">
                                {arg}
                            </Badge>
                        ))}
                    </div>
                </TableCell>
                <TableCell>
                    <div className="flex items-center space-x-2">
                        <Switch
                            checked={prompt.enabled}
                            onCheckedChange={() => togglePrompt(prompt.name)}
                        />
                        <span className="text-sm text-muted-foreground w-12">
                            {prompt.enabled ? "Active" : "Off"}
                        </span>
                    </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
