/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Layers, Cuboid, RefreshCcw } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { CreateStackDialog } from "@/components/stacks/create-stack-dialog";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchStacks = async () => {
    setLoading(true);
    try {
        const collections = await apiClient.listCollections();
        setStacks(collections || []);
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
        <div className="flex items-center gap-2">
            <Button variant="outline" size="icon" onClick={fetchStacks} title="Refresh">
                <RefreshCcw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            </Button>
            <CreateStackDialog onStackCreated={fetchStacks} />
        </div>
      </div>

      {loading && stacks.length === 0 ? (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
              {[1, 2, 3].map((i) => (
                  <Card key={i} className="animate-pulse h-[200px] bg-muted/20" />
              ))}
          </div>
      ) : stacks.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-[400px] border-2 border-dashed rounded-lg bg-muted/10">
              <Layers className="h-12 w-12 text-muted-foreground opacity-20 mb-4" />
              <h3 className="text-lg font-medium">No stacks found</h3>
              <p className="text-sm text-muted-foreground mb-6">Create a new stack to get started.</p>
              <CreateStackDialog onStackCreated={fetchStacks} />
          </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.map((stack) => (
            <Link key={stack.name} href={`/stacks/${stack.name}`}>
                <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50 h-full flex flex-col">
                <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                    <CardTitle className="text-sm font-medium text-muted-foreground">
                        Stack
                    </CardTitle>
                    <Cuboid className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                </CardHeader>
                <CardContent className="flex-1 flex flex-col">
                    <div className="flex items-center gap-3 mb-4">
                        <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                            <Layers className="h-6 w-6 text-primary" />
                        </div>
                        <div className="overflow-hidden">
                            <div className="text-2xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                            {stack.description && (
                                <div className="text-xs text-muted-foreground truncate" title={stack.description}>{stack.description}</div>
                            )}
                        </div>
                    </div>

                    <div className="mt-auto pt-4 border-t flex items-center justify-between text-xs text-muted-foreground">
                        <div className="flex items-center gap-1.5">
                            <span className="relative flex h-2 w-2">
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                            </span>
                            Active
                        </div>
                        <div>
                            {(stack.services || []).length} Services
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
