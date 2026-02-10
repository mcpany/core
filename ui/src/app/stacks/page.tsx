/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Layers, Cuboid, Plus, RefreshCw, Trash2, Edit } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";

interface Stack {
  name: string;
  description?: string;
  services?: any[];
  // Add other properties if available
}

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<Stack[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  const fetchStacks = async () => {
    setLoading(true);
    try {
      const res = await apiClient.listCollections();
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

  const handleDelete = async (e: React.MouseEvent, name: string) => {
    e.preventDefault(); // Prevent link navigation
    e.stopPropagation();
    if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;

    try {
        await apiClient.deleteCollection(name);
        toast({
            title: "Stack Deleted",
            description: `Stack ${name} has been removed.`
        });
        fetchStacks();
    } catch (e) {
        console.error("Failed to delete stack", e);
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
        <div className="flex items-center gap-2">
            <Button variant="outline" size="icon" onClick={fetchStacks} disabled={loading}>
                <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            </Button>
            <Link href="/stacks/new">
                <Button>
                    <Plus className="mr-2 h-4 w-4" />
                    New Stack
                </Button>
            </Link>
        </div>
      </div>

      {stacks.length === 0 && !loading ? (
           <div className="flex flex-col items-center justify-center py-20 border-2 border-dashed rounded-lg bg-muted/20">
              <Layers className="h-10 w-10 text-muted-foreground mb-4 opacity-50" />
              <h3 className="text-lg font-medium">No stacks found</h3>
              <p className="text-sm text-muted-foreground mb-4">Create a new stack to get started.</p>
              <Link href="/stacks/new">
                <Button>Create Stack</Button>
              </Link>
          </div>
      ) : (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.map((stack) => (
            <Link key={stack.name} href={`/stacks/${stack.name}`} className="group">
                <Card className="hover:shadow-md transition-all cursor-pointer h-full border-transparent shadow-sm bg-card hover:bg-muted/50 flex flex-col">
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
                        <div className="overflow-hidden">
                            <div className="text-2xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                            {stack.description && (
                                <div className="text-xs text-muted-foreground truncate" title={stack.description}>
                                    {stack.description}
                                </div>
                            )}
                        </div>
                    </div>
                </CardContent>
                <CardFooter className="pt-2 border-t text-xs text-muted-foreground flex justify-between items-center">
                     <div>
                         {stack.services?.length || 0} Services
                     </div>
                     <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                        <Button variant="ghost" size="icon" className="h-6 w-6" asChild>
                             {/* Technically link click propagates, but button style is nice */}
                             <span><Edit className="h-3 w-3" /></span>
                        </Button>
                        <Button variant="ghost" size="icon" className="h-6 w-6 text-destructive hover:text-destructive" onClick={(e) => handleDelete(e, stack.name)}>
                             <Trash2 className="h-3 w-3" />
                        </Button>
                     </div>
                </CardFooter>
                </Card>
            </Link>
            ))}
        </div>
      )}
    </div>
  );
}
