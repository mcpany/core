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

interface Resource {
  name: string;
  type?: string;
  service?: string;
  description?: string;
}

export default function ResourcesPage() {
  const [resources, setResources] = useState<Resource[]>([]);

  useEffect(() => {
    async function fetchResources() {
      try {
        const { resources } = await apiClient.listResources();
        setResources(resources);
      } catch (e) {
        console.error("Failed to fetch resources", e);
      }
    }
    fetchResources();
  }, []);

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Resources</h2>
      </div>
      <Card className="backdrop-blur-sm bg-background/50">
        <CardHeader>
          <CardTitle>Resources</CardTitle>
          <CardDescription>Managed resources available to the system.</CardDescription>
        </CardHeader>
        <CardContent>
           <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Service</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {resources.map((res) => (
                <TableRow key={res.name}>
                  <TableCell className="font-medium">{res.name}</TableCell>
                  <TableCell>{res.type}</TableCell>
                  <TableCell><Badge variant="outline">{res.service}</Badge></TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
