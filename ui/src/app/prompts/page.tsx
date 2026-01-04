/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient, PromptDefinition } from "@/lib/client";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { MessageSquare } from "lucide-react";

export default function PromptsPage() {
  const [prompts, setPrompts] = useState<PromptDefinition[]>([]);

  useEffect(() => {
    fetchPrompts();
  }, []);

  const fetchPrompts = async () => {
    try {
      const res: any = await apiClient.listPrompts();
      if (Array.isArray(res)) {
          setPrompts(res);
      } else {
          setPrompts(res.prompts || []);
      }
    } catch (e) {
      console.error("Failed to fetch prompts", e);
    }
  };

  const togglePrompt = async (name: string, currentStatus: boolean) => {
    // Optimistic update
    setPrompts(prompts.map(p => p.name === name ? { ...p, disable: !currentStatus } : p));

    try {
        await apiClient.setPromptStatus(name, !currentStatus);
    } catch (e) {
        console.error("Failed to toggle prompt", e);
        fetchPrompts(); // Revert
    }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Prompts</h2>
      </div>

      <Card className="backdrop-blur-sm bg-background/50">
        <CardHeader>
          <CardTitle>Prompt Templates</CardTitle>
          <CardDescription>Manage available prompt templates.</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Service</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {prompts.map((prompt) => (
                <TableRow key={prompt.name}>
                  <TableCell className="font-medium flex items-center">
                    <MessageSquare className="h-4 w-4 mr-2 text-muted-foreground" />
                    {prompt.name}
                  </TableCell>
                  <TableCell>{prompt.description}</TableCell>
                  <TableCell>
                      <Badge variant="outline">{(prompt as any).serviceId}</Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                        <Switch
                            checked={!prompt.disable}
                            onCheckedChange={() => togglePrompt(prompt.name, !prompt.disable)}
                        />
                        <span className="text-sm text-muted-foreground w-16">
                            {!prompt.disable ? "Enabled" : "Disabled"}
                        </span>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
