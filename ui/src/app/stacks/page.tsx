/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Layers, Cuboid, Plus } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";

interface Stack {
  id: string;
  name: string;
  status: string;
  services: number;
  type: string;
}

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStacks = async () => {
      try {
        const collections = await apiClient.listCollections();
        // Collection format from backend: { name: string, services: [] }
        // We map it to the format expected by the UI or just use it directly
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const mappedStacks: Stack[] = collections.map((c: any) => ({
          id: c.name, // Name is unique ID for collections usually
          name: c.name,
          status: "active", // Default status, can be refined later
          services: c.services ? c.services.length : 0,
          type: "Standard"
        }));
        setStacks(mappedStacks);
      } catch (error) {
        console.error("Failed to fetch stacks", error);
      } finally {
        setLoading(false);
      }
    };

    fetchStacks();
  }, []);

  if (loading) {
    return (
      <div className="space-y-6">
        <div className="flex flex-col gap-2">
           <Skeleton className="h-8 w-48" />
           <Skeleton className="h-4 w-96" />
        </div>
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
           {[1, 2, 3].map((i) => (
             <Skeleton key={i} className="h-48 w-full rounded-xl" />
           ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-2">
        <div className="flex items-center justify-between">
           <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
           <Button variant="outline" size="sm" className="hidden">
              <Plus className="mr-2 h-4 w-4" /> Create Stack
           </Button>
        </div>
        <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
      </div>

      {stacks.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 border rounded-lg bg-muted/10 border-dashed">
             <div className="p-4 rounded-full bg-muted/30 mb-4">
                 <Layers className="h-8 w-8 text-muted-foreground" />
             </div>
             <h3 className="text-lg font-medium">No stacks found</h3>
             <p className="text-sm text-muted-foreground mt-1 mb-4">Create your first configuration stack to get started.</p>
             <Button>Create Stack</Button>
          </div>
      ) : (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.map((stack) => (
              <Link key={stack.id} href={`/stacks/${stack.id}`}>
                <Card className="hover:shadow-md transition-all cursor-pointer group border-transparent shadow-sm bg-card hover:bg-muted/50">
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
                            <div className="text-2xl font-bold tracking-tight truncate max-w-[150px]" title={stack.name}>{stack.name}</div>
                            <div className="text-xs text-muted-foreground font-mono truncate max-w-[150px]">{stack.id}</div>
                        </div>
                    </div>

                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                        <div className="flex items-center gap-1.5">
                            <span className="relative flex h-2 w-2">
                              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                              <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                            </span>
                            Online
                        </div>
                        <div>
                            {stack.services} Services
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
