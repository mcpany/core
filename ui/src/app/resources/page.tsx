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

interface Resource {
    id: string;
    name: string;
    type: string;
    service: string;
}

export default function ResourcesPage() {
    const [resources, setResources] = useState<Resource[]>([]);

    useEffect(() => {
        async function fetchResources() {
            const res = await fetch("/api/resources");
            if (res.ok) {
                setResources(await res.json());
            }
        }
        fetchResources();
    }, []);

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Resources</h2>
      </div>
      <Card className="backdrop-blur-sm bg-background/50">
        <CardHeader>
            <CardTitle>Available Resources</CardTitle>
            <CardDescription>
                Static or dynamic resources exposed by services.
            </CardDescription>
        </CardHeader>
        <CardContent>
            <div className="rounded-md border">
                <Table>
                <TableHeader>
                    <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Service</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {resources.map((resource) => (
                    <TableRow key={resource.id}>
                        <TableCell className="font-medium">{resource.name}</TableCell>
                        <TableCell>
                            <Badge variant="outline">{resource.type}</Badge>
                        </TableCell>
                        <TableCell>{resource.service}</TableCell>
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
