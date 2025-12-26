/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

interface Prompt {
  name: string;
  description?: string;
  service?: string;
}

export default function PromptsPage() {
  const [prompts, setPrompts] = useState<Prompt[]>([]);

  useEffect(() => {
    async function fetchPrompts() {
      // Need to aggregate prompts from services as there is no direct listPrompts API yet
      try {
        const { services } = await apiClient.listServices();
        const allPrompts = services.flatMap(s =>
            (s.grpc_service?.prompts ||
             s.http_service?.prompts ||
             s.command_line_service?.prompts ||
             s.mcp_service?.prompts ||
             [])
        );
        setPrompts(allPrompts);
      } catch (e) {
         console.error("Failed to fetch prompts", e);
      }
    }
    fetchPrompts();
  }, []);

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Prompts</h2>
      </div>
      <Card className="backdrop-blur-sm bg-background/50">
        <CardHeader>
            <CardTitle>System Prompts</CardTitle>
            <CardDescription>Pre-configured prompts available for agents.</CardDescription>
        </CardHeader>
        <CardContent>
           <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Description</TableHead>
                <TableHead>Service</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {prompts.map((p) => (
                <TableRow key={p.name}>
                  <TableCell className="font-medium">{p.name}</TableCell>
                  <TableCell>{p.description}</TableCell>
                  <TableCell><Badge variant="outline">{p.service}</Badge></TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
