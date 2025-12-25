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
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";

interface Prompt {
    name: string;
    description: string;
    service: string;
}

export default function PromptsPage() {
    const [prompts, setPrompts] = useState<Prompt[]>([]);

    useEffect(() => {
        async function fetchPrompts() {
            const res = await fetch("/api/prompts");
            if (res.ok) {
                setPrompts(await res.json());
            }
        }
        fetchPrompts();
    }, []);

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <h2 className="text-3xl font-bold tracking-tight">Prompts</h2>
      </div>
      <Card className="backdrop-blur-sm bg-background/50">
          <CardHeader>
              <CardTitle>Available Prompts</CardTitle>
              <CardDescription>
                  Pre-defined prompts for LLM interaction.
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
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {prompts.map((prompt) => (
                    <TableRow key={prompt.name}>
                        <TableCell className="font-medium">{prompt.name}</TableCell>
                        <TableCell>{prompt.description}</TableCell>
                        <TableCell>{prompt.service}</TableCell>
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
