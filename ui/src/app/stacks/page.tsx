/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { Layers, Cuboid, Plus, Loader2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { ServiceCollection } from "@/lib/marketplace-service";

/**
 * StacksPage component for listing all available stacks.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<ServiceCollection[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function loadStacks() {
      try {
        const data = await apiClient.listCollections();
        setStacks(data);
      } catch (error) {
        console.error("Failed to load stacks", error);
      } finally {
        setLoading(false);
      }
    }
    loadStacks();
  }, []);

  if (loading) {
      return <div className="flex items-center justify-center h-full"><Loader2 className="animate-spin" /></div>;
  }

  return (
    <div className="space-y-6 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <Link href="/stacks/new">
            <Button>
                <Plus className="mr-2 h-4 w-4" /> Create Stack
            </Button>
        </Link>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {stacks.map((stack) => (
          <Link key={stack.name} href={`/stacks/${stack.name}`}>
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
                    <div className="truncate">
                        <div className="text-2xl font-bold tracking-tight truncate">{stack.name}</div>
                        <div className="text-xs text-muted-foreground font-mono truncate">{stack.version || "latest"}</div>
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
        {stacks.length === 0 && (
            <div className="col-span-full text-center py-12 text-muted-foreground border-2 border-dashed rounded-lg">
                No stacks found. Create one to get started.
            </div>
        )}
      </div>
    </div>
  );
}
