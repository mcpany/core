/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Layers, Cuboid, Loader2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { apiClient } from "@/lib/client";
import { CreateStackDialog } from "@/components/stacks/create-stack-dialog";
import { toast } from "sonner";

interface StackSummary {
    id: string;
    name: string;
    services: number;
    status: string;
}

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<StackSummary[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchStacks();
  }, []);

  const fetchStacks = async () => {
    try {
      const res = await apiClient.listCollections();
      // Map collection response to UI model
      const mapped = (res || []).map((c: any) => ({
          id: c.name,
          name: c.name,
          services: c.services ? c.services.length : 0,
          status: "active"
      }));
      setStacks(mapped);
    } catch (e) {
      console.error("Failed to fetch stacks", e);
      toast.error("Failed to load stacks. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
         <div className="flex flex-col gap-2">
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
         </div>
         <CreateStackDialog />
      </div>

      {loading ? (
          <div className="flex items-center justify-center h-64">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
      ) : (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.length === 0 && (
                <div className="col-span-full text-center p-12 border border-dashed rounded-lg bg-muted/10">
                    <p className="text-muted-foreground mb-4">No stacks found. Create one to get started.</p>
                    <CreateStackDialog />
                </div>
            )}
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
                            <div className="text-xl font-bold tracking-tight truncate max-w-[200px]" title={stack.name}>{stack.name}</div>
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
