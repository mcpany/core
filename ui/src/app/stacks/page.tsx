/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Layers, Cuboid, Plus, RefreshCw, Trash2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

interface Stack {
  name: string;
  description?: string;
  services: any[];
}

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  const loadStacks = async () => {
    setLoading(true);
    try {
      const res = await apiClient.listCollections();
      setStacks(res || []);
    } catch (e) {
      console.error("Failed to list stacks", e);
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
    loadStacks();
  }, []);

  const handleDelete = async (e: React.MouseEvent, name: string) => {
    e.preventDefault();
    e.stopPropagation();
    if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;

    try {
      await apiClient.deleteCollection(name);
      toast({
        title: "Stack Deleted",
        description: `Stack ${name} has been removed.`
      });
      loadStacks();
    } catch (err) {
      console.error("Failed to delete stack", err);
      toast({
        variant: "destructive",
        title: "Error",
        description: "Failed to delete stack."
      });
    }
  };

  return (
    <div className="space-y-6 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
          <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
          <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <div className="flex gap-2">
            <Button variant="outline" onClick={loadStacks} disabled={loading}>
                <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                Refresh
            </Button>
            <Link href="/stacks/new">
                <Button>
                    <Plus className="mr-2 h-4 w-4" />
                    Create Stack
                </Button>
            </Link>
        </div>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {stacks.map((stack) => (
          <Link key={stack.name} href={`/stacks/${stack.name}`} className="group">
             <Card className="hover:shadow-md transition-all cursor-pointer h-full flex flex-col border-transparent shadow-sm bg-card hover:bg-muted/50 hover:border-border">
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
                        <div className="text-xs text-muted-foreground font-mono truncate max-w-[200px]">{stack.description || "No description"}</div>
                    </div>
                 </div>

                 <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                    <div className="flex items-center gap-1.5">
                        <Badge variant="secondary" className="font-normal">
                             {stack.services?.length || 0} Services
                        </Badge>
                    </div>
                    <div>
                       Active
                    </div>
                 </div>
               </CardContent>
               <CardFooter className="pt-0 pb-4 justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                    <Button variant="ghost" size="sm" onClick={(e) => handleDelete(e, stack.name)} className="text-destructive hover:text-destructive hover:bg-destructive/10 h-8 w-8 p-0">
                        <Trash2 className="h-4 w-4" />
                    </Button>
               </CardFooter>
             </Card>
          </Link>
        ))}
        {stacks.length === 0 && !loading && (
             <div className="col-span-full flex flex-col items-center justify-center py-20 border-2 border-dashed rounded-lg bg-muted/20 text-muted-foreground">
                 <Layers className="h-12 w-12 mb-4 opacity-50" />
                 <h3 className="text-lg font-medium">No stacks found</h3>
                 <p className="text-sm mb-4">Create a new stack to get started.</p>
                 <Link href="/stacks/new">
                    <Button variant="outline">Create Stack</Button>
                 </Link>
             </div>
        )}
      </div>
    </div>
  );
}
