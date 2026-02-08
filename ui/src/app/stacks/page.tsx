/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { Layers, Cuboid, Loader2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";
import { Collection } from "@proto/config/v1/collection";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<Collection[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchStacks();
  }, []);

  const fetchStacks = async () => {
    try {
      const collections = await apiClient.listCollections();
      // Collections response is an array of configv1.Collection
      setStacks(collections || []);
    } catch (e) {
      console.error("Failed to fetch stacks", e);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-[calc(100vh-4rem)]">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="space-y-6 flex flex-col h-[calc(100vh-4rem)] p-8 pt-6">
      <div className="flex flex-col gap-2 flex-none">
        <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
        <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
      </div>

      <div className="flex-1 overflow-y-auto">
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 pb-8">
          {stacks.length === 0 ? (
             <div className="col-span-full flex flex-col items-center justify-center py-12 text-muted-foreground border-2 border-dashed rounded-lg">
                <Layers className="h-12 w-12 mb-4 opacity-50" />
                <p>No stacks found.</p>
                <p className="text-sm mt-2">Create a new stack or import a collection.</p>
             </div>
          ) : (
            stacks.map((stack) => (
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
                            <div className="text-xs text-muted-foreground font-mono truncate" title={stack.name}>{stack.name}</div>
                        </div>
                     </div>

                     <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                        <div className="flex items-center gap-1.5">
                            {/* We don't have real-time status here yet, so maybe just show "Configured" or remove the dot */}
                            <span className="relative flex h-2 w-2">
                              <span className="relative inline-flex rounded-full h-2 w-2 bg-muted-foreground/50"></span>
                            </span>
                            Configured
                        </div>
                        <div>
                            {stack.services ? stack.services.length : 0} Services
                        </div>
                     </div>
                   </CardContent>
                 </Card>
              </Link>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
