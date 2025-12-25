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
import { Search, FileText } from "lucide-react";

const initialResources = [
  { name: "api_spec.yaml", description: "OpenAPI Specification", type: "text/yaml", uri: "file:///docs/api.yaml", enabled: true },
  { name: "logo.png", description: "Company Logo", type: "image/png", uri: "https://cdn.example.com/logo.png", enabled: true },
  { name: "system_logs", description: "System Logs (Last 24h)", type: "text/plain", uri: "logs://system", enabled: false },
];

export default function ResourcesPage() {
  const [resources, setResources] = useState(initialResources);
  const [search, setSearch] = useState("");

  const filteredResources = resources.filter(res =>
    res.name.toLowerCase().includes(search.toLowerCase()) ||
    res.description.toLowerCase().includes(search.toLowerCase())
  );

  const toggleResource = (name: string) => {
    setResources(resources.map(r => r.name === name ? { ...r, enabled: !r.enabled } : r));
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Resources</h2>
         <div className="relative w-64">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                    placeholder="Search resources..."
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
              <TableHead>Type</TableHead>
              <TableHead>URI</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredResources.map((res) => (
              <TableRow key={res.name}>
                <TableCell className="font-medium flex items-center gap-2">
                    <FileText className="h-4 w-4 text-muted-foreground" />
                    {res.name}
                </TableCell>
                <TableCell>{res.description}</TableCell>
                <TableCell>
                    <Badge variant="outline">{res.type}</Badge>
                </TableCell>
                 <TableCell className="font-mono text-xs text-muted-foreground">
                    {res.uri}
                </TableCell>
                 <TableCell>
                    <div className="flex items-center space-x-2">
                        <Switch
                            checked={res.enabled}
                            onCheckedChange={() => toggleResource(res.name)}
                        />
                         <span className="text-sm text-muted-foreground w-12">
                            {res.enabled ? "Active" : "Off"}
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
