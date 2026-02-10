/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Layers, Cuboid, Plus, RefreshCw, Trash2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { useToast } from "@/hooks/use-toast";
import { Badge } from "@/components/ui/badge";
import { ServiceCollection } from "@/lib/marketplace-service";

/**
 * StacksPage component.
 * Lists all available service collections (stacks) and provides options to create or delete them.
 * @returns The rendered page.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<ServiceCollection[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  const fetchStacks = async () => {
    setLoading(true);
    try {
      const data = await apiClient.listCollections();
      setStacks(data);
    } catch (e) {
      console.error("Failed to fetch stacks", e);
      toast({
        variant: "destructive",
        title: "Error",
        description: "Failed to load stacks.",
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStacks();
  }, []);

  const handleDelete = async (e: React.MouseEvent, name: string) => {
    e.preventDefault(); // Prevent navigation
    if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;

    try {
      await apiClient.deleteCollection(name);
      toast({ title: "Stack Deleted", description: `Stack ${name} removed.` });
      fetchStacks();
    } catch (e) {
      console.error("Failed to delete stack", e);
       toast({
        variant: "destructive",
        title: "Error",
        description: "Failed to delete stack.",
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
            <Button variant="outline" size="sm" onClick={fetchStacks} disabled={loading}>
                <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                Refresh
            </Button>
            <Link href="/stacks/new">
                <Button>
                    <Plus className="mr-2 h-4 w-4" /> Create Stack
                </Button>
            </Link>
        </div>
      </div>

      {stacks.length === 0 && !loading ? (
        <div className="flex flex-col items-center justify-center py-20 border-2 border-dashed rounded-lg bg-muted/20">
            <Layers className="h-10 w-10 text-muted-foreground mb-4 opacity-50" />
            <h3 className="text-lg font-medium">No stacks found</h3>
            <p className="text-sm text-muted-foreground mb-4">Create a stack to manage multiple services together.</p>
            <Link href="/stacks/new">
                <Button>Create Stack</Button>
            </Link>
        </div>
      ) : (
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
                    <div>
                        <div className="flex items-center gap-3 mb-4">
                            <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform">
                                <Layers className="h-6 w-6 text-primary" />
                            </div>
                            <div>
                                <div className="text-xl font-bold tracking-tight truncate max-w-[200px]" title={stack.name}>{stack.name}</div>
                                <div className="text-xs text-muted-foreground line-clamp-1">{stack.description || "No description"}</div>
                            </div>
                        </div>
                    </div>

                    <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                        <div className="flex items-center gap-1.5">
                            <Badge variant="secondary" className="font-normal">
                                {stack.services?.length || 0} Services
                            </Badge>
                        </div>
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 text-muted-foreground hover:text-destructive"
                            onClick={(e) => handleDelete(e, stack.name)}
                        >
                            <Trash2 className="h-4 w-4" />
                        </Button>
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
