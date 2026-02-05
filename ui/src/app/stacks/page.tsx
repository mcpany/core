/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { Layers, Cuboid, Plus, Trash2, RefreshCcw } from "lucide-react";
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { apiClient } from "@/lib/client";
import { ServiceCollection } from "@/lib/marketplace-service";
import { useToast } from "@/hooks/use-toast";
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<ServiceCollection[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();
  const [stackToDelete, setStackToDelete] = useState<string | null>(null);

  const fetchStacks = useCallback(async () => {
    setLoading(true);
    try {
      const res = await apiClient.listCollections();
      // Ensure we handle potential null/undefined returns safely
      setStacks(res || []);
    } catch (e) {
      console.error("Failed to load stacks", e);
      toast({
        variant: "destructive",
        title: "Error",
        description: "Failed to load stacks."
      });
    } finally {
        setLoading(false);
    }
  }, [toast]);

  useEffect(() => {
    fetchStacks();
  }, [fetchStacks]);

  const handleDelete = async () => {
      if (!stackToDelete) return;
      try {
          await apiClient.deleteCollection(stackToDelete);
          toast({
              title: "Stack Deleted",
              description: `Stack ${stackToDelete} has been removed.`
          });
          setStackToDelete(null);
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
    <div className="space-y-6 p-8">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
            <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
            <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={fetchStacks} disabled={loading}>
                <RefreshCcw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                Refresh
            </Button>
            <Link href="/marketplace">
                <Button>
                    <Plus className="mr-2 h-4 w-4" /> New Stack
                </Button>
            </Link>
        </div>
      </div>

      {loading && stacks.length === 0 ? (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
              {[...Array(3)].map((_, i) => (
                  <div key={i} className="h-48 rounded-lg border bg-card animate-pulse" />
              ))}
          </div>
      ) : (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {stacks.map((stack) => (
                 <Card key={stack.name} className="hover:shadow-md transition-all group border-transparent shadow-sm bg-card hover:bg-muted/50 flex flex-col justify-between card">
                   <Link href={`/stacks/${stack.name}`} className="flex-1">
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
                                <div className="text-2xl font-bold tracking-tight truncate max-w-[200px]" title={stack.name}>{stack.name}</div>
                                <div className="text-xs text-muted-foreground font-mono truncate max-w-[200px]">{stack.version || "latest"}</div>
                            </div>
                         </div>
                         <div className="text-sm text-muted-foreground line-clamp-2 min-h-[40px]">
                             {stack.description || "No description provided."}
                         </div>
                       </CardContent>
                   </Link>
                   <CardFooter className="pt-4 border-t flex items-center justify-between">
                        <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                            <span className="relative flex h-2 w-2">
                              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                              <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                            </span>
                            {stack.services?.length || 0} Services
                        </div>
                        <Button
                            variant="ghost"
                            size="icon"
                            className="text-muted-foreground hover:text-destructive transition-colors"
                            onClick={(e) => {
                                e.stopPropagation();
                                setStackToDelete(stack.name);
                            }}
                        >
                            <Trash2 className="h-4 w-4" />
                        </Button>
                   </CardFooter>
                 </Card>
            ))}
            {stacks.length === 0 && (
                <div className="col-span-full flex flex-col items-center justify-center p-12 border-2 border-dashed rounded-lg bg-muted/10">
                    <Layers className="h-12 w-12 text-muted-foreground opacity-20 mb-4" />
                    <h3 className="text-lg font-medium">No stacks found</h3>
                    <p className="text-muted-foreground text-sm mb-4">Create a new stack to get started.</p>
                    <Link href="/marketplace">
                        <Button variant="outline">Browse Marketplace</Button>
                    </Link>
                </div>
            )}
          </div>
      )}

      <AlertDialog open={!!stackToDelete} onOpenChange={(open) => !open && setStackToDelete(null)}>
        <AlertDialogContent>
            <AlertDialogHeader>
                <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                <AlertDialogDescription>
                    This will permanently delete the stack "{stackToDelete}" and its configuration.
                    Running services associated with this stack might stop.
                </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
                <AlertDialogCancel>Cancel</AlertDialogCancel>
                <AlertDialogAction onClick={handleDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                    Delete
                </AlertDialogAction>
            </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
