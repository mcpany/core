/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Layers, Cuboid, Loader2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { apiClient } from "@/lib/client";
import { ServiceCollection } from "@/lib/marketplace-service";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<ServiceCollection[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiClient.listCollections()
        .then((data) => setStacks(data))
        .catch(console.error)
        .finally(() => setLoading(false));
  }, []);

  if (loading) {
      return <div className="flex h-full items-center justify-center"><Loader2 className="h-8 w-8 animate-spin" /></div>;
  }

  return (
    <div className="space-y-6 p-8">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
        <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
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
                        <div className="text-xs text-muted-foreground font-mono">{stack.version || "1.0.0"}</div>
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
                        {stack.services?.length || 0} Services
                    </div>
                 </div>
               </CardContent>
             </Card>
          </Link>
        ))}
        {stacks.length === 0 && (
            <div className="col-span-full flex flex-col items-center justify-center p-12 border-2 border-dashed rounded-lg text-muted-foreground">
                <Layers className="h-12 w-12 mb-4 opacity-50" />
                <p>No stacks found.</p>
            </div>
        )}
      </div>
    </div>
  );
}
