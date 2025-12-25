/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

interface Tool {
    name: string;
    description: string;
    service: string;
    type: string;
}

export default function ToolsPage() {
  const [tools, setTools] = useState<Tool[]>([]);

  useEffect(() => {
    async function fetchTools() {
        const res = await fetch("/api/tools");
        if (res.ok) {
            setTools(await res.json());
        }
    }
    fetchTools();
  }, []);

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Tools</h2>
      </div>
      <Card className="backdrop-blur-sm bg-background/50">
          <CardHeader>
              <CardTitle>Available Tools</CardTitle>
              <CardDescription>
                  List of all tools exposed by the connected services.
              </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="rounded-md border">
                <Table>
                <TableHeader>
                    <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Description</TableHead>
                    <TableHead>Service</TableHead>
                    <TableHead>Type</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {tools.map((tool) => (
                    <TableRow key={tool.name}>
                        <TableCell className="font-medium">{tool.name}</TableCell>
                        <TableCell>{tool.description}</TableCell>
                        <TableCell>{tool.service}</TableCell>
                        <TableCell>
                            <Badge variant="secondary">{tool.type}</Badge>
                        </TableCell>
                    </TableRow>
                    ))}
                </TableBody>
                </Table>
            </div>
          </CardContent>
      </Card>
    </div>
  );
}
