/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Layers, Cuboid, Loader2, AlertCircle } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { apiClient } from "@/lib/client";
import { CreateStackDialog } from "@/components/stacks/create-stack-dialog";

interface Stack {
    name: string;
    services?: any[];
}

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchStacks();
  }, []);

  const fetchStacks = async () => {
    try {
      setLoading(true);
      const res = await apiClient.listCollections();
      // listCollections returns generic array of collections
      setStacks(res || []);
      setError(null);
    } catch (e) {
      console.error("Failed to fetch stacks", e);
      setError("Failed to load stacks.");
    } finally {
      setLoading(false);
    }
  };

  if (loading && stacks.length === 0) {
      return (
          <div className="flex items-center justify-center h-[calc(100vh-4rem)] text-muted-foreground gap-2">
              <Loader2 className="h-6 w-6 animate-spin" /> Loading stacks...
          </div>
      );
  }

  if (error) {
      return (
          <div className="flex flex-col items-center justify-center h-[calc(100vh-4rem)] gap-4">
              <div className="flex items-center gap-2 text-destructive">
                  <AlertCircle className="h-6 w-6" />
                  <span className="text-lg font-medium">Error</span>
              </div>
              <p className="text-muted-foreground">{error}</p>
              <button onClick={fetchStacks} className="text-primary hover:underline">Try Again</button>
          </div>
      );
  }

  return (
    <div className="space-y-6 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <CreateStackDialog />
      </div>

      {stacks.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 border-2 border-dashed rounded-lg bg-muted/10">
              <div className="p-4 bg-muted rounded-full mb-4">
                  <Layers className="h-8 w-8 text-muted-foreground" />
              </div>
              <h3 className="text-lg font-medium">No stacks found</h3>
              <p className="text-sm text-muted-foreground mb-6 text-center max-w-sm">
                  Stacks allow you to group and manage multiple services together. Create your first stack to get started.
              </p>
              <CreateStackDialog />
          </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.map((stack) => (
            <Link key={stack.name} href={`/stacks/${encodeURIComponent(stack.name)}`}>
                <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50 backdrop-blur-sm">
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
                            <div className="text-xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                            <div className="text-xs text-muted-foreground font-mono truncate opacity-70">
                                {stack.services ? `${stack.services.length} services` : "Empty"}
                            </div>
                        </div>
                    </div>

                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                        <div className="flex items-center gap-1.5">
                            {/* Simple status heuristic - if services > 0, assume active for now, or check real status later */}
                            <span className="relative flex h-2 w-2">
                                <span className={`animate-ping absolute inline-flex h-full w-full rounded-full opacity-75 ${stack.services && stack.services.length > 0 ? "bg-green-400" : "bg-gray-400"}`}></span>
                                <span className={`relative inline-flex rounded-full h-2 w-2 ${stack.services && stack.services.length > 0 ? "bg-green-500" : "bg-gray-500"}`}></span>
                            </span>
                            {stack.services && stack.services.length > 0 ? "Active" : "Empty"}
                        </div>
                        <div>
                            {/* Type is usually Compose/Collection */}
                            Collection
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
