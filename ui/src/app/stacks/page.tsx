/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Layers, Cuboid, Plus, Loader2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchStacks = async () => {
      try {
        const data = await apiClient.listCollections();
        setStacks(Array.isArray(data) ? data : []);
      } catch (error) {
        console.error("Failed to fetch stacks", error);
      } finally {
        setIsLoading(false);
      }
    };
    fetchStacks();
  }, []);

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex flex-col gap-2">
          <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
          <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <div className="flex h-[400px] items-center justify-center">
             <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
            <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <Button>
            <Plus className="mr-2 h-4 w-4" /> Create Stack
        </Button>
      </div>

      {stacks.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-[400px] border border-dashed rounded-lg bg-muted/10">
            <Layers className="h-12 w-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-semibold">No Stacks Found</h3>
            <p className="text-muted-foreground text-sm max-w-sm text-center mb-6">
                You haven't created any configuration stacks yet. Start by creating one to organize your MCP services.
            </p>
            <Button>
                <Plus className="mr-2 h-4 w-4" /> Create First Stack
            </Button>
        </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.map((stack) => (
            <Link key={stack.name} href={`/stacks/${stack.name}`}>
                <Card className="stack-card hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50">
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
                        <div>
                            <div className="text-2xl font-bold tracking-tight">{stack.name}</div>
                        </div>
                    </div>

                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                        <div className="flex items-center gap-1.5">
                            <span className="relative flex h-2 w-2">
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                            </span>
                            Active
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
