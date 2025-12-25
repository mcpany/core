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
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Search } from "lucide-react";

// Mock data based on Tool definition
const initialTools = [
  { name: "stripe_charge", description: "Create a charge on Stripe", service: "Payment Gateway", type: "function", enabled: true },
  { name: "get_user", description: "Retrieve user details", service: "User Service", type: "function", enabled: true },
  { name: "search_docs", description: "Search internal documentation", service: "Search Indexer", type: "read", enabled: false },
  { name: "send_email", description: "Send transactional email", service: "Email Service", type: "function", enabled: true },
];

export default function ToolsPage() {
  const [tools, setTools] = useState(initialTools);
  const [search, setSearch] = useState("");

  const filteredTools = tools.filter(tool =>
    tool.name.toLowerCase().includes(search.toLowerCase()) ||
    tool.description.toLowerCase().includes(search.toLowerCase())
  );

  const toggleTool = (name: string) => {
      setTools(tools.map(t => t.name === name ? { ...t, enabled: !t.enabled } : t));
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Tools</h2>
        <div className="flex items-center space-x-2">
            <div className="relative w-64">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                    placeholder="Search tools..."
                    className="pl-8"
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                />
            </div>
        </div>
      </div>
      <div className="rounded-md border bg-background">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Service</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredTools.map((tool) => (
              <TableRow key={tool.name}>
                <TableCell className="font-medium">{tool.name}</TableCell>
                <TableCell>{tool.description}</TableCell>
                <TableCell>{tool.service}</TableCell>
                <TableCell>
                    <Badge variant="outline">{tool.type}</Badge>
                </TableCell>
                <TableCell>
                    <div className="flex items-center space-x-2">
                        <Switch
                            checked={tool.enabled}
                            onCheckedChange={() => toggleTool(tool.name)}
                        />
                        <span className="text-sm text-muted-foreground w-12">
                            {tool.enabled ? "Active" : "Off"}
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
