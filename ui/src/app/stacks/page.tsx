/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Layers, Cuboid, Plus } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const list = await apiClient.listCollections();
        setStacks(list);
      } catch (e) {
        console.error("Failed to load stacks", e);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  if (loading) {
      return <div className="p-8 text-muted-foreground">Loading stacks...</div>;
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
                        <div className="text-2xl font-bold tracking-tight">{stack.name}</div>
                        {/* Assuming description or version might be useful here */}
                        <div className="text-xs text-muted-foreground font-mono truncate max-w-[200px]" title={stack.description}>
                            {stack.description || "No description"}
                        </div>
                    </div>
                 </div>

                 <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                    <div className="flex items-center gap-1.5">
                        <span className="relative flex h-2 w-2">
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
             <div className="col-span-full text-center py-12 border-2 border-dashed rounded-lg text-muted-foreground">
                 No stacks found. Create one to get started.
             </div>
        )}
      </div>
    </div>
  );
}
