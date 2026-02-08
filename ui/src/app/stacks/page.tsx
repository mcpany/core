/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { Layers, Cuboid, Loader2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { useState, useEffect } from "react";
import { apiClient } from "@/lib/client";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchStacks();
  }, []);

  const fetchStacks = async () => {
    try {
      const res = await apiClient.listCollections();
      // Ensure we handle array or object response if wrapped
      const list = Array.isArray(res) ? res : (res.collections || []);
      setStacks(list);
    } catch (e) {
      console.error("Failed to fetch stacks", e);
      setError("Failed to load stacks.");
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center p-8 h-[calc(100vh-4rem)]">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex-1 flex items-center justify-center p-8 h-[calc(100vh-4rem)] text-destructive">
        {error}
      </div>
    );
  }

  return (
    <div className="space-y-6 p-8 pt-6">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
        <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
      </div>

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
               <CardContent className="flex-1 flex flex-col justify-between">
                 <div className="flex items-center gap-3 mb-4">
                    <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                        <Layers className="h-6 w-6 text-primary" />
                    </div>
                    <div className="overflow-hidden">
                        <div className="text-2xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                        {/* Use sanitized name or id if available, mostly name is ID for collections */}
                    </div>
                 </div>

                 <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                    <div className="flex items-center gap-1.5">
                        {/* Status is not readily available in collection object, assuming configured means valid/ready */}
                        <span className="relative flex h-2 w-2">
                          <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
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
        ))}
        {stacks.length === 0 && (
            <div className="col-span-full text-center p-12 border border-dashed rounded-lg text-muted-foreground">
                No stacks found. Create one to get started.
            </div>
        )}
      </div>
    </div>
  );
}
