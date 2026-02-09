/**
 * Copyright 2025 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Layers, Cuboid, Plus, Trash2 } from "lucide-react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { apiClient } from "@/lib/client";
import { CreateStackDialog } from "@/components/stacks/create-stack-dialog";
import { useToast } from "@/hooks/use-toast";
import { Collection } from "@proto/config/v1/collection";

/**
 * StacksPage component.
 * @returns The rendered component.
 */
export default function StacksPage() {
  const [stacks, setStacks] = useState<Collection[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const { toast } = useToast();

  const fetchStacks = async () => {
    try {
      const collections = await apiClient.listCollections();
      setStacks(collections || []);
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

  const handleCreate = async (name: string) => {
    try {
      // Create empty stack
      const newStack = {
        name: name,
        description: "New Stack",
        author: "User", // TODO: Get from auth context
        version: "1.0.0",
        services: []
      };
      await apiClient.saveCollection(newStack);
      toast({
        title: "Stack Created",
        description: `Stack ${name} has been created.`
      });
      fetchStacks();
    } catch (e) {
      console.error("Failed to create stack", e);
      toast({
        variant: "destructive",
        title: "Error",
        description: "Failed to create stack."
      });
      throw e; // Re-throw to let dialog handle loading state if needed
    }
  };

  const handleDelete = async (e: React.MouseEvent, name: string) => {
    e.preventDefault(); // Prevent navigation
    e.stopPropagation();

    if (!confirm(`Are you sure you want to delete stack "${name}"?`)) return;

    try {
      await apiClient.deleteCollection(name);
      toast({
        title: "Stack Deleted",
        description: `Stack ${name} has been removed.`
      });
      setStacks(prev => prev.filter(s => s.name !== name));
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
    <div className="space-y-6 flex-1 p-8 pt-6">
      <div className="flex items-center justify-between">
        <div className="flex flex-col gap-2">
          <h1 className="text-3xl font-bold tracking-tight">Stacks</h1>
          <p className="text-muted-foreground">Manage your MCP Any configuration stacks.</p>
        </div>
        <Button onClick={() => setIsCreateOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Stack
        </Button>
      </div>

      {loading ? (
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin h-8 w-8 border-4 border-primary border-t-transparent rounded-full" />
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
                 <CardContent className="flex-1 flex flex-col">
                   <div className="flex items-center gap-3 mb-4 flex-1">
                      <div className="p-2.5 bg-primary/10 rounded-lg group-hover:scale-105 transition-transform h-fit">
                          <Layers className="h-6 w-6 text-primary" />
                      </div>
                      <div className="min-w-0">
                          <div className="text-2xl font-bold tracking-tight truncate" title={stack.name}>{stack.name}</div>
                          <div className="text-xs text-muted-foreground font-mono truncate" title={stack.description}>{stack.description || "No description"}</div>
                      </div>
                   </div>

                   <div className="flex items-center justify-between text-xs text-muted-foreground mt-4 pt-4 border-t">
                      <div className="flex items-center gap-1.5">
                          <span className="relative flex h-2 w-2">
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
                          </span>
                          Ready
                      </div>
                      <div className="flex items-center gap-4">
                          <span>
                              {stack.services ? stack.services.length : 0} Services
                          </span>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 text-muted-foreground hover:text-destructive z-10"
                            onClick={(e) => handleDelete(e, stack.name)}
                          >
                              <Trash2 className="h-4 w-4" />
                          </Button>
                      </div>
                   </div>
                 </CardContent>
               </Card>
            </Link>
          ))}
          {stacks.length === 0 && (
            <div className="col-span-full flex flex-col items-center justify-center p-12 border-2 border-dashed rounded-lg text-muted-foreground bg-muted/10">
              <Layers className="h-10 w-10 mb-4 opacity-50" />
              <h3 className="text-lg font-medium">No stacks found</h3>
              <p className="text-sm mb-4">Create a new stack to get started.</p>
              <Button variant="outline" onClick={() => setIsCreateOpen(true)}>Create Stack</Button>
            </div>
          )}
        </div>
      )}

      <CreateStackDialog
        open={isCreateOpen}
        onOpenChange={setIsCreateOpen}
        onCreate={handleCreate}
      />
    </div>
  );
}
