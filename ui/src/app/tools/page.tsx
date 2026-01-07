/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { UpstreamServiceConfig } from "@proto/config/v1/upstream_service";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Wrench, Play } from "lucide-react";
import { ToolDefinition } from "@proto/config/v1/tool";
import { ToolInspector } from "@/components/tools/tool-inspector";

export default function ToolsPage() {
  const [tools, setTools] = useState<ToolDefinition[]>([]);
  const [selectedTool, setSelectedTool] = useState<ToolDefinition | null>(null);
  const [inspectorOpen, setInspectorOpen] = useState(false);

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

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Tools</h2>
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
                <TableHead>Name</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Service</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {tools.map((tool) => (
                <TableRow key={tool.name} className="group">
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
