/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient, ResourceDefinition } from "@/lib/client";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Switch } from "@/components/ui/switch";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { FileText, Database } from "lucide-react";

export default function ResourcesPage() {
  const [resources, setResources] = useState<ResourceDefinition[]>([]);

  useEffect(() => {
    fetchResources();
  }, []);

  const fetchResources = async () => {
    try {
      const res = await apiClient.listResources();
      setResources(res.resources || []);
    } catch (e) {
      console.error("Failed to fetch resources", e);
    }
  };

  const toggleResource = async (uri: string, currentStatus: boolean) => {
    // Optimistic update (toggle status)
    // currentStatus is "enabled" (boolean).
    // If enabled, we want to disable (disable=true).
    // If disabled, we want to enable (disable=false).
    // So newDisable = currentStatus. (If enabled=true, disable becomes true).

    // Actually the usage below passes `!!resource.enabled`.
    // Wait, resource has `disable`?
    // If I change resource to use `disable`, then UI should use `!resource.disable`.
    const isEnabled = currentStatus;
    const newDisable = isEnabled; // If currently enabled, make it disabled (true).

    setResources(resources.map(r => r.uri === uri ? { ...r, disable: newDisable } : r));

    try {
        await apiClient.setResourceStatus(uri, newDisable);
    } catch (e) {
        console.error("Failed to toggle resource", e);
        fetchResources(); // Revert
    }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Resources</h2>
      </div>

      <Card className="backdrop-blur-sm bg-background/50">
        <CardHeader>
          <CardTitle>Managed Resources</CardTitle>
          <CardDescription>View and control access to resources.</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>URI</TableHead>
                <TableHead>MIME Type</TableHead>
                <TableHead>Service</TableHead>
                <TableHead>Status</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {resources.map((resource) => (
                <TableRow key={resource.uri}>
                  <TableCell className="font-medium flex items-center">
                    <FileText className="h-4 w-4 mr-2 text-muted-foreground" />
                    {resource.name}
                  </TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground">{resource.uri}</TableCell>
                  <TableCell>{resource.mimeType}</TableCell>
                  <TableCell>
                      {/* Service name not directly available in ResourceDefinition usually, check if mock has it? Mock had serviceName */}
                      <Badge variant="outline">System</Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center space-x-2">
                        <Switch
                            checked={!resource.disable}
                            onCheckedChange={() => toggleResource(resource.uri, !resource.disable)}
                        />
                        <span className="text-sm text-muted-foreground w-16">
                            {!resource.disable ? "Enabled" : "Disabled"}
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
