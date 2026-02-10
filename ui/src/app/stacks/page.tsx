/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Layers, Cuboid, Plus, Trash2, Edit } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { ServiceCollection } from "@/lib/marketplace-service";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<ServiceCollection[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  const fetchStacks = async () => {
    try {
      setLoading(true);
      const res = await apiClient.listCollections();
      // Ensure we map the response correctly if it differs
      setStacks(res || []);
    } catch (e) {
      console.error("Failed to fetch stacks", e);
      toast({
        variant: "destructive",
        title: "Error",
        description: "Failed to load stacks."
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStacks();
  }, []);

  const deleteStack = async (name: string) => {
      if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;
      try {
          await apiClient.deleteCollection(name);
          fetchStacks();
          toast({ title: "Stack Deleted", description: `Stack ${name} has been removed.` });
      } catch (e) {
          toast({ variant: "destructive", title: "Error", description: "Failed to delete stack." });
      }
  };

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div>
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground mt-2">Manage your MCP Any configuration stacks.</p>
        </div>
        <Link href="/stacks/new">
            <Button>
                <Plus className="mr-2 h-4 w-4" /> Add Stack
            </Button>
        </Link>
      </div>

      {loading ? (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
              {[...Array(3)].map((_, i) => (
                  <div key={i} className="h-40 bg-muted animate-pulse rounded-lg" />
              ))}
          </div>
      ) : (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.map((stack) => (
                 <Card key={stack.name} className="hover:shadow-md transition-all group border-transparent shadow-sm bg-card hover:bg-muted/50 flex flex-col">
                   <CardHeader className="flex flex-row items-center justify-between pb-2 space-y-0">
                      <CardTitle className="text-sm font-medium text-muted-foreground">
                        Stack
                      </CardTitle>
                      <Cuboid className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                   </CardHeader>
                   <CardContent className="flex-1">
                     <div className="flex items-center gap-3 mb-4">
                        <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                            <Layers className="h-6 w-6 text-primary" />
                        </div>
                        <div>
                            <div className="text-xl font-bold tracking-tight truncate max-w-[200px]" title={stack.name}>{stack.name}</div>
                            <div className="text-xs text-muted-foreground font-mono truncate max-w-[200px]">{stack.author || "Unknown Author"}</div>
                        </div>
                     </div>
                     <p className="text-sm text-muted-foreground line-clamp-2 min-h-[40px]">
                         {stack.description || "No description provided."}
                     </p>

                     <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                        <div className="flex items-center gap-1.5">
                            <Badge variant="outline">{stack.version || "0.0.1"}</Badge>
                        </div>
                        <div>
                            {stack.services?.length || 0} Services
                        </div>
                     </div>
                   </CardContent>
                   <CardFooter className="pt-0 gap-2">
                       <Link href={`/stacks/${stack.name}`} className="flex-1">
                           <Button variant="outline" className="w-full">
                               <Edit className="mr-2 h-4 w-4" /> Manage
                           </Button>
                       </Link>
                       <Button variant="ghost" size="icon" className="text-destructive hover:bg-destructive/10" onClick={() => deleteStack(stack.name)}>
                           <Trash2 className="h-4 w-4" />
                       </Button>
                   </CardFooter>
                 </Card>
            ))}
            {stacks.length === 0 && (
                <div className="col-span-full flex flex-col items-center justify-center p-12 border-2 border-dashed rounded-lg text-muted-foreground">
                    <Layers className="h-12 w-12 mb-4 opacity-20" />
                    <p>No stacks found. Create one to get started.</p>
                </div>
            )}
          </div>
      )}
    </div>
  );
}
