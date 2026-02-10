/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import Link from "next/link";
import { useState, useEffect } from "react";
import { Layers, Cuboid, Loader2, AlertCircle } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { apiClient } from "@/lib/client";
import { CreateStackDialog } from "@/components/stacks/create-stack-dialog";

interface Stack {
  name: string;
  services?: any[]; // Services are complex, keep any or define simplified Service type if needed
}

export default function StacksPage() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStacks = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await apiClient.listCollections();
      setStacks(data);
    } catch (e) {
      console.error("Failed to list stacks", e);
      setError("Failed to load stacks.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStacks();
  }, []);

  return (
    <div className="space-y-6 flex flex-col h-[calc(100vh-4rem)] p-8">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
            <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <CreateStackDialog onStackCreated={fetchStacks} />
      </div>

      {loading && (
          <div className="flex-1 flex items-center justify-center text-muted-foreground">
              <Loader2 className="h-8 w-8 animate-spin mr-2" /> Loading stacks...
          </div>
      )}

      {error && (
          <div className="flex items-center p-4 text-red-500 bg-red-50 dark:bg-red-900/10 border border-red-200 dark:border-red-900 rounded-md">
              <AlertCircle className="h-5 w-5 mr-2" />
              {error}
          </div>
      )}

      {!loading && !error && stacks.length === 0 && (
          <div className="flex-1 flex flex-col items-center justify-center text-muted-foreground border-2 border-dashed rounded-lg bg-muted/20">
              <Layers className="h-12 w-12 mb-4 opacity-50" />
              <h3 className="text-lg font-medium">No Stacks Found</h3>
              <p className="text-sm">Create a stack to organize your services.</p>
          </div>
      )}

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
                        <div className="text-xl font-bold tracking-tight truncate max-w-[200px]" title={stack.name}>{stack.name}</div>
                        <div className="text-xs text-muted-foreground font-mono">{stack.services?.length || 0} Services</div>
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
                 </div>
               </CardContent>
             </Card>
          </Link>
        ))}
      </div>
    </div>
  );
}
