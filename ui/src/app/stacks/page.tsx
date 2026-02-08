/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Layers, Cuboid, Loader2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { apiClient, UpstreamServiceConfig } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

interface StackCollection {
  name: string;
  services: UpstreamServiceConfig[];
}

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [collections, setCollections] = useState<StackCollection[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  useEffect(() => {
    const fetchCollections = async () => {
      try {
        const data = await apiClient.listCollections();
        setCollections(data);
      } catch (error) {
        console.error("Failed to fetch collections", error);
        toast({
            title: "Error",
            description: "Failed to load stacks. Please try again later.",
            variant: "destructive",
        });
      } finally {
        setLoading(false);
      }
    };
    fetchCollections();
  }, [toast]);

  if (loading) {
    return (
      <div className="flex flex-1 items-center justify-center p-8">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="space-y-6 flex-1 p-8 pt-6">
      <div className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold tracking-tight">Stacks</h1>
        <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
      </div>

      {collections.length === 0 ? (
         <div className="flex flex-col items-center justify-center p-12 text-center border-2 border-dashed rounded-lg bg-muted/20">
            <Layers className="h-12 w-12 text-muted-foreground/50 mb-4" />
            <h3 className="text-lg font-semibold">No Stacks Found</h3>
            <p className="text-muted-foreground mt-2 max-w-sm">
                You haven't created any stacks yet. Create a stack using the CLI or import a collection.
            </p>
         </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {collections.map((stack) => (
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
                          <div className="text-xs text-muted-foreground font-mono">{stack.name}</div>
                      </div>
                   </div>

                   <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                      <div className="flex items-center gap-1.5">
                          {/* Placeholder for status - assuming Active if listed for now */}
                          <span className="relative flex h-2 w-2">
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                          </span>
                          Available
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
