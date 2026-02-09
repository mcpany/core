/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { Layers, Cuboid, Loader2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { useEffect, useState } from "react";
import { apiClient } from "@/lib/client";
import { CreateStackDialog } from "@/components/stacks/create-stack-dialog";

interface Stack {
    name: string;
    description?: string;
    services: any[];
}

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchStacks = async () => {
    setLoading(true);
    try {
        const res = await apiClient.listCollections();
        setStacks(res || []);
    } catch (e) {
        console.error("Failed to list stacks", e);
    } finally {
        setLoading(false);
    }
  };

  useEffect(() => {
    fetchStacks();
  }, []);

  return (
    <div className="space-y-6 p-8">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
            <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <CreateStackDialog onStackCreated={fetchStacks} />
      </div>

      {loading ? (
          <div className="flex items-center justify-center h-40">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      ) : stacks.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 border-2 border-dashed rounded-lg bg-muted/20 text-muted-foreground">
              <Layers className="h-12 w-12 mb-4 opacity-50" />
              <h3 className="text-lg font-medium">No stacks found</h3>
              <p className="text-sm mb-4">Create your first stack to get started.</p>
              <CreateStackDialog onStackCreated={fetchStacks} />
          </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.map((stack) => (
            <Link key={stack.name} href={`/stacks/${stack.name}`}>
                <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50 h-full">
                <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                    <CardTitle className="text-sm font-medium text-muted-foreground">
                        Stack
                    </CardTitle>
                    <Cuboid className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                </CardHeader>
                <CardContent>
                    <div className="flex items-center gap-3 mb-4">
                        <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                            <Layers className="h-6 w-6 text-primary" />
                        </div>
                        <div className="overflow-hidden">
                            <div className="text-2xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                            <div className="text-xs text-muted-foreground truncate">{stack.description || "No description"}</div>
                        </div>
                    </div>

                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                        <div className="flex items-center gap-1.5">
                            {/* We don't really know status of the whole stack easily without more queries, assume config exists = valid for now */}
                            <span className="relative flex h-2 w-2">
                                <span className="relative inline-flex rounded-full h-2 w-2 bg-slate-400"></span>
                            </span>
                            Configured
                        </div>
                        <div>
                            {stack.services?.length || 0} Services
                        </div>
                    </div>
                </CardContent>
                </Card>
            </Link>
            ))}
        </div>
      )}
    </div>
  );
}
