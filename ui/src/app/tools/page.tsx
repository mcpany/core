/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Wrench, Play, Star } from "lucide-react";
import { ToolDefinition } from "@proto/config/v1/tool";
import { ToolInspector } from "@/components/tools/tool-inspector";
import { usePinnedTools } from "@/hooks/use-pinned-tools";

/**
 * ToolsPage component.
 * @returns The rendered component.
 */
export default function ToolsPage() {
  const [tools, setTools] = useState<ToolDefinition[]>([]);
  const [selectedTool, setSelectedTool] = useState<ToolDefinition | null>(null);
  const [inspectorOpen, setInspectorOpen] = useState(false);
  const { isPinned, togglePin, isLoaded } = usePinnedTools();
  const [showPinnedOnly, setShowPinnedOnly] = useState(false);

  useEffect(() => {
    fetchTools();
  }, []);

  const fetchTools = async () => {
    try {
      const res = await apiClient.listTools();
      setTools(res.tools || []);
    } catch (e) {
      console.error("Failed to fetch tools", e);
    }
  };

  const toggleTool = async (name: string, currentStatus: boolean) => {
    // Optimistic update
    setTools(tools.map(t => t.name === name ? { ...t, disable: currentStatus } : t));

    try {
        await apiClient.setToolStatus(name, !currentStatus);
    } catch (e) {
        console.error("Failed to toggle tool", e);
        fetchTools(); // Revert
    }
  };

  const openInspector = (tool: ToolDefinition) => {
      setSelectedTool(tool);
      setInspectorOpen(true);
  };

  const filteredTools = tools
    .filter((t) => !showPinnedOnly || isPinned(t.name))
    .sort((a, b) => {
      const aPinned = isPinned(a.name);
      const bPinned = isPinned(b.name);
      if (aPinned && !bPinned) return -1;
      if (!aPinned && bPinned) return 1;
      return a.name.localeCompare(b.name);
    });

  if (!isLoaded) {
      return null;
  }

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Tools</h2>
        <div className="flex items-center space-x-2">
            <Switch
                id="show-pinned"
                checked={showPinnedOnly}
                onCheckedChange={setShowPinnedOnly}
            />
            <label htmlFor="show-pinned" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                Show Pinned Only
            </label>
        </div>
      </div>

      <Card className="backdrop-blur-sm bg-background/50">
        <CardHeader>
          <CardTitle>Available Tools</CardTitle>
          <CardDescription>Manage exposed tools from connected services.</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[30px]"></TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Service</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredTools.map((tool) => (
                <TableRow key={tool.name} className="group">
                  <TableCell>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-6 w-6"
                        onClick={() => togglePin(tool.name)}
                        aria-label={`Pin ${tool.name}`}
                      >
                          <Star className={`h-4 w-4 ${isPinned(tool.name) ? "fill-yellow-400 text-yellow-400" : "text-muted-foreground"}`} />
                      </Button>
                  </TableCell>
                  <TableCell className="font-medium flex items-center">
                    <Wrench className="h-4 w-4 mr-2 text-muted-foreground" />
                    {tool.name}
                  </TableCell>
                  <TableCell>{tool.description}</TableCell>
                  <TableCell>
                      <Badge variant="outline">{tool.serviceId}</Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                        <Switch
                            checked={!tool.disable}
                            onCheckedChange={() => toggleTool(tool.name, !tool.disable)}
                        />
                        <span className="text-sm text-muted-foreground w-16">
                            {!tool.disable ? "Enabled" : "Disabled"}
                        </span>
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                      <Button variant="outline" size="sm" onClick={() => openInspector(tool)}>
                          <Play className="h-3 w-3 mr-1" /> Inspect
                      </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <ToolInspector
        tool={selectedTool}
        open={inspectorOpen}
        onOpenChange={setInspectorOpen}
      />
    </div>
  );
}
